# Roadmap: Wallpaper CLI Tool

## Milestones

- ✅ **v1.0/1.1 Core CLI** - Phases 1-2 (shipped)
- ✅ **v1.2 Desktop Integration** - Phase 3 (shipped)
- ✅ **v1.3 Sources, API & Downloads** - Phases 4-9 (shipped)

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

<details>
<summary>✅ v1.3 Sources, API & Downloads (Phases 4-9) — SHIPPED</summary>

3 new source adapters (Danbooru, Konachan, Zerochan), source tag harvesting, stable JSON API contract with NDJSON event stream, resumable downloads, retry with backoff, parallel multi-source fetch.

**Note:** Zerochan adapter code is complete but blocked by TLS-fingerprint anti-bot. Deferred to GUI app which has a browser engine.

### Phase 4: Foundation
**Goal**: Source interface/registry, output emitter, tag schema, per-source rate limiter
**Status**: Complete

### Phase 5: Danbooru Adapter
**Goal**: Danbooru fetch with tag harvesting and category metadata
**Status**: Complete

### Phase 6: Konachan Adapter
**Goal**: Konachan fetch reusing shared booru pattern
**Status**: Complete

### Phase 7: Zerochan Adapter
**Goal**: Zerochan with 404-as-empty, 2-call pattern, User-Agent handling
**Status**: Complete (code; anti-bot deferred to GUI)

### Phase 8: JSON API Contract
**Goal**: --json on all commands, NDJSON event stream, capabilities
**Status**: Complete

### Phase 9: Download Pipeline
**Goal**: Resumable downloads, retry/backoff, parallel multi-source fetch
**Status**: Complete

</details>

---

## Progress

| Phase | Milestone | Status | Completed |
|-------|-----------|--------|-----------|
| 1. Core Download Pipeline | v1.0/1.1 | ✅ Complete | - |
| 2. Source Adapters v1 | v1.0/1.1 | ✅ Complete | - |
| 3. Desktop Integration | v1.2 | ✅ Complete | - |
| 4. Foundation | v1.3 | ✅ Complete | 2026-04-04 |
| 5. Danbooru Adapter | v1.3 | ✅ Complete | 2026-04-04 |
| 6. Konachan Adapter | v1.3 | ✅ Complete | 2026-04-04 |
| 7. Zerochan Adapter | v1.3 | ✅ Complete | 2026-04-04 |
| 8. JSON API Contract | v1.3 | ✅ Complete | 2026-04-04 |
| 9. Download Pipeline | v1.3 | ✅ Complete | 2026-04-04 |

## Backlog

| ID | Title | Priority | Notes |
|----|-------|----------|-------|
| M004 | GUI App | High | Separate app on top of CLI JSON API; handles Zerochan via browser engine |
| M005 | AI Auto-Tagging | Low | Local model or API-based; source tags already harvested |
| M006 | Multi-Monitor Support | Medium | Per-display wallpaper setting |

---

*Roadmap maintained by GSD workflow*
*Last updated: 2026-04-05*
