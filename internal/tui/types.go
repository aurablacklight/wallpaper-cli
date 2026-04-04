package tui

import (
	"fmt"
	"os"
	"strings"
)

// ImageMethod represents the terminal image rendering method
type ImageMethod int

const (
	MethodUnknown ImageMethod = iota
	MethodKitty
	MethodIterm2
	MethodSixel
	MethodHalfBlocks
	MethodNone
)

// detectImageMethod detects the best available image rendering method
func detectImageMethod() ImageMethod {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	// Check for Kitty
	if term == "xterm-kitty" || os.Getenv("KITTY_WINDOW_ID") != "" {
		return MethodKitty
	}

	// Check for iTerm2
	if termProgram == "iTerm.app" {
		return MethodIterm2
	}

	// Check for SIXEL support
	if hasSixelSupport() {
		return MethodSixel
	}

	// Default to half-blocks (universal)
	return MethodHalfBlocks
}

// hasSixelSupport checks if terminal supports SIXEL
func hasSixelSupport() bool {
	term := os.Getenv("TERM")
	sixelTerms := []string{"mlterm", "yaft", "xterm-256color"}
	for _, t := range sixelTerms {
		if strings.Contains(term, t) {
			return true
		}
	}
	return false
}

// WallpaperItem is the legacy type kept for backward compatibility
type WallpaperItem struct {
	Path      string
	Name      string
	Source    string
	Thumbnail string
	HasError  bool
}

// Title returns the item title for the list
func (w WallpaperItem) Title() string {
	return w.Name
}

// Description returns the item description
func (w WallpaperItem) Description() string {
	if w.HasError {
		return "⚠️ Error loading thumbnail"
	}
	source := w.Source
	if source == "" {
		source = "unknown"
	}
	return fmt.Sprintf("Source: %s", source)
}

// FilterValue returns the value to filter on
func (w WallpaperItem) FilterValue() string {
	return w.Name
}
