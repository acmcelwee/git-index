package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/adam/git-index/internal/config"
	"github.com/adam/git-index/internal/indexer"
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

		// Use a map to deduplicate DisplayNames
		seen := make(map[string]bool)
		for _, repo := range repos {
			if !seen[repo.DisplayName] {
				// Append a trailing slash to make menu selection visually distinct
				fmt.Println(repo.DisplayName + "/")
				seen[repo.DisplayName] = true
			}
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
