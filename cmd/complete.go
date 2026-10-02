package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/acmcelwee/git-index/internal/config"
	"github.com/acmcelwee/git-index/internal/frecency"
	"github.com/acmcelwee/git-index/internal/indexer"
	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Output list of repository basenames for shell completion",
	Run: func(cmd *cobra.Command, args []string) {
		checkAndTriggerBackgroundRebuild()

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

		frecencyData, _ := frecency.Load()
		halfLife := frecency.GetHalfLife()

		type completionItem struct {
			name  string
			score float64
		}

		nameScores := make(map[string]float64)
		for _, repo := range repos {
			score := frecency.Score(frecencyData[repo.Path], halfLife)
			if current, exists := nameScores[repo.DisplayName]; !exists || score > current {
				nameScores[repo.DisplayName] = score
			}
		}

		var items []completionItem
		for name, score := range nameScores {
			items = append(items, completionItem{name: name, score: score})
		}

		sort.SliceStable(items, func(i, j int) bool {
			// Alphabetical tiebreaker
			if items[i].score == items[j].score {
				return items[i].name < items[j].name
			}
			return items[i].score > items[j].score
		})

		for _, item := range items {
			fmt.Println(item.name + "/")
		}
	},
}

func checkAndTriggerBackgroundRebuild() {
	cfg, err := config.LoadConfig()
	if err != nil {
		return
	}

	intervalStr := cfg.AutoRebuildInterval
	if intervalStr == "" {
		intervalStr = "24h"
	}
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		interval = 24 * time.Hour
	}

	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return
	}

	indexPath := filepath.Join(cacheDir, "index.json")
	stat, err := os.Stat(indexPath)
	if err != nil {
		return
	}

	if time.Since(stat.ModTime()) > interval {
		// Touch the file immediately to prevent multiple forks
		currentTime := time.Now().Local()
		os.Chtimes(indexPath, currentTime, currentTime)

		// Fork background rebuild
		execPath, err := os.Executable()
		if err == nil {
			cmd := exec.Command(execPath, "index")
			cmd.Start()
		}
	}
}

func init() {
	rootCmd.AddCommand(completeCmd)
}
