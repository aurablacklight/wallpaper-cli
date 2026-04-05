# Architecture Research

**Domain:** Go CLI wallpaper tool — multi-source adapter integration, JSON API layer, download pipeline
**Researched:** 2026-04-04
**Confidence:** HIGH (based on direct codebase inspection + API documentation)

---

## Existing Architecture (Baseline)

Understanding the current structure is the foundation for integration decisions.

### Current System Map

```
┌─────────────────────────────────────────────────────────────┐
│                        cmd/ (Cobra handlers)                 │
│  fetch.go   set.go   list.go   export.go   stats.go  ...    │
│     │          │        │         │            │             │
│     └──────────┴────────┴─────────┴────────────┘            │
│                         │                                    │
├─────────────────────────┼────────────────────────────────────┤
│               internal/ (domain packages)                    │
│                         │                                    │
│  sources/               │  download/      dedup/             │
│  ├── wallhaven/         │  manager.go     checker.go         │
│  │   client.go          │  progress.go                       │
│  │   query.go           │                                    │
│  │   pagination.go      │  data/          config/            │
│  │   types.go           │  db.go          config.go          │
│  └── reddit/            │                                    │
│      client.go          │  collections/   platform/          │
│      types.go           │  ...            ...                │
│                         │                                    │
│  model/wallpaper.go     │  validate/      utils/             │
├─────────────────────────┴────────────────────────────────────┤
│                      SQLite (modernc)                         │
│  images  favorites  ratings  playlists  playlist_items  config│
└─────────────────────────────────────────────────────────────┘
```

### Observed Patterns in Existing Code

**No shared source interface exists.** Each source (wallhaven, reddit) has its own `Client` struct, its own `SearchOptions`, and its own result types. The cmd/fetch.go wires them via a `switch source { case "wallhaven": ... case "reddit": ... }` block. Adding new sources means adding new `case` branches — manageable for 2 sources, untenable at 5.

**No common `Wallpaper` adapter type is used at the source boundary.** `model.Wallpaper` exists but is not returned by any source client. Instead, cmd/fetch.go receives wallhaven-specific or reddit-specific types and converts inline. This is the primary coupling to fix.

**The download manager is decoupled.** `download.Manager` accepts `[]DownloadJob` (URL + filename pairs). It knows nothing about sources. This is correct and should not change.

**The dedup checker is a dependency of the download manager.** It receives a `*data.DB` handle. No change needed.

**Config uses `map[string]SourceConfig` keyed by source name.** New sources plug in without schema changes.

---

## New Architecture: Target State

### System Map with New Components

```
┌─────────────────────────────────────────────────────────────┐
│                     cmd/ (Cobra handlers)                    │
│  fetch.go   list.go   export.go   config.go  stats.go  ...  │
│      │                                                       │
│      │  writes JSON lines to stdout (event stream)           │
│      │  --json flag or GUI mode enables structured output    │
├──────┴──────────────────────────────────────────────────────┤
│                  internal/output/ (NEW)                      │
│  emitter.go  — JSONLinesEmitter / TextEmitter interface      │
│  events.go   — event type definitions (started, progress,   │
│                 found, downloaded, skipped, error, done)     │
├─────────────────────────────────────────────────────────────┤
│                  internal/sources/ (EXPANDED)                │
│                                                              │
│  interface.go     — Source interface (NEW)                   │
│  registry.go      — source registry map (NEW)                │
│                                                              │
│  wallhaven/       reddit/       danbooru/  zerochan/         │
│  (existing)       (existing)    (NEW)      (NEW)             │
│                                            konachan/         │
│                                            (NEW)             │
│                                                              │
│  All adapters return []model.Wallpaper + []model.Tag         │
├─────────────────────────────────────────────────────────────┤
│                  internal/model/ (EXPANDED)                  │
│  wallpaper.go    (existing — minor additions)                │
│  tag.go          (NEW — shared Tag type with source/category)│
├─────────────────────────────────────────────────────────────┤
│                  internal/download/ (MODIFIED)               │
│  manager.go      — add resumable + retry + event emission    │
│  retry.go        (NEW) — exponential backoff helper          │
├─────────────────────────────────────────────────────────────┤
│                  internal/data/ (MODIFIED)                   │
│  db.go           — add source_tags table + queries           │
├─────────────────────────────────────────────────────────────┤
│              SQLite (existing + source_tags table)           │
└─────────────────────────────────────────────────────────────┘
```

