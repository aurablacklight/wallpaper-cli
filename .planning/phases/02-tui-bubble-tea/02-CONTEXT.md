# Phase 02: TUI with Bubble Tea - Context

**Gathered:** 2026-04-04
**Status:** Ready for planning

<domain>
## Phase Boundary

Interactive wallpaper browser with thumbnail preview using Bubble Tea TUI framework. Browse downloaded wallpapers in a terminal UI: arrow keys to navigate, Enter to select and set as wallpaper, 'q' to quit. Includes macOS WallpaperEngine integration hint. Fuzzy search is explicitly deferred to Phase 03.

**Goal:** Users can visually browse their wallpaper collection and set wallpapers interactively from the terminal.

**Out of Scope:**
- Fuzzy search (Phase 03)
- Split pane layout (potential future enhancement)
- Advanced filtering/sorting in TUI

</domain>

<decisions>
## Implementation Decisions

### Thumbnail Rendering Strategy
- **D-01:** Use high-quality image rendering with terminal capability detection
- **D-02:** Primary library: `blacktop/go-termimg` (supports Kitty, iTerm2, SIXEL, half-blocks)
- **D-03:** Detect terminal capability at runtime and use best available method:
  1. Kitty Graphics Protocol (highest quality)
  2. iTerm2 Inline Images (macOS specific, high quality)
  3. SIXEL (moderate quality, multiple terminals)
  4. Half-blocks/ASCII (universal fallback)
- **D-04:** Generate thumbnails at 256x256 resolution for performance
- **D-05:** Cache thumbnails in `~/.cache/wallpaper-cli/thumbs/`

### TUI Layout
- **D-06:** Start with **List view** (simple, familiar, efficient)
- **D-07:** Each item shows: thumbnail + filename + resolution + source
- **D-08:** Status bar at bottom with: current selection count, help hint
- **D-09:** Split pane layout may be explored later if list view is unsatisfying (not in this phase)

### Navigation & Interaction
- **D-10:** Arrow keys for navigation (up/down)
- **D-11:** Enter to select and immediately set wallpaper
- **D-12:** 'q' or Escape to quit
- **D-13:** '?' for help overlay
- **D-14:** No fuzzy search in this phase (deferred to Phase 03 per user decision)
- **D-15:** Mouse support optional (click to select)

### macOS WallpaperEngine Integration
- **D-16:** Include T09 hint feature
- **D-17:** Detect macOS: `runtime.GOOS == "darwin"`
- **D-18:** Check if `/Applications/WallpaperEngine.app` exists
- **D-19:** Show dismissible hint banner: "💡 Tip: Open WallpaperEngine.app for live wallpapers (press 'o')"
- **D-20:** 'o' keybinding opens WallpaperEngine.app via `open` command
- **D-21:** 'd' keybinding dismisses hint for session
- **D-22:** Persist dismiss preference? Not for MVP (session-only)

### Performance Strategy
- **D-23:** Use Bubble Tea's built-in list component with virtual scrolling
- **D-24:** Lazy-load thumbnails (generate on first view if not cached)
- **D-25:** Pre-generate thumbnails in background when TUI starts
- **D-26:** Target: Handle 1000+ wallpapers without lag
- **D-27:** Memory target: <50MB for TUI with large collection

### Image Processing
- **D-28:** Use standard Go image libraries (image/jpeg, image/png)
- **D-29:** Resize algorithm: Lanczos or bilinear (quality vs speed)
- **D-30:** Skip generating thumbnails for non-image files

### Error Handling in TUI
- **D-31:** Show errors in status bar (not modal dialogs)
- **D-32:** Gracefully handle: missing thumbnails, permission denied, invalid images
- **D-33:** Allow TUI to continue even if some images fail to load

### OpenCode's Discretion
- Exact color scheme and styling (use Lipgloss defaults initially)
- Thumbnail generation concurrency (suggest 4 parallel workers)
- List item height (suggest 3 lines: thumbnail + text)
- Help overlay content and formatting
- Cache expiration policy (suggest 30 days)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase Specification
- `.planning/milestones/M002/slices/S02/S02-PLAN.md` — Full task breakdown
- `.planning/milestones/M002/M002-ROADMAP.md` — Milestone context and boundary map

