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

func CollectPaths(root string, maxDepth int, includeFiles bool, includeDotfiles bool) []node {
	var paths []node

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Don't include the root itself
		if path == root {
			return nil
		}

		// Calculate relative depth
		relPath, _ := filepath.Rel(root, path)
		depth := len(strings.Split(relPath, string(os.PathSeparator)))

		// Enforce max depth
		if maxDepth >= 0 && depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip dotfiles and dotdirs unless included
		if !includeDotfiles && strings.HasPrefix(d.Name(), ".") {
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

func PrintTreeWithPaths(paths []node, root, prefix string, noAI bool, model string, maxDepth int, endpoint string, promptInstruction string) {
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

		color := "\033[1;32m" // green for files by default
		if n.isDir {
			color = "\033[1;34m" // blue for directories
		}
		reset := "\033[0m"

		desc := ""
		if !noAI {
			if _, err := os.Stat(n.path); err == nil {
				desc = ai.Describe(n.path, n.isDir, model, endpoint, promptInstruction)
			}
		}

		fmt.Printf("%s%s%s%s %s\n", prefix, color, base, reset, desc)
	}
}
