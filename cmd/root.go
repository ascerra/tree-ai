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
)

var rootCmd = &cobra.Command{
	Use:   "tree-ai",
	Short: "AI-enhanced tree command",
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		// Pass verbosity to AI layer
		ai.Verbose = verbose

		// Collect tree paths
		paths := tree.CollectPaths(dir, maxDepth, includeFiles, includeDotfiles)
		ai.SetTotalFiles(len(paths))

		// Render tree with AI summaries
		tree.PrintTreeWithPaths(paths, dir, "", noAI, model, maxDepth, endpoint, promptInstruction)
	},
}

func init() {
	rootCmd.Flags().BoolVar(&noAI, "no-ai", false, "Disable AI-generated descriptions")
	rootCmd.Flags().StringVar(&model, "model", "granite-3-1-8b-instruct-w4a16", "Model to use for AI descriptions")
	rootCmd.Flags().IntVar(&maxDepth, "max-depth", -1, "Limit the depth of the directory tree (default: -1 for unlimited)")
	rootCmd.Flags().BoolVar(&includeFiles, "include-files", true, "Include files in the output (default: true)")
	rootCmd.Flags().StringVar(&endpoint, "endpoint", "", "Custom model endpoint URL (overrides default Granite endpoint)")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging (default: off)")
	rootCmd.Flags().BoolVar(&includeDotfiles, "include-dotfiles", false, "Include dotfiles and dotdirs like `tree -a`")
	rootCmd.Flags().StringVar(&promptInstruction, "prompt-instruction", "", "Custom prompt instruction to append after the file/directory contents")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
