# ROADMAP: Wallpaper CLI Tool

**Active Milestone:** M002 - Desktop Integration (v1.2)  
**Last Updated:** 2026-04-04 (Phase 01 planned)

**Project Vision:** A resource-efficient, single-binary CLI tool for downloading high-quality anime wallpapers with smart filtering, deduplication, organization, and desktop integration.

---

## Milestones

| ID | Title | Version | Status | Progress |
|----|-------|---------|--------|----------|
| M001 | Wallpaper CLI Tool | v1.0 | ✅ Complete | 8/8 slices |
| M002 | Desktop Integration | v1.2 | 🚧 Planned | 0/3 slices |

---

## ✅ M001: Wallpaper CLI Tool (COMPLETE)

**Status:** All slices delivered and verified via smoke tests

### Delivered Slices

| ID | Title | Status |
|----|-------|--------|
| S01 | Project Foundation & CLI Scaffold | ✅ |
| S02 | CLI Interface & Config System | ✅ |
| S03 | Wallhaven Source Adapter | ✅ |
| S04 | Download Manager | ✅ |
| S05 | Deduplication System | ✅ |
| S06 | Organization & Storage | ✅ |
| S07 | Cross-Platform Builds | ✅ |
| S08 | Reddit Source Adapter (v1.1) | ✅ |

### Key Deliverables
- ✅ Go project scaffold with Cobra CLI
- ✅ Config system (JSON persistence)
- ✅ Wallhaven API adapter with pagination
- ✅ Concurrent download manager (5 parallel)
- ✅ Perceptual hash deduplication (pHash + SQLite)
- ✅ Organization modes (source/date/tags)
- ✅ Cross-platform builds (macOS/Linux/Windows)
- ✅ **v1.1:** Progress bar with schollz/progressbar
- ✅ **v1.1:** Reddit source adapter
- ✅ **v1.1:** Sorting by popularity (top, favorites, views)
- ✅ **v1.1:** Time period filtering (day, week, month, year)

### Metrics
- Binary size: 11MB (target: <20MB) ✅
- Memory: <10MB at idle ✅
- Cross-platform: 5 targets ✅
- Smoke tests: 10/10 passed ✅

---

## 🚧 M002: Desktop Integration (v1.2)

**Goal:** Transform the CLI into a complete wallpaper management solution with auto-set capabilities and interactive TUI.

### Planned Slices

| ID | Title | Goal | Risk | Status | Depends On |
|----|-------|------|------|--------|------------|
| S01 | Cross-Platform Wallpaper Setting | `set` command for macOS, Linux, Windows | high | [◆] planned | - |
| S02 | TUI with Bubble Tea | Interactive wallpaper browser with thumbnails | medium | [ ] todo | S01 |
| S03 | Fuzzy Search & Integration | Fuzzy finder + set command integration | low | [ ] todo | S02 |

### Phase 01: Cross-Platform Wallpaper Setting (S01)

**Phase Status:** Planned — 3 plans ready for execution

| Plan | Objective | Wave | Files |
|------|-----------|------|-------|
| [01-01](./phases/01-cross-platform-wallpaper-setting/01-01-PLAN.md) | Platform detection + macOS/Linux backends | 1 | `internal/platform/*.go` |
| [01-02](./phases/01-cross-platform-wallpaper-setting/01-02-PLAN.md) | Windows backend + CLI set command | 2 | `internal/platform/windows.go`, `cmd/set.go` |
| [01-03](./phases/01-cross-platform-wallpaper-setting/01-03-PLAN.md) | Config persistence + comprehensive tests | 3 | `internal/config/config.go`, `*_test.go` |

**Plan Dependencies:**
```
01-01 (Platform Base) ──► 01-02 (CLI + Windows) ──► 01-03 (Config + Tests)
```

**Artifacts:**
- `01-CONTEXT.md` — Implementation decisions (AppleScript, DE detection, config persistence)
- `01-RESEARCH.md` — Platform-specific implementation research
- `01-VALIDATION.md` — Test strategy and verification checkpoints

### Planned Features
1. **Auto-Set Wallpaper**
   - `wallpaper-cli set <path>` - Set specific wallpaper
   - `wallpaper-cli set --random` - Set random from collection
   - `wallpaper-cli set --latest` - Set most recent download
   - Cross-platform: macOS, Linux (GNOME/KDE/XFCE), Windows

2. **TUI Browser**
   - `wallpaper-cli browse` - Interactive terminal UI
   - Thumbnail previews
   - Arrow key navigation
   - Real-time fuzzy search
   - Enter to set wallpaper

3. **Libraries**
   - Bubble Tea for TUI framework
   - Lipgloss for styling
   - sahilm/fuzzy for search

### Estimated: 6-8 hours

---

## 📋 Future Milestones (Backlog)

| ID | Title | Version | Priority |
|----|-------|---------|----------|
| M003 | Additional Sources | v1.3 | Low |
| M004 | AI Auto-Tagging | v1.4 | Low |
| M005 | Multi-Monitor Support | v1.5 | Medium |
| M006 | Wallpaper Rotation | v1.5 | Low |
| M007 | Web UI | v2.0 | Very Low |

---

## Navigation

### Current Milestone (M002)
- [M002 ROADMAP](./milestones/M002/M002-ROADMAP.md)
- [S01: Cross-Platform Wallpaper Setting](./milestones/M002/slices/S01/S01-PLAN.md)
- [S02: TUI with Bubble Tea](./milestones/M002/slices/S02/S02-PLAN.md)
- [S03: Fuzzy Search & Integration](./milestones/M002/slices/S03/S03-PLAN.md)

### Completed Milestones
- [M001 ROADMAP](./milestones/M001/M001-ROADMAP.md)
- [M001 SUMMARY](./milestones/M001/M001-SUMMARY.md)

### Requirements
- [Requirements](./REQUIREMENTS.md) (from SPEC.md)

---

*Roadmap updated: v1.0/v1.1 verified complete, v1.2 planned*
