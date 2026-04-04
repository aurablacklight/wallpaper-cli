package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all user configuration
type Config struct {
	DefaultSource      string            `json:"default_source"`
	DefaultResolution  string            `json:"default_resolution"`
	OutputDirectory    string            `json:"output_directory"`
	Organization       string            `json:"organization"`
	Format             string            `json:"format"`
	Dedup              bool              `json:"dedup"`
	DedupThreshold     int               `json:"dedup_threshold"`
	ConcurrentDownloads int              `json:"concurrent_downloads"`
	Sources            map[string]SourceConfig `json:"sources"`
}

// SourceConfig holds per-source configuration
type SourceConfig struct {
	Enabled    bool     `json:"enabled"`
	APIKey     string   `json:"api_key,omitempty"`
	Subreddits []string `json:"subreddits,omitempty"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		DefaultSource:       "wallhaven",
		DefaultResolution:   "4k",
		OutputDirectory:     filepath.Join(home, "Pictures", "wallpapers"),
		Organization:        "source",
		Format:              "original",
		Dedup:               true,
		DedupThreshold:      10,
		ConcurrentDownloads: 5,
		Sources: map[string]SourceConfig{
			"wallhaven": {
				Enabled: true,
			},
			"reddit": {
				Enabled:    true,
				Subreddits: []string{"Animewallpaper"},
			},
		},
	}
}

// Load reads configuration from the specified path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// Save writes configuration to the specified path
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "wallpaper-cli", "config.json")
}
