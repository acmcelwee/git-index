package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type SearchDir struct {
	Path     string `mapstructure:"path" toml:"path"`
	MaxDepth int    `mapstructure:"max_depth" toml:"max_depth"`
}

type Config struct {
	Ignore              []string    `mapstructure:"ignore" toml:"ignore"`
	SearchDirs          []SearchDir `mapstructure:"search_dirs" toml:"search_dirs"`
	AutoRebuildInterval string      `mapstructure:"auto_rebuild_interval" toml:"auto_rebuild_interval"`
}

func GetConfigDir() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "git-index"), nil
}

func GetCacheDir() (string, error) {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheDir = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheDir, "git-index"), nil
}

func InitConfig() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(dir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Write a default config
			viper.Set("ignore", []string{"**/archive/**", "**/node_modules/**", "**/.terraform/**"})
			viper.Set("auto_rebuild_interval", "24h")
			
			home, _ := os.UserHomeDir()
			viper.Set("search_dirs", []SearchDir{
				{Path: filepath.Join(home, "src"), MaxDepth: 4},
			})
			
			if err := viper.SafeWriteConfig(); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func AddSearchDir(path string, maxDepth int) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Check if already exists
	for _, dir := range cfg.SearchDirs {
		if dir.Path == path {
			return fmt.Errorf("directory %s is already in config", path)
		}
	}

	cfg.SearchDirs = append(cfg.SearchDirs, SearchDir{
		Path:     path,
		MaxDepth: maxDepth,
	})

	viper.Set("search_dirs", cfg.SearchDirs)
	return viper.WriteConfig()
}
