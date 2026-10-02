package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/acmcelwee/git-index/internal/config"
	"github.com/acmcelwee/git-index/internal/indexer"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Add a single repository as a search dir (max_depth 0)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error resolving path:", err)
			os.Exit(1)
		}

		// Verify it's a git repo
		if stat, err := os.Stat(filepath.Join(absPath, ".git")); err != nil || !stat.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: %s is not a git repository\n", absPath)
			os.Exit(1)
		}

		if err := config.AddSearchDir(absPath, 0); err != nil {
			fmt.Fprintln(os.Stderr, "Error adding to config:", err)
			os.Exit(1)
		}

		fmt.Printf("Added %s to config.\n", absPath)

		// Rebuild index
		if err := indexer.Index(); err != nil {
			fmt.Fprintln(os.Stderr, "Error rebuilding index:", err)
			os.Exit(1)
		}
		fmt.Println("Index rebuilt successfully.")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
