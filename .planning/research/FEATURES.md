# Feature Research

**Domain:** Multi-source wallpaper CLI with Booru adapters, JSON event stream, and resumable download pipeline
**Researched:** 2026-04-04
**Confidence:** HIGH (API docs verified from official sources; Go patterns from official blog and active libraries)

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features the GUI consumer and power users assume exist. Missing these = the milestone isn't done.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Danbooru adapter: tag search + pagination | Danbooru is the reference Booru; any anime wallpaper tool must support it | MEDIUM | REST JSON API; unauthenticated gets 2 tags max; authenticated unlocks more; max 200 posts/page; pages via `?page=N` or cursor (`b<id>`/`a<id>`) |
| Konachan adapter: tag search + pagination | Konachan is Moebooru-based; widely used for high-res anime wallpapers | MEDIUM | API is Moebooru-compatible (Danbooru 1.13); `post.json?tags=...&limit=100&page=N`; max 100 posts/page; login + SHA1 password hash for auth |
| Zerochan adapter: tag search + pagination | Zerochan is a major anime wallpaper source with large dimensions filter | MEDIUM | Read-only REST; `/?tags&json&l=250&p=N`; User-Agent header with project+username is mandatory (ban risk otherwise); 60 req/min cap |
| Tag harvesting: persist source tags to DB | Milestone explicitly requires this; feeds future AI tagging | MEDIUM | Danbooru tags have categories (general, character, copyright, artist, meta); Zerochan/Konachan return tag lists per-post; store tag + category + source |
| Stable JSON output for all commands | GUI app depends on a parseable contract; must not break between runs | LOW | Use `--json` flag pattern; stable field names, ISO 8601 timestamps; additive changes only — never remove fields |
| JSON lines event stream for progress | GUI needs real-time download progress without polling; pipe-friendly | MEDIUM | Write NDJSON to stdout: one JSON object per newline; event types: `started`, `progress`, `completed`, `error`; flush after each line |
| Resumable downloads | Large wallpaper files (4K+) can fail mid-download; retry should not re-download from zero | MEDIUM | HTTP Range requests (`Range: bytes=N-`); detect server support via `Accept-Ranges: bytes`; save `.part` file; rename on completion |
| Retry with exponential backoff | Booru APIs rate-limit aggressively; a flat retry loop causes bans | LOW | Retry on 429, 503, 504; parse `Retry-After` header; backoff: `min(2^n + jitter, 60s)`; max 5 retries; use `hashicorp/go-retryablehttp` or stdlib |
| Parallel multi-source fetching | Users querying 3 sources at once should not wait serially | MEDIUM | Fan-out pattern: one goroutine per source, merge results via fan-in channel; per-source rate limiter to avoid bans |

### Differentiators (Competitive Advantage)

Features that go beyond the table stakes and make the tool meaningfully better.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Per-source configurable rate limiter | Prevents bans without user needing to guess safe delays; each source has different tolerance | LOW | Token bucket per source in config; Zerochan: 60/min, Danbooru: 10/s, Konachan: conservative default 2/s |
| Cursor-based pagination (not offset) | Offset pagination on Danbooru breaks with new posts; cursor gives stable results | LOW | Danbooru supports `b<id>` (before) and `a<id>` (after); use last post ID as cursor; simpler to implement correctly |
| Tag category metadata in DB | Enables richer filtering later (search only "character" tags, etc.); foundation for future AI work | LOW | Danbooru has 5 categories: 0=general, 1=artist, 3=copyright, 4=character, 5=meta; store as integer, display as string |
| Graceful partial-result emission | Even if one source errors, return results from the others instead of failing the whole command | LOW | Fan-in collects results + errors separately; emit partial results with per-source error in JSON output |
| Source capability advertisement | GUI can query what each source supports (dimensions filter, rating filter, sort options) before showing controls | LOW | Static capability map per adapter: `{"supports_rating": true, "supports_dimensions": true, "max_tags": 2}` — returned in `--json` source list |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| WebSocket/SSE event stream | Seems more "real-time" than NDJSON | Requires an HTTP server running alongside CLI; huge complexity for no gain when reading from pipe or subprocess | JSON lines to stdout: GUI reads from subprocess stdout line by line — simpler, no server needed |
| Scraping Zerochan HTML as fallback | Some image URLs are only in page HTML | ToS gray area; brittle vs HTML changes; maintenance burden | Use official Zerochan JSON API exclusively; flag unsupported entries, do not scrape |
| Danbooru "explicit" content by default | Source allows it | Requires user opt-in; rating filter should default to `s` (safe) or `q` (questionable); unauthenticated API already restricts | `rating:safe` default; expose `--rating` flag for authenticated users who opt in |
| Unbounded parallelism across sources | Faster fetching | Danbooru bans IPs that exceed 10 req/s; Zerochan bans at >60 req/min; goroutine storm with no semaphore causes bans | Worker pool with per-source semaphore; configurable concurrency cap, default conservative |
| Caching API responses to disk | Reduces API calls on repeat queries | Stale cache serves removed/updated images; cache invalidation is complex; not worth it for a download tool | Trust pHash deduplication + SQLite tracking already in place — the source of truth is what was downloaded, not what the API returned |

