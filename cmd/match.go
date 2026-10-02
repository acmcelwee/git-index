package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/acmcelwee/git-index/internal/frecency"
	"github.com/acmcelwee/git-index/internal/indexer"
	"github.com/spf13/cobra"
)

var matchCmd = &cobra.Command{
	Use:   "match [query]",
	Short: "Match a query to a repository path",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkAndTriggerBackgroundRebuild()

		// Strip trailing slashes so auto-completed paths work correctly
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

		var exactMatches []indexer.RepoEntry
		var startsWithMatches []indexer.RepoEntry
		var containsMatches []indexer.RepoEntry
		var pathMatches []indexer.RepoEntry

		for _, repo := range repos {
			display := strings.ToLower(repo.DisplayName)
			if display == query {
				exactMatches = append(exactMatches, repo)
			} else if strings.HasPrefix(display, query) {
				startsWithMatches = append(startsWithMatches, repo)
			} else if strings.Contains(display, query) {
				containsMatches = append(containsMatches, repo)
			} else if strings.Contains(strings.ToLower(repo.Path), query) {
				pathMatches = append(pathMatches, repo)
			}
		}

		var finalMatches []indexer.RepoEntry
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

		// JIT existence check
		var validMatches []string
		for _, match := range finalMatches {
			if stat, err := os.Stat(match.Path); err == nil && stat.IsDir() {
				validMatches = append(validMatches, match.Path)
			} else {
				// Path no longer exists, remove from index
				fmt.Fprintf(os.Stderr, "Removing stale path from index: %s\n", match.Path)
				indexer.RemoveFromIndex(match.Path)
			}
		}

		if len(validMatches) == 0 {
			os.Exit(1)
		}

		frecencyData, _ := frecency.Load()
		halfLife := frecency.GetHalfLife()

		sort.SliceStable(validMatches, func(i, j int) bool {
			scoreI := frecency.Score(frecencyData[validMatches[i]], halfLife)
			scoreJ := frecency.Score(frecencyData[validMatches[j]], halfLife)
			return scoreI > scoreJ
		})

		for _, match := range validMatches {
			fmt.Println(match)
		}
	},
}

func init() {
	rootCmd.AddCommand(matchCmd)
}
