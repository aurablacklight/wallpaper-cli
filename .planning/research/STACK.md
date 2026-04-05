# Stack Research

**Domain:** Go CLI — source adapters, JSON event streaming, resumable downloads
**Researched:** 2026-04-04
**Confidence:** HIGH (all library decisions are stdlib or verified via pkg.go.dev; API behaviors verified via official docs)

---

## Recommended Stack

### New Dependencies

| Library | Version | Purpose | Why Recommended |
|---------|---------|---------|-----------------|
| `github.com/cenkalti/backoff/v4` | v4.3.0 | Exponential backoff for retry logic | Widely used Go port of Google's algorithm; clean API (`RetryWithData`, `NewExponentialBackOff`); avoids writing custom jitter logic; fits cleanly inside `downloadOne` without restructuring the manager |
| `github.com/hashicorp/go-retryablehttp` | v0.7.8 | Retryable HTTP client with 429/503 `Retry-After` awareness | Handles rewindable request bodies across retries; respects `Retry-After` response header automatically; wraps `net/http.Client` so existing source clients need minimal change |

### No New Dependencies Needed (Stdlib Covers This)

| Capability | Stdlib Package | Rationale |
|-----------|----------------|-----------|
| JSON lines / NDJSON event stream | `encoding/json` — `json.NewEncoder(os.Stdout)` | `Encoder.Encode()` appends `\n` automatically; each call produces one complete JSON line; no third-party streaming lib needed |
| Resumable downloads (HTTP Range) | `net/http`, `os`, `io` | Set `Range: bytes=<offset>-` header, open file with `os.O_APPEND`, handle `206 Partial Content`; pure stdlib, zero new deps |
| Parallel multi-source fetching | `sync`, `context` | Existing `sync.WaitGroup` worker pool pattern in `download/manager.go` already handles this; extend, don't replace |
| Danbooru adapter HTTP client | `net/http` | Danbooru REST JSON API — standard GET with `login` + `api_key` query params; same pattern as Wallhaven adapter |
| Konachan adapter HTTP client | `net/http` | Moebooru-compatible REST API (`/post.json`, `/tag.json`) — same GET pattern; shares struct shapes with Danbooru |
| Zerochan adapter HTTP client | `net/http` | Read-only GET API (`/TagName?json`); requires custom `User-Agent` header only |

---

## API Integration Reference

### Danbooru (`danbooru.donmai.us`)

- **Base URL:** `https://danbooru.donmai.us`
- **Post search:** `GET /posts.json?tags=<tags>&limit=<n>&page=<n>&login=<u>&api_key=<k>`
- **Tag search:** `GET /tags.json?search[name_matches]=<pattern>&limit=<n>`
- **Auth:** URL params `login` + `api_key` OR HTTP Basic Auth
- **Rate limit:** 10 read requests/second (anonymous: ~500/hr); use `time.Ticker` at ~6 req/s
- **Tag search limit:** Anonymous/Basic = 2 tags max; Gold = 6 tags; use authenticated account for harvesting
- **Key response fields:** `id`, `file_url`, `large_file_url`, `tag_string`, `tag_string_artist`, `tag_string_character`, `tag_string_copyright`, `image_width`, `image_height`, `rating` (`g`/`s`/`q`/`e`), `score`, `created_at`
- **Pagination:** `page` param (integer) OR `b<id>` (before-ID) / `a<id>` (after-ID) for stable cursor-based iteration
- **Reuse:** Wallhaven adapter pattern applies directly — same `http.NewRequestWithContext`, `waitForRateLimit`, `json.NewDecoder`

### Konachan (`konachan.com`)

