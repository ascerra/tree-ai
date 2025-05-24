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
	"regexp"
	"strings"
	"time"
)

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
		instruction = fmt.Sprintf("In 1 sentence, explain the purpose of this %s **as it relates to the whole project**. Respond only with the explanation. Avoid repeating the file name or type.", itemType)
	}

	prompt := fmt.Sprintf(`You are a senior developer helping onboard a new teammate. You are summarizing project components.
This is a %s named "%s". Its contents are:
%s

%s`, itemType, target, content, instruction)

	if Verbose {
		fmt.Fprintf(os.Stderr, "[tree-ai] prompt for %s:\n%s\n", path, prompt)
	}

	endpoint := userEndpoint
	if endpoint == "" {
		endpoint = os.Getenv("TREE_AI_ENDPOINT")
	}

	if endpoint == "" {
		if Verbose {
			fmt.Fprintln(os.Stderr, "[tree-ai] no remote endpoint configured, falling back to local model.")
		}
		return formatFinalResponse(target, cleanModelResponse(fallback(target, isDir, model, prompt), target, isDir), isDir)
	}

	healthURL := strings.Replace(endpoint, "/v1/completions", "/health", 1)
	if !isEndpointAvailable(healthURL) {
		if Verbose {
			fmt.Fprintf(os.Stderr, "[tree-ai] remote endpoint %s not available, falling back to local model.\n", endpoint)
		}
		return formatFinalResponse(target, cleanModelResponse(fallback(target, isDir, model, prompt), target, isDir), isDir)
	}

	payload := fmt.Sprintf(`{
		"model": %q,
		"prompt": %q,
		"max_tokens": 100,
		"temperature": 0.7
	}`, model, prompt)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return formatFinalResponse(target, cleanModelResponse(fallback(target, isDir, model, prompt), target, isDir), isDir)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := os.Getenv("TREE_AI_API_KEY"); key != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return formatFinalResponse(target, cleanModelResponse(fallback(target, isDir, model, prompt), target, isDir), isDir)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return formatFinalResponse(target, cleanModelResponse(fallback(target, isDir, model, prompt), target, isDir), isDir)
	}

	var result struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil || len(result.Choices) == 0 {
		return formatFinalResponse(target, cleanModelResponse(fallback(target, isDir, model, prompt), target, isDir), isDir)
	}

	return formatFinalResponse(target, cleanModelResponse(result.Choices[0].Text, target, isDir), isDir)
}

func cleanModelResponse(rawText string, target string, isDir bool) string {
	original := strings.TrimSpace(rawText)
	text := original

	pattern := regexp.MustCompile(`(?i)^((this|the)\s+)?(.*\b` + regexp.QuoteMeta(target) + `\b.*?)\s*(file|directory|script|module|document)?\s*(,|is|provides|serves|:|-)*\s*`)
	temp := pattern.ReplaceAllString(text, "")
	if strings.TrimSpace(temp) != "" {
		text = temp
	}

	if i := strings.Index(text, "."); i > 5 {
		text = text[:i+1]
	}

	text = strings.TrimLeft(text, "\"',:; ")
	if strings.TrimSpace(text) == "" {
		text = original
	}

	return summarizeToOneLine(text)
}

func formatFinalResponse(label string, desc string, isDir bool) string {
	arrow := "⇒"
	desc = strings.TrimSpace(desc)
	if isDir {
		desc = strings.TrimPrefix(desc, ". ")
	}
	return fmt.Sprintf("%s %s", arrow, desc)
}

func fallback(target string, isDir bool, model string, fullPrompt string) string {
	cmd := exec.Command(".venv/bin/python", "model/granite_infer.py", "--prompt", fullPrompt)
	cmd.Env = append(os.Environ(), "TRANSFORMERS_CACHE=.hf-cache")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return string(output)
	}
	if isDir {
		return "Internal directory for project logic."
	}
	return "Internal project file."
}

func isEndpointAvailable(url string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	if key := os.Getenv("TREE_AI_API_KEY"); key != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func collectContent(path string, isDir bool) string {
	var builder strings.Builder
	const maxTotalBytes = 6000

	addFile := func(p string) {
		if isBinary(p) {
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

func summarizeToOneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
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
		if b == 0 {
			return true
		}
	}
	return false
}