### Research Findings
- Bubble Tea GitHub issue #163 — Image display discussion
- `blacktop/go-termimg` — Terminal image library documentation
- `charmbracelet/bubbles/list` — Virtual scrolling list component

### Prior Phase Context
- `.planning/phases/01-cross-platform-wallpaper-setting/01-CONTEXT.md` — Platform setting interface
- `internal/platform/platform.go` — Setter interface to use from TUI

### Integration Spec
- `.planning/milestones/M002/INTEGRATION-macOS-WallpaperEngine.md` — S04 context for T09 hint

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **internal/platform/** — Use `platform.Get()` and `Setter` interface for setting wallpaper from TUI
- **internal/utils/image.go** — `FindWallpapers()`, `IsImageFile()` for discovering wallpapers
- **internal/config/config.go** — Config for output directory path, current wallpaper tracking
- **cmd/set.go** — Pattern for wallpaper setting with error handling

### Established Patterns
- **Cobra commands** — TUI will be a new command `wallpaper-cli browse`
- **Config persistence** — Follow `config.Load()` and `cfg.Save()` pattern
- **Error handling** — Return errors wrapped with context

### Integration Points
- **New package**: `internal/tui/` — Bubble Tea model, view, update logic
- **New package**: `internal/thumbs/` — Thumbnail generation and caching
- **New command**: `cmd/browse.go` — Cobra command to launch TUI
- **Integration with platform**: TUI calls `platform.Get()` and `setter.SetWallpaper()`

### Technology Stack
- **Bubble Tea** — Core TUI framework (Elm architecture: Model, Update, View)
- **Lipgloss** — Styling and layouts
- **Bubbles** — List component with virtual scrolling
- **go-termimg** — Terminal image rendering
- **Standard library**: image/jpeg, image/png for thumbnail generation

</code_context>

<specifics>
## Specific Ideas

### TUI Mockup (List View)
```
┌─────────────────────────────────────────────────────┐
│ Browse Wallpapers                     [? help] [q quit]│
├─────────────────────────────────────────────────────┤
│  [▓▓▓▓] 01_anime_sunset_4k.jpg    3840x2160  WH  │
│▶ [▓▓▓▓] 02_city_night_4k.jpg      3840x2160  RD  │
│  [▓▓▓▓] 03_forest_mist_4k.jpg     3840x2160  WH  │
│  [    ] 04_invalid_file.txt         -        -    │
│                                                     │
├─────────────────────────────────────────────────────┤
│ 2 of 47 selected | Enter to set | 💡 Press 'o' for WE│
└─────────────────────────────────────────────────────┘
```

### Thumbnail Cache Structure
```
~/.cache/wallpaper-cli/thumbs/
├── 256x256/
│   ├── a1b2c3d4.jpg  (SHA256 hash of original path)
│   ├── e5f6g7h8.png
│   └── ...
└── cache.json  (metadata: original path, modtime, size)
```

### Keybindings Reference (for help overlay)
```
Navigation:
  ↑/↓ or j/k     Navigate up/down
  Enter          Set selected wallpaper
  q/Esc          Quit

macOS Integration:
  o              Open WallpaperEngine.app
  d              Dismiss hint

General:
  ?              Toggle help
```

### Terminal Detection Logic
```go
func DetectImageMethod() ImageMethod {
    if os.Getenv("TERM") == "xterm-kitty" {
        return KittyGraphics
    }
    if os.Getenv("TERM_PROGRAM") == "iTerm.app" {
        return Iterm2Inline
    }
    if hasSixelSupport() {
        return Sixel
    }
    return HalfBlocks // Universal fallback
}
```

</specifics>

<deferred>
## Deferred Ideas

- **Fuzzy search** — Explicitly deferred to Phase 03 (per user decision)
- **Split pane layout** — May explore if list view is unsatisfying (not in this phase)
- **Advanced filtering** — By resolution, source, date (future phase)
- **Multi-select** — Select multiple wallpapers for batch operations
- **Preview pane** — Full image preview in separate panel
- **Sorting options** — Sort by name, date, size in TUI
- **Theme customization** — User-configurable colors
- **Mouse drag scrolling** — Advanced mouse interaction
- **Thumbnail size options** — User-configurable thumbnail dimensions

</deferred>

---

*Phase: 02-tui-bubble-tea*  
*Context gathered: 2026-04-04*  
*Research: Terminal image libraries evaluated, Bubble Tea components assessed*