---

## Feature Dependencies

```
Danbooru Adapter
    └──requires──> Stable JSON output contract  (adapter must emit standard post schema)
    └──requires──> Retry with exponential backoff  (rate limit recovery)
    └──enhances──> Tag harvesting  (tag category metadata richer on Danbooru)

Konachan Adapter
    └──requires──> Stable JSON output contract
    └──requires──> Retry with exponential backoff
    └──shares patterns with──> Danbooru Adapter  (Moebooru/Booru-style APIs nearly identical)

Zerochan Adapter
    └──requires──> Stable JSON output contract
    └──requires──> Retry with exponential backoff
    └──requires──> Per-source rate limiter  (60 req/min hard cap; violations cause bans)

Tag Harvesting
    └──requires──> Danbooru Adapter OR Konachan Adapter OR Zerochan Adapter  (at least one source)
    └──requires──> SQLite schema extension  (new tags table or tags column in existing posts table)

JSON Lines Event Stream
    └──requires──> Stable JSON output contract  (event envelope must match overall schema)
    └──enhances──> Resumable Downloads  (emit progress events during resume)
    └──enhances──> Parallel Multi-Source Fetching  (emit per-source progress independently)

Resumable Downloads
    └──requires──> Retry with exponential backoff  (resume is a retry-class operation)
    └──builds on──> Existing Download Manager  (internal/download/manager.go)

Parallel Multi-Source Fetching
    └──requires──> All three adapters exist
    └──requires──> Per-source rate limiter  (parallel without rate limiting = instant bans)
    └──enhances──> JSON Lines Event Stream  (real-time per-source progress events)

Per-Source Rate Limiter
    └──required by──> Zerochan Adapter  (hard cap)
    └──required by──> Parallel Multi-Source Fetching
```

### Dependency Notes

- **Adapters require stable JSON contract first:** The output schema must be locked before adapters are built, otherwise each adapter invents its own field names and the GUI breaks.
- **Konachan shares Danbooru patterns:** Moebooru API is described as "mostly compatible with Danbooru API 1.13." Build Danbooru first, Konachan reuses 80% of the pattern.
- **Zerochan requires rate limiter before parallelism:** Its 60 req/min limit is strict and violations cause bans. The rate limiter must be in place before multi-source parallel fetch is turned on.
- **Tag harvesting requires schema work:** The existing posts SQLite table needs a companion tags table (or JSONB column) before any adapter can persist tags.

---

## MVP Definition

This is a subsequent milestone, not a greenfield project. "MVP" here means: what must ship together for the milestone to be valid?

### Must Ship Together (v1.3 Complete)

- [ ] Stable JSON output contract documented and implemented (`--json` flag on fetch/list/export) — all downstream features depend on this
- [ ] Danbooru adapter — primary target; establishes the booru adapter pattern
- [ ] Konachan adapter — shares Danbooru pattern; low marginal cost after Danbooru
- [ ] Zerochan adapter — different API shape; independent effort
- [ ] Tag harvesting — persists tags from all three adapters into SQLite
- [ ] JSON lines event stream — `--stream` flag emits NDJSON progress events; GUI readiness
- [ ] Resumable downloads — `.part` file pattern + Range request detection
- [ ] Retry with exponential backoff — required for booru API reliability
- [ ] Per-source rate limiter — required before parallelism is safe

### Add After Core Ships (v1.3.x)

- [ ] Parallel multi-source fetching — depends on all adapters + rate limiter being stable
- [ ] Source capability advertisement — informational, not blocking
- [ ] Cursor-based pagination (Danbooru `b<id>`) — improve offset pagination first, optimize later

### Future Consideration (v2+)

