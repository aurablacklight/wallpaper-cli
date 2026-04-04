package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/user/wallpaper-cli/internal/utils"
)

func TestSetCommandHelp(t *testing.T) {
	// Reset output
	buf := new(bytes.Buffer)
	setCmd.SetOut(buf)
	setCmd.SetErr(buf)

	err := setCmd.Help()
	if err != nil {
		t.Errorf("Help() error = %v", err)
	}

	help := buf.String()
	if !strings.Contains(help, "set") {
		t.Error("Help should mention 'set'")
	}
	if !strings.Contains(help, "random") {
		t.Error("Help should mention --random flag")
	}
	if !strings.Contains(help, "latest") {
		t.Error("Help should mention --latest flag")
	}
	if !strings.Contains(help, "current") {
		t.Error("Help should mention --current flag")
	}
}

func TestSetCommandValidation(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "wallpaper-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with non-existent file
	err = setCmd.RunE(setCmd, []string{"/nonexistent/image.jpg"})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with non-image file
	txtFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(txtFile, []byte("not an image"), 0644)

	err = setCmd.RunE(setCmd, []string{txtFile})
	if err == nil {
		t.Error("Expected error for non-image file")
	}

	// Test with directory
	err = setCmd.RunE(setCmd, []string{tmpDir})
	if err == nil {
		t.Error("Expected error for directory")
	}
}

func TestSetFlags(t *testing.T) {
	// Reset flags
	randomFlag = false
	latestFlag = false
	currentFlag = false

	// Parse flags
	setCmd.ParseFlags([]string{"--random"})
	if !randomFlag {
		t.Error("--random flag should be set")
	}

	// Reset and test --latest
	randomFlag = false
	setCmd.ParseFlags([]string{"--latest"})
	if !latestFlag {
		t.Error("--latest flag should be set")
	}

	// Reset and test --current
	latestFlag = false
	setCmd.ParseFlags([]string{"--current"})
	if !currentFlag {
		t.Error("--current flag should be set")
	}
}

func TestIsImageFile(t *testing.T) {
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

	for _, ext := range validExtensions {
		t.Run("valid_"+ext, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-*"+ext)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			if !utils.IsImageFile(tmpFile.Name()) {
				t.Errorf("IsImageFile should return true for %s", ext)
			}
		})
	}

	invalidExtensions := []string{".txt", ".pdf", ".doc", ".go"}
	for _, ext := range invalidExtensions {
		t.Run("invalid_"+ext, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-*"+ext)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			if utils.IsImageFile(tmpFile.Name()) {
				t.Errorf("IsImageFile should return false for %s", ext)
			}
		})
	}
}

func TestFindWallpapers(t *testing.T) {
	// Create temp directory with some images
	tmpDir, err := os.MkdirTemp("", "wallpaper-find-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create image files
	for _, name := range []string{"img1.jpg", "img2.png", "img3.gif"} {
		f, _ := os.Create(filepath.Join(tmpDir, name))
		f.Close()
	}

	// Create non-image file
	f, _ := os.Create(filepath.Join(tmpDir, "readme.txt"))
	f.Close()

	// Create subdirectory with image
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	f, _ = os.Create(filepath.Join(subDir, "nested.jpg"))
	f.Close()

	wallpapers, err := utils.FindWallpapers(tmpDir)
	if err != nil {
		t.Fatalf("FindWallpapers failed: %v", err)
	}

	if len(wallpapers) != 4 {
		t.Errorf("Expected 4 wallpapers, got %d", len(wallpapers))
	}
}

func TestGetLatestWallpaper(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wallpaper-latest-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Empty directory should return error
	_, err = utils.GetLatestWallpaper(tmpDir)
	if err == nil {
		t.Error("Expected error for empty directory")
	}

	// Create images with delays
	f1, _ := os.Create(filepath.Join(tmpDir, "old.jpg"))
	f1.Close()

	// Small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	f2, _ := os.Create(filepath.Join(tmpDir, "new.png"))
	f2.Close()

	latest, err := utils.GetLatestWallpaper(tmpDir)
	if err != nil {
		t.Fatalf("GetLatestWallpaper failed: %v", err)
	}

	if !strings.Contains(latest, "new.png") {
		t.Errorf("Expected latest to be new.png, got %s", latest)
	}
}

func TestGetRandomWallpaper(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wallpaper-random-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Empty directory should return error
	_, err = utils.GetRandomWallpaper(tmpDir)
	if err == nil {
		t.Error("Expected error for empty directory")
	}

	// Create images
	for i := 0; i < 5; i++ {
		f, _ := os.Create(filepath.Join(tmpDir, fmt.Sprintf("img%d.jpg", i)))
		f.Close()
	}

	// Should return one of the images
	random, err := utils.GetRandomWallpaper(tmpDir)
	if err != nil {
		t.Fatalf("GetRandomWallpaper failed: %v", err)
	}

	if !strings.HasSuffix(random, ".jpg") {
		t.Errorf("Expected .jpg file, got %s", random)
	}
}
