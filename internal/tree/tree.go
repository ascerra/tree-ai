package tree

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tree-ai/internal/ai"
)

type node struct {
	path  string
	isDir bool
	depth int
}

func CollectPaths(root string, maxDepth int, includeFiles bool) []node {
	if maxDepth == -1 {
		maxDepth = 0 // default: root contents only
	}
	ignoreSet := loadIgnoreFiles(root)
	var paths []node
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		depth := len(strings.Split(rel, string(os.PathSeparator))) - 1

		if maxDepth >= 0 && depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if rel == "." || ignoreSet[rel] {
			return nil
		}

		if depth == 1 && strings.HasPrefix(filepath.Base(path), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		if info.IsDir() || includeFiles {
			paths = append(paths, node{path: path, isDir: info.IsDir(), depth: depth})
		}
		return nil
	})
	return paths
}

func loadIgnoreFiles(root string) map[string]bool {
	ignoreSet := make(map[string]bool)
	ignoreFiles := []string{".gitignore", ".dockerignore", ".ignore"}

	for _, filename := range ignoreFiles {
		ignorePath := filepath.Join(root, filename)
		data, err := os.ReadFile(ignorePath)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			globMatches, _ := filepath.Glob(filepath.Join(root, line))
			for _, match := range globMatches {
				relPath, err := filepath.Rel(root, match)
				if err == nil {
					ignoreSet[relPath] = true
				}
			}
		}
	}
	return ignoreSet
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func PrintTreeWithPaths(paths []node, root, prefix string, noAI bool, model string, maxDepth int, endpoint string) {
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].path < paths[j].path
	})

	for i, n := range paths {
		base := filepath.Base(n.path)
		isLast := i == len(paths)-1 || paths[i+1].depth <= n.depth

		prefix := strings.Repeat("│   ", max(n.depth-1, 0))
		if isLast {
			prefix += "└── "
		} else {
			prefix += "├── "
		}

		var icon, color string
		if n.isDir {
			icon = "💼"
			color = "\033[1;34m"
		} else {
			icon = "📄"
			color = "\033[1;32m"
		}
		reset := "\033[0m"

		desc := ""
		if !noAI {
			if _, err := os.Stat(n.path); err == nil {
				desc = ai.Describe(n.path, n.isDir, model, endpoint)
			}
		}

		fmt.Printf("%s%s%s %s%s %s\n", prefix, color, icon, base, reset, desc)
	}
}
