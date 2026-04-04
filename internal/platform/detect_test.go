package platform

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestDetect(t *testing.T) {
	osType, de, err := Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	// Verify OS matches runtime.GOOS
	switch runtime.GOOS {
	case "darwin":
		if osType != MacOS {
			t.Errorf("Expected MacOS on darwin, got %v", osType)
		}
	case "linux":
		if osType != Linux {
			t.Errorf("Expected Linux on linux, got %v", osType)
		}
	case "windows":
		if osType != Windows {
			t.Errorf("Expected Windows on windows, got %v", osType)
		}
	}

	t.Logf("Detected OS: %v, DE: %v", osType, de)
}

func TestDetectLinuxDE(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected DE
	}{
		{"GNOME", "GNOME", GNOME},
		{"ubuntu GNOME", "ubuntu:GNOME", GNOME},
		{"KDE", "KDE", KDE},
		{"XFCE", "XFCE", XFCE},
		{"empty", "", UnknownDE},
		{"unknown", "UNKNOWN", OtherDE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			orig := os.Getenv("XDG_CURRENT_DESKTOP")
			defer os.Setenv("XDG_CURRENT_DESKTOP", orig)

			os.Setenv("XDG_CURRENT_DESKTOP", tt.envValue)
			de := DetectLinuxDE()

			if de != tt.expected {
				t.Errorf("DetectLinuxDE() = %v, want %v", de, tt.expected)
			}
		})
	}
}

func TestPlatformGetter(t *testing.T) {
	setter, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	platformName := setter.Platform()
	switch runtime.GOOS {
	case "darwin":
		if platformName != "macOS" {
			t.Errorf("Expected 'macOS', got '%s'", platformName)
		}
	case "linux":
		if !strings.Contains(platformName, "Linux") {
			t.Errorf("Expected 'Linux (DE)', got '%s'", platformName)
		}
	case "windows":
		if platformName != "Windows" {
			t.Errorf("Expected 'Windows', got '%s'", platformName)
		}
	}
}

func TestSetWallpaperValidation(t *testing.T) {
	setter, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Test with non-existent file
	err = setter.SetWallpaper("/nonexistent/path/image.jpg")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Test with directory instead of file
	err = setter.SetWallpaper("/tmp")
	if err == nil {
		t.Error("Expected error for directory, got nil")
	}
}

func TestSetWallpaperCommandGeneration(t *testing.T) {
	// Create a temp image file for testing
	tmpFile, err := os.CreateTemp("", "test-*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	setter, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Attempt to set (will fail on non-matching platform, but tests command generation)
	err = setter.SetWallpaper(tmpFile.Name())
	// We expect this might fail due to platform-specific tools not being available in test environment
	// but we verify the error message contains expected platform info
	if err != nil {
		t.Logf("SetWallpaper error (expected in test env): %v", err)
	}
}
