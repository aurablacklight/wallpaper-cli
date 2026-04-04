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
