package download

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressBar provides a nice visual progress bar
type ProgressBar struct {
	bar       *progressbar.ProgressBar
	total     int
	started   time.Time
	completed int
	errors    int
	skipped   int
}

// NewProgressBar creates a new progress bar for batch downloads
func NewProgressBar(total int, description string) *ProgressBar {
	pb := &ProgressBar{
		total:   total,
		started: time.Now(),
	}

	pb.bar = progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	)

	return pb
}

// OnStart is called when a download starts
func (pb *ProgressBar) OnStart(url string, index int, total int) {
	// Progress bar already shows count, no need for individual messages
}

// OnProgress is called when download progresses (not used for bar)
func (pb *ProgressBar) OnProgress(url string, downloaded, total int64) {
	// Handled internally by io.TeeReader
}

// OnComplete is called when a download completes
func (pb *ProgressBar) OnComplete(url string, path string, size int64) {
	pb.completed++
	pb.bar.Add(1)
}

// OnError is called when a download fails
func (pb *ProgressBar) OnError(url string, err error) {
	pb.errors++
	pb.bar.Add(1)
	fmt.Fprintf(os.Stderr, "\n✗ %v\n", err)
}

// OnSkip is called when a download is skipped (duplicate)
func (pb *ProgressBar) OnSkip(url string, reason string) {
	pb.skipped++
	pb.bar.Add(1)
}

// PrintSummary prints a summary of all downloads
func (pb *ProgressBar) PrintSummary() {
	elapsed := time.Since(pb.started)
	
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "✓ Downloaded: %d/%d", pb.completed, pb.total)
	if pb.skipped > 0 {
		fmt.Fprintf(os.Stderr, " | Skipped: %d", pb.skipped)
	}
	if pb.errors > 0 {
		fmt.Fprintf(os.Stderr, " | Failed: %d", pb.errors)
	}
	fmt.Fprintf(os.Stderr, " | Time: %s\n", elapsed.Round(time.Second))
}

// GetBar returns the underlying progress bar for io.TeeReader
func (pb *ProgressBar) GetBar() *progressbar.ProgressBar {
	return pb.bar
}

// WrapReader wraps an io.Reader to track progress
func (pb *ProgressBar) WrapReader(r io.Reader, size int64) io.Reader {
	return io.TeeReader(r, pb.bar)
}
