package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type linuxSetter struct {
	de DE
}

func (l *linuxSetter) Platform() string {
	switch l.de {
	case GNOME:
		return "Linux (GNOME)"
	case KDE:
		return "Linux (KDE)"
	case XFCE:
		return "Linux (XFCE)"
	default:
		return "Linux (Unknown DE)"
	}
}

func (l *linuxSetter) SetWallpaper(path string) error {
	// Validate file exists per D-13
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("wallpaper file not found: %w", err)
	}

	// Get absolute path for URI construction
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("getting absolute path: %w", err)
	}

	switch l.de {
	case GNOME:
		return l.setGNOME(absPath)
	case KDE:
		return l.setKDE(absPath)
	case XFCE:
		return l.setXFCE(absPath)
	default:
		return l.setFallback(absPath)
	}
}

func (l *linuxSetter) setGNOME(path string) error {
	// Per D-03: gsettings set org.gnome.desktop.background picture-uri "file://path"
	uri := "file://" + path
	cmd := exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri", uri)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gsettings failed: %w (output: %s)", err, output)
	}
	return nil
}

func (l *linuxSetter) setKDE(path string) error {
	// Per D-03: Use qdbus with PlasmaShell script
	uri := "file://" + path
	script := fmt.Sprintf(`
var allDesktops = desktops();
for (var i = 0; i < allDesktops.length; i++) {
    var desktop = allDesktops[i];
    desktop.wallpaperPlugin = "org.kde.image";
    desktop.currentConfigGroup = ["Wallpaper", "org.kde.image", "General"];
    desktop.writeConfig("Image", %q);
}`, uri)

	cmd := exec.Command("qdbus", "org.kde.plasmashell", "/PlasmaShell", "evaluateScript", script)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("kde wallpaper set failed: %w (output: %s)", err, output)
	}
	return nil
}

func (l *linuxSetter) setXFCE(path string) error {
	// Per D-03: Use xfconf-query
	// First, list all backdrop properties
	listCmd := exec.Command("xfconf-query", "-c", "xfce4-desktop", "-l")
	output, err := listCmd.Output()
	if err != nil {
		// Try setting common property paths directly
		return l.setXFCEFallback(path)
	}

	// Parse output and set all last-image properties
	lines := strings.Split(string(output), "\n")
	var lastErr error
	setCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "last-image") && strings.HasPrefix(line, "/backdrop/") {
			cmd := exec.Command("xfconf-query", "-c", "xfce4-desktop", "-p", line, "-s", path)
			if err := cmd.Run(); err != nil {
				lastErr = err
			} else {
				setCount++
			}
		}
	}

	if setCount == 0 && lastErr != nil {
		return l.setXFCEFallback(path)
	}

	return nil
}

func (l *linuxSetter) setXFCEFallback(path string) error {
	// Try common XFCE property paths
	paths := []string{
		"/backdrop/screen0/monitor0/workspace0/last-image",
		"/backdrop/screen0/monitor0/workspace1/last-image",
		"/backdrop/screen0/monitor0/workspace2/last-image",
		"/backdrop/screen0/monitor0/workspace3/last-image",
	}

	for _, prop := range paths {
		cmd := exec.Command("xfconf-query", "-c", "xfce4-desktop", "-p", prop, "-s", path)
		cmd.Run() // Ignore errors, try all paths
	}

	return nil
}

func (l *linuxSetter) setFallback(path string) error {
	// Per D-04: Try feh first, then nitrogen
	if _, err := exec.LookPath("feh"); err == nil {
		cmd := exec.Command("feh", "--bg-fill", path)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("feh failed: %w", err)
		}
		return nil
	}

	if _, err := exec.LookPath("nitrogen"); err == nil {
		cmd := exec.Command("nitrogen", "--set-zoom-fill", path)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("nitrogen failed: %w", err)
		}
		return nil
	}

	return fmt.Errorf("no supported wallpaper setter found (tried: gsettings, feh, nitrogen). Desktop environment: %s", l.de)
}
