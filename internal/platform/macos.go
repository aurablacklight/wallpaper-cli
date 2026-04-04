package platform

import (
	"fmt"
	"os"
	"os/exec"
)

type macOSSetter struct{}

func (m *macOSSetter) Platform() string {
	return "macOS"
}

func (m *macOSSetter) SetWallpaper(path string) error {
	// Validate file exists per D-13
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("wallpaper file not found: %w", err)
	}

	// Use AppleScript per D-01, D-11
	// Command: osascript -e 'tell application "Finder" to set desktop picture to POSIX file "path"'
	cmd := exec.Command("osascript", "-e",
		fmt.Sprintf(`tell application "Finder" to set desktop picture to POSIX file %q`, path))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set wallpaper: %w (output: %s)", err, output)
	}
	return nil
}
