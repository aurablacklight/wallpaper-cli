# Pitfalls Research

**Domain:** Go CLI tool — adding Booru adapters, JSON streaming API, resumable downloads, tag storage
**Researched:** 2026-04-04
**Confidence:** HIGH (code-verified + official API docs + community issue tracker evidence)

---

## Critical Pitfalls

### Pitfall 1: Danbooru's 2-Tag Search Limit Breaks Tag-Based UX

**What goes wrong:**
Anonymous and basic-tier Danbooru users are limited to 2 tags per search query. Sending 3+ tags returns a silent error or empty result rather than a clear user-facing message. A user typing `--tags "rem,rezero,1920x1080"` will get zero results and no explanation why.

**Why it happens:**
Developers test with 1-2 tags, the feature works, and the constraint isn't discovered until a user tries real multi-tag queries. Danbooru's API returns a 422 or empty set rather than a descriptive error for the over-limit case. The limit is tier-dependent (anonymous: 2, Gold: 6, Platinum+: 12).

**How to avoid:**
- Parse the `tags` input and count before sending the request
- If anonymous (no API key configured), enforce a 2-tag maximum and emit a clear error: "Danbooru allows 2 tags without an API key; provide credentials for up to 6"
- Document the constraint in `--help` text for the Danbooru source
- Store the user's Danbooru API key + login in Viper config under `sources.danbooru.api_key` / `sources.danbooru.login`

**Warning signs:**
- Danbooru searches silently return 0 results when the user passes 3+ tags
- `--dry-run` shows correct query construction but live fetch returns nothing

**Phase to address:**
Danbooru adapter phase — validate tag count before every API call, not just at CLI flag parse time.

---

### Pitfall 2: Zerochan Returns 404 for Empty Results (Not a Normal "No Results")

**What goes wrong:**
When a Zerochan tag search returns no results, the API responds with HTTP 404 Not Found — not an empty result set. If the adapter treats any 4xx as a hard error (the current pattern in `wallhaven/client.go` line 158: `if resp.StatusCode != http.StatusOK`), the user sees a download error instead of "no wallpapers found."

**Why it happens:**
Every other booru (and the existing Wallhaven/Reddit adapters) treat 404 as a real error. Zerochan is the exception. The existing error-handling pattern `fmt.Errorf("API error: %s (status %d)", resp.Status, resp.StatusCode)` will propagate this as a failure.

**How to avoid:**
- In the Zerochan client, handle 404 as a special case that returns an empty result set, not an error
- Add a code comment explaining this intentional deviation: `// Zerochan returns 404 when no results match — treat as empty, not as error`
- Write a unit test specifically for the 404 → empty-results path

**Warning signs:**
- `fetch --source zerochan --tags "nonexistent_character"` prints an error instead of "0 wallpapers found"
- Integration tests for Zerochan fail on empty-tag searches

**Phase to address:**
Zerochan adapter phase — the 404 handling must be the first thing written in the client, before any pagination logic.

---

### Pitfall 3: Progress Bar Written to Stderr Breaks JSON Output Mode

**What goes wrong:**
The existing progress bar is explicitly written to `os.Stderr` (confirmed in `cmd/fetch.go` line 456: `progressbar.OptionSetWriter(os.Stderr)`). The summary lines (`✓ Downloaded: X/Y`) are also written to `os.Stderr` (lines 517-524). This is correct for human mode. However, when JSON event streaming is active, even stderr output can corrupt a GUI app that reads both streams, and the summary text lines will need to be replaced by a terminal JSON event.

The danger is inconsistency: if JSON mode suppresses the progress bar but still emits the `fmt.Fprintf(os.Stderr, "✓ Downloaded...")` summary, the GUI receives a mixed signal.

**Why it happens:**
The JSON flag is added after the progress bar code is written. Developers add `if !jsonMode { bar.Finish() }` but forget the 6 other `fmt.Fprintf(os.Stderr, ...)` calls scattered through `downloadWallpapers` and `downloadRedditPosts`.

**How to avoid:**
- Introduce a single output abstraction early: `type OutputMode int; const (HumanMode OutputMode = iota; JSONMode)`
- Route all terminal output — progress bar, summary, status lines — through this abstraction
- In JSON mode, suppress the progress bar entirely and emit only JSON lines to stdout
- In human mode, emit nothing to stdout from the download functions (stderr only)
- Never emit bare `fmt.Println` / `fmt.Printf` to stdout inside download logic if JSON mode is possible

