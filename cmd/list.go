package cmd

import (
	"fmt"
	"os"

	"github.com/adam/git-index/internal/indexer"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all repositories currently in the index",
	Run: func(cmd *cobra.Command, args []string) {
		repos, err := indexer.LoadIndex()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error loading index:", err)
			os.Exit(1)
		}

		if len(repos) == 0 {
			fmt.Println("Index is empty. Try running 'git-index index' to build it.")
			return
		}

		for _, repo := range repos {
			if repo.Type != "repo" {
				fmt.Printf("%s\t[%s]\n", repo.Path, repo.DisplayName)
			} else {
				fmt.Println(repo.Path)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
