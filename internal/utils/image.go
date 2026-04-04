package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Supported image extensions per D-13
var imageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

// IsImageFile checks if a file has a supported image extension
func IsImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, valid := range imageExtensions {
		if ext == valid {
			return true
		}
	}
	return false
}

// FindWallpapers recursively finds all image files in a directory
// Used by --random flag per D-15
func FindWallpapers(dir string) ([]string, error) {
	var wallpapers []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip inaccessible files/directories but continue walking
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if IsImageFile(path) {
			wallpapers = append(wallpapers, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scanning directory: %w", err)
	}

	return wallpapers, nil
}

// GetLatestWallpaper returns the most recently modified wallpaper
// Used by --latest flag per D-16
func GetLatestWallpaper(dir string) (string, error) {
	wallpapers, err := FindWallpapers(dir)
	if err != nil {
		return "", err
	}

	if len(wallpapers) == 0 {
		return "", fmt.Errorf("no wallpapers found in %s", dir)
	}

	var latest string
	var latestTime time.Time

	for _, path := range wallpapers {
		info, err := os.Stat(path)
		if err != nil {
			continue // Skip files we can't stat
		}

		if info.ModTime().After(latestTime) || latest == "" {
			latest = path
			latestTime = info.ModTime()
		}
	}

	if latest == "" {
		return "", fmt.Errorf("no accessible wallpapers found in %s", dir)
	}

	return latest, nil
}

// GetRandomWallpaper returns a random wallpaper from the directory
// Used by --random flag per D-15
func GetRandomWallpaper(dir string) (string, error) {
	wallpapers, err := FindWallpapers(dir)
	if err != nil {
		return "", err
	}

	if len(wallpapers) == 0 {
		return "", fmt.Errorf("no wallpapers found in %s", dir)
	}

	// Use nanosecond timestamp as simple random source
	idx := time.Now().UnixNano() % int64(len(wallpapers))
	return wallpapers[idx], nil
}