**Warning signs:**
- `wallpaper-cli fetch --json | jq` crashes because non-JSON text appears in the stream
- The GUI app receives both a `{"event":"complete"}` line and a stray "✓ Downloaded: 10/10" string

**Phase to address:**
JSON output phase — must refactor output routing before adding any JSON events, not after.

---

### Pitfall 4: Tags Stored as JSON Blob in `images.tags` Cannot Be Queried

**What goes wrong:**
The current schema stores tags as a JSON string in `images.tags TEXT` (confirmed in `internal/data/db.go` line 91). This works for display but makes tag-based queries impossible without full table scans and in-process JSON parsing. When tag harvesting is added for 3 new sources, users will want to filter their collection by tag — queries like "show me all wallpapers tagged 'rem'" will be either impossible or extremely slow.

**Why it happens:**
The JSON blob approach was fast to implement for the MVP. Adding tag harvesting without normalizing the schema first means all the harvested tags end up in the same queryable-only-by-full-scan column.

**How to avoid:**
- Migrate to a normalized `tags` table and `image_tags` join table before adding tag harvesting
- Schema: `tags(id, name, source, source_tag_id)` and `image_tags(image_hash, tag_id)`
- The existing `images.tags` column can remain as a denormalized cache for fast display — populate it from the join table
- Write the migration as a versioned schema upgrade (check schema version at startup, apply if needed)

**Warning signs:**
- Tag filtering requires `WHERE tags LIKE '%"rem"%'` — a red flag that normalization is needed
- Adding an index on `images.tags` provides no benefit for substring searches

**Phase to address:**
Tag harvesting phase — schema migration must run before the first tag is inserted, not retrofitted after.

---

### Pitfall 5: Per-Source Rate Limiters Not Isolated — Sources Throttle Each Other

**What goes wrong:**
The existing `wallhaven/client.go` embeds a rate limiter (`time.Ticker`) in the client struct. When `--source all` is used, the current implementation calls `fetchFromWallhaven` then `fetchFromReddit` sequentially (confirmed in `cmd/fetch.go` line 358-374). When parallel multi-source fetching is added, a naive global rate limiter shared across sources will either throttle one source unnecessarily or allow another to exceed its own limit.

More concretely: Zerochan allows 60 req/min. Danbooru allows 10 req/sec (global IP-level). Konachan is undocumented but uses 421 throttle responses. A single shared limiter cannot express these different policies.

**Why it happens:**
Developers copy the Wallhaven client pattern (1 limiter per client instance) but then instantiate all clients from a single manager, sharing one limiter.

**How to avoid:**
- Each source adapter owns its own rate limiter with source-specific parameters
- Use `golang.org/x/time/rate.Limiter` (already in the Go standard extended library) rather than `time.Ticker` — Limiter is goroutine-safe and supports burst
- The download `Manager` is source-agnostic — rate limiting belongs in the source client, not the downloader
- When `--source all` runs sources in parallel, each source's goroutine blocks only on its own limiter

**Warning signs:**
- Zerochan gets a ban or 429 after a successful multi-source run
- Danbooru returns 503 or rate-limit headers during `--source all` runs
- Rate-limit sleep in one source blocks progress reporting for other sources

**Phase to address:**
Multi-source parallel fetching phase — rate limiter isolation must be designed before parallelism is introduced.

---

### Pitfall 6: Resumable Downloads Require Server-Side `Accept-Ranges` — Not Guaranteed

**What goes wrong:**
Resumable downloads depend on the server responding with `Accept-Ranges: bytes` and honoring `Range: bytes=X-` requests. Image CDNs serving Danbooru, Zerochan, and Konachan images may not support range requests (especially if behind Cloudflare or a content-gating CDN). If the adapter sends a `Range` header to a server that ignores it, the server responds with 200 and the full file — and the client appends to the partial file, producing a corrupt image.

**Why it happens:**
Developers implement resumable downloads, test against a server that supports ranges, and don't test the fallback. The CDN in front of the booru strips `Range` headers or responds 200 instead of 206.

