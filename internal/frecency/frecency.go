package frecency

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/acmcelwee/git-index/internal/config"
)

type Event struct {
	Path      string `json:"path"`
	Timestamp int64  `json:"timestamp"`
}

type Entry struct {
	Path         string
	Count        int
	LastAccessed int64
}

func getPath() (string, error) {
	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "frecency.jsonl"), nil
}

func Load() (map[string]Entry, error) {
	path, err := getPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]Entry), nil
		}
		return nil, err
	}
	defer file.Close()

	entries := make(map[string]Entry)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue // Skip invalid lines
		}

		entry := entries[event.Path]
		entry.Path = event.Path
		entry.Count++
		if event.Timestamp > entry.LastAccessed {
			entry.LastAccessed = event.Timestamp
		}
		entries[event.Path] = entry
	}

	return entries, nil
}

func Track(path string) error {
	logPath, err := getPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(logPath)
	os.MkdirAll(dir, 0700)

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	event := Event{
		Path:      path,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = file.Write(data)
	return err
}

func Score(entry Entry) float64 {
	if entry.Count == 0 {
		return 0
	}

	now := time.Now().Unix()
	dt := now - entry.LastAccessed

	// dt is in seconds
	var multiplier float64
	if dt < 3600 { // 1 hour
		multiplier = 4.0
	} else if dt < 86400 { // 1 day
		multiplier = 2.0
	} else if dt < 604800 { // 1 week
		multiplier = 0.5
	} else {
		multiplier = 0.1
	}

	return float64(entry.Count) * multiplier
}