---

## Component Responsibilities

| Component | Responsibility | New or Modified |
|-----------|----------------|-----------------|
| `internal/sources/interface.go` | Defines the `Source` interface all adapters satisfy | NEW |
| `internal/sources/registry.go` | Map of name → Source factory; `cmd/fetch.go` looks up by name | NEW |
| `internal/sources/danbooru/` | Danbooru API client + adapter implementing Source | NEW |
| `internal/sources/zerochan/` | Zerochan API client + adapter implementing Source | NEW |
| `internal/sources/konachan/` | Konachan/Moebooru API client + adapter implementing Source | NEW |
| `internal/sources/wallhaven/` | Existing — add `Tags() []model.Tag` harvesting method | MODIFIED |
| `internal/sources/reddit/` | Existing — minimal change; Reddit has no structured tags | UNCHANGED |
| `internal/model/tag.go` | `model.Tag` — shared tag type with Source, Name, Category, Count | NEW |
| `internal/output/emitter.go` | `Emitter` interface with `TextEmitter` and `JSONLinesEmitter` impls | NEW |
| `internal/output/events.go` | Typed event structs: `EventFound`, `EventProgress`, `EventDone` etc. | NEW |
| `internal/download/manager.go` | Add resumable (Range header), retry with backoff, emit events | MODIFIED |
| `internal/download/retry.go` | Exponential backoff + jitter, context-aware, no new deps | NEW |
| `internal/data/db.go` | Add `source_tags` table, `SaveTags()`, `GetTags()` methods | MODIFIED |
| `cmd/fetch.go` | Replace switch/case with registry lookup; thread emitter through | MODIFIED |

---

## Pattern 1: Shared Source Interface

**What:** All source adapters satisfy a single Go interface. `cmd/fetch.go` holds a `Source` and calls methods on it without knowing which source it is.

**When to use:** Any time a new source is added. The interface is the contract.

**Trade-offs:** The interface must be general enough to accommodate all sources. Danbooru tags are first-class; Reddit has no tags. Use zero values / empty slices for missing capabilities rather than optional interface methods — keeps the dispatch simple.

```go
// internal/sources/interface.go

package sources

import (
    "context"
    "github.com/user/wallpaper-cli/internal/model"
)

// SearchOptions is the source-agnostic search request.
// Sources ignore fields they don't support (e.g., Reddit ignores Resolution).
type SearchOptions struct {
    Tags       []string
    Resolution string
    Ratio      string
    Limit      int
    Page       int
    Sort       string // "newest", "popular", "random"
    TimeRange  string // "day", "week", "month", "year", "all"
    Rating     string // "safe", "questionable", "explicit"
    APIKey     string
}

// Result is what a source returns from a single Search call.
type Result struct {
    Wallpapers []model.Wallpaper
    Tags       []model.Tag      // Empty if source doesn't provide tags
    NextPage   int              // 0 means no more pages
    Total      int              // 0 if unknown
}

// Source is the interface every adapter must implement.
type Source interface {
    // Name returns the canonical source identifier ("danbooru", "wallhaven", etc.)
    Name() string
    // Search fetches one page of results.
    Search(ctx context.Context, opts SearchOptions) (Result, error)
    // Supports declares optional capabilities for UI display.
    Supports() Capabilities
}

type Capabilities struct {
    Tags       bool // source provides per-image tag lists
    Rating     bool // source supports rating filter (safe/nsfw)
    Resolution bool // source supports resolution filtering
    Pagination bool // source supports page-based pagination
}
```

**Registry (internal/sources/registry.go):**

```go
// SourceFactory is a constructor that receives config and returns a Source.
type SourceFactory func(cfg config.SourceConfig) (Source, error)

var registry = map[string]SourceFactory{
    "wallhaven": wallhaven.NewSource,
    "reddit":    reddit.NewSource,
    "danbooru":  danbooru.NewSource,
    "zerochan":  zerochan.NewSource,
    "konachan":  konachan.NewSource,
}

func Get(name string, cfg config.SourceConfig) (Source, error) {
    factory, ok := registry[name]
    if !ok {
        return nil, fmt.Errorf("unknown source: %q", name)
    }
    return factory(cfg)
}
```

