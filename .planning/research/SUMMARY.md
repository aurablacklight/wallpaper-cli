# Project Research Summary

**Project:** wallpaper-cli-tool v1.3
**Domain:** Go CLI — multi-source Booru adapter integration, JSON event stream, resumable download pipeline, tag harvesting
**Researched:** 2026-04-04
**Confidence:** HIGH

## Executive Summary

This milestone adds three Booru image source adapters (Danbooru, Konachan, Zerochan) to an existing Go CLI wallpaper tool with established patterns for Wallhaven and Reddit. All three APIs are RESTful JSON and well-documented; no scraping or reverse engineering is required. The recommended approach is to introduce a shared `Source` interface backed by a registry, allowing `cmd/fetch.go` to shed its growing `switch/case` block in favor of `sources.Get(name, cfg)`. Konachan is Moebooru-compatible with Danbooru (80% shared struct shapes), so building Danbooru first delivers the second adapter nearly for free. Zerochan deviates most (tag-in-URL-path, 2-call pattern for full image URLs, mandatory custom User-Agent) and should be built last.

The output layer must be designed first, before any adapter work begins. The existing codebase mixes human-readable output with stdout in ways that will corrupt a JSON event stream if not refactored. A single `Emitter` abstraction (JSON lines to stdout, human text to stderr) eliminates an entire class of GUI-breaking bugs and is the foundational dependency for every other feature in this milestone. Similarly, the SQLite tag schema must be normalized (a `source_tags` table with proper indexes) before any adapter persists tags — retrofitting this after insertion code exists is costly.

The primary risks are API-specific: Danbooru silently returns empty results (not an error) when anonymous users send more than 2 tags; Zerochan returns HTTP 404 (not empty JSON) when a tag search has no results; and all three Booru CDNs may silently ignore HTTP Range headers, producing corrupt images if the fallback to full re-download is not implemented. Each of these is well-documented and straightforward to prevent with targeted defensive code. The milestone is low architectural risk — the existing download manager, dedup checker, and SQLite layer are all source-agnostic and require only additive changes.

---

## Key Findings

### Recommended Stack

The milestone requires only two new direct dependencies. `github.com/cenkalti/backoff/v4` (v4.3.0) provides battle-tested exponential backoff with jitter for the retry layer. `github.com/hashicorp/go-retryablehttp` (v0.7.8) wraps `net/http` with automatic `Retry-After` header parsing for 429/503 responses, which is critical for Danbooru's rate limiter. Everything else — NDJSON event streaming, HTTP Range requests for resumable downloads, parallel fan-out, and tag DB storage — is covered by stdlib or extends existing dependencies already in the project.

**Core technologies:**
- `github.com/cenkalti/backoff/v4` v4.3.0: retry with jitter — avoids writing custom exponential backoff (~14 lines of config vs. ~80 lines custom)
- `github.com/hashicorp/go-retryablehttp` v0.7.8: retryable HTTP client — built-in `Retry-After` header parsing, rewindable request bodies on retry
- `encoding/json` (stdlib): NDJSON event stream — `Encoder.Encode()` appends `\n` per call, which is the NDJSON spec; no third-party lib needed
- `net/http` + `os` (stdlib): resumable downloads via `Range: bytes=N-` header and `O_APPEND` file open; ~20 lines, no lib needed
- `modernc.org/sqlite` (existing): tag storage in new `source_tags` table; no new DB engine, no CGO

### Expected Features

**Must have (table stakes — v1.3 complete):**
- Danbooru adapter: tag search + pagination — primary Booru target, establishes the adapter pattern
- Konachan adapter: tag search + pagination — shares 80% of Danbooru implementation via shared `booru` base package
- Zerochan adapter: tag search + pagination — independent API shape, build after Danbooru/Konachan
- Tag harvesting: persist source tags to `source_tags` SQLite table — foundation for future AI tagging
- Stable JSON output contract (`--json` flag): locked field names, ISO 8601 timestamps, additive-only — GUI depends on this
- JSON lines event stream: NDJSON to stdout for real-time download progress, flush after every line
- Resumable downloads: `.part` file + HTTP Range request with 206/200 fallback detection
- Retry with exponential backoff: handles 429, 503, 504; parses `Retry-After` header; max 5 retries
- Per-source rate limiter: isolated `rate.Limiter` per adapter (Zerochan: 1/s, Danbooru: ~6/s, Konachan: conservative 0.5/s)

