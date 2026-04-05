# Project State: Wallpaper CLI Tool

**Status:** Idle — Post-cleanup, awaiting M003 definition
**Last Updated:** 2026-04-04

---

## Current State

The CLI underwent a major cleanup after M002. Stripped features that don't belong in a CLI tool (TUI, schedule/daemon) and fixed architectural issues (broken data layer, filesystem-scanning commands).

### Active Command Surface

| Command | Description |
|---------|-------------|
| `fetch` | Download wallpapers from Wallhaven/Reddit with filtering |
| `set` | Set desktop wallpaper (macOS/Linux/Windows) |
| `list` | Query downloaded wallpapers from DB |
| `export` | Export metadata as JSON |
| `favorite` | Toggle favorites |
| `rate` | Rate wallpapers 1-5 |
| `playlist` | Manage playlists (create/add/list/show) |
| `config` | Manage CLI configuration |
| `stats` | Collection statistics |

### Key Metrics (Post-Cleanup)

- **Go source files:** 39
- **Dependencies:** 32 (direct + indirect)
- **Tests:** 47 passing across 7 packages
- **Removed:** TUI (Bubble Tea), schedule/daemon (cron), stub commands

### Post-M002 Cleanup (2026-04-04)

| Change | Commit | Impact |
|--------|--------|--------|
| Remove schedule/daemon | `5e92dc7` | Fixed broken cron dep, -968 lines |
| Strip TUI + Bubble Tea | `62a04ff` | -3,238 lines, dropped 25+ deps |
| Remove stub commands | `59c44f1` | search/update/clean were empty |
| Fix data layer + tests | `9d29d4d` | Added missing DB schema, exposed SQL methods |
| List/export use DB | `b3d1d79` | Replaced filesystem scanning |
| Add test coverage | `d57f912` | validate, config, cmd registration tests |
| Rewrite README | `2590172` | Matches actual CLI scope |

---

## Completed Milestones

### M001: Wallpaper CLI Tool (v1.0/1.1) — Complete

Core download pipeline: Wallhaven + Reddit sources, concurrent downloads, pHash deduplication, SQLite tracking, cross-platform builds, progress bars.

### M002: Desktop Integration (v1.2) — Complete (Trimmed)

Cross-platform `set` command, `list`/`export`/`stats` commands, collections (favorites/ratings/playlists). TUI and schedule features were stripped during cleanup — they belong in a separate GUI app.

---

## Next Steps

No active milestone. Direction: keep CLI focused as a backend tool. A separate GUI app will be built on top for browsing/scheduling.

Candidates for M003:
- Additional wallpaper sources (Zerochan, Danbooru, etc.)
- Better export format / stable JSON API contract
- Wallpaper tagging / metadata enrichment
- Performance: parallel source fetching, resumable downloads

---

*State maintained by GSD workflow*
