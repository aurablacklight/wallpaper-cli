# Phase 01: Cross-Platform Wallpaper Setting - Context

**Gathered:** 2026-04-04
**Status:** Ready for planning

<domain>
## Phase Boundary

Implement the `wallpaper-cli set` command with platform-specific backends for macOS, Linux, and Windows. Users can set specific wallpaper files, random wallpapers from their collection, or the most recent download. Cross-platform testing ensures it works on all target platforms.

</domain>

<decisions>
## Implementation Decisions

### macOS Implementation Approach
- **D-01:** Use AppleScript (`osascript`) for initial implementation — tell Finder to set desktop picture
- **D-02:** Keep native API (NSWorkspace via CGO) as future enhancement path if multi-monitor or styling control is needed
- **Rationale:** AppleScript requires no CGO, simpler builds, works on all macOS versions

### Linux Desktop Environment Support
- **D-03:** Support GNOME, KDE, and XFCE as primary desktop environments
- **D-04:** Implement fallback using `feh` or `nitrogen` for unsupported DEs
- **D-05:** Auto-detect DE via `$XDG_CURRENT_DESKTOP` environment variable

### Windows Implementation
- **D-06:** Use PowerShell approach initially: `Set-ItemProperty` on Registry + `rundll32.exe` refresh
- **D-07:** Evaluate Win32 API (`SystemParametersInfo`) as alternative if PowerShell has reliability issues

### Config Persistence Strategy
- **D-08:** Add `current_wallpaper` field to config.json storing absolute path to active wallpaper
- **D-09:** Add `wallpaper_history` array storing last N (suggest 10) set wallpapers with timestamps
- **D-10:** Enables future features: `set --previous`, `set --history`, displaying current wallpaper
- **Rationale:** Aligns with existing config pattern in codebase (config.go uses JSON persistence)

### Error Handling Strategy
- **D-11:** Fail fast with clear error message on any platform-specific failure
- **D-12:** Exit with non-zero status code for CLI conventions and scripting compatibility
- **D-13:** Validate image file exists and is supported format before attempting platform-specific set

### Command Interface
- **D-14:** `wallpaper-cli set <path>` — Set specific wallpaper file
- **D-15:** `wallpaper-cli set --random` — Pick random wallpaper from collection
- **D-16:** `wallpaper-cli set --latest` — Set most recently downloaded wallpaper
- **D-17:** `wallpaper-cli set --current` — Show currently set wallpaper path (reads from config)

### Multi-Monitor Handling (Initial)
- **D-18:** For v1.2 S01, set wallpaper on all displays using simplest available method per platform
- **D-19:** Per-display control deferred to future milestone (requires more complex platform APIs)

### OpenCode's Discretion
- Platform detection implementation details (runtime.GOOS, environment variables)
- Exact wallpaper styling options (fill, fit, stretch, tile, center) — use OS defaults initially
- Image format validation approach (extension check vs MIME type vs image decoding)
- History array size (suggest 10, but implementor can adjust based on config patterns)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase Specification
- `.planning/milestones/M002/slices/S01/S01-PLAN.md` — Full task breakdown and implementation notes

### Project Documentation
- `.planning/milestones/M002/M002-ROADMAP.md` — Milestone definition, success criteria, verification contracts
- `.planning/ROADMAP.md` — Project roadmap with milestone context

### Integration Specification
- `.planning/INTEGRATION-macOS-WallpaperEngine.md` — S04 macOS app integration context

### Existing Code Patterns
- `internal/config/config.go` — Config structure and persistence pattern to follow
- `cmd/config.go` — Command implementation pattern using Cobra
- `cmd/root.go` — CLI root command setup with viper integration

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **Config system** (`internal/config/config.go`): Has Load/Save pattern with JSON marshaling. Add `CurrentWallpaper` and `WallpaperHistory` fields following existing pattern.
- **Cobra command structure** (`cmd/`): Each command has its own file with `RunE` error handling. Create `cmd/set.go` following `cmd/config.go` pattern.
- **Path utilities** (`internal/utils/path.go`): May have helpers for wallpaper directory resolution.

### Established Patterns
- **Error handling**: Commands return errors from `RunE` and let Cobra print them. Use `fmt.Errorf()` with wrapping.
- **Config persistence**: JSON files at `~/.config/wallpaper-cli/config.json`. Config struct uses json tags.
- **Platform detection**: No existing pattern yet — this is new territory for the codebase.

### Integration Points
- **New package**: `internal/platform/` — Create platform detection and wallpaper setting abstractions
- **New command**: `cmd/set.go` — Add `set` subcommand to root, register in `init()`
- **Config extension**: Add fields to `Config` struct in `internal/config/config.go`

</code_context>

<specifics>
## Specific Ideas

- macOS AppleScript approach: `tell application "Finder" to set desktop picture to POSIX file "/path/to/image.jpg"`
- Linux GNOME: `gsettings set org.gnome.desktop.background picture-uri file:///path/to/image.jpg`
- Linux KDE: Use qdbus to call PlasmaShell script
- Linux XFCE: Use xfconf-query
- Windows: PowerShell `Set-ItemProperty -Path "HKCU:\Control Panel\Desktop" -Name Wallpaper -Value "path"`
- History format: Array of `{path, timestamp, source}` objects where source indicates how it was set (manual, random, latest)

</specifics>

<deferred>
## Deferred Ideas

- Per-display wallpaper control — Future milestone (requires platform-specific multi-monitor APIs)
- Wallpaper styling options (fill/fit/stretch/tile/center) — Can be added per-platform later
- Native macOS API (NSWorkspace via CGO) — Enhancement after AppleScript baseline works
- Scheduled wallpaper rotation — Separate phase M006 per roadmap
- Multi-monitor support — Separate phase M005 per roadmap

</deferred>

---

*Phase: 01-cross-platform-wallpaper-setting*
*Context gathered: 2026-04-04*