**Should have (competitive — v1.3.x after core ships):**
- Parallel multi-source fetching (`--source all`): fan-out goroutines per source, shared dedup-aware download batch
- Source capability advertisement: static `Capabilities` struct per adapter for GUI control rendering
- Cursor-based pagination (Danbooru `b<id>`): more stable than offset pagination for large harvests

**Defer (v2+):**
- AI-powered auto-tagging — source tag harvesting this milestone is the prerequisite
- Multi-monitor wallpaper setting — separate platform concern
- TUI / terminal browser — belongs in the GUI app layer

### Architecture Approach

The target architecture introduces a `Source` interface in `internal/sources/interface.go` and a registry in `internal/sources/registry.go` that maps source names to factory functions. All five sources (wallhaven, reddit, danbooru, konachan, zerochan) implement the same interface and return `[]model.Wallpaper` + `[]model.Tag`. The command layer (`cmd/fetch.go`) resolves a source by name via `sources.Get()` and calls `source.Search()`, eliminating all source-specific branching from the command file. An `Emitter` interface in `internal/output/` routes structured events to stdout (JSON lines) and human text to stderr, and is threaded through both the fetch loop and the download manager. The existing download manager, dedup checker, and config system are all additive changes only.

**Major components:**
1. `internal/sources/interface.go` + `registry.go` — Source interface and factory registry; eliminates switch/case in cmd/fetch.go
2. `internal/sources/danbooru/`, `konachan/`, `zerochan/` — new adapters (client.go, types.go, adapter.go each); danbooru/konachan share a `booru` base package
3. `internal/output/emitter.go` + `events.go` — Emitter interface with JSONLinesEmitter (stdout, mutex-protected) and TextEmitter (stderr)
4. `internal/model/tag.go` + `internal/data/db.go` (`source_tags` table) — tag model and normalized storage with SaveTags/GetTags/SearchTags
5. `internal/download/manager.go` (resumable) + `retry.go` (new) — Range header + 206/200 fallback, exponential backoff wrapper

### Critical Pitfalls

1. **Danbooru 2-tag anonymous limit** — Silently returns empty results (not an error) for 3+ tags without an API key. Enforce tag count before every API call; emit a clear user-facing error pointing to the `api_key` config field.

2. **Zerochan 404 = empty results, not an error** — Every other source returns 200 with empty array for no results. Zerochan returns 404. The existing error-handling pattern will propagate this as a failure. Catch 404 in the Zerochan client and return an empty `Result{}` instead.

3. **JSON event stream buffering** — Go's `os.Stdout` is fully buffered when piped. Events written with `json.NewEncoder(os.Stdout)` will not arrive at the GUI until the buffer fills (4KB) or process exits. Wrap in `bufio.Writer` and call `Flush()` after every `Emit()` call. Verify with `fetch --json | head -1` completing without hanging.

4. **Resumable downloads assume server Range support** — Booru CDNs (especially behind Cloudflare) may return HTTP 200 instead of 206 when a `Range` header is sent, ignoring the range. Appending this full response to a partial `.tmp` file produces a corrupt image. Always check `Accept-Ranges: bytes` in the response; fall back to deleting the `.tmp` and restarting if the server returns 200.

5. **Per-source rate limiters must be isolated** — A shared global rate limiter across sources allows one slow source to block others or one fast source to exhaust the shared budget. Each source client owns its own `golang.org/x/time/rate.Limiter` initialized in `NewClient()`. This is mandatory before parallel multi-source fetching is enabled.

6. **Tag schema normalization before first insert** — The existing `images.tags TEXT` column stores a JSON blob that cannot be indexed for tag queries. A new `source_tags` table with `UNIQUE(source, name)` and indexed columns must be created via a versioned migration before any adapter writes tags. Retrofitting this after insert code exists is high cost.

---

## Implications for Roadmap

Research reveals a clear dependency chain that dictates phase order. The output layer and schema must precede adapter work. Danbooru must precede Konachan (shared pattern). Zerochan is independent but most divergent. Parallel multi-source fetching requires all adapters and rate limiters to be stable first.

