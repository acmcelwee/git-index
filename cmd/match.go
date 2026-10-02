package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/acmcelwee/git-index/internal/indexer"
	"github.com/spf13/cobra"
)

var matchCmd = &cobra.Command{
	Use:   "match [query]",
	Short: "Match a query to a repository path",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Strip trailing slashes so auto-completed paths (like "container/") work correctly
		query := strings.TrimRight(strings.ToLower(args[0]), "/")
		
		repos, err := indexer.LoadIndex()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error loading index:", err)
			os.Exit(1)
		}

		if len(repos) == 0 {
			// Try to build index once if empty
			if err := indexer.Index(); err == nil {
				repos, _ = indexer.LoadIndex()
			}
		}

		var exactMatches []string
		var startsWithMatches []string
		var containsMatches []string
		var pathMatches []string

		for _, repo := range repos {
			base := strings.ToLower(filepath.Base(repo))
			if base == query {
				exactMatches = append(exactMatches, repo)
			} else if strings.HasPrefix(base, query) {
				startsWithMatches = append(startsWithMatches, repo)
			} else if strings.Contains(base, query) {
				containsMatches = append(containsMatches, repo)
			} else if strings.Contains(strings.ToLower(repo), query) {
				pathMatches = append(pathMatches, repo)
			}
		}

		var finalMatches []string
		if len(exactMatches) > 0 {
			finalMatches = exactMatches
		} else if len(startsWithMatches) > 0 {
			finalMatches = startsWithMatches
		} else if len(containsMatches) > 0 {
			finalMatches = containsMatches
		} else if len(pathMatches) > 0 {
			finalMatches = pathMatches
		}

		if len(finalMatches) == 0 {
			os.Exit(1)
		}

		for _, match := range finalMatches {
			fmt.Println(match)
		}
	},
}

func init() {
	rootCmd.AddCommand(matchCmd)
}