`cmd/fetch.go` becomes:
```go
src, err := sources.Get(source, cfg.Sources[source])
// then: src.Search(ctx, opts)
```

The `validSources` slice in cmd/fetch.go is replaced by `sources.RegisteredNames()`.

---

## Pattern 2: Booru Adapter Structure (Danbooru / Konachan / Zerochan)

**What:** Each Booru source is a self-contained package: `client.go` (HTTP), `types.go` (API response structs), `adapter.go` (implements `Source` interface, converts to `model.Wallpaper`).

**API endpoint differences:**

| Source | Base URL | Posts endpoint | Tag field | Auth |
|--------|----------|----------------|-----------|------|
| Danbooru | `https://danbooru.donmai.us` | `GET /posts.json` | `tag_string` (space-separated) | Optional `api_key` + `login` |
| Konachan | `https://konachan.com` | `GET /post.json` | `tags` (space-separated) | Optional login/SHA1 hash |
| Zerochan | `https://www.zerochan.net` | `GET /{tag}?json` | Tag in URL path | Requires User-Agent with username |

**Danbooru key response fields:** `id`, `file_url`, `large_file_url`, `tag_string`, `tag_string_character`, `tag_string_artist`, `image_width`, `image_height`, `file_size`, `file_ext`, `rating`, `score`, `created_at`. Rate: 2 tags max for anonymous; up to 200 results per page.

**Konachan/Moebooru key response fields:** `id`, `file_url`, `preview_url`, `tags`, `width`, `height`, `file_size`, `source`, `rating`, `score`, `created_at`. Rate: max 100 posts per request; HTTP 421 = throttled.

**Zerochan key response fields:** `id`, `src` (full image URL), `small` (thumbnail), `tag` (primary tag), `tags` (array), `width`, `height`, `size`, `source`, `primary`. Rate: 60 req/min. Pagination via `p` param; max 250 per page.

**Example Danbooru adapter structure:**

```go
// internal/sources/danbooru/adapter.go

type Source struct{ client *Client }

func NewSource(cfg config.SourceConfig) (sources.Source, error) {
    return &Source{client: NewClient(cfg.APIKey, cfg.Login)}, nil
}

func (s *Source) Name() string { return "danbooru" }

func (s *Source) Search(ctx context.Context, opts sources.SearchOptions) (sources.Result, error) {
    posts, err := s.client.Posts(ctx, PostsRequest{
        Tags:  strings.Join(opts.Tags, " "),
        Limit: opts.Limit,
        Page:  opts.Page,
    })
    if err != nil {
        return sources.Result{}, err
    }
    return sources.Result{
        Wallpapers: toWallpapers(posts),
        Tags:       toTags(posts),
        NextPage:   opts.Page + 1,
    }, nil
}

func (s *Source) Supports() sources.Capabilities {
    return sources.Capabilities{Tags: true, Rating: true, Resolution: false, Pagination: true}
}
```

**Danbooru tag constraint:** Anonymous requests allow max 2 tags. With `api_key` (free account), up to 6 tags. Config `api_key` field unlocks this. The adapter should warn when >2 tags are requested without a key.

---

## Pattern 3: JSON Lines Event Stream

**What:** All commands emit structured JSON lines to stdout. Progress events go to stdout as JSON objects (one per line). The `--json` flag is not needed — this becomes the default output contract. Human-readable text goes to stderr (progress bars, status messages).

**Why JSON lines over alternatives:** No buffering issues (each line is a complete message), pipe-friendly (`wallpaper-cli fetch | jq .`), trivially parseable by GUI app (`bufio.Scanner` on stdout), no WebSocket complexity.

**Event contract (internal/output/events.go):**

```go
type EventType string

const (
    EventTypeStarted    EventType = "started"
    EventTypeFound      EventType = "found"
    EventTypeProgress   EventType = "progress"
    EventTypeDownloaded EventType = "downloaded"
    EventTypeSkipped    EventType = "skipped"
    EventTypeError      EventType = "error"
    EventTypeDone       EventType = "done"
)

type Event struct {
    Type      EventType   `json:"type"`
    Timestamp time.Time   `json:"ts"`
    Payload   interface{} `json:"data"`
}

// Example payloads:
type FoundPayload struct {
    Source string `json:"source"`
    Count  int    `json:"count"`
    Total  int    `json:"total,omitempty"`
}

type DownloadedPayload struct {
    URL        string `json:"url"`
    LocalPath  string `json:"path"`
    SourceID   string `json:"source_id"`
    Source     string `json:"source"`
    Resolution string `json:"resolution"`
    Tags       []string `json:"tags,omitempty"`
    FileSize   int64  `json:"file_size"`
}

type DonePayload struct {
    Downloaded int `json:"downloaded"`
    Skipped    int `json:"skipped"`
    Errors     int `json:"errors"`
    Duration   string `json:"duration"`
}
```

