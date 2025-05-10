// internal/ai/describe_test.go
package ai

import (
	"testing"
)

func TestDescribe(t *testing.T) {
	if got := Describe("/tmp/config.yaml", false); got == "" {
		t.Errorf("expected description for file, got empty string")
	}
	if got := Describe("/tmp/app", true); got == "" {
		t.Errorf("expected description for directory, got empty string")
	}
}