package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func ModelIsCached() bool {
	cachePath := filepath.Join(".hf-cache", "models--ibm-granite")
	entries, err := os.ReadDir(cachePath)
	return err == nil && len(entries) > 0
}

var fileCounter int
var totalFiles int

func SetTotalFiles(n int) {
	totalFiles = n
}

var Verbose bool = false

func Describe(path string, isDir bool, model, userEndpoint, userInstruction string) string {
	itemType := map[bool]string{true: "directory", false: "file"}[isDir]
	target := filepath.Base(path)
	content := collectContent(path, isDir)
	
	instruction := userInstruction
	if instruction == "" {
		instruction = fmt.Sprintf("In 2 sentence, explain the purpose of this %s **as it relates to the whole project**.\nRespond only with the explanation. Do not repeat the prompt.", itemType)
	}
	
	prompt := fmt.Sprintf(`You are a senior developer helping onboard a new teammate.
	
	This is a %s named "%s". Below are its contents:
	---
	%s
	---
	
	%s`, itemType, target, content, instruction)

	if Verbose {
		fmt.Fprintf(os.Stderr, "[tree-ai] prompt for %s:\n%s\n", path, prompt)
	}

	endpoint := userEndpoint
	if endpoint == "" {
		endpoint = "https://granite-3-1-8b-instruct-w4a16-maas-apicast-production.apps.prod.rhoai.rh-aiservices-bu.com/v1/completions"
	}
	healthURL := strings.Replace(endpoint, "/v1/completions", "/health", 1)

	if !isEndpointAvailable(healthURL) {
		return fallback(target, isDir, model)
	}

	payload := fmt.Sprintf(`{
		"model": %q,
		"prompt": %q,
		"max_tokens": 100,
		"temperature": 0.7
	}`, model, prompt)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		if Verbose {
			fmt.Fprintf(os.Stderr, "[tree-ai] request creation failed: %v\n", err)
		}
		return fallback(target, isDir, model)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := os.Getenv("TREE_AI_API_KEY"); key != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if Verbose {
			fmt.Fprintf(os.Stderr, "[tree-ai] request failed: %v\n", err)
		}
		return fallback(target, isDir, model)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if Verbose {
			fmt.Fprintf(os.Stderr, "[tree-ai] read failed: %v\n", err)
		}
		return fallback(target, isDir, model)
	}

	var result struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil || len(result.Choices) == 0 {
		if Verbose {
			fmt.Fprintf(os.Stderr, "[tree-ai] failed to parse response: %v\n", err)
		}
		return fallback(target, isDir, model)
	}

	return cleanModelResponse(prompt, result.Choices[0].Text)
}

func fallback(target string, isDir bool, model string) string {
	prompt := fmt.Sprintf("For %s named '%s' read everything underneath it and tell me a specific detail that is most important as it relates to this entire repo. No more than 100 characters",
		map[bool]string{true: "directory", false: "file"}[isDir], target)

	cmd := exec.Command(".venv/bin/python", "model/granite_infer.py", "--prompt", prompt)
	cmd.Env = append(os.Environ(), "TRANSFORMERS_CACHE=.hf-cache")
	output, err := cmd.Output()

	if err == nil && len(output) > 0 {
		return summarizeToOneLine(string(output))
	}

	// true last-resort fallback
	if isDir {
		return fmt.Sprintf("(Directory for managing %s using model %s)", target, model)
	}
	return fmt.Sprintf("(File handling %s functionality using model %s)", strings.TrimSuffix(target, filepath.Ext(target)), model)
}

func summarizeToOneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

func isEndpointAvailable(url string) bool {
	client := http.Client{Timeout: 2 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		if Verbose {
			fmt.Fprintf(os.Stderr, "[tree-ai] health check request creation failed: %v\n", err)
		}
		return false
	}
	if key := os.Getenv("TREE_AI_API_KEY"); key != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	}

	resp, err := client.Do(req)
	if err != nil {
		if Verbose {
			fmt.Fprintf(os.Stderr, "[tree-ai] health check failed: %v\n", err)
		}
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		if Verbose {
			fmt.Fprintf(os.Stderr, "[tree-ai] health check returned status: %s\n", resp.Status)
		}
		return false
	}

	return true
}

func collectContent(path string, isDir bool) string {
	var builder strings.Builder
	const maxTotalBytes = 6000

	addFile := func(p string) {
		if isBinary(p) {
			if Verbose {
				fmt.Fprintf(os.Stderr, "[tree-ai] skipping binary file: %s\n", p)
			}
			return
		}
	
		data, err := os.ReadFile(p)
		if err != nil || len(data) == 0 {
			return
		}
	
		builder.WriteString(fmt.Sprintf("\n--- %s ---\n", filepath.Base(p)))
		builder.Write(data)

		if builder.Len() > maxTotalBytes {
			builder.WriteString("\n... [truncated]")
		}
	}

	if !isDir {
		addFile(path)
	} else {
		filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			addFile(p)
			if builder.Len() > maxTotalBytes {
				return io.EOF
			}
			return nil
		})
	}

	result := builder.String()
	if len(result) > maxTotalBytes {
		result = result[:maxTotalBytes] + "\n... [truncated]"
	}
	return result
}

func cleanModelResponse(prompt, rawText string) string {
	text := strings.TrimSpace(rawText)

	// Remove verbatim prompt if included
	if strings.HasPrefix(text, prompt) {
		text = strings.TrimPrefix(text, prompt)
	}

	// Heuristic: find start of real answer
	starts := []string{"The ", "This ", "A ", "An ", "In ", "It ", "There ", "Directory ", "File "}
	foundIdx := -1
	for _, s := range starts {
		idx := strings.Index(text, s)
		if idx != -1 && (foundIdx == -1 || idx < foundIdx) {
			foundIdx = idx
		}
	}

	if foundIdx != -1 {
		text = text[foundIdx:]
	}

	text = strings.TrimSpace(text)
	return summarizeToOneLine(text)
}

func isBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 800)
	n, _ := f.Read(buf)
	if n == 0 {
		return false
	}

	for _, b := range buf[:n] {
		if b == 0 { // null byte = likely binary
			return true
		}
	}
	return false
}