### Phase 1: Foundation — Output Layer + Schema
**Rationale:** The JSON event emitter and normalized tag schema are foundational dependencies for every subsequent phase. Building adapters before the emitter exists causes either (a) adapters inventing their own output format, or (b) a painful refactor after all three adapters are written. The schema migration must run before any insert code is written.
**Delivers:** `internal/output/` (Emitter interface, JSONLinesEmitter, TextEmitter, event structs); `internal/model/tag.go`; `source_tags` SQLite table with migration; `internal/download/retry.go`; `internal/sources/interface.go` + `registry.go`; wallhaven/reddit wrapped as Source implementors with all existing tests passing
**Addresses:** JSON output contract, event stream infrastructure, tag storage foundation
**Avoids:** Progress bar polluting JSON stream (Pitfall 3), tag JSON blob anti-pattern (Pitfall 4), JSON stream buffering (Pitfall 7)

### Phase 2: Danbooru Adapter
**Rationale:** Danbooru is the reference Booru with the richest tag category metadata (5 categories: general, character, copyright, artist, meta). Building it first establishes the booru adapter pattern that Konachan will reuse at 80% code overlap. Tag harvesting can be validated here against well-structured data.
**Delivers:** `internal/sources/danbooru/` (client.go, types.go, adapter.go); `internal/sources/booru/` shared BooruPost type; Danbooru wired into registry; tag harvesting validated end-to-end with `db.SaveTags()`
**Uses:** `hashicorp/go-retryablehttp` for Retry-After handling, `cenkalti/backoff/v4` for retries, `sources.Source` interface from Phase 1
**Avoids:** 2-tag anonymous limit (Pitfall 1) — enforce tag count before every request; API key exposure (use HTTP Basic Auth not query params)

### Phase 3: Konachan Adapter
**Rationale:** Moebooru-compatible API shares 80% of the Danbooru pattern. Low marginal implementation cost after Phase 2. Can share `BooruPost` struct shapes from `internal/sources/booru/`. Delivers a second validated source before tackling the more divergent Zerochan.
**Delivers:** `internal/sources/konachan/` adapter; Konachan wired into registry; SHA1 password hash auth support; HTTP 421 handled as retryable throttle
**Uses:** Booru base package from Phase 2; rate limiter at conservative 0.5 req/s default

### Phase 4: Zerochan Adapter
**Rationale:** Zerochan deviates most from the established pattern: tag-in-URL-path search, mandatory custom User-Agent, 2-call pattern (list → single-entry for full image URL), and 404-as-empty-results behavior. Building it last means the interface is proven and the team knows what edge cases look like.
**Delivers:** `internal/sources/zerochan/` adapter; 60 req/min rate limiter; strict mode support; Zerochan wired into registry
**Avoids:** 404-as-empty-results (Pitfall 2) — handle as `Result{}` not error; User-Agent omission (ban risk); meta tag contamination (use strict mode)

### Phase 5: Resumable Downloads
**Rationale:** Infrastructure improvement independent of source adapters. Can be built in parallel with Phase 3 or 4 if capacity allows, but depends on Phase 1 (emitter for progress events). Groups naturally with download pipeline changes.
**Delivers:** Resumable `.part` file logic in `download/manager.go`; `Accept-Ranges` check with 200/206 fallback; `If-Range` ETag validation; per-file progress events via emitter
**Avoids:** Range-request corruption (Pitfall 6) — HEAD-check-then-range-or-restart is the foundation, not an afterthought

### Phase 6: Parallel Multi-Source Fetching
**Rationale:** This phase depends on all adapters (Phases 2-4) being stable and all rate limiters being per-source isolated. Fan-out pattern is straightforward given the Source interface, but enabling parallelism before rate limiter isolation causes Zerochan/Danbooru bans.
**Delivers:** `--source all` goroutine fan-out; per-source goroutine with own rate limiter; shared dedup-aware download batch; parallel `EventFound` emission; source capability advertisement
**Avoids:** Rate limiter cross-contamination (Pitfall 5); blocking parallel sources on downloads (Architecture anti-pattern 3)

### Phase Ordering Rationale

- Output layer first because every adapter must emit events through the same channel — retrofitting this after 3 adapters exist is a significant merge/refactor cost
- Schema migration before any tag inserts — the `source_tags` table with `UNIQUE(source, name)` constraint and indexes must exist before the first adapter writes tags
- Danbooru before Konachan because Konachan reuses the booru base package; building in this order avoids defining shared types twice
- Zerochan last because it is the most divergent API shape; the interface is proven before encountering its edge cases
- Resumable downloads are independent of adapters and can slot between Phase 4 and Phase 6 without blocking
- Parallel fetch last because it requires all adapters, all rate limiters, and the emitter to be stable

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 4 (Zerochan):** The 2-call pattern (list → single entry for full URL) and the strict mode behavior interact in ways that may require rate limit budget planning beyond the documented 60 req/min — verify actual CDN behavior and whether `full` URL is available in list responses
- **Phase 6 (Parallel Multi-Source):** Dedup coordination across simultaneously-executing source goroutines and the shared download batch needs careful design to avoid race conditions on the pHash checker