**How to avoid:**
- Always send a HEAD request first (or check the initial GET response headers) for `Accept-Ranges: bytes`
- If absent, fall back to full re-download — delete the partial `.tmp` file and start clean
- If present, send `Range: bytes=<partial_size>-` and verify the response is `206 Partial Content` before appending
- Use `If-Range` with the ETag from the original response to prevent appending to a stale partial file
- Log a debug message when range requests are not supported so users can diagnose unexpectedly slow re-runs

**Warning signs:**
- Downloaded files are larger than expected (double the correct size)
- Images fail to open after a resumed download
- `file --mime-type` reports wrong type on a "resumed" file

**Phase to address:**
Resumable download phase — the HEAD-check-then-range-or-fallback pattern must be the implementation foundation, not an afterthought.

---

### Pitfall 7: JSON Event Stream Buffering — Events Arrive Late or Never

**What goes wrong:**
When writing JSON lines to stdout in a long-running command, Go's `os.Stdout` is buffered by default when piped. A GUI app waiting for `{"event":"progress","downloaded":5}` will not receive it until the buffer flushes (typically at 4KB or process exit). This makes the stream appear frozen until the command finishes.

**Why it happens:**
Go's `fmt.Println` to a pipe does not auto-flush. A developer tests by printing to a terminal (which is line-buffered) and everything works; piping to a GUI app (which is fully buffered) shows stale or delayed events.

**How to avoid:**
- Use `bufio.NewWriter(os.Stdout)` and call `.Flush()` after every JSON line write
- Or use `json.NewEncoder(os.Stdout)` and wrap it in an explicit-flush encoder
- Better: create a `JSONEmitter` type that wraps `bufio.Writer` and flushes on every `Emit()` call — the single abstraction from Pitfall 3 handles this naturally
- Verify with `wallpaper-cli fetch --json | cat` — if `cat` receives lines incrementally, buffering is correct

**Warning signs:**
- GUI app shows no progress until the command exits, then all events arrive at once
- `wallpaper-cli fetch --json | head -1` hangs indefinitely (head is waiting for the first line that never flushes)

**Phase to address:**
JSON output phase — flush behavior must be verified in the first implementation of JSON event emission, not discovered later by the GUI consumer.

---

### Pitfall 8: Source-ID Uniqueness Assumption Breaks Cross-Source Dedup

**What goes wrong:**
The current schema uses `(source, source_id)` as a logical unique key (index confirmed in `db.go` line 96-97). This is correct. However, Danbooru post IDs are integers (e.g., `6543210`), Zerochan entry IDs are integers, and Konachan post IDs are integers. If tag harvesting stores tags with source-scoped IDs, a bug where the source column is not set correctly will cause Danbooru tag ID `1234` to be treated as the same record as Konachan tag ID `1234`.

**Why it happens:**
When adding the third and fourth source, the source column gets hardcoded incorrectly in copy-pasted code, or a default value is left empty, silently colliding IDs across sources.

**How to avoid:**
- Define source name constants (`const SourceDanbooru = "danbooru"`) and use them everywhere — never raw string literals
- Add a `CHECK` constraint or application-level assertion that `source` is never empty before any insert
- For the tags table, use `(source, source_tag_id)` as a unique key on the normalized table
- Write a test that inserts tags from two sources with the same numeric ID and verifies no collision

**Warning signs:**
- `SELECT COUNT(*) FROM image_tags WHERE tag_id = X` returns unexpectedly high counts
- A Konachan wallpaper shows Danbooru tag names in its metadata

