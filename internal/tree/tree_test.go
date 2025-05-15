package tree

import (
	"testing"
)

func TestCollectPaths_RootLevel(t *testing.T) {
	paths := CollectPaths(".", 1, true, false)
	if len(paths) == 0 {
		t.Fatal("expected to collect at least one path")
	}
}