**Emitter interface (internal/output/emitter.go):**

```go
type Emitter interface {
    Emit(e Event)
    Close()
}

// JSONLinesEmitter writes one JSON object per line to w (os.Stdout).
// Thread-safe.
type JSONLinesEmitter struct {
    w   io.Writer
    mu  sync.Mutex
    enc *json.Encoder
}

// TextEmitter writes human-readable lines to w (os.Stderr).
// Used as fallback when stdout is a terminal and no GUI is consuming.
type TextEmitter struct { w io.Writer }
```

**Threading the emitter:** The `Emitter` is created once in `cmd/fetch.go` (or root) and passed down to the download manager and any source fetch loop. Commands write to stdout via emitter; progress bars remain on stderr.

**Query/CRUD commands** (list, export, stats, config) emit a single JSON object or JSON array to stdout and exit. They don't use the event stream.

---

## Pattern 4: Resumable Downloads

**What:** Before creating a temp file, check if a `.tmp` partial file already exists for that job. If it does, get its current byte offset and set `Range: bytes=N-` on the HTTP request.

**Where it fits:** Entirely within `internal/download/manager.go` (`downloadOne` method). No interface changes needed elsewhere.

**Implementation steps in `downloadOne`:**

1. Check if `job.Filename + ".tmp"` exists and get its size.
2. If size > 0, set `Range: bytes={size}-` header on the request.
3. Open file with `O_APPEND|O_CREATE|O_WRONLY` instead of creating fresh.
4. Accept both `200 OK` (server ignored Range) and `206 Partial Content`.
5. If server returns 200 (no Range support), truncate the file and restart.
6. Existing atomic rename (`os.Rename(tempPath, finalPath)`) stays unchanged.

**Server-side support check:** Read `Accept-Ranges: bytes` from the first response. Cache this per-host if needed, but for simplicity just handle both status codes gracefully.

---

## Pattern 5: Retry with Exponential Backoff

**What:** HTTP requests in source clients and the download manager retry on transient failures (5xx, network timeout, 429 rate-limit) using exponential backoff with jitter.

**Where it fits:** `internal/download/retry.go` — a small self-contained helper with no new dependencies.

```go
// internal/download/retry.go

// Retryable wraps fn with exponential backoff.
// Retries on: network errors, 429, 500, 502, 503, 504.
// Does not retry on: 401, 403, 404, context cancellation.
func Retryable(ctx context.Context, maxAttempts int, fn func() error) error {
    base := 500 * time.Millisecond
    for attempt := 0; attempt < maxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        if !isRetryable(err) || attempt == maxAttempts-1 {
            return err
        }
        // jitter: base * 2^attempt * (0.5 + rand.Float64()*0.5)
        sleep := time.Duration(float64(base<<attempt) * (0.5 + rand.Float64()*0.5))
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(sleep):
        }
    }
    return nil
}
```

Source clients (Danbooru, Konachan, Zerochan) each rate-limit themselves differently. Wrap each client's HTTP call with `Retryable`. The max-attempts default is 3; configurable via `Options`.

---

## Pattern 6: Tag Harvesting

**What:** When a source returns tags alongside wallpapers, persist them to a `source_tags` table. Tags are available for `--tags` flag autocomplete and future AI tagging features.

**Where it fits:** `internal/data/db.go` gets a new table and two methods. The source fetch loop in `cmd/fetch.go` calls `db.SaveTags()` after each page of results.

**New SQLite table:**

```sql
CREATE TABLE IF NOT EXISTS source_tags (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    source      TEXT NOT NULL,           -- "danbooru", "wallhaven", etc.
    name        TEXT NOT NULL,
    category    TEXT DEFAULT '',          -- "character", "artist", "copyright", "general"
    post_count  INTEGER DEFAULT 0,
    harvested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source, name)
);
CREATE INDEX IF NOT EXISTS idx_source_tags_source ON source_tags(source);
CREATE INDEX IF NOT EXISTS idx_source_tags_name ON source_tags(name);
```

