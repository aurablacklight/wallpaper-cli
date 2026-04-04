package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/user/wallpaper-cli/internal/dedup"
)

// Manager orchestrates concurrent downloads
type Manager struct {
	concurrency int
	httpClient  *http.Client
	progress    ProgressTracker
	bar         *progressbar.ProgressBar
	checker     *dedup.Checker
	mu          sync.RWMutex
	active      int
	completed   int
	errors      int
	skipped     int
	useBar      bool
}

// ProgressTracker receives download progress updates (legacy interface)
type ProgressTracker interface {
	OnStart(url string, index int, total int)
	OnProgress(url string, downloaded, total int64)
	OnComplete(url string, path string, size int64)
	OnError(url string, err error)
	OnSkip(url string, reason string)
}

// Options configures the download manager
type Options struct {
	Concurrency int
	OutputDir   string
	Timeout     time.Duration
	EnableDedup bool
}

// DefaultOptions returns default download options
func DefaultOptions() *Options {
	return &Options{
		Concurrency: 5,
		OutputDir:   "",
		Timeout:     30 * time.Second,
		EnableDedup: true,
	}
}

// NewManager creates a new download manager with legacy progress tracker
func NewManager(opts *Options, progress ProgressTracker, checker *dedup.Checker) *Manager {
	if opts == nil {
		opts = DefaultOptions()
	}

	return &Manager{
		concurrency: opts.Concurrency,
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
		progress: progress,
		checker:  checker,
		useBar:   false,
	}
}

// NewManagerWithBar creates a new download manager with progress bar
func NewManagerWithBar(opts *Options, bar *progressbar.ProgressBar, checker *dedup.Checker) *Manager {
	if opts == nil {
		opts = DefaultOptions()
	}

	return &Manager{
		concurrency: opts.Concurrency,
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
		bar:      bar,
		checker:  checker,
		useBar:   true,
	}
}

// DownloadJob represents a single download task
type DownloadJob struct {
	URL      string
	Filename string
	Index    int
	Total    int
}

// DownloadResult represents the result of a download
type DownloadResult struct {
	URL      string
	Path     string
	Size     int64
	Skipped  bool
	Error    error
}

// Stats returns download statistics
func (m *Manager) Stats() (completed, skipped, errors int) {
	return m.completed, m.skipped, m.errors
}

// DownloadBatch downloads multiple files concurrently
func (m *Manager) DownloadBatch(ctx context.Context, jobs []DownloadJob) []DownloadResult {
	results := make([]DownloadResult, len(jobs))
	
	// Create job channel
	jobCh := make(chan DownloadJob, len(jobs))
	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)

	// Create result channel
	resultCh := make(chan struct {
		idx int
		res DownloadResult
	}, len(jobs))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < m.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				res := m.downloadOne(ctx, job)
				resultCh <- struct {
					idx int
					res DownloadResult
				}{job.Index, res}
			}
		}()
	}

	// Collect results in background
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	for r := range resultCh {
		results[r.idx] = r.res
		if r.res.Error != nil {
			m.errors++
		} else if r.res.Skipped {
			m.skipped++
		} else {
			m.completed++
		}
		// Don't update progress bar here if using TeeReader (it's updated during download)
		if m.useBar && m.bar != nil {
			// Bar is updated via TeeReader during download, just ensure it reflects completion
		}
	}

	return results
}

// downloadOne downloads a single file
func (m *Manager) downloadOne(ctx context.Context, job DownloadJob) DownloadResult {
	// Notify start
	if !m.useBar && m.progress != nil {
		m.progress.OnStart(job.URL, job.Index, job.Total)
	}

	// Create temp file
	tempPath := job.Filename + ".tmp"
	
	// Ensure directory exists
	dir := filepath.Dir(job.Filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "\n✗ Error creating directory: %v\n", err)
		return DownloadResult{URL: job.URL, Error: err}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", job.URL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n✗ Error creating request: %v\n", err)
		return DownloadResult{URL: job.URL, Error: err}
	}

	// Perform request
	resp, err := m.httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n✗ Error downloading: %v\n", err)
		return DownloadResult{URL: job.URL, Error: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("HTTP %d", resp.StatusCode)
		fmt.Fprintf(os.Stderr, "\n✗ Error: %v\n", err)
		return DownloadResult{URL: job.URL, Error: err}
	}

	// Create temp file
	file, err := os.Create(tempPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n✗ Error creating file: %v\n", err)
		return DownloadResult{URL: job.URL, Error: err}
	}

	// Copy with optional progress bar (bytes)
	var written int64
	if m.useBar && m.bar != nil && resp.ContentLength > 0 {
		// Use progress bar's reader wrapper for byte tracking
		reader := &progressBarReader{
			Reader: resp.Body,
			Bar:    m.bar,
		}
		written, err = io.Copy(file, reader)
	} else {
		written, err = io.Copy(file, resp.Body)
	}
	file.Close()

	if err != nil {
		os.Remove(tempPath)
		fmt.Fprintf(os.Stderr, "\n✗ Error writing file: %v\n", err)
		return DownloadResult{URL: job.URL, Error: err}
	}

	// Check for duplicates if checker is enabled
	if m.checker != nil {
		result, err := m.checker.CheckFile(tempPath)
		if err != nil {
			os.Remove(tempPath)
			fmt.Fprintf(os.Stderr, "\n✗ Error checking duplicate: %v\n", err)
			return DownloadResult{URL: job.URL, Error: err}
		}

		if result.IsDuplicate {
			os.Remove(tempPath)
			if !m.useBar && m.progress != nil {
				m.progress.OnSkip(job.URL, fmt.Sprintf("duplicate of %s", result.Existing.LocalPath))
			}
			return DownloadResult{URL: job.URL, Skipped: true}
		}
	}

	// Atomic rename
	if err := os.Rename(tempPath, job.Filename); err != nil {
		os.Remove(tempPath)
		fmt.Fprintf(os.Stderr, "\n✗ Error finishing download: %v\n", err)
		return DownloadResult{URL: job.URL, Error: err}
	}

	// Notify complete
	if !m.useBar && m.progress != nil {
		m.progress.OnComplete(job.URL, job.Filename, written)
	}

	return DownloadResult{
		URL:  job.URL,
		Path: job.Filename,
		Size: written,
	}
}

// PrintSummary prints a summary of all downloads (legacy only)
func (m *Manager) PrintSummary() {
	if m.progress != nil {
		if tp, ok := m.progress.(*TextProgress); ok {
			tp.PrintSummary()
		}
	}
}

// progressBarReader wraps an io.Reader to update progress bar by bytes
type progressBarReader struct {
	Reader io.Reader
	Bar    *progressbar.ProgressBar
}

func (pr *progressBarReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if n > 0 && pr.Bar != nil {
		pr.Bar.Add64(int64(n))
	}
	return n, err
}

// progressReader wraps an io.Reader for progress tracking (legacy)
type progressReader struct {
	Reader     io.Reader
	Total      int64
	OnProgress func(int64)
	read       int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.read += int64(n)
	pr.OnProgress(pr.read)
	return n, err
}
