package indexer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/acmcelwee/git-index/internal/config"
	"github.com/bmatcuk/doublestar/v4"
)

type RepoEntry struct {
	Path        string `json:"path"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"` // "repo", "worktree", "submodule"
}

func expandHome(path string) (string, error) {
	if path == "~" {
		return os.UserHomeDir()
	}
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

func parseGitFile(dotGitPath string, repoPath string) RepoEntry {
	baseName := filepath.Base(repoPath)
	content, err := os.ReadFile(dotGitPath)
	if err != nil {
		return RepoEntry{Path: repoPath, DisplayName: baseName, Type: "repo"}
	}

	contentStr := strings.TrimSpace(string(content))
	if strings.HasPrefix(contentStr, "gitdir: ") {
		gitdir := strings.TrimPrefix(contentStr, "gitdir: ")
		// Typically looks like: /path/to/parent/.git/worktrees/child
		// or /path/to/parent/.git/modules/child
		parts := strings.Split(gitdir, "/.git/")
		if len(parts) == 2 {
			parentPath := parts[0]
			parentName := filepath.Base(parentPath)
			
			if strings.HasPrefix(parts[1], "worktrees/") {
				return RepoEntry{
					Path:        repoPath,
					DisplayName: parentName + "+" + baseName,
					Type:        "worktree",
				}
			} else if strings.HasPrefix(parts[1], "modules/") {
				return RepoEntry{
					Path:        repoPath,
					DisplayName: parentName + "@" + baseName,
					Type:        "submodule",
				}
			}
		}
	}

	return RepoEntry{Path: repoPath, DisplayName: baseName, Type: "repo"}
}

func Index() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var allRepos []RepoEntry
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
				dotGit := filepath.Join(rootPath, ".git")
				if stat, err := os.Stat(dotGit); err == nil {
					var entry RepoEntry
					if stat.IsDir() {
						entry = RepoEntry{Path: rootPath, DisplayName: filepath.Base(rootPath), Type: "repo"}
					} else {
						entry = parseGitFile(dotGit, rootPath)
					}
					mu.Lock()
					allRepos = append(allRepos, entry)
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

				rel, _ := filepath.Rel(rootPath, path)
				if rel == "." {
					return nil
				}
				depth := strings.Count(rel, string(os.PathSeparator)) + 1

				if depth > sdir.MaxDepth {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}

				if d.Name() == ".git" {
					repoPath := filepath.Dir(path)
					var entry RepoEntry
					if d.IsDir() {
						entry = RepoEntry{Path: repoPath, DisplayName: filepath.Base(repoPath), Type: "repo"}
					} else {
						entry = parseGitFile(path, repoPath)
					}
					
					mu.Lock()
					allRepos = append(allRepos, entry)
					mu.Unlock()
					
					if d.IsDir() {
						return filepath.SkipDir // Don't descend into .git
					}
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
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return err
	}

	file, err := os.CreateTemp(cacheDir, "index-*.json.tmp")
	if err != nil {
		return err
	}
	tempPath := file.Name()

	if err := os.Chmod(tempPath, 0600); err != nil {
		file.Close()
		os.Remove(tempPath)
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(allRepos); err != nil {
		file.Close()
		os.Remove(tempPath)
		return err
	}
	
	file.Close()

	indexPath := filepath.Join(cacheDir, "index.json")
	return os.Rename(tempPath, indexPath)
}

func LoadIndex() ([]RepoEntry, error) {
	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return nil, err
	}

	indexPath := filepath.Join(cacheDir, "index.json")
	
	// Fallback to old index.txt if JSON doesn't exist yet
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		oldPath := filepath.Join(cacheDir, "index.txt")
		if _, err := os.Stat(oldPath); err == nil {
			return migrateOldIndex(oldPath)
		}
		return nil, nil // No index yet
	}

	content, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var repos []RepoEntry
	if err := json.Unmarshal(content, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func RemoveFromIndex(pathToRemove string) error {
	repos, err := LoadIndex()
	if err != nil {
		return err
	}
	
	var filtered []RepoEntry
	for _, repo := range repos {
		if repo.Path != pathToRemove {
			filtered = append(filtered, repo)
		}
	}
	
	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return err
	}
	
	file, err := os.CreateTemp(cacheDir, "index-*.json.tmp")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	os.Chmod(tempPath, 0600)
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(filtered); err != nil {
		file.Close()
		os.Remove(tempPath)
		return err
	}
	file.Close()
	
	indexPath := filepath.Join(cacheDir, "index.json")
	return os.Rename(tempPath, indexPath)
}

func migrateOldIndex(oldPath string) ([]RepoEntry, error) {
	content, err := os.ReadFile(oldPath)
	if err != nil {
		return nil, err
	}
	
	var repos []RepoEntry
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) != "" {
			repos = append(repos, RepoEntry{
				Path: line,
				DisplayName: filepath.Base(line),
				Type: "repo",
			})
		}
	}
	return repos, nil
}
