package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/acmcelwee/git-index/internal/config"
	"github.com/acmcelwee/git-index/internal/indexer"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var addRootCmd = &cobra.Command{
	Use:   "add-root [path]",
	Short: "Add a new recursive search root",
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

		depth, _ := cmd.Flags().GetInt("depth")

		if err := config.AddSearchDir(absPath, depth); err != nil {
			fmt.Fprintln(os.Stderr, "Error adding to config:", err)
			os.Exit(1)
		}

		fmt.Printf("Added %s (max_depth: %d) to config.\n", absPath, depth)

		// Rebuild index
		if err := indexer.Index(); err != nil {
			fmt.Fprintln(os.Stderr, "Error rebuilding index:", err)
			os.Exit(1)
		}
		fmt.Println("Index rebuilt successfully.")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	addRootCmd.Flags().Int("depth", 3, "Maximum depth to search")
	configCmd.AddCommand(addRootCmd)
}
