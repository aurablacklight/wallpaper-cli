# Requirements: Wallpaper CLI v1.3

**Defined:** 2026-04-04
**Completed:** 2026-04-05
**Core Value:** Reliably fetch, deduplicate, and organize wallpapers from any supported source

## v1.3 Requirements

### Foundation

- [x] **FND-01**: Source interface and registry — all sources implement a common interface, cmd/fetch.go resolves by name
- [x] **FND-02**: Output emitter abstraction — JSON lines to stdout, human text to stderr, flush after every event
- [x] **FND-03**: Normalized `source_tags` SQLite table with migration — before any adapter writes tags
- [x] **FND-04**: Per-source rate limiter — isolated token bucket per source client

### Source Adapters

- [x] **SRC-01**: User can fetch wallpapers from Danbooru with tag search and pagination
- [x] **SRC-02**: User can authenticate with Danbooru API key to bypass 2-tag limit
- [x] **SRC-03**: User can fetch wallpapers from Konachan with tag search and pagination
- [x] **SRC-04**: User can fetch wallpapers from Zerochan with tag search and pagination *(code complete; blocked by TLS-fingerprint anti-bot — deferred to GUI app)*
- [x] **SRC-05**: Each source advertises its capabilities (supported filters, max tags, auth options)

### Tags

- [x] **TAG-01**: Tags from all source APIs are harvested and persisted to the database during fetch
- [x] **TAG-02**: Tag category metadata is stored (general, character, copyright, artist, meta)

### JSON API

- [x] **API-01**: All commands support `--json` flag for structured JSON output
- [x] **API-02**: Fetch emits JSON lines event stream with started/progress/completed/error events
- [x] **API-03**: Partial results are emitted when one source fails in multi-source mode

### Download Pipeline

- [x] **DL-01**: Downloads are resumable via HTTP Range requests with .part file recovery
- [x] **DL-02**: Failed requests retry with exponential backoff and Retry-After header support
- [x] **DL-03**: User can fetch from all sources in parallel with `--source all`

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### AI Tagging

- **AI-01**: Local model auto-tags wallpapers missing source tags
- **AI-02**: External AI API enriches wallpaper descriptions

### Multi-Monitor

- **MON-01**: User can set different wallpapers per display
- **MON-02**: User can detect connected monitors

## Out of Scope

| Feature | Reason |
|---------|--------|
| TUI / terminal browser | Belongs in separate GUI app |
| Schedule / daemon / auto-rotation | Belongs in separate GUI app |
| AI-powered auto-tagging | Harvest source tags first; defer AI to future milestone |
| Cloud sync | Distributed systems complexity, no current need |
| WebSocket/SSE event stream | JSON lines to stdout is simpler, pipe-friendly, no server needed |
| Scraping HTML fallback | ToS gray area, brittle; use official APIs only |
| Response caching | pHash dedup + SQLite tracking is the source of truth |
| Zerochan CLI fetch | Anti-bot uses TLS fingerprinting; requires browser engine (GUI app) |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| FND-01 | Phase 4 | Complete |
| FND-02 | Phase 4 | Complete |
| FND-03 | Phase 4 | Complete |
| FND-04 | Phase 4 | Complete |
| SRC-01 | Phase 5 | Complete |
| SRC-02 | Phase 5 | Complete |
| TAG-01 | Phase 5 | Complete |
| TAG-02 | Phase 5 | Complete |
| SRC-03 | Phase 6 | Complete |
| SRC-04 | Phase 7 | Complete (code; blocked by anti-bot) |
| API-01 | Phase 8 | Complete |
| API-02 | Phase 8 | Complete |
| API-03 | Phase 8 | Complete |
| SRC-05 | Phase 8 | Complete |
| DL-01 | Phase 9 | Complete |
| DL-02 | Phase 9 | Complete |
| DL-03 | Phase 9 | Complete |

**Coverage:**
- v1.3 requirements: 17 total
- Complete: 17
- Unmapped: 0

---
*Requirements defined: 2026-04-04*
*Last updated: 2026-04-05 after milestone v1.3 completion*