- **Base URL:** `https://konachan.com` (safe: `konachan.net`)
- **Post search:** `GET /post.json?tags=<tags>&limit=<n>&page=<n>`
- **Tag search:** `GET /tag.json?name=<pattern>&limit=<n>&order=count`
- **Auth:** `login` + `password_hash` params (SHA1 of salted password); optional for read-only browsing
- **Rate limit:** HTTP 421 "User Throttled" on violation; no documented rate number — use 1 req/2s (same as Wallhaven)
- **Hard limit:** 100 posts per request
- **Key response fields:** `id`, `file_url`, `sample_url`, `preview_url`, `tags`, `rating` (`s`/`q`/`e`), `score`, `width`, `height`, `file_size`, `created_at`
- **Moebooru compatibility:** Struct shapes nearly identical to Danbooru v1 — define shared `BooruPost` struct and embed/alias per source
- **Reuse:** Danbooru adapter code shares 80%+ of logic; extract a `booru` base package

### Zerochan (`zerochan.net`)

- **Base URL:** `https://www.zerochan.net`
- **Tag search:** `GET /TagName?json&l=50&p=<n>&s=id`
- **Strict mode:** `GET /TagName?json&strict` — only entries where tag is primary (better for wallpaper quality)
- **Single entry:** `GET /<id>?json` — returns full metadata including direct image URL
- **Auth:** None (read-only API); **mandatory:** `User-Agent: wallpaper-cli - <username>` header
- **Rate limit:** 60 req/min; chronic violations → IP ban
- **Key response fields (list):** `id`, `thumbnail`, `small`, `medium`, `large`, `full` (URL variants), `tag` (primary tag string), `tags` (array)
- **Key response fields (single entry):** adds `width`, `height`, `size`, direct download URL
- **Caveat:** List endpoint returns thumbnail URLs; full-resolution URL requires a second GET on each entry ID. Account for 2× API calls per image in rate limit budget.
- **Tag harvesting:** Zerochan uses hierarchical tags; `strict` mode avoids tag contamination for wallpaper searches

---

## Architecture: What Changes and What Doesn't

### Changes to Existing Code

**`internal/download/manager.go` — Resume + Retry**

Add resumable support inside `downloadOne`:
1. `os.Stat(tempPath)` → get existing byte offset
2. If offset > 0, set `Range: bytes=<offset>-` on request
3. Open file with `os.O_APPEND|os.O_CREATE|os.O_WRONLY`
4. Handle both `200 OK` (no resume support) and `206 Partial Content`
5. Wrap retry loop using `cenkalti/backoff/v4` `RetryWithData` — retries on transient errors (connection refused, 5xx, timeout); does NOT retry 4xx

Replace the flat `http.Client` in `Manager` with `hashicorp/go-retryablehttp` client for source adapter HTTP calls (handles `Retry-After` on 429 from Danbooru automatically).

**`internal/model/wallpaper.go` — Tag Fields**

Add `TagCategories map[string][]string` to carry source-structured tags (artist, character, copyright, general) for DB storage without breaking existing `Tags []string`.

**`internal/data/db.go` — Tag Table**

New `source_tags` table: `(source TEXT, tag TEXT, category TEXT, count INTEGER, PRIMARY KEY (source, tag))`. No new deps — SQLite via existing `modernc.org/sqlite`.

**`cmd/fetch.go` — JSON Output**

Switch from `fmt.Println` progress to `json.NewEncoder(os.Stdout).Encode(event)` when `--json` flag is set. Event struct:
```go
type Event struct {
    Type    string `json:"type"`    // "start" | "progress" | "complete" | "error" | "skip" | "summary"
    Payload any    `json:"payload"`
}
```
No external library — `encoding/json` `Encoder` writes one JSON line + `\n` per call, which is the NDJSON spec.

### New Packages

```
internal/sources/danbooru/   — client.go, types.go
internal/sources/konachan/   — client.go, types.go (shares booru types)
internal/sources/zerochan/   — client.go, types.go
internal/sources/booru/      — shared types (BooruPost) for Danbooru + Konachan
internal/output/             — JSON lines emitter, wraps json.Encoder
```

---

## Installation

```bash
# New direct dependencies only
go get github.com/cenkalti/backoff/v4@v4.3.0
go get github.com/hashicorp/go-retryablehttp@v0.7.8
```

