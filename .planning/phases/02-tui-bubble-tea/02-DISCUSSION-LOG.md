# Phase 02: TUI with Bubble Tea - Discussion Log

> **Audit trail only.** Decisions captured in CONTEXT.md — this log preserves the analysis.

**Date:** 2026-04-04
**Phase:** 02-tui-bubble-tea
**Mode:** Research + Discussion

---

## Research Conducted

### Terminal Image Libraries Evaluated
| Library | Stars | Protocols | Decision |
|---------|-------|-----------|----------|
| blacktop/go-termimg | 52 | Kitty, iTerm2, SIXEL, half-blocks | **Selected** |
| srlehn/termimg | 30+ | Multiple | Alternative option |
| dolmen-go/kittyimg | 37 | Kitty only | Too limited |
| BourgeoisBear/rasterm | 108 | iTerm2, Kitty, SIXEL | Good alternative |

### Bubble Tea Components Researched
- **List** (`charmbracelet/bubbles/list`) — Built-in virtual scrolling ✅
- **Viewport** — For scrollable content areas
- **Image support** — Issue #163 (still open since 2021) — requires external library

### Performance Findings
- Bubble Tea list handles 1000+ items with virtual scrolling
- Issue #810 mentions paginator performance considerations
- Thumbnail caching critical for large collections

---

## Decisions Presented

### Decision 1: Thumbnail Rendering
| Option | User Choice |
|--------|-------------|
| High-quality images with terminal detection | ✅ **SELECTED** |
| ASCII art thumbnails (universal) | |
| No thumbnails - metadata only | |

**Rationale:** User wants actual image previews, willing to handle terminal detection complexity.

### Decision 2: TUI Layout
| Option | User Choice |
|--------|-------------|
| List view (like gh cli) | ✅ **START WITH THIS** |
| Split pane (like lazygit) | May explore later |
| Minimal list + status bar | |

**Rationale:** Start simple, evaluate before investing in split pane complexity.

### Decision 3: Fuzzy Search
| Option | User Choice |
|--------|-------------|
| Include fuzzy search now | |
| Arrow keys only - fuzzy in S03 | ✅ **SELECTED** |

**Rationale:** Defer to maintain focus, follow roadmap phases.

### Decision 4: macOS WallpaperEngine Hint
| Option | User Choice |
|--------|-------------|
| Include macOS WallpaperEngine hint | ✅ **SELECTED** |
| Skip hint - core TUI only | |

**Rationale:** User wants the integration feature from T09.

---

## Final Decisions Summary

| Area | Decision |
|------|----------|
| **Images** | High-quality with terminal detection (go-termimg library) |
| **Layout** | List view initially, split pane potential future |
| **Navigation** | Arrow keys + Enter, no fuzzy search |
| **macOS hint** | Include WallpaperEngine integration |
| **Performance** | Virtual scrolling, lazy thumbnail loading |
| **Caching** | 256x256 thumbnails in ~/.cache/wallpaper-cli/thumbs/ |

---

## OpenCode's Discretion Areas

User deferred to OpenCode for:
- Color scheme and styling (use Lipgloss defaults)
- Thumbnail generation concurrency
- List item height
- Help overlay formatting
- Cache expiration policy

---

*Discussion complete. Ready for planning.*
