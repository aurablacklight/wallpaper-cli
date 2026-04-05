# Requirements: Wallpaper CLI v1.3

**Defined:** 2026-04-04
**Core Value:** Reliably fetch, deduplicate, and organize wallpapers from any supported source

## v1.3 Requirements

Requirements for this milestone. Each maps to roadmap phases.

### Foundation

- [ ] **FND-01**: Source interface and registry — all sources implement a common interface, cmd/fetch.go resolves by name
- [ ] **FND-02**: Output emitter abstraction — JSON lines to stdout, human text to stderr, flush after every event
- [ ] **FND-03**: Normalized `source_tags` SQLite table with migration — before any adapter writes tags
- [ ] **FND-04**: Per-source rate limiter — isolated token bucket per source client

### Source Adapters

- [ ] **SRC-01**: User can fetch wallpapers from Danbooru with tag search and pagination
- [ ] **SRC-02**: User can authenticate with Danbooru API key to bypass 2-tag limit
- [ ] **SRC-03**: User can fetch wallpapers from Konachan with tag search and pagination
- [ ] **SRC-04**: User can fetch wallpapers from Zerochan with tag search and pagination
- [ ] **SRC-05**: Each source advertises its capabilities (supported filters, max tags, auth options)

### Tags

- [ ] **TAG-01**: Tags from all source APIs are harvested and persisted to the database during fetch
- [ ] **TAG-02**: Tag category metadata is stored (general, character, copyright, artist, meta)

### JSON API

- [ ] **API-01**: All commands support `--json` flag for structured JSON output
- [ ] **API-02**: Fetch emits JSON lines event stream with started/progress/completed/error events
- [ ] **API-03**: Partial results are emitted when one source fails in multi-source mode

### Download Pipeline

- [ ] **DL-01**: Downloads are resumable via HTTP Range requests with .part file recovery
- [ ] **DL-02**: Failed requests retry with exponential backoff and Retry-After header support
- [ ] **DL-03**: User can fetch from all sources in parallel with `--source all`

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

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| FND-01 | Phase 4 | Pending |
| FND-02 | Phase 4 | Pending |
| FND-03 | Phase 4 | Pending |
| FND-04 | Phase 4 | Pending |
| SRC-01 | Phase 5 | Pending |
| SRC-02 | Phase 5 | Pending |
| TAG-01 | Phase 5 | Pending |
| TAG-02 | Phase 5 | Pending |
| SRC-03 | Phase 6 | Pending |
| SRC-04 | Phase 7 | Pending |
| API-01 | Phase 8 | Pending |
| API-02 | Phase 8 | Pending |
| API-03 | Phase 8 | Pending |
| SRC-05 | Phase 8 | Pending |
| DL-01 | Phase 9 | Pending |
| DL-02 | Phase 9 | Pending |
| DL-03 | Phase 9 | Pending |

**Coverage:**
- v1.3 requirements: 17 total
- Mapped to phases: 17
- Unmapped: 0

**Notes:**
- TAG-01 is assigned to Phase 5 (first adapter to write tags). Phases 6 and 7 extend tag harvesting for their respective sources — TAG-01 is fully satisfied when all three adapters are complete.

---
*Requirements defined: 2026-04-04*
*Last updated: 2026-04-04 after roadmap creation (phases 4-9)*
