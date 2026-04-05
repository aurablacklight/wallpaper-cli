package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultSource != "wallhaven" {
		t.Errorf("DefaultSource = %q, want wallhaven", cfg.DefaultSource)
	}
	if cfg.ConcurrentDownloads != 5 {
		t.Errorf("ConcurrentDownloads = %d, want 5", cfg.ConcurrentDownloads)
	}
	if !cfg.Dedup {
		t.Error("Dedup should default to true")
	}
	if cfg.Sources["wallhaven"].Enabled != true {
		t.Error("wallhaven source should be enabled by default")
	}
	if cfg.Sources["reddit"].Enabled != true {
		t.Error("reddit source should be enabled by default")
	}
	if cfg.WallpaperHistory == nil {
		t.Error("WallpaperHistory should be initialized, not nil")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := DefaultConfig()
	cfg.DefaultResolution = "4k"

	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.DefaultResolution != "4k" {
		t.Errorf("DefaultResolution = %q, want 4k", loaded.DefaultResolution)
	}
	if loaded.DefaultSource != "wallhaven" {
		t.Errorf("DefaultSource = %q, want wallhaven", loaded.DefaultSource)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/config.json")
	if err != nil {
		t.Fatalf("Load missing file should return default, got error: %v", err)
	}
	if cfg.DefaultSource != "wallhaven" {
		t.Error("missing file should return default config")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("{invalid"), 0644)

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestAddWallpaper(t *testing.T) {
	cfg := DefaultConfig()

	cfg.AddWallpaper("/tmp/wall1.jpg", "manual")
	if cfg.CurrentWallpaper != "/tmp/wall1.jpg" {
		t.Errorf("CurrentWallpaper = %q, want /tmp/wall1.jpg", cfg.CurrentWallpaper)
	}
	if len(cfg.WallpaperHistory) != 1 {
		t.Fatalf("history len = %d, want 1", len(cfg.WallpaperHistory))
	}

	// Add 11 items — should cap at 10
	for i := 0; i < 11; i++ {
		cfg.AddWallpaper("/tmp/wall.jpg", "random")
	}
	if len(cfg.WallpaperHistory) != 10 {
		t.Errorf("history len = %d, want 10 (capped)", len(cfg.WallpaperHistory))
	}
}

func TestAddWallpaper_NewestFirst(t *testing.T) {
	cfg := DefaultConfig()

	cfg.AddWallpaper("/tmp/old.jpg", "manual")
	cfg.AddWallpaper("/tmp/new.jpg", "manual")

	if cfg.WallpaperHistory[0].Path != "/tmp/new.jpg" {
		t.Error("newest wallpaper should be first in history")
	}
}
