package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/adam/git-index/internal/indexer"
	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Output list of repository basenames for shell completion",
	Run: func(cmd *cobra.Command, args []string) {
		repos, err := indexer.LoadIndex()
		if err != nil || len(repos) == 0 {
			// Try to build index once if empty, but silently
			if err := indexer.Index(); err == nil {
				repos, _ = indexer.LoadIndex()
			}
		}

		// Add flags
		fmt.Println("--rebuild")
		fmt.Println("--add")
		fmt.Println("--add-root")
		fmt.Println("--list")

		// Use a map to deduplicate basenames
		seen := make(map[string]bool)
		for _, repo := range repos {
			base := filepath.Base(repo)
			if !seen[base] {
				// Append a trailing slash to make menu selection visually distinct
				fmt.Println(base + "/")
				seen[base] = true
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}
