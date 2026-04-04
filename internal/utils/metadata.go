package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// SetRedditURL sets the Reddit source URL as extended file attribute (macOS/Linux)
func SetRedditURL(filepath string, url string) error {
	if url == "" {
		return nil
	}

	// On macOS, use xattr to set extended attribute
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		// Use xattr command to set custom attribute
		cmd := exec.Command("xattr", "-w", "user.reddit_url", url, filepath)
		err := cmd.Run()
		if err != nil {
			// Try alternative: set com.apple.metadata attribute for macOS Finder
			cmd2 := exec.Command("xattr", "-w", "com.apple.metadata:kMDItemWhereFroms", url, filepath)
			_ = cmd2.Run() // Ignore error if this also fails
			return err
		}
		return nil
	}

	// Windows not supported for now
	return fmt.Errorf("extended attributes not supported on %s", runtime.GOOS)
}

// GetRedditURL retrieves the Reddit source URL from file extended attributes
func GetRedditURL(filepath string) (string, error) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return "", fmt.Errorf("extended attributes not supported on %s", runtime.GOOS)
	}

	// Try user.reddit_url first
	cmd := exec.Command("xattr", "-p", "user.reddit_url", filepath)
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		return string(out), nil
	}

	// Fallback to com.apple.metadata:kMDItemWhereFroms
	cmd2 := exec.Command("xattr", "-p", "com.apple.metadata:kMDItemWhereFroms", filepath)
	out2, err2 := cmd2.Output()
	if err2 == nil && len(out2) > 0 {
		return string(out2), nil
	}

	return "", fmt.Errorf("no Reddit URL found in file attributes")
}

// SetWallhavenURL sets the Wallhaven source URL as extended file attribute
func SetWallhavenURL(filepath string, url string) error {
	if url == "" {
		return nil
	}

	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		cmd := exec.Command("xattr", "-w", "user.wallhaven_url", url, filepath)
		return cmd.Run()
	}

	return fmt.Errorf("extended attributes not supported on %s", runtime.GOOS)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
