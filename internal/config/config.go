package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds all user configuration
type Config struct {
	DefaultSource       string            `json:"default_source"`
	DefaultResolution   string            `json:"default_resolution"`
	OutputDirectory     string            `json:"output_directory"`
	Organization        string            `json:"organization"`
	Format              string            `json:"format"`
	Dedup               bool              `json:"dedup"`
	DedupThreshold      int               `json:"dedup_threshold"`
	ConcurrentDownloads int               `json:"concurrent_downloads"`
	Sources             map[string]SourceConfig `json:"sources"`

	// Wallpaper state per D-08, D-09
	CurrentWallpaper string            `json:"current_wallpaper,omitempty"`
	WallpaperHistory []WallpaperRecord `json:"wallpaper_history,omitempty"`
}

// WallpaperRecord stores a wallpaper setting event per D-09
type WallpaperRecord struct {
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // "manual", "random", "latest"
}

// SourceConfig holds per-source configuration
type SourceConfig struct {
	Enabled    bool     `json:"enabled"`
	APIKey     string   `json:"api_key,omitempty"`
	Login      string   `json:"login,omitempty"`      // Danbooru login username
	Username   string   `json:"username,omitempty"`    // Zerochan username for User-Agent
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
			"danbooru": {
				Enabled: true,
			},
			"konachan": {
				Enabled: true,
			},
			"zerochan": {
				Enabled: true,
			},
		},
		WallpaperHistory: make([]WallpaperRecord, 0), // Initialize empty
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

// AddWallpaper adds a wallpaper to history per D-09
// Maintains max 10 entries (oldest removed)
func (c *Config) AddWallpaper(path string, source string) {
	record := WallpaperRecord{
		Path:      path,
		Timestamp: time.Now(),
		Source:    source,
	}

	// Prepend to history (newest first)
	c.WallpaperHistory = append([]WallpaperRecord{record}, c.WallpaperHistory...)

	// Limit to 10 entries per D-09
	if len(c.WallpaperHistory) > 10 {
		c.WallpaperHistory = c.WallpaperHistory[:10]
	}

	// Update current wallpaper per D-08
	c.CurrentWallpaper = path
}
