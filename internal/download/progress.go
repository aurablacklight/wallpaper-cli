package download

import (
	"fmt"
	"os"
	"sync"
)

// TextProgress provides simple text-based progress output
type TextProgress struct {
	mu        sync.Mutex
	downloads map[string]*downloadStatus
	skipped   int
}

type downloadStatus struct {
	url        string
	index      int
	total      int
	downloaded int64
	totalSize  int64
	complete   bool
	err        error
}

// NewTextProgress creates a new text progress tracker
func NewTextProgress() *TextProgress {
	return &TextProgress{
		downloads: make(map[string]*downloadStatus),
	}
}

// OnStart is called when a download starts
func (tp *TextProgress) OnStart(url string, index int, total int) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.downloads[url] = &downloadStatus{
		url:   url,
		index: index,
		total: total,
	}

	fmt.Fprintf(os.Stderr, "[%d/%d] Starting download...\n", index+1, total)
}

// OnProgress is called when download progresses
func (tp *TextProgress) OnProgress(url string, downloaded, total int64) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if status, ok := tp.downloads[url]; ok {
		status.downloaded = downloaded
		status.totalSize = total
	}
}

// OnComplete is called when a download completes
func (tp *TextProgress) OnComplete(url string, path string, size int64) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if status, ok := tp.downloads[url]; ok {
		status.complete = true
		status.totalSize = size
	}

	// Format size
	sizeStr := formatBytes(size)
	fmt.Fprintf(os.Stderr, "✓ Downloaded %s (%s)\n", path, sizeStr)
}

// OnError is called when a download fails
func (tp *TextProgress) OnError(url string, err error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if status, ok := tp.downloads[url]; ok {
		status.err = err
	}

	fmt.Fprintf(os.Stderr, "✗ Failed to download: %v\n", err)
}

// OnSkip is called when a download is skipped (duplicate)
func (tp *TextProgress) OnSkip(url string, reason string) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.skipped++
	fmt.Fprintf(os.Stderr, "⊘ Skipped (duplicate: %s)\n", reason)
}

// PrintSummary prints a summary of all downloads
func (tp *TextProgress) PrintSummary() {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	var completed, failed int
	var totalBytes int64

	for _, status := range tp.downloads {
		if status.err != nil {
			failed++
		} else if status.complete {
			completed++
			totalBytes += status.totalSize
		}
	}

	total := len(tp.downloads)
	fmt.Fprintf(os.Stderr, "\nDownloaded: %d/%d (%s total)\n", completed, total, formatBytes(totalBytes))
	if tp.skipped > 0 {
		fmt.Fprintf(os.Stderr, "Skipped (duplicates): %d\n", tp.skipped)
	}
	if failed > 0 {
		fmt.Fprintf(os.Stderr, "Failed: %d\n", failed)
	}
}

// formatBytes formats byte size to human readable
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// SilentProgress is a no-op progress tracker
type SilentProgress struct{}

func (sp *SilentProgress) OnStart(url string, index int, total int)   {}
func (sp *SilentProgress) OnProgress(url string, downloaded, total int64) {}
func (sp *SilentProgress) OnComplete(url string, path string, size int64) {}
func (sp *SilentProgress) OnError(url string, err error)                   {}
func (sp *SilentProgress) OnSkip(url string, reason string)                 {}
