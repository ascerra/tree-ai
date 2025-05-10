// internal/tree/printer_test.go
package tree

import (
	"os"
	"testing"
)

func TestPrintTree(t *testing.T) {
	tempDir := t.TempDir()
	file, err := os.CreateTemp(tempDir, "testfile*.go")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	file.Close()

	err = PrintTree(tempDir, "")
	if err != nil {
		t.Errorf("PrintTree returned error: %v", err)
	}
}