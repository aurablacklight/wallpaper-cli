# Wallpaper CLI

## What This Is

A resource-efficient CLI tool for downloading and managing anime wallpapers from multiple sources (Wallhaven, Reddit, Danbooru, Konachan). Designed as a headless backend — a separate GUI app will sit on top, consuming structured JSON output. Power users can also use it directly from the terminal.

## Core Value

Reliably fetch, deduplicate, and organize wallpapers from any supported source — the CLI is the single source of truth for the wallpaper collection.

## Requirements

### Validated

- ✓ Wallhaven source adapter with filtering/sorting — v1.0
- ✓ Reddit source adapter (r/Animewallpaper) — v1.1
- ✓ Concurrent download manager with progress bars — v1.0
- ✓ pHash deduplication with SQLite tracking — v1.0
- ✓ Cross-platform wallpaper setting (macOS/Linux/Windows) — v1.2
- ✓ Collection management (favorites, ratings, playlists) — v1.2
- ✓ List/export from database with filtering — v1.2
- ✓ Config management — v1.2
- ✓ Collection statistics — v1.2
- ✓ Source interface + registry (plug-in new sources) — v1.3
- ✓ Output emitter (NDJSON to stdout, text to stderr) — v1.3
- ✓ Normalized source_tags table with tag harvesting — v1.3
- ✓ Per-source rate limiters (isolated) — v1.3
- ✓ Danbooru adapter with tag search, pagination, auth — v1.3
- ✓ Konachan adapter (Moebooru-compatible) — v1.3
- ✓ Tag category metadata (general, character, copyright, artist, meta) — v1.3
- ✓ --json flag on fetch/list/stats with stable contract — v1.3
- ✓ JSON lines event stream with capabilities — v1.3
- ✓ Partial results on multi-source failure — v1.3
- ✓ Source capability advertisement — v1.3
- ✓ Resumable downloads (.part + HTTP Range) — v1.3
- ✓ Retry with exponential backoff + Retry-After — v1.3
- ✓ Parallel multi-source fetch (--source all) — v1.3

### Active

(None — awaiting next milestone)

### Out of Scope

- TUI / terminal browser — belongs in separate GUI app
- Schedule / daemon / auto-rotation — belongs in separate GUI app
- AI-powered auto-tagging — deferred to future milestone, harvest source tags first
- Cloud sync — distributed systems complexity, no current need
- Multi-monitor wallpaper setting — deferred to future milestone
- Zerochan live fetching from CLI — anti-bot uses TLS fingerprinting, requires browser engine (GUI app)

## Context

- Go CLI using Cobra, Viper config, SQLite (modernc.org/sqlite, CGO-free)
- 55+ Go source files, 124 tests across 13 packages, 34 dependencies
- 5 source adapters: Wallhaven, Reddit, Danbooru, Konachan, Zerochan (code complete, blocked by anti-bot)
- Shared booru base package for Danbooru/Konachan pattern reuse
- NDJSON event stream ready for GUI app consumption
- Cookies.txt parser available for future source auth needs

## Constraints

- **Tech stack**: Go, single binary, CGO-free — cross-platform builds must work
- **Binary size**: Target < 25MB
- **No new heavy deps**: Prefer stdlib and existing deps where possible
- **Backwards compatible**: Existing commands must keep working; JSON output is additive

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| CLI stays headless | GUI is a separate app; CLI is the backend | ✓ Good |
| CGO-free SQLite (modernc.org/sqlite) | Cross-compilation without C toolchain | ✓ Good |
| Strip TUI and daemon | Don't belong in CLI tool | ✓ Good |
| Source tags only (no AI this milestone) | Harvest what's free before investing in AI | ✓ Good |
| JSON lines for event stream | Simple, pipe-friendly, no WebSocket complexity | ✓ Good |
| Defer Zerochan to GUI app | Anti-bot uses TLS fingerprinting, can't bypass from CLI | ✓ Good |
| Shared booru package | Danbooru/Konachan share 80% code | ✓ Good |
| retryablehttp for API calls | Built-in Retry-After parsing, rewindable bodies | ✓ Good |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-05 after milestone v1.3 completion*