**New model type (internal/model/tag.go):**

```go
type Tag struct {
    Source    string
    Name      string
    Category  string // "character", "artist", "copyright", "general", ""
    PostCount int
}
```

**New DB methods:**

```go
func (db *DB) SaveTags(tags []model.Tag) error       // upsert, update post_count
func (db *DB) GetTags(source string) ([]model.Tag, error)
func (db *DB) SearchTags(query string) ([]model.Tag, error) // prefix search
```

**Tag flow per search page:**

```
Source.Search() → Result.Tags → cmd/fetch.go → db.SaveTags(result.Tags)
```

Reddit returns no tags. Wallhaven tags are on each wallpaper's Tags array (existing `wallhaven.Tag` structs). Danbooru and Konachan return tags in `tag_string` fields (space-separated). Zerochan returns a `tags` array per entry.

---

## Data Flow

### Fetch Command — New Flow

```
wallpaper-cli fetch --source danbooru --tags "landscape" --limit 20

cmd/fetch.go
  │
  ├── sources.Get("danbooru", cfg) → danbooru.Source
  │
  ├── loop pages until limit reached:
  │     Source.Search(ctx, opts) → Result{Wallpapers, Tags, NextPage}
  │     emitter.Emit(EventFound{...})
  │     db.SaveTags(result.Tags)            ← tag harvesting
  │     build []download.DownloadJob
  │
  ├── download.Manager.DownloadBatch(ctx, jobs)
  │     per job: check .tmp exists → set Range header if resuming
  │     retry.Retryable(fn) wraps HTTP call
  │     on complete: emitter.Emit(EventDownloaded{...})
  │     on skip:     emitter.Emit(EventSkipped{...})
  │     on error:    emitter.Emit(EventError{...})
  │
  ├── db.SaveImage(record) per downloaded file
  │
  └── emitter.Emit(EventDone{downloaded, skipped, errors, duration})
```

### JSON Lines Stdout (consumed by GUI)

```
{"type":"started","ts":"...","data":{"source":"danbooru","limit":20}}
{"type":"found","ts":"...","data":{"source":"danbooru","count":20,"total":4823}}
{"type":"downloaded","ts":"...","data":{"url":"...","path":"/...","source_id":"123","resolution":"3840x2160","tags":["landscape"],"file_size":4200000}}
{"type":"skipped","ts":"...","data":{"url":"...","reason":"duplicate"}}
{"type":"done","ts":"...","data":{"downloaded":18,"skipped":2,"errors":0,"duration":"12.4s"}}
```

### Parallel Multi-Source Fetch

```
--source all
  │
  ├── goroutine: sources.Get("wallhaven") → fetch → emit
  ├── goroutine: sources.Get("danbooru") → fetch → emit
  ├── goroutine: sources.Get("zerochan") → fetch → emit
  │
  └── download.Manager.DownloadBatch (shared, dedup across sources)
```

The emitter is thread-safe (`sync.Mutex` on the JSON encoder). Each goroutine emits `EventFound` independently; the download batch is collected and submitted after all sources finish their search phase. This avoids download interleaving issues with dedup.

---

## Recommended Project Structure (Delta from Current)

```
internal/
├── sources/
│   ├── interface.go       NEW — Source interface, SearchOptions, Result, Capabilities
│   ├── registry.go        NEW — registered source factories
│   ├── wallhaven/         existing (add Tags() to types, minor adapter.go wrapper)
│   ├── reddit/            existing (unchanged)
│   ├── danbooru/
│   │   ├── client.go      NEW — HTTP client, rate limiting (1 req/s anon, adjustable)
│   │   ├── types.go       NEW — Post struct, field mapping from API JSON
│   │   └── adapter.go     NEW — implements sources.Source
│   ├── zerochan/
│   │   ├── client.go      NEW — HTTP client, 60 req/min rate limit
│   │   ├── types.go       NEW — Entry struct, tag array
│   │   └── adapter.go     NEW — implements sources.Source
│   └── konachan/
│       ├── client.go      NEW — Moebooru HTTP client (same structure as danbooru)
│       ├── types.go       NEW — Post struct (Moebooru format)
│       └── adapter.go     NEW — implements sources.Source
├── model/
│   ├── wallpaper.go       existing (no required changes)
│   └── tag.go             NEW — model.Tag{Source, Name, Category, PostCount}
├── output/
│   ├── emitter.go         NEW — Emitter interface, JSONLinesEmitter, TextEmitter
│   └── events.go          NEW — event type constants and payload structs
├── download/
│   ├── manager.go         MODIFIED — resumable downloads, event emission hook
│   ├── retry.go           NEW — Retryable() with exponential backoff + jitter
│   ├── progress.go        existing
│   └── progressbar.go     existing
└── data/
    └── db.go              MODIFIED — source_tags table, SaveTags(), GetTags(), SearchTags()
```

