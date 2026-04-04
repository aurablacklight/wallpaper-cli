package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ to the user's home directory and converts to absolute path
func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = strings.Replace(path, "~", home, 1)
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	// Convert to absolute path
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(wd, path)
	}

	// Clean the path
	path = filepath.Clean(path)

	return path, nil
}

// ParseWallpaperFilename extracts metadata from wallpaper filename
// Expected format: {index}_{id}_{resolution}.{ext}
// Example: 01_8g5dp1_3840x2160.jpg
func ParseWallpaperFilename(filename string) (id, resolution string) {
	parts := SplitFilename(filename)
	if len(parts) >= 3 {
		return parts[1], parts[2]
	}
	return "", ""
}

// SplitFilename splits a filename by underscore, removing extension
// Example: "01_8g5dp1_3840x2160.jpg" -> ["01", "8g5dp1", "3840x2160"]
func SplitFilename(filename string) []string {
	// Remove extension
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			filename = filename[:i]
			break
		}
	}

	// Split by underscore
	return strings.Split(filename, "_")
}
