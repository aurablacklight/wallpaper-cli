# Roadmap: Wallpaper CLI Tool

## Milestones

- ✅ **v1.0/1.1 Core CLI** - Phases 1-2 (shipped)
- ✅ **v1.2 Desktop Integration** - Phase 3 (shipped)
- 🚧 **v1.3 Sources, API & Downloads** - Phases 4-9 (in progress)

---

## Phases

<details>
<summary>✅ v1.0/1.1 Core CLI (Phases 1-2) — SHIPPED</summary>

Wallhaven + Reddit source adapters, concurrent download manager, pHash deduplication, SQLite tracking, cross-platform builds, progress bars, sorting/filtering.

### Phase 1: Core Download Pipeline
**Goal**: Users can download wallpapers from Wallhaven and Reddit with deduplication
**Plans**: Complete

### Phase 2: Source Adapters v1
**Goal**: Wallhaven and Reddit are fully featured with filtering, sorting, and pagination
**Plans**: Complete

</details>

<details>
<summary>✅ v1.2 Desktop Integration (Phase 3) — SHIPPED</summary>

Cross-platform `set` command, `list`/`export`/`stats` commands, collections (favorites/ratings/playlists), config management. TUI and schedule features stripped post-release.

### Phase 3: Desktop Integration
**Goal**: Users can manage their wallpaper collection and set the desktop wallpaper
**Plans**: Complete

</details>

---

### 🚧 v1.3 Sources, API & Downloads (In Progress)

**Milestone Goal:** Expand the CLI into a multi-source, API-driven backend with a robust download pipeline — ready for a GUI app to consume.

## Phases

**Phase Numbering:**
- Integer phases (4-9): v1.3 milestone work
- Decimal phases: Urgent insertions if needed (e.g., 4.1)

- [ ] **Phase 4: Foundation** - Output emitter, source interface/registry, tag schema, and per-source rate limiter
- [ ] **Phase 5: Danbooru Adapter** - Full Danbooru source with tag harvesting and category metadata
- [ ] **Phase 6: Konachan Adapter** - Konachan source reusing the booru base pattern
- [ ] **Phase 7: Zerochan Adapter** - Zerochan source with divergent API shape handled
- [ ] **Phase 8: JSON API Contract** - Stable `--json` flag and event stream across all commands
- [ ] **Phase 9: Download Pipeline** - Resumable downloads, retry with backoff, parallel multi-source fetch

## Phase Details

### Phase 4: Foundation
**Goal**: The shared infrastructure is in place so adapters can be written against a stable interface and emit structured output without inventing their own formats
**Depends on**: Nothing (first v1.3 phase)
**Requirements**: FND-01, FND-02, FND-03, FND-04
**Success Criteria** (what must be TRUE):
  1. `fetch --source wallhaven` and `fetch --source reddit` work identically to before — all existing tests pass
  2. A new source can be registered by implementing one Go interface and adding one line to the registry — no changes to `cmd/fetch.go`
  3. JSON output to stdout and human text to stderr are fully separated — piping `fetch --json` produces no human-readable text on stdout
  4. The `source_tags` SQLite table exists after migration — `SELECT * FROM source_tags` returns no error on a fresh database
  5. Each source client holds its own rate limiter — one source being throttled does not block another source's requests
**Plans**: TBD

### Phase 5: Danbooru Adapter
**Goal**: Users can search and download wallpapers from Danbooru with full tag support, and all tags are persisted to the database
**Depends on**: Phase 4
**Requirements**: SRC-01, SRC-02, TAG-01, TAG-02
**Success Criteria** (what must be TRUE):
  1. `fetch --source danbooru --tags "scenery sky"` returns results and downloads images
  2. Running the same fetch with a configured API key returns more results than anonymous (bypasses 2-tag limit)
  3. After a fetch, `SELECT * FROM source_tags WHERE source = 'danbooru'` returns harvested tags with non-null categories (general, character, copyright, artist, meta)
  4. Requesting 3+ tags without an API key emits a clear error message pointing to the `api_key` config field — it does not silently return zero results
**Plans**: TBD