Everything else (NDJSON streaming, resumable downloads, parallel fetching, tag DB storage) is stdlib or extends existing dependencies.

---

## Alternatives Considered

| Recommended | Alternative | Why Not |
|-------------|-------------|---------|
| `cenkalti/backoff/v4` | Write custom backoff | Custom jitter + max-elapsed-time logic is subtle to get right; backoff v4 is 14 lines of config vs. ~80 lines custom |
| `hashicorp/go-retryablehttp` | `projectdiscovery/retryablehttp-go` | HashiCorp's is more widely used (3,000+ importers), better maintained, `Retry-After` header parsing is built-in |
| `hashicorp/go-retryablehttp` | Wrap `cenkalti/backoff` around `net/http` manually | `go-retryablehttp` also handles rewindable request bodies on retry — important for POST auth calls to Konachan |
| `encoding/json` Encoder for NDJSON | `goccy/go-json` or `francoispqt/gojay` | Throughput is not a bottleneck for CLI event output; stdlib avoids a dep for no measurable gain |
| Stdlib Range requests | Third-party resumable download lib | Range + 206 handling is ~20 lines; no lib needed, and existing `Manager` structure stays intact |
| `go.felesatra.moe/danbooru` (Go client) | — | Wraps a limited subset; we need tag harvesting and consistent error handling aligned with our adapter pattern — writing our own is 200 lines and gives full control |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| CGO-based SQLite libs | Breaks cross-compilation | `modernc.org/sqlite` (already in use) |
| `database/sql` + `lib/pq` or MySQL | No new DB engine needed | Existing SQLite schema, just add `source_tags` table |
| WebSocket or SSE for event stream | Overkill for CLI-to-GUI IPC; GUI reads stdout | JSON lines (`encoding/json` Encoder to stdout) |
| Full scraping framework (colly, goquery) | Zerochan has a real JSON API | `net/http` + `encoding/json` — no HTML parsing needed |
| `avast/retry-go` | Slightly less ergonomic for HTTP-specific retry; `go-retryablehttp` already handles HTTP status checking | `hashicorp/go-retryablehttp` |
| `go.felesatra.moe/danbooru` (pre-built client) | No tag category harvesting support; couples us to third-party struct definitions | Custom adapter in `internal/sources/danbooru/` |

---

## Version Compatibility

| Package | Go Version | Notes |
|---------|------------|-------|
| `cenkalti/backoff/v4` v4.3.0 | Go 1.18+ | Uses generics in `RetryWithData`; project already on Go 1.26 |
| `hashicorp/go-retryablehttp` v0.7.8 | Go 1.13+ | No compatibility issues with Go 1.26 |
| `encoding/json` (stdlib) | All Go versions | `json.Encoder` NDJSON behavior is stable and unchanged |

---

## Sources

- Danbooru official API: https://danbooru.donmai.us/wiki_pages/help:api — rate limits, auth, endpoint parameters (HIGH confidence)
- Konachan official API docs: https://konachan.com/help/api — endpoints, throttle code, response format (HIGH confidence)
- Zerochan official API: https://www.zerochan.net/api — endpoints, strict mode, rate limit, user-agent requirement (HIGH confidence)
- `hashicorp/go-retryablehttp` v0.7.8: https://pkg.go.dev/github.com/hashicorp/go-retryablehttp — verified version, Retry-After support (HIGH confidence)
- `cenkalti/backoff/v4` v4.3.0: https://pkg.go.dev/github.com/cenkalti/backoff/v4 — verified version, API surface (HIGH confidence)
- Resumable download pattern: https://transloadit.com/devtips/build-a-resumable-file-downloader-in-go-with-concurrent-chunks/ — confirmed stdlib-only implementation (MEDIUM confidence)
- NDJSON Go pattern: Go standard library `encoding/json` docs — `Encoder.Encode()` appends `\n` per call by design (HIGH confidence)

---

*Stack research for: wallpaper-cli-tool v1.3 — source adapters, JSON API contract, download improvements*
*Researched: 2026-04-04*
