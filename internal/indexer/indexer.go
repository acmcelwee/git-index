package indexer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adam/git-index/internal/config"
	"github.com/bmatcuk/doublestar/v4"
)

func expandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return filepath.Abs(path)
}

func isIgnored(path string, ignores []string) bool {
	for _, pattern := range ignores {
		if matched, _ := doublestar.Match(pattern, path); matched {
			return true
		}
	}
	return false
}

func Index() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var allRepos []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, sdir := range cfg.SearchDirs {
		wg.Add(1)
		go func(sdir config.SearchDir) {
			defer wg.Done()

			rootPath, err := expandHome(sdir.Path)
			if err != nil {
				return
			}

			// If max depth is 0, we just check this directory
			if sdir.MaxDepth == 0 {
				if stat, err := os.Stat(filepath.Join(rootPath, ".git")); err == nil && stat.IsDir() {
					mu.Lock()
					allRepos = append(allRepos, rootPath)
					mu.Unlock()
				}
				return
			}

			_ = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}

				if isIgnored(path, cfg.Ignore) {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}

				// Calculate depth
				rel, _ := filepath.Rel(rootPath, path)
				if rel == "." {
					return nil
				}
				depth := len(strings.Split(rel, string(os.PathSeparator)))

				if depth > sdir.MaxDepth {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}

				if d.IsDir() && d.Name() == ".git" {
					repoPath := filepath.Dir(path)
					mu.Lock()
					allRepos = append(allRepos, repoPath)
					mu.Unlock()
					return filepath.SkipDir // Don't descend into .git
				}

				return nil
			})
		}(sdir)
	}

	wg.Wait()

	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	indexPath := filepath.Join(cacheDir, "index.txt")
	file, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, repo := range allRepos {
		_, _ = writer.WriteString(repo + "\n")
	}
	return writer.Flush()
}

func LoadIndex() ([]string, error) {
	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return nil, err
	}

	indexPath := filepath.Join(cacheDir, "index.txt")
	file, err := os.Open(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No index yet
		}
		return nil, err
	}
	defer file.Close()

	var repos []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			repos = append(repos, line)
		}
	}
	return repos, scanner.Err()
}