- [ ] AI-powered auto-tagging — explicitly deferred; harvest source tags first (this milestone)
- [ ] Multi-monitor wallpaper setting — separate platform concern
- [ ] TUI / terminal browser — belongs in the GUI app, not the CLI

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Stable JSON output contract | HIGH | LOW | P1 |
| Danbooru adapter | HIGH | MEDIUM | P1 |
| Konachan adapter | HIGH | LOW (reuses Danbooru pattern) | P1 |
| Zerochan adapter | HIGH | MEDIUM | P1 |
| Retry with exponential backoff | HIGH | LOW | P1 |
| Per-source rate limiter | HIGH | LOW | P1 |
| Tag harvesting | HIGH | MEDIUM | P1 |
| JSON lines event stream | HIGH | MEDIUM | P1 |
| Resumable downloads | MEDIUM | MEDIUM | P1 |
| Parallel multi-source fetching | MEDIUM | MEDIUM | P2 |
| Source capability advertisement | LOW | LOW | P2 |
| Cursor-based pagination (Danbooru) | LOW | LOW | P3 |

---

## API Contract Notes (Verified from Official Sources)

These inform what the adapter code must handle — not opinions, but observed API behavior.

### Danbooru

- Base URL: `https://danbooru.donmai.us/posts.json`
- Auth: `login` + `api_key` query params, or HTTP Basic Auth (username:api_key)
- Unauthenticated: max 2 tags per search; `rating:safe` counts as a free third tag
- Pagination: `?page=N` (numbered) or `?page=b<id>` (before ID) / `?page=a<id>` (after ID)
- Max per page: 200 posts
- Rate limit: 10 requests/second; `x-rate-limit` response headers
- Tag search: space-delimited tags, supports wildcards (`*`), range syntax (`score:>100`)
- Tag listing: `GET /tags.json?search[name_matches]=...&search[category]=...`
- Tag categories: 0=general, 1=artist, 3=copyright, 4=character, 5=meta
- **Confidence: HIGH** (verified via danbooru.donmai.us/wiki_pages/help:api)

### Konachan

- Base URL: `https://konachan.com/post.json`
- Auth: `login` + `password_hash` (SHA1 of `"So-I-Heard-You-Like-Mupkids-?--_password_\--"`)
- Pagination: `?page=N&limit=100`
- Max per page: 100 posts
- Rate limit: HTTP 421 = "User Throttled"; no published req/s number — treat conservatively (2 req/s)
- Tag search: `?tags=tag1+tag2` (space-delimited, URL-encoded as `+`)
- Tag listing: `/tag.json?name=...&name_pattern=...`
- Same response shape as Danbooru 1.x (field names differ slightly: `file_url`, `sample_url`, `preview_url`)
- **Confidence: HIGH** (verified via konachan.com/help/api)

### Zerochan

- Base URL: `https://www.zerochan.net/`
- Auth: none for read; User-Agent MUST include project name + Zerochan username (e.g. `"WallpaperCLI - myusername"`)
- Pagination: `?p=N&l=250` (page number, up to 250 entries per page)
- Rate limit: 60 requests/minute (hard); bans for chronic violations
- Tag search: `/{Tag+Name}?json` for single tag; `/{Tag1,Tag2}?json` for multiple
- Sort: `?s=id` (recent) or `?s=fav` (popular); `?t=0|1|2` for time range on popularity
- Dimension filter: `?d=large|huge|landscape|portrait|square`
- Response: array of entries with `id`, `width`, `height`, `size`, `primary`, `src`, `full`
- **Confidence: HIGH** (verified via zerochan.net/api)

---

## Sources

- [Danbooru API Help Page](https://danbooru.donmai.us/wiki_pages/help:api) — verified 2026-04-04
- [Konachan API Help Page](https://konachan.com/help/api) — verified 2026-04-04
- [Zerochan API Page](https://www.zerochan.net/api) — verified 2026-04-04
- [Pybooru Moebooru API Reference](https://pybooru.readthedocs.io/en/stable/api_moebooru.html) — cross-reference for Konachan/Moebooru endpoint shapes
- [Go Pipelines and Cancellation (Official Blog)](https://go.dev/blog/pipelines) — fan-out/fan-in pattern authority
- [hashicorp/go-retryablehttp](https://pkg.go.dev/github.com/hashicorp/go-retryablehttp) — exponential backoff + Retry-After header parsing
- [Go encoding/json stream.go](https://go.dev/src/encoding/json/stream.go) — stdlib JSON encoder for NDJSON output
- [Transloadit: Resumable file downloader in Go](https://transloadit.com/devtips/build-a-resumable-file-downloader-in-go-with-concurrent-chunks/) — Range request patterns
- [Efficient JSON Streaming in Go (2025)](https://tiagomelo.info/golang/rest/api/streaming/2025/02/26/efficient-json-streaming-rest-api.html) — NDJSON memory efficiency patterns

---

*Feature research for: wallpaper-cli v1.3 — Booru adapters, JSON API contract, download pipeline*
*Researched: 2026-04-04*
