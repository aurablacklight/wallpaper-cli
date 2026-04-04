# S04: Download Manager

**Goal:** Concurrent download manager with progress tracking and resume support

**Success Criteria:**
- Concurrent downloads (5 parallel by default, configurable)
- Stream to disk (no memory buffering)
- Progress bar or status output
- HTTP Range request support for resume
- Connection pooling
- Rate limiting between requests
- Temporary file handling (atomic rename on complete)

---

## Integration Closure

Download pipeline from source adapter to filesystem complete

## Observability Impact

Download speed, active connections, queue depth observable

## Proof Level

L2 - Integration complexity

---

## Dependencies

- S03: Wallhaven Source Adapter

---

## Risk

Medium - concurrent goroutine handling, proper resource cleanup

---

## Demo

Download 10 images concurrently, progress shown, stream to disk

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched |
|----|-------|------|-------|--------|---------------|
| T01 | Design download manager architecture | 20m | Spec | internal/download/manager.go design | internal/download/manager.go |
| T02 | Implement worker pool | 30m | T01 | internal/download/worker.go | internal/download/worker.go |
| T03 | Implement streaming download | 30m | T02 | internal/download/downloader.go | internal/download/downloader.go |
| T04 | Add progress tracking | 25m | T02 | internal/download/progress.go | internal/download/progress.go |
| T05 | Add resume support | 20m | T03 | Range request handling | downloader.go updates |
| T06 | Integrate with fetch command | 15m | T02-T04 | cmd/fetch.go | cmd/fetch.go |

---

## Plan

### T01: Design Download Manager Architecture
**Estimate:** 20m

**Description:**
Design the download manager with proper abstractions for concurrent downloading.

**Architecture:**
```
Manager (orchestrator)
  └── Worker Pool (N goroutines)
        └── Downloader (per-file)
              ├── HTTP GET with streaming
              ├── Progress updates
              └── Atomic file write
```

**Key Design Decisions:**
1. Worker pool pattern with fixed concurrency
2. Channel-based job queue
3. Progress reporting via callbacks/channels
4. Context for cancellation

**Files Likely Touched:**
- internal/download/manager.go (design only)

**Expected Output:**
- Interface definitions
- Type signatures
- Flow diagram (in comments)

---

### T02: Implement Worker Pool
**Estimate:** 30m

**Description:**
Implement the worker pool for concurrent downloads.

**Steps:**
1. Create Manager struct with:
   - concurrency int (default 5)
   - jobs channel
   - results channel
   - WaitGroup for sync
2. Implement Start/Stop lifecycle
3. Add job submission
4. Implement worker goroutine
5. Add graceful shutdown

**Files Likely Touched:**
- internal/download/manager.go
- internal/download/job.go

**Expected Output:**
- Workers can process jobs concurrently
- Configurable concurrency level
- No race conditions

**Verification:**
```go
// Test with dummy jobs
m := download.NewManager(5)
m.Start()
// Submit jobs...
m.Stop()
```

---

### T03: Implement Streaming Download
**Estimate:** 30m

**Description:**
Implement the actual file download with streaming to disk.

**Steps:**
1. Create downloader.go with Download function
2. Use http.Client with timeout
3. Stream response body to file (no buffering)
4. Use io.Copy with buffer pooling
5. Atomic write (download to temp, rename on success)
6. Handle HTTP errors (404, 403, etc.)

**Key Implementation:**
```go
func Download(ctx context.Context, url string, destPath string) error {
    // Create temp file
    // HTTP GET with ctx
    // Stream to temp
    // Rename on success
}
```

**Files Likely Touched:**
- internal/download/downloader.go

**Expected Output:**
- Files downloaded without loading into memory
- Atomic writes (no partial files on interrupt)
- Proper error handling

**Verification:**
```bash
# Download a test file
./wallpaper-cli fetch --limit 1 --source wallhaven
# Check memory usage stays low
```

---

### T04: Add Progress Tracking
**Estimate:** 25m

**Description:**
Add progress reporting for downloads (bar or status lines).

**Steps:**
1. Create progress.go with ProgressTracker
2. Implement callback interface:
   - OnStart(url)
   - OnProgress(url, downloaded, total)
   - OnComplete(url, path)
   - OnError(url, error)
3. Add simple text output (progress bar as stretch)
4. Make it optional (--quiet flag support)

**Output Format:**
```
Downloading 5 wallpapers...
[1/5] landscape_12345.jpg 2.4MB/4.1MB 58%
[2/5] anime_67890.jpg 1.1MB/3.2MB 34%
...
```

**Files Likely Touched:**
- internal/download/progress.go
- internal/download/manager.go (integrate callbacks)

**Expected Output:**
- Real-time progress visible
- Clean output format
- Optional (can be disabled)

**Verification:**
```bash
./wallpaper-cli fetch --limit 5
# See progress output
```

---

### T05: Add Resume Support
**Estimate:** 20m

**Description:**
Add HTTP Range request support for resuming interrupted downloads.

**Steps:**
1. Check for existing partial file
2. Send HEAD request to verify server support
3. Add Range header if resuming: `Range: bytes=1024-`
4. Open file in append mode for resume
5. Fallback to full download if server doesn't support ranges

**Files Likely Touched:**
- internal/download/downloader.go (resume logic)

**Expected Output:**
- Interrupted downloads can resume
- Servers without Range support still work

**Verification:**
```bash
# Start download, interrupt with Ctrl+C, restart
./wallpaper-cli fetch --limit 1
# ^C during download
./wallpaper-cli fetch --limit 1  # Should resume
```

---

### T06: Integrate with Fetch Command
**Estimate:** 15m

**Description:**
Wire up the download manager to the fetch command.

**Steps:**
1. Update cmd/fetch.go run function
2. After getting wallpaper list from source:
   - Create download manager with config
   - Submit each wallpaper as job
   - Wait for completion
   - Handle results/errors
3. Pass organization settings to download manager

**Files Likely Touched:**
- cmd/fetch.go
- internal/download/options.go (download options)

**Expected Output:**
- `wallpaper-cli fetch` downloads actual files
- Progress shown
- Files saved to output directory

**Verification:**
```bash
./wallpaper-cli fetch --source wallhaven --limit 5 --output ~/test-wallpapers
ls ~/test-wallpapers
# Should see 5 downloaded images
```
