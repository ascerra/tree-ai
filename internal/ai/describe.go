package ai

import (
	"bufio"
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
	"sync"
	"time"
)

var (
	fileCounter          int
	totalFiles           int
	localModelProcess    *exec.Cmd
	Verbose              bool = false
	TruncateDescriptions bool = true
	modelWriter          io.WriteCloser
	modelReader          *bufio.Reader
	modelLock            sync.Mutex
)

func SetTotalFiles(n int) {
	totalFiles = n
}

func Describe(path string, isDir bool, model, userEndpoint, userInstruction string) string {
	itemType := map[bool]string{true: "directory", false: "file"}[isDir]
	target := filepath.Base(path)
	content := collectContent(path, isDir)
	if Verbose {
		fmt.Fprintf(os.Stderr, "[tree-ai] collected content for %s (%s):\n%s\n", path, itemType, content)
	}

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

	if endpoint == "" || !isEndpointAvailable(strings.Replace(endpoint, "/v1/completions", "/health", 1)) {
		if Verbose {
			fmt.Fprintln(os.Stderr, "[tree-ai] falling back to local model.")
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

func fallback(target string, isDir bool, model string, fullPrompt string) string {
	var outputBuf bytes.Buffer

	if err := ensureModelRunning(); err != nil {
		if Verbose {
			fmt.Fprintf(os.Stderr, "[tree-ai] local model init failed: %v\n", err)
		}
		return defaultFallbackText(isDir)
	}

	modelLock.Lock()
	defer modelLock.Unlock()

	if Verbose {
		fmt.Fprintf(os.Stderr, "[tree-ai] Sending prompt to local model for %s:\n%s\n", target, fullPrompt)
	}

	fmt.Fprintln(modelWriter, fullPrompt)
	fmt.Fprintln(modelWriter, "<<END>>")

	readStart := time.Now()
	for {
		if time.Since(readStart) > 5*time.Second {
			if Verbose {
				fmt.Fprintf(os.Stderr, "[tree-ai] Timeout reading model output for %s\n", target)
			}
			break
		}
		line, err := modelReader.ReadString('\n')
		if err != nil && err != io.EOF {
			if Verbose {
				fmt.Fprintf(os.Stderr, "[tree-ai] Error reading response for %s: %v\n", target, err)
			}
			break
		}
		outputBuf.WriteString(line)
		if strings.TrimSpace(line) == "<<END>>" || strings.HasSuffix(line, "\n\n") {
			break
		}
	}

	response := strings.TrimSpace(outputBuf.String())
	if Verbose {
		fmt.Fprintf(os.Stderr, "[tree-ai] Final raw model output for %s:\n%s\n", target, response)
	}

	if response != "" && response != "." {
		return response
	}
	return defaultFallbackText(isDir)
}

func ensureModelRunning() error {
	modelLock.Lock()
	defer modelLock.Unlock()

	if localModelProcess != nil {
		return nil
	}

	fmt.Fprintln(os.Stderr, "🔄 Launching local AI model...")

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}
	projectRoot := filepath.Dir(filepath.Dir(exePath))
	pythonScriptPath := filepath.Join(projectRoot, "model", "granite_stream.py")
	venvPythonPath := filepath.Join(projectRoot, ".venv", "bin", "python")

	cmd := exec.Command(venvPythonPath, pythonScriptPath)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe error: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe error: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe error: %w", err)
	}

	go func() { io.Copy(os.Stderr, stderr) }()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start error: %w", err)
	}

	modelWriter = stdin
	modelReader = bufio.NewReader(stdout)
	localModelProcess = cmd

	fmt.Fprintln(os.Stderr, "⏳ Waiting for model to become ready...")

	line, err := modelReader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("model failed to start: %v", err)
	}

	if !strings.Contains(line, "[READY]") {
		return fmt.Errorf("unexpected output from model: %q", line)
	}

	fmt.Fprintln(os.Stderr, "✅ Local model ready.")
	return nil
}

func defaultFallbackText(isDir bool) string {
	if isDir {
		return "Internal directory for project logic."
	}
	return "Internal project file."
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

	if TruncateDescriptions {
		return summarizeToOneLine(text)
	}
	return text
}

func summarizeToOneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.Join(strings.Fields(s), " ")
	s = strings.TrimSpace(s)

	const maxLen = 120
	if len(s) > maxLen {
		s = s[:maxLen]
		if i := strings.LastIndex(s, " "); i > 0 {
			s = s[:i]
		}
		s += "..."
	}
	return s
}

func formatFinalResponse(label string, desc string, isDir bool) string {
	arrow := "\033[38;5;208m➤\033[0m"
	desc = strings.TrimSpace(desc)
	if isDir {
		desc = strings.TrimPrefix(desc, ". ")
	}
	return fmt.Sprintf("%s %s", arrow, desc)
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