**Phase to address:**
Tag harvesting phase — source constant definitions and uniqueness constraints belong in the schema design, before any insert logic is written.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Keep `images.tags` as JSON blob and query with `LIKE` | Zero migration effort | O(n) tag queries, no index, unusable for tag filtering | Never — schema migration before tag harvesting is required |
| One shared `http.Client` timeout across all sources | Less code | A slow Zerochan response blocks Danbooru downloads in parallel mode | Never in parallel mode; acceptable for sequential-only |
| Skip `Accept-Ranges` check, always send `Range` header | Simpler resume logic | Corrupt images when CDN ignores range requests | Never |
| Hardcode source name strings instead of constants | Fast to write | Cross-source ID collisions, hard to rename | Never |
| Emit JSON events to stderr instead of stdout | Avoids buffering decisions | Breaks pipe consumers; stderr is for errors | Never |
| Add `--json` flag but leave `fmt.Println` calls in command bodies | Additive change, no refactor | Human text leaks into JSON stream | Only if JSON mode is truly opt-in and documented as experimental |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Danbooru API | Send 3+ tags without API key, get silent empty result | Count tags before request; enforce 2-tag limit for anonymous; surface clear error |
| Danbooru API | Use `api_key` in URL query params (exposed in logs) | Prefer HTTP Basic Auth: `Authorization: Basic base64(login:api_key)` |
| Zerochan API | Treat 404 as a request error | Catch 404, return empty result set with no error |
| Zerochan API | Omit custom User-Agent | Always send `User-Agent: wallpaper-cli/<version> (<contact>)` — anonymous requests risk bans |
| Zerochan API | Query a meta tag directly | Meta tags (e.g., "Anime") are not queryable; use character/series tags |
| Konachan API | Send plain-text password | Password must be SHA1-hashed with Konachan's salt; authenticate for read-only if needed, but most searches work unauthenticated |
| Konachan API | Exceed undocumented rate limit | Treat HTTP 421 as a retryable throttle, not a permanent error; back off and retry |
| All booru CDNs | Send `Range` header without checking `Accept-Ranges` | HEAD first; only use range resume if server confirms support |
| JSON event stream | Write events with `fmt.Println` to a pipe | Use `bufio.Writer` with explicit `Flush()` after each event |
| Multi-source parallel | Share one rate limiter across sources | Each source client owns its own `golang.org/x/time/rate.Limiter` |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Loading all tag blobs into memory to filter by tag | High RSS during `list --tag X` with large collections | Normalize tags into a join table with an index | ~500+ images with tags |
| Fetching all pages from all sources before starting downloads | Slow to first download, high memory for large `--limit` | Pipeline: stream pages into download queue as they arrive | `--limit 200+` across 3 sources |
| No per-source concurrency cap when using `--source all` | One slow source starves bandwidth for faster sources | Per-source goroutine with its own semaphore | Any parallel multi-source run |
| Checking pHash dedup against full DB on every download in a batch | DB lock contention, slow batch completion | Pre-fetch all known hashes into a local map before the batch starts | Batches >50 with existing collection |
| Downloading image to temp file then computing pHash before rename | Double I/O for every image | Compute pHash during the streaming write (tee the download) | High-concurrency downloads |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Logging the full Danbooru request URL (which contains `api_key=` query param) | API key leaked to log files or terminal scrollback | Use HTTP Basic Auth instead of query params; redact keys in debug logs |
| Storing Konachan password in Viper config as plaintext | Password visible in config file | Store Konachan hashed password or use per-session hash; document the risk clearly |
| Not validating the download URL scheme before fetching | A malformed source_url in the DB could trigger a file:// or ssrf-adjacent request | Validate all download URLs start with `https://` before the `http.NewRequest` call |
| Blind temp-file atomic rename across filesystem boundaries | `os.Rename` fails cross-filesystem (e.g., /tmp to /mnt/nas) | Create temp file in same directory as final destination, not in os.TempDir() |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Adding `danbooru`, `zerochan`, `konachan` as valid `--source` values but not listing them in `fetch --help` | Users don't know new sources exist | Update `validSources` slice and help text in the same commit as the adapter |
| JSON mode emits no progress signal for long downloads | GUI app shows spinner with no data for minutes | Emit `{"event":"progress","source":"danbooru","queued":X}` events even during the search phase, before downloads start |
| Tag harvesting silently drops tags that exceed column width or have unexpected unicode | Collection metadata is incomplete with no user-facing notice | No column width limit in SQLite TEXT; sanitize only control characters, preserve all unicode |
| Resume logic always re-verifies pHash even for already-known files | Wastes CPU on files already in DB | Check `(source, source_id)` in DB before downloading; skip entirely if already tracked |
| `--source all` with `--limit 10` gives 10 per source, not 10 total | User expects 10 total, gets 30+ | Document this behavior explicitly; add `--limit-total` flag or change semantics in the milestone |

---

## "Looks Done But Isn't" Checklist