---

## Integration Points: New vs. Modified

### What is NEW (no existing code touched)

| Package | File(s) | Purpose |
|---------|---------|---------|
| `sources` | `interface.go`, `registry.go` | Source abstraction + discovery |
| `sources/danbooru` | `client.go`, `types.go`, `adapter.go` | Danbooru adapter |
| `sources/zerochan` | `client.go`, `types.go`, `adapter.go` | Zerochan adapter |
| `sources/konachan` | `client.go`, `types.go`, `adapter.go` | Konachan adapter (Moebooru-style) |
| `model` | `tag.go` | Shared tag type |
| `output` | `emitter.go`, `events.go` | JSON lines event system |
| `download` | `retry.go` | Retry with backoff |

### What is MODIFIED (existing code changes)

| Package | File | Change |
|---------|------|--------|
| `cmd/fetch.go` | fetch.go | Replace switch/case with registry; add emitter; thread through |
| `sources/wallhaven` | `adapter.go` (new thin wrapper file) | Implement `Source` interface |
| `sources/reddit` | `adapter.go` (new thin wrapper file) | Implement `Source` interface |
| `download/manager.go` | manager.go | Add resumable logic in `downloadOne`; accept emitter hook |
| `data/db.go` | db.go | Add `source_tags` table in `createTables()`; add 3 tag methods |
| `config/config.go` | config.go | Add per-source fields: `Login`, `Username` for Danbooru/Zerochan auth |

### What is UNCHANGED

- `internal/dedup/checker.go` — no changes
- `internal/collections/` — no changes
- `internal/platform/` — no changes
- `internal/validate/` — add new source names to valid list, otherwise unchanged
- All other cmd/ handlers (set, list, export, stats, etc.)

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Source-Specific Logic in cmd/fetch.go

**What people do:** Keep growing the `switch source { case "danbooru": ... }` block in cmd/fetch.go.

**Why it's wrong:** Every new source requires touching the command file. cmd/fetch.go becomes the coupling point for all source API differences.

**Do this instead:** The interface + registry pattern. `cmd/fetch.go` calls `sources.Get(name, cfg)` and then calls `source.Search()`. Source-specific knowledge stays inside `internal/sources/<name>/`.

### Anti-Pattern 2: Writing Events to Both Stdout and Stderr

**What people do:** Mix JSON events and human-readable output on the same stream.

**Why it's wrong:** The GUI app reads stdout. If progress text leaks to stdout, JSON parsing breaks.

**Do this instead:** All human-readable output (progress bars, status messages) goes to stderr. All structured events go to stdout via `JSONLinesEmitter`. These are separate streams. The progress bar library (`schollz/progressbar`) already supports `OptionSetWriter(os.Stderr)` — keep it there.

### Anti-Pattern 3: Blocking Parallel Source Fetches on Downloads

**What people do:** For `--source all`, fetch source A → download A → fetch source B → download B sequentially.

**Why it's wrong:** Wastes time; sources can be queried in parallel.

**Do this instead:** Fan out source searches in goroutines, collect all `[]model.Wallpaper` results (with a wait group), then submit a single `DownloadBatch` to the manager. This keeps dedup correct (all URLs are known before any download starts) and maximizes throughput.

### Anti-Pattern 4: Hardcoding Rate Limits as Constants

**What people do:** `const DanbooruRateLimit = time.Second`.

**Why it's wrong:** Rate limits differ between anonymous and authenticated access. Danbooru: more lenient with API key. Zerochan: 60 req/min flat. These should be configurable, or at minimum set from the client constructor based on whether auth credentials are present.

