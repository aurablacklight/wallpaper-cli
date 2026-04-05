# Wallpaper CLI

## What This Is

A resource-efficient CLI tool for downloading and managing anime wallpapers from multiple sources. Designed as a headless backend — a separate GUI app will sit on top, consuming structured JSON output. Power users can also use it directly from the terminal.

## Core Value

Reliably fetch, deduplicate, and organize wallpapers from any supported source — the CLI is the single source of truth for the wallpaper collection.

## Current Milestone: v1.3 Sources, API & Downloads

**Goal:** Expand the CLI into a multi-source, API-driven backend with a robust download pipeline — ready for a GUI app to consume.

**Target features:**
- 3 new source adapters (Danbooru, Zerochan, Konachan)
- Source tag harvesting from all APIs
- Stable JSON API contract with event stream
- Download pipeline improvements (resumable, parallel, retry)

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

### Active

- [ ] Danbooru source adapter with tag-based search
- [ ] Zerochan source adapter
- [ ] Konachan source adapter
- [ ] Source tag harvesting — persist tags from all source APIs in DB
- [ ] Structured JSON output for all commands (query, CRUD)
- [ ] JSON lines event stream for real-time progress
- [ ] Resumable downloads (partial file recovery)
- [ ] Parallel multi-source fetching
- [ ] Retry with exponential backoff

### Out of Scope

- TUI / terminal browser — belongs in separate GUI app
- Schedule / daemon / auto-rotation — belongs in separate GUI app
- AI-powered auto-tagging — deferred to future milestone, harvest source tags first
- Cloud sync — distributed systems complexity, no current need
- Multi-monitor wallpaper setting — deferred to future milestone

## Context

- Go CLI using Cobra, Viper config, SQLite (modernc.org/sqlite, CGO-free)
- 39 Go source files, 47 tests across 7 packages, 32 dependencies
- Post-cleanup: TUI (Bubble Tea), schedule/daemon, and stub commands were stripped
- Danbooru and Konachan share Moebooru/Booru-style APIs — adapter code can share patterns
- Zerochan is less documented but follows similar tag-based conventions
- The GUI app will consume CLI output, so JSON contract stability matters

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
| Source tags only (no AI this milestone) | Harvest what's free before investing in AI | — Pending |
| JSON lines for event stream | Simple, pipe-friendly, no WebSocket complexity | — Pending |

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
*Last updated: 2026-04-04 after milestone v1.3 initialization*
