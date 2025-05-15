package ai

import "testing"

func TestDescribeFallback_File(t *testing.T) {
	desc := Describe("go.sum", false, "mock-model", "", "")
	if desc == "" {
		t.Fatal("expected a non-empty description for file fallback")
	}
}

func TestDescribeFallback_Dir(t *testing.T) {
	desc := Describe(".", true, "mock-model", "", "")
	if desc == "" {
		t.Fatal("expected a non-empty description for directory fallback")
	}
}
