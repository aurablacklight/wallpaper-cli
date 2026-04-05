# Project State: Wallpaper CLI Tool

**Status:** Active — Milestone v1.3 roadmap created
**Last Updated:** 2026-04-04

---

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-04)

**Core value:** Reliably fetch, deduplicate, and organize wallpapers from any supported source
**Current focus:** Phase 4 — Foundation (v1.3 Sources, API & Downloads)

---

## Current Position

Phase: 4 of 9 (Foundation)
Plan: — of — (not yet planned)
Status: Ready to plan
Last activity: 2026-04-04 — v1.3 roadmap created (phases 4-9 defined)

Progress: [░░░░░░░░░░] 0% (v1.3)

---

## Active Command Surface

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

---

## Accumulated Context

### Decisions

- [v1.3]: JSON lines to stdout, human text to stderr — emitter abstraction required before any adapter work
- [v1.3]: `source_tags` SQLite table must be created via migration before first tag insert — retrofitting is costly
- [v1.3]: Danbooru before Konachan — 80% code overlap via shared booru base package
- [v1.3]: Zerochan last — most divergent API (tag-in-URL-path, 2-call pattern, 404-as-empty, mandatory User-Agent)
- [v1.3]: Parallel fetch (`--source all`) after all adapters stable — rate limiter isolation is a prerequisite
- [v1.3]: 2 new deps approved: `cenkalti/backoff/v4`, `hashicorp/go-retryablehttp`

### Research Flags for Planning

- **Phase 7 (Zerochan):** Verify if `full` URL is available in list responses (may eliminate 2-call pattern)
- **Phase 9 (Parallel fetch):** Resolve `--limit` semantics — per-source or total? Consider `--limit-total` flag
- **Phase 6 (Konachan):** Empirically validate 0.5 req/s default; expose config knob for user tuning

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

---

## Performance Metrics

**Velocity:**
- Total plans completed: 0 (v1.3)
- Average duration: —
- Total execution time: —

*Updated after each plan completion*

---

## Session Continuity

Last session: 2026-04-04
Stopped at: Roadmap created — ready to plan Phase 4
Resume file: None
