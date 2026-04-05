# ROADMAP: Wallpaper CLI Tool

**Last Updated:** 2026-04-04
**Project Vision:** A resource-efficient CLI tool for downloading and managing anime wallpapers — designed to be consumed by a separate GUI app.

---

## Milestones

| ID | Title | Version | Status |
|----|-------|---------|--------|
| M001 | Core CLI | v1.0/1.1 | Complete |
| M002 | Desktop Integration | v1.2 | Complete (trimmed) |
| M003 | TBD | v1.3 | Not started |

---

## M001: Core CLI (v1.0/1.1) — Complete

Wallhaven + Reddit source adapters, concurrent download manager, pHash deduplication, SQLite tracking, cross-platform builds, progress bars, sorting/filtering.

**Slices:** S01–S08 all delivered.

---

## M002: Desktop Integration (v1.2) — Complete

Cross-platform `set` command (macOS/Linux/Windows), `list`, `export`, `stats`, collections (favorites/ratings/playlists), config management.

**Note:** TUI browser and schedule/daemon were built during M002 but stripped in a post-release cleanup. The CLI stays focused as a backend; UI belongs in a separate app.

**Slices:** S01–S04 delivered. S02/S03 (TUI) code removed.

---

## M003: TBD — Not Started

No milestone defined yet. Candidates:
- Additional sources (Zerochan, Danbooru, Konachan)
- Stable JSON API contract for GUI app consumption
- Metadata enrichment / auto-tagging
- Resumable downloads / parallel source fetching

---

## Backlog

| ID | Title | Priority | Notes |
|----|-------|----------|-------|
| M004 | AI Auto-Tagging | Low | Local model or API-based |
| M005 | Multi-Monitor Support | Medium | Per-display wallpaper setting |
| M006 | Web UI | Very Low | Deferred — separate GUI app planned |

---

*Roadmap maintained by GSD workflow*