Phases with standard patterns (skip research-phase):
- **Phase 1 (Foundation):** Output emitter and SQLite schema migration are well-established Go patterns; the Architecture research provides sufficient implementation detail
- **Phase 2 (Danbooru):** API fully documented; Wallhaven adapter is the proven pattern to follow
- **Phase 3 (Konachan):** Moebooru compatibility is documented; Danbooru phase establishes the template
- **Phase 5 (Resumable Downloads):** HTTP Range request pattern is well-specified (RFC 7233); implementation is ~20 lines of stdlib

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All library decisions verified via pkg.go.dev and official API docs; only 2 new deps, both widely used |
| Features | HIGH | All three API contracts verified from official documentation; feature dependencies are clear and traceable |
| Architecture | HIGH | Based on direct codebase inspection + official API docs; existing patterns (wallhaven adapter, download manager) are well understood |
| Pitfalls | HIGH | Most pitfalls verified via official API docs, gallery-dl issue tracker evidence, and direct code inspection; not inferred |

**Overall confidence:** HIGH

### Gaps to Address

- **Zerochan CDN `full` URL availability in list responses:** Research notes the list endpoint returns thumbnail URLs and full-resolution requires a second GET per entry ID. This 2x API call pattern needs to be validated against the rate limit budget during Phase 4 implementation. If `full` URL is included in the paginated list response, the 2-call pattern is unnecessary.
- **Konachan rate limit specifics:** HTTP 421 "User Throttled" is documented but the actual request rate that triggers it is not published. The conservative default of 0.5 req/s needs empirical validation during Phase 3 implementation; the config should allow the user to tune this.
- **`--source all` with `--limit` semantics:** Pitfalls research flags that `--limit 10` with `--source all` produces 10 results per source (30+ total), which is counterintuitive. The Phase 6 roadmap should explicitly resolve whether `--limit` means per-source or total, and add documentation or a `--limit-total` flag accordingly.
- **Zerochan User-Agent format:** The API requires `wallpaper-cli - <zerochan_username>`. The username must come from config (not hardcoded). Config shape for `sources.zerochan.username` needs to be added to `internal/config/config.go` in Phase 1 to avoid a later breaking config change.

---

## Sources

### Primary (HIGH confidence)
- Danbooru API: https://danbooru.donmai.us/wiki_pages/help:api — rate limits, auth, tag categories, pagination cursor
- Konachan API: https://konachan.com/help/api — endpoints, 421 throttle code, response format
- Zerochan API: https://www.zerochan.net/api — endpoints, strict mode, User-Agent requirement, rate limit
- `hashicorp/go-retryablehttp` v0.7.8: https://pkg.go.dev/github.com/hashicorp/go-retryablehttp — Retry-After support, rewindable bodies
- `cenkalti/backoff/v4` v4.3.0: https://pkg.go.dev/github.com/cenkalti/backoff/v4 — API surface, generics requirement (Go 1.18+)
- Existing codebase: `internal/data/db.go`, `internal/download/manager.go`, `cmd/fetch.go` — direct inspection
- gallery-dl issue #8313: https://github.com/mikf/gallery-dl/issues/8313 — Zerochan 404 = empty results (confirmed)
- gallery-dl issue #209: https://github.com/mikf/gallery-dl/issues/209 — Danbooru 2-tag limit (confirmed)
- HTTP Range Requests: https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Range_requests — RFC 7233 spec
- Go Pipelines blog: https://go.dev/blog/pipelines — fan-out/fan-in pattern authority

### Secondary (MEDIUM confidence)
- Transloadit Go resumable downloader: https://transloadit.com/devtips/build-a-resumable-file-downloader-in-go-with-concurrent-chunks/ — Range + O_APPEND pattern (stdlib implementation confirmed)
- Pybooru Moebooru API reference: https://pybooru.readthedocs.io/en/stable/api_moebooru.html — cross-reference for Konachan/Moebooru endpoint shapes
- Go JSON streaming pitfalls: https://medium.com/geekculture/pitfalls-of-golang-interface-streaming-to-json-part1-1a067c9bb3cd — buffering behavior

---
*Research completed: 2026-04-04*
*Ready for roadmap: yes*