- [ ] **Zerochan adapter:** Returns empty result (not error) for 404 responses — verify with a search for a nonexistent tag
- [ ] **Danbooru adapter:** Enforces 2-tag limit for anonymous users and surfaces a clear error message — verify with 3-tag query and no API key
- [ ] **Tag normalization:** `image_tags` join table has `(image_hash, tag_id)` UNIQUE constraint — verify with duplicate insert attempt
- [ ] **JSON event stream:** Events arrive incrementally when piped — verify with `wallpaper-cli fetch --json | head -5` completing without hanging
- [ ] **Resumable downloads:** Corrupt partial files are detected and re-downloaded, not appended to — verify by truncating a `.tmp` file mid-download and re-running
- [ ] **Rate limiters:** Each source client has its own limiter — verify with a parallel `--source all` run that does not trigger 429/421 on any source
- [ ] **Source constants:** No raw source-name string literals in insert code — verify with `grep -r '"danbooru"' internal/` finding only the constant definition
- [ ] **Backward compatibility:** All existing `fetch --source wallhaven` and `fetch --source reddit` commands produce identical output in non-JSON mode — verify with existing tests

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Tags stored as JSON blob, queries too slow | HIGH | Write schema migration v2; backfill `image_tags` from existing `images.tags` JSON; keep blob as cache |
| Corrupt images from range-request fallback not implemented | LOW | Delete partial `.tmp` files; fix the fallback; re-run fetch |
| JSON events buffered, GUI app received nothing | LOW | Add flush calls; re-test with piped output |
| Cross-source source_id collision in tag table | MEDIUM | Add `source` column to tags table if missing; re-harvest tags with correct source values |
| Rate limit bans from Zerochan or Danbooru | MEDIUM | Wait for ban to expire (hours to days); add per-source rate limiter; reduce concurrency |
| `os.Rename` fails cross-filesystem for temp files | LOW | Change temp file creation to use `filepath.Dir(destination)` instead of `os.TempDir()` |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Danbooru 2-tag limit | Danbooru adapter implementation | Test with 3-tag query, no API key — must return clear error |
| Zerochan 404 = empty results | Zerochan adapter implementation | Test with nonexistent tag — must return 0 results, no error |
| Progress bar polluting JSON stream | JSON output phase (before adapter phases) | Pipe fetch output through `jq` — must parse without error |
| Tags stored as JSON blob | Schema migration (before tag harvesting) | `EXPLAIN QUERY PLAN SELECT WHERE tag='x'` must use index |
| Per-source rate limiter isolation | Multi-source parallel fetch phase | Parallel run must not trigger 429/421 on any source |
| Resumable download range-request assumptions | Download pipeline improvement phase | Truncate partial file mid-download, re-run — file must be valid |
| JSON event stream buffering | JSON output phase | `fetch --json \| head -1` must return immediately |
| Cross-source source_id collision | Tag harvesting schema phase | Insert same numeric ID for two sources — no collision |

---

## Sources

- Danbooru API docs: https://danbooru.donmai.us/wiki_pages/help:api — rate limits, pagination, authentication (HIGH confidence)
- Konachan API docs: https://konachan.com/help/api — pagination limits (hard cap 100/page), 421 throttle status (HIGH confidence)
- Zerochan API docs: https://www.zerochan.net/api — 60 req/min limit, User-Agent requirement, meta tag restriction (HIGH confidence)
- gallery-dl issue #8313: Zerochan returns 404 for empty results — https://github.com/mikf/gallery-dl/issues/8313 (HIGH confidence)
- gallery-dl issue #209: Danbooru 2-tag search limit — https://github.com/mikf/gallery-dl/issues/209 (HIGH confidence)
- Existing codebase: `internal/data/db.go`, `internal/download/manager.go`, `cmd/fetch.go` — direct code inspection (HIGH confidence)
- HTTP Range Requests spec — MDN: https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Range_requests (HIGH confidence)
- Go JSON streaming pitfalls: https://medium.com/geekculture/pitfalls-of-golang-interface-streaming-to-json-part1-1a067c9bb3cd (MEDIUM confidence)

---
*Pitfalls research for: wallpaper-cli v1.3 — Booru adapters, JSON API, resumable downloads, tag storage*
*Researched: 2026-04-04*
