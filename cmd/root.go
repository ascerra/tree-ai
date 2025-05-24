// ✅ Updated root.go (truncate default false, --truncate sets it true)
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tree-ai/internal/ai"
	"tree-ai/internal/tree"
)

var (
	noAI              bool
	model             string
	maxDepth          int
	includeFiles      bool
	endpoint          string
	verbose           bool
	includeDotfiles   bool
	promptInstruction string
	truncate          bool
)

var rootCmd = &cobra.Command{
	Use:   "tree-ai",
	Short: "AI-enhanced tree command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stdout, "⚠️  AI-generated summaries may be inaccurate or outdated. Always verify important details.")
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		ai.Verbose = verbose
		ai.TruncateDescriptions = truncate

		paths := tree.CollectPaths(dir, maxDepth, includeFiles, includeDotfiles)
		ai.SetTotalFiles(len(paths))
		tree.PrintTreeWithPaths(paths, dir, "", noAI, model, maxDepth, endpoint, promptInstruction)
	},
}

func init() {
	rootCmd.Flags().BoolVar(&noAI, "no-ai", false, "Disable AI-generated descriptions")
	rootCmd.Flags().StringVar(&model, "model", "", "Model to use for AI descriptions (required with --endpoint)")	
	rootCmd.Flags().IntVar(&maxDepth, "max-depth", -1, "Limit the depth of the directory tree (default: -1 for unlimited)")
	rootCmd.Flags().BoolVar(&includeFiles, "include-files", true, "Include files in the output (default: true)")
	rootCmd.Flags().StringVar(&endpoint, "endpoint", "", "Custom model endpoint URL (overrides default Granite endpoint)")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging (default: off)")
	rootCmd.Flags().BoolVar(&includeDotfiles, "include-dotfiles", false, "Include dotfiles and dotdirs like `tree -a`")
	rootCmd.Flags().StringVar(&promptInstruction, "prompt", "", "Custom prompt instruction to append after the file/directory contents")
	rootCmd.Flags().BoolVar(&truncate, "truncate", false, "Truncate AI descriptions to one line (set true to enable)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