**Do this instead:** Each client computes its rate limit in `NewClient()` based on whether credentials are provided. Wallhaven's existing `time.Ticker` approach is the correct pattern — copy it per source.

### Anti-Pattern 5: Storing Tags as JSON Blob in `images.tags`

**What people do:** Continue using `images.tags TEXT` (the existing JSON array column) as the only tag store.

**Why it's wrong:** Not queryable for autocomplete or filtering. Cannot aggregate tag frequency across sources. Cannot track which tags exist without downloading images.

**Do this instead:** Keep `images.tags` for per-image tag display (backward compatible). Add the separate `source_tags` table for the harvested tag catalog. The two coexist without conflict.

---

## Build Order (Suggested)

This order minimizes merge conflicts and lets each step be independently testable.

**Step 1: model.Tag + data/db.go source_tags table**
- No cmd changes. Purely additive. Can be tested with DB unit tests.

**Step 2: internal/output/ (emitter + events)**
- No downstream deps. Testable in isolation with a bytes.Buffer.

**Step 3: Source interface + registry (without new adapters)**
- Wrap existing wallhaven and reddit as `Source` implementors.
- Update cmd/fetch.go to use registry.
- All existing tests should pass. No behavior change yet.

**Step 4: Danbooru adapter**
- Builds on Step 3 interface. Add to registry. Integration test against live API.
- Tag harvesting can be validated here (Danbooru tags are rich and well-structured).

**Step 5: Konachan adapter**
- Moebooru API is nearly identical to Danbooru API response format. Konachan's `client.go` can share type patterns from Danbooru with minor field differences.

**Step 6: Zerochan adapter**
- Slightly different (tag-in-URL-path pattern, User-Agent auth). Build last because it deviates most from the Booru pattern.

**Step 7: download/retry.go + resumable in manager.go**
- Independent of source adapters. Add retry wrapper + Range header logic.
- Backward compatible: if no .tmp file exists, behavior is identical to current.

**Step 8: JSON lines output wired through cmd/fetch.go**
- Thread the emitter into the fetch loop and download manager.
- TextEmitter as default (stderr), JSONLinesEmitter when stdout is not a terminal or `--json` flag is set.

---

## Integration Points Summary

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `cmd/fetch.go` ↔ `sources.Source` | Direct method call via interface | Registry resolves name → impl |
| `sources.Source` → `output.Emitter` | Not direct — cmd threads emitter | Source doesn't know about emitter |
| `cmd/fetch.go` ↔ `download.Manager` | Existing `[]DownloadJob` stays | Manager gets optional emitter hook |
| `download.Manager` → `output.Emitter` | Direct call: `emitter.Emit(event)` | Manager emits per-file events |
| `sources.Result.Tags` → `data.DB` | cmd calls `db.SaveTags(result.Tags)` | Tags persisted after each search page |
| `download.retry.Retryable` ↔ clients | Closure wrapping HTTP calls | Each source client calls Retryable internally |
| `model.Wallpaper` | Common type returned by all adapters | Existing struct; no breaking changes |
| `model.Tag` | New type for tag harvesting | Flows from Source → cmd → DB |

---

## Sources

- Direct inspection of `/internal/sources/wallhaven/`, `/internal/sources/reddit/`, `/internal/download/manager.go`, `/internal/data/db.go`, `/internal/config/config.go`, `/cmd/fetch.go`
- [Konachan/Moebooru API reference](https://konachan.com/help/api) — confirmed endpoints, rate limits (100 posts max, HTTP 421 throttle)
- [Zerochan API reference](https://www.zerochan.net/api) — confirmed endpoints, query params (`p`, `l`, `s`, `d`), 60 req/min limit
- [Danbooru API](https://danbooru.pw/wiki_pages/help:api) — 2-tag anon limit, `/posts.json` up to 200 per page, `api_key` + `login` auth
- [Go resumable download pattern](https://transloadit.com/devtips/build-a-resumable-file-downloader-in-go-with-concurrent-chunks/) — Range header, O_APPEND, 206 handling
- [cenkalti/backoff](https://pkg.go.dev/github.com/cenkalti/backoff/v4) — exponential backoff; stdlib-only implementation preferred to avoid new deps

---

*Architecture research for: wallpaper-cli-tool v1.3 — sources, API, downloads*
*Researched: 2026-04-04*
