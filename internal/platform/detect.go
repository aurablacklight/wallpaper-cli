package platform

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Detect returns the current OS and Linux DE (if applicable)
func Detect() (OS, DE, error) {
	switch runtime.GOOS {
	case "darwin":
		return MacOS, UnknownDE, nil
	case "linux":
		de := DetectLinuxDE()
		return Linux, de, nil
	case "windows":
		return Windows, UnknownDE, nil
	default:
		return UnknownOS, UnknownDE, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// DetectLinuxDE detects the Linux desktop environment from XDG_CURRENT_DESKTOP
func DetectLinuxDE() DE {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	desktop = strings.ToUpper(desktop)

	// Handle variants like "ubuntu:GNOME"
	if strings.Contains(desktop, "GNOME") {
		return GNOME
	}

	switch desktop {
	case "KDE":
		return KDE
	case "XFCE":
		return XFCE
	case "":
		return UnknownDE
	default:
		return OtherDE
	}
}
