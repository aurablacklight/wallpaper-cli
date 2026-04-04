package platform

import (
	"fmt"
	"os"
	"os/exec"
)

type windowsSetter struct{}

func (w *windowsSetter) Platform() string {
	return "Windows"
}

func (w *windowsSetter) SetWallpaper(path string) error {
	// Validate file exists per D-13
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("wallpaper file not found: %w", err)
	}

	// Per D-06: Use PowerShell to set registry
	// Command: Set-ItemProperty -Path "HKCU:\Control Panel\Desktop" -Name Wallpaper -Value "path"
	psCmd := fmt.Sprintf(
		`Set-ItemProperty -Path "HKCU:\Control Panel\Desktop" -Name Wallpaper -Value %q`,
		path,
	)

	cmd := exec.Command("powershell", "-Command", psCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set wallpaper registry: %w (output: %s)", err, output)
	}

	// Refresh desktop to apply changes
	// Per D-06: rundll32.exe user32.dll,UpdatePerUserSystemParameters
	refresh := exec.Command("rundll32.exe", "user32.dll,UpdatePerUserSystemParameters")
	if err := refresh.Run(); err != nil {
		// Log warning but don't fail - wallpaper may still update
		// Some Windows versions apply without explicit refresh
	}

	return nil
}
