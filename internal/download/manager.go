package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	maxRetries  int
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
	MaxRetries  int
}

// DefaultOptions returns default download options
func DefaultOptions() *Options {
	return &Options{
		Concurrency: 5,
		OutputDir:   "",
		Timeout:     30 * time.Second,
		EnableDedup: true,
		MaxRetries:  3,
	}
}

// NewManager creates a new download manager with legacy progress tracker
func NewManager(opts *Options, progress ProgressTracker, checker *dedup.Checker) *Manager {
	if opts == nil {
		opts = DefaultOptions()
	}
	maxRetries := opts.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &Manager{
		concurrency: opts.Concurrency,
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
		progress:   progress,
		checker:    checker,
		useBar:     false,
		maxRetries: maxRetries,
	}
}

// NewManagerWithBar creates a new download manager with progress bar
func NewManagerWithBar(opts *Options, bar *progressbar.ProgressBar, checker *dedup.Checker) *Manager {
	if opts == nil {
		opts = DefaultOptions()
	}
	maxRetries := opts.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &Manager{
		concurrency: opts.Concurrency,
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
		bar:        bar,
		checker:    checker,
		useBar:     true,
		maxRetries: maxRetries,
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
	URL     string
	Path    string
	Size    int64
	Skipped bool
	Error   error
}

// Stats returns download statistics
func (m *Manager) Stats() (completed, skipped, errors int) {
	return m.completed, m.skipped, m.errors
}

// DownloadBatch downloads multiple files concurrently
func (m *Manager) DownloadBatch(ctx context.Context, jobs []DownloadJob) []DownloadResult {
	results := make([]DownloadResult, len(jobs))

	jobCh := make(chan DownloadJob, len(jobs))
	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)

	resultCh := make(chan struct {
		idx int
		res DownloadResult
	}, len(jobs))

	var wg sync.WaitGroup
	for i := 0; i < m.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				res := m.downloadWithRetry(ctx, job)
				resultCh <- struct {
					idx int
					res DownloadResult
				}{job.Index, res}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for r := range resultCh {
		results[r.idx] = r.res
		if r.res.Error != nil {
			m.errors++
		} else if r.res.Skipped {
			m.skipped++
		} else {
			m.completed++
		}
	}

	return results
}

// downloadWithRetry wraps downloadOne with retry logic and backoff.
func (m *Manager) downloadWithRetry(ctx context.Context, job DownloadJob) DownloadResult {
	var lastResult DownloadResult

	for attempt := 0; attempt <= m.maxRetries; attempt++ {
		if attempt > 0 {
			// Check for Retry-After from last attempt
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return DownloadResult{URL: job.URL, Error: ctx.Err()}
			}
		}

		lastResult = m.downloadOne(ctx, job)
		if lastResult.Error == nil || lastResult.Skipped {
			return lastResult
		}

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return lastResult
		}
	}

	return lastResult
}

// downloadOne downloads a single file, supporting resume via HTTP Range.
func (m *Manager) downloadOne(ctx context.Context, job DownloadJob) DownloadResult {
	if !m.useBar && m.progress != nil {
		m.progress.OnStart(job.URL, job.Index, job.Total)
	}

	partPath := job.Filename + ".part"

	dir := filepath.Dir(job.Filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return DownloadResult{URL: job.URL, Error: err}
	}

	// Check for existing .part file for resume
	var existingSize int64
	if info, err := os.Stat(partPath); err == nil {
		existingSize = info.Size()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", job.URL, nil)
	if err != nil {
		return DownloadResult{URL: job.URL, Error: err}
	}

	// Set Range header for resume
	if existingSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return DownloadResult{URL: job.URL, Error: err}
	}
	defer resp.Body.Close()

	// Handle HTTP 429 with Retry-After
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return DownloadResult{
			URL:   job.URL,
			Error: fmt.Errorf("HTTP 429: retry after %v", retryAfter),
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return DownloadResult{URL: job.URL, Error: fmt.Errorf("HTTP %d", resp.StatusCode)}
	}

	// Handle resume: if server returned 200 instead of 206, it doesn't support Range
	// Discard partial file and start fresh
	var file *os.File
	if existingSize > 0 && resp.StatusCode == http.StatusOK {
		// Server ignored Range header — restart
		os.Remove(partPath)
		existingSize = 0
	}

	if existingSize > 0 && resp.StatusCode == http.StatusPartialContent {
		// Append to existing .part file
		file, err = os.OpenFile(partPath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		file, err = os.Create(partPath)
	}
	if err != nil {
		return DownloadResult{URL: job.URL, Error: err}
	}

	// Copy with optional progress bar
	var written int64
	if m.useBar && m.bar != nil && resp.ContentLength > 0 {
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
		// Keep .part file for potential resume
		return DownloadResult{URL: job.URL, Error: err}
	}

	totalSize := existingSize + written

	// Check for duplicates
	if m.checker != nil {
		result, err := m.checker.CheckFile(partPath)
		if err != nil {
			os.Remove(partPath)
			return DownloadResult{URL: job.URL, Error: err}
		}

		if result.IsDuplicate {
			os.Remove(partPath)
			if !m.useBar && m.progress != nil {
				m.progress.OnSkip(job.URL, fmt.Sprintf("duplicate of %s", result.Existing.LocalPath))
			}
			return DownloadResult{URL: job.URL, Skipped: true}
		}
	}

	// Atomic rename .part → final
	if err := os.Rename(partPath, job.Filename); err != nil {
		os.Remove(partPath)
		return DownloadResult{URL: job.URL, Error: err}
	}

	if !m.useBar && m.progress != nil {
		m.progress.OnComplete(job.URL, job.Filename, totalSize)
	}

	return DownloadResult{
		URL:  job.URL,
		Path: job.Filename,
		Size: totalSize,
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

// parseRetryAfter parses the Retry-After header value into a duration.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 5 * time.Second
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	return 5 * time.Second
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
