package model

import (
	"fmt"
	"time"
)

// Wallpaper represents a wallpaper to download
type Wallpaper struct {
	ID          string
	Source      string // "wallhaven", "reddit", etc.
	SourceID    string // ID from source
	URL         string // Direct download URL
	Title       string
	Resolution  string // "3840x2160"
	AspectRatio string // "16x9"
	Tags        []string
	FileSize    int64
	Format      string // jpg, png, webp
	Purity      string // sfw, sketchy, nsfw
	Category    string
	CreatedAt   time.Time
}

// GetFilename generates a filename for the wallpaper
func (w *Wallpaper) GetFilename() string {
	// Use source_id + resolution + extension
	ext := w.Format
	if ext == "" {
		ext = "jpg"
	}
	return fmt.Sprintf("%s_%s.%s", w.SourceID, w.Resolution, ext)
}
