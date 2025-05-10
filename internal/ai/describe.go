// internal/ai/describe.go
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"os/exec"
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

func Describe(path string, isDir bool, model string, userEndpoint string) string {
	target := filepath.Base(path)
	itemType := "file"
	if isDir {
		itemType = "directory"
	}
	prompt := fmt.Sprintf("For %s named '%s' read everything underneath it and tell me a specific detail that is most important as it relates to this entire repo. No more than 100 characters", itemType, target)

	// Use custom endpoint if provided, otherwise default Granite model
	endpoint := userEndpoint
	if endpoint == "" {
		endpoint = "https://granite-3-1-8b-instruct-w4a16-maas-apicast-production.apps.prod.rhoai.rh-aiservices-bu.com/v1/completions"
	}
	healthURL := strings.Replace(endpoint, "/v1/completions", "/health", 1)

	if !isEndpointAvailable(healthURL) {
		return fallback(target, isDir, model)
	}

	if Verbose {
		fmt.Fprintln(os.Stderr, "[tree-ai] querying endpoint...")
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

	return summarizeToOneLine(result.Choices[0].Text)
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
