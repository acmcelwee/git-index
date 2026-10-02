package frecency

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/acmcelwee/git-index/internal/config"
)

type Event struct {
	Path       string  `json:"path"`
	Timestamp  int64   `json:"timestamp,omitempty"`  // Legacy fallback
	Timestamps []int64 `json:"timestamps,omitempty"` // Array of timestamps
}

type Entry struct {
	Path       string
	Timestamps []int64
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
			continue
		}

		entry := entries[event.Path]
		entry.Path = event.Path
		
		if event.Timestamp > 0 {
			entry.Timestamps = append(entry.Timestamps, event.Timestamp)
		}
		if len(event.Timestamps) > 0 {
			entry.Timestamps = append(entry.Timestamps, event.Timestamps...)
		}
		
		entries[event.Path] = entry
	}

	return entries, nil
}

func Compact() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	maxAgeStr := cfg.MaxFrecencyAge
	if maxAgeStr == "" {
		maxAgeStr = "2160h" // 90 days default
	}
	
	maxAge, err := time.ParseDuration(maxAgeStr)
	if err != nil {
		maxAge = 90 * 24 * time.Hour
	}
	
	cutoff := time.Now().Add(-maxAge).Unix()

	entries, err := Load()
	if err != nil {
		return err
	}

	logPath, err := getPath()
	if err != nil {
		return err
	}

	tmpPath := logPath + ".tmp"
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	for path, entry := range entries {
		// Prune stale paths
		if stat, err := os.Stat(path); err != nil || !stat.IsDir() {
			continue
		}

		// Prune old timestamps
		var validTimestamps []int64
		for _, ts := range entry.Timestamps {
			if ts >= cutoff {
				validTimestamps = append(validTimestamps, ts)
			}
		}

		if len(validTimestamps) == 0 {
			continue
		}

		event := Event{
			Path:       path,
			Timestamps: validTimestamps,
		}
		
		data, err := json.Marshal(event)
		if err == nil {
			data = append(data, '\n')
			file.Write(data)
		}
	}
	file.Close()

	return os.Rename(tmpPath, logPath)
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

	event := Event{
		Path:       path,
		Timestamps: []int64{time.Now().Unix()},
	}

	data, err := json.Marshal(event)
	if err != nil {
		file.Close()
		return err
	}

	data = append(data, '\n')
	_, err = file.Write(data)
	
	// Check size for compaction before closing
	stat, statErr := file.Stat()
	file.Close()
	
	if statErr == nil && err == nil {
		cfg, cfgErr := config.LoadConfig()
		var maxSize int64 = 50 * 1024
		if cfgErr == nil && cfg.MaxFrecencyLogSize > 0 {
			maxSize = cfg.MaxFrecencyLogSize
		}
		
		if stat.Size() > maxSize {
			Compact()
		}
	}

	return err
}

func GetHalfLife() float64 {
	cfg, err := config.LoadConfig()
	if err != nil {
		return 72 * 3600 // 3 days default
	}

	hlStr := cfg.FrecencyHalfLife
	if hlStr == "" {
		hlStr = "72h"
	}

	duration, err := time.ParseDuration(hlStr)
	if err != nil {
		return 72 * 3600
	}

	return duration.Seconds()
}

func Score(entry Entry, halfLifeSecs float64) float64 {
	if len(entry.Timestamps) == 0 {
		return 0
	}

	now := time.Now().Unix()
	var score float64

	for _, ts := range entry.Timestamps {
		dt := float64(now - ts)
		if dt < 0 {
			dt = 0 // Future? Shouldn't happen, but safe
		}

		// Calculate continuous exponential decay: score = 1.0 * (0.5 ^ (dt / halfLife))
		score += math.Pow(0.5, dt/halfLifeSecs)
	}

	return score
}
