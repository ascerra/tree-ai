// model/granite-runner.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Calls the local Python script running IBM Granite 4.0 Tiny Preview model.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: granite-runner <prompt>")
		os.Exit(1)
	}

	prompt := strings.Join(os.Args[1:], " ")
	cmd := exec.Command("python3", "model/granite_infer.py", "--prompt", prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "granite-runner error: %v\n%s\n", err, output)
		os.Exit(1)
	}
	fmt.Println(string(output))
}