### Phase 6: Konachan Adapter
**Goal**: Users can search and download wallpapers from Konachan, reusing the Danbooru adapter pattern with Konachan-specific authentication
**Depends on**: Phase 5
**Requirements**: SRC-03
**Success Criteria** (what must be TRUE):
  1. `fetch --source konachan --tags "scenery"` returns results and downloads images
  2. Konachan tags are persisted to `source_tags` during fetch — existing Danbooru tags in the table are not affected
  3. HTTP 421 responses from Konachan are treated as retryable throttle errors — the CLI backs off and retries rather than failing immediately
**Plans**: TBD

### Phase 7: Zerochan Adapter
**Goal**: Users can search and download wallpapers from Zerochan, with all edge cases (404-as-empty, 2-call image URL pattern, mandatory User-Agent) handled correctly
**Depends on**: Phase 6
**Requirements**: SRC-04
**Success Criteria** (what must be TRUE):
  1. `fetch --source zerochan --tags "landscape"` returns results and downloads full-resolution images (not thumbnails)
  2. A tag search with no matching results returns zero wallpapers with no error — it does not propagate Zerochan's 404 as a failure
  3. Zerochan requests include a valid custom User-Agent derived from the config — omitting the Zerochan username from config produces a clear error before any requests are sent
  4. `fetch --source zerochan --tags "landscape" --strict` returns only results tagged exactly with the given tag
**Plans**: TBD

### Phase 8: JSON API Contract
**Goal**: Every CLI command emits stable, structured JSON when `--json` is passed, and fetch emits a real-time JSON lines event stream that a GUI can consume reliably
**Depends on**: Phase 7
**Requirements**: API-01, API-02, API-03, SRC-05
**Success Criteria** (what must be TRUE):
  1. `fetch --source danbooru --tags "sky" --json | head -1` prints the first event immediately — it does not hang until the fetch completes
  2. Every event in the stream has a consistent shape: `type`, `source`, `timestamp` (ISO 8601) fields — a GUI can parse any event without branching on missing fields
  3. `fetch --source danbooru --source reddit --json` with one source failing still emits partial results from the succeeding source before the error event
  4. `list --json` and `stats --json` emit structured JSON objects with stable field names — no human-readable text appears on stdout
  5. Running `fetch --source danbooru --json` outputs a capabilities object at stream start — a GUI can read supported filters and auth options before showing UI controls
**Plans**: TBD

### Phase 9: Download Pipeline
**Goal**: Downloads survive network interruptions, respect source rate limits under load, and users can fetch all sources in one command
**Depends on**: Phase 8
**Requirements**: DL-01, DL-02, DL-03
**Success Criteria** (what must be TRUE):
  1. Interrupting a download mid-file and re-running the same fetch resumes from where it stopped — the partial `.part` file is reused, not redownloaded from byte 0
  2. A source returning HTTP 429 causes the CLI to wait the `Retry-After` duration and retry — it does not fail immediately or hammer the server
  3. `fetch --source all --tags "scenery"` fetches from Danbooru, Konachan, Zerochan, Wallhaven, and Reddit concurrently — results from all sources appear in the output
  4. A source CDN ignoring the `Range` header (returning HTTP 200 instead of 206) is detected — the partial file is discarded and the download restarts cleanly rather than producing a corrupt image
**Plans**: TBD

---

## Progress

**Execution Order:**
Phases execute in numeric order: 4 → 5 → 6 → 7 → 8 → 9

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Core Download Pipeline | v1.0/1.1 | — | Complete | - |
| 2. Source Adapters v1 | v1.0/1.1 | — | Complete | - |
| 3. Desktop Integration | v1.2 | — | Complete | - |
| 4. Foundation | v1.3 | 0/? | Not started | - |
| 5. Danbooru Adapter | v1.3 | 0/? | Not started | - |
| 6. Konachan Adapter | v1.3 | 0/? | Not started | - |
| 7. Zerochan Adapter | v1.3 | 0/? | Not started | - |
| 8. JSON API Contract | v1.3 | 0/? | Not started | - |
| 9. Download Pipeline | v1.3 | 0/? | Not started | - |

---

*Roadmap maintained by GSD workflow*
*Last updated: 2026-04-04*
