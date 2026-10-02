package cmd

import (
	"fmt"
	"os"

	"github.com/adam/git-index/internal/indexer"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Rebuild the repository index",
	Run: func(cmd *cobra.Command, args []string) {
		if err := indexer.Index(); err != nil {
			fmt.Fprintln(os.Stderr, "Error building index:", err)
			os.Exit(1)
		}
		fmt.Println("Index rebuilt successfully.")
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
