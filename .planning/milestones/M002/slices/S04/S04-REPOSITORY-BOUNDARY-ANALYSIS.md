# S04 macOS App Integration: Repository Boundary Analysis

**Date:** 2026-04-04  
**Phase:** M002-S04 - macOS App Integration  
**Analyst:** GSD Research Agent  

---

## Executive Summary

The S04 "macOS App Integration" slice has a **fundamental architectural boundary**: the wallpaper-cli-tool repository (Go CLI) and the macOS WallpaperEngine app (Swift) are **separate codebases**. This analysis identifies what CAN be done in this repository to support integration vs. what is **blocked on external dependency** (macOS app changes).

**Bottom Line:**
- ~20% of S04 can be completed in this repo (CLI-side enhancements)
- ~80% requires changes to the external macOS app
- The planned tasks T02 and T03 from S04-PLAN.md **cannot be done in this repository**

---

## Current State Analysis

### CLI Already Does (No Changes Needed)

| Feature | Implementation | Location |
|---------|---------------|----------|
| Output to structured folders | `~/Pictures/wallpapers/{wallhaven,reddit}/` | `fetch.go:organizeBy` |
| File naming convention | `{index}_{id}_{resolution}.{ext}` | `generateWallhavenFilename()` |
| SQLite metadata storage | `~/.local/share/wallpaper-cli/wallpapers.db` | `data/db.go` |
| Extended file attributes | `xattr -w user.wallhaven_url` on macOS | `utils/metadata.go` |
| TUI with thumbnails | Bubble Tea + go-termimg | `internal/tui/` |
| Platform wallpaper setting | osascript for macOS | `platform/macos.go` |

### Folder Structure (Ready for App Integration)

```
~/Pictures/wallpapers/
├── wallhaven/
│   ├── 01_8g5dp1_3840x2160.jpg   # File naming: {rank}_{id}_{res}.{ext}
│   ├── 02_abc123_1920x1080.png
│   └── ...
└── reddit/
    ├── 01_xyz789_2560x1440.jpg
    └── ...
```

---

## Repository Boundary: What Can vs. Cannot Be Done

### ✅ CAN DO IN THIS REPO (CLI-Side Enhancements)

#### 1. Integration Protocol Documentation
**Status:** ✅ ALREADY COMPLETE  
**File:** `.planning/INTEGRATION-macOS-WallpaperEngine.md`

Documents:
- Folder structure convention
- File naming patterns
- SQLite schema for metadata sharing
- Integration architecture

---

#### 2. Metadata Export Command (`wallpaper-cli export`)
**Value:** MEDIUM — Enables rich metadata display in macOS app without SQLite sharing  
**Effort:** 2-3 hours  

**New Command:**
```bash
# Export metadata to JSON for macOS app consumption
wallpaper-cli export --format json --output ~/Pictures/wallpapers/metadata.json

# Export specific source
wallpaper-cli export --source wallhaven --since "7d"
```

**Output Format:**
```json
{
  "version": "1.0",
  "generated_at": "2026-04-04T10:30:00Z",
  "cli_version": "1.2.0",
  "wallpapers": [
    {
      "id": "8g5dp1",
      "source": "wallhaven",
      "local_path": "~/Pictures/wallpapers/wallhaven/01_8g5dp1_3840x2160.jpg",
      "url": "https://wallhaven.cc/w/8g5dp1",
      "resolution": "3840x2160",
      "aspect_ratio": "16:9",
      "tags": ["anime", "landscape", "sunset"],
      "downloaded_at": "2026-04-04T08:15:00Z",
      "file_size": 2847563
    }
  ]
}
```

**Why It Helps:** macOS app can read this JSON instead of implementing SQLite access in Swift.

---

#### 3. Change Notification (`wallpaper-cli watch`)
**Value:** HIGH — macOS app can refresh when new wallpapers arrive  
**Effort:** 4-6 hours  

**New Command:**
```bash
# Watch mode: notify macOS app of changes
wallpaper-cli watch --notify-app

# Or one-time notification after fetch
wallpaper-cli fetch --limit 10 --notify-app
```

**Implementation Options:**

**Option A: macOS Notification Center (easiest)**
```go
// Use osascript to post notification
exec.Command("osascript", "-e", `display notification "10 new wallpapers downloaded" with title "wallpaper-cli"`)
```

**Option B: XPC / NSDistributedNotificationCenter (better)**
```swift
// In macOS app: listen for this
NotificationCenter.default.post(
    name: NSNotification.Name("com.wallpaper-cli.newdownloads"),
    object: nil,
    userInfo: ["count": 10, "path": "~/Pictures/wallpapers"]
)
```

**Option C: Touch file trigger**
```bash
# CLI touches a sentinel file that macOS app watches
touch ~/Pictures/wallpapers/.last_updated
```

---

#### 4. JSON Output for All Commands
**Value:** MEDIUM — Enables scripting and integration  
**Effort:** 2-3 hours  

**Add `--json` flag:**
```bash
wallpaper-cli fetch --limit 5 --json
wallpaper-cli list --json
wallpaper-cli config list --json
```

**Example output:**
```json
{
  "success": true,
  "downloaded": 5,
  "skipped": 0,
  "failed": 0,
  "wallpapers": [...]
}
```

---

#### 5. Metadata Sidecar Files (Alternative to SQLite)
**Value:** MEDIUM — macOS app can read `.json` sidecar files without SQLite dependency  
**Effort:** 3-4 hours  

**Implementation:**
```
~/Pictures/wallpapers/wallhaven/
├── 01_8g5dp1_3840x2160.jpg
├── 01_8g5dp1_3840x2160.json    # Sidecar metadata
├── 02_abc123_1920x1080.png
└── 02_abc123_1920x1080.json
```

**Contents of sidecar:**
```json
{
  "source": "wallhaven",
  "source_id": "8g5dp1",
  "source_url": "https://wallhaven.cc/w/8g5dp1",
  "tags": ["anime", "sunset"],
  "downloaded_at": "2026-04-04T08:15:00Z",
  "resolution": "3840x2160",
  "file_size": 2847563,
  "hash": "a1b2c3d4..."
}
```

**Enable with flag:**
```bash
wallpaper-cli fetch --sidecar-metadata
```

---

#### 6. List Command Completion
**Status:** ⚠️ STUB EXISTS — Needs implementation  
**File:** `cmd/list.go` (currently returns `not yet implemented`)

**Implementation:**
```bash
# List with filters
wallpaper-cli list --source wallhaven --since 7d
wallpaper-cli list --resolution 4k
wallpaper-cli list --path-only  # For piping to other tools
```

---

### ❌ CANNOT DO IN THIS REPO (Blocked on External Dependency)

#### Blocked Task: T02 — Add Auto-Discovery to macOS App
**Status:** ❌ BLOCKED  
**Why:** Requires modifying `AppDelegate.swift` in macOS WallpaperEngine app  

From S04-PLAN.md:
```swift
// This code goes in macOS app's AppDelegate.swift
private func addCLIContentSources(to library: WallpaperLibrary) {
    // ... implementation
}
```

**What's needed in macOS app:**
- Modify `AppDelegate.swift` to call `addCLIContentSources()` on launch
- Modify `LocalFolderContentSource` to accept custom `displayName` parameter

---

#### Blocked Task: T03 — Source Label Enhancement
**Status:** ❌ BLOCKED  
**Why:** Requires modifying `LocalFolderContentSource.swift` in macOS app

From S04-PLAN.md:
```swift
// This code goes in macOS app's LocalFolderContentSource.swift
init(folderURL: URL, id: String = UUID().uuidString, displayName: String? = nil) {
    // ... implementation
}
```

**What's needed in macOS app:**
- Add optional `displayName` parameter to init
- UI to show "Wallhaven" instead of "wallhaven"

---

#### Blocked Task: T04 — End-to-End Testing
**Status:** ❌ BLOCKED (partially)  
**Why:** Requires working macOS app integration

Can do partial testing (CLI side), but full E2E requires macOS app changes.

---

## Recommended Implementation Plan

### Immediate Actions (In This Repo)

| Priority | Task | Effort | Value |
|----------|------|--------|-------|
| P1 | Implement `list` command | 2h | HIGH — Currently stubbed |
| P2 | Add `--json` global flag | 2h | MEDIUM — Enables integration |
| P3 | Add `export` command | 3h | MEDIUM — Metadata sharing |
| P4 | Add `--notify-app` flag to fetch | 2h | HIGH — Change notification |

### External Dependency Actions (macOS App Repo)

| Priority | Task | Location | PR Required |
|----------|------|----------|-------------|
| P1 | Auto-discover CLI folders | `AppDelegate.swift` | Yes |
| P2 | Custom displayName for LocalFolderContentSource | `LocalFolderContentSource.swift` | Yes |
| P3 | Listen for CLI notifications (optional) | `AppDelegate.swift` | Yes |

---

## S04 Plan Revision

### Original Plan (Current S04-PLAN.md)

| ID | Title | Status | Owner |
|----|-------|--------|-------|
| T01 | Document integration protocol | ✅ Complete | CLI repo |
| T02 | Add auto-discovery to macOS app | ❌ Blocked | macOS app repo |
| T03 | Source label enhancement | ❌ Blocked | macOS app repo |
| T04 | End-to-end testing | ❌ Blocked | Both repos |

### Revised Plan (Realistic)

| ID | Title | Status | Repository |
|----|-------|--------|------------|
| T01 | Document integration protocol | ✅ Complete | CLI repo |
| **T01b** | **Implement `list` command** | 📝 Todo | **CLI repo** |
| **T01c** | **Add `--json` output flag** | 📝 Todo | **CLI repo** |
| **T01d** | **Add `export` command** | 📝 Todo | **CLI repo** |
| T02 | Auto-discovery in macOS app | ⏳ External | macOS app repo |
| T03 | Source label enhancement | ⏳ External | macOS app repo |
| T04 | E2E testing | ⏳ External | Both repos |

---

## CLI Enhancements Detailed Spec

### 1. List Command (Complete Implementation)

**File:** `cmd/list.go` (replace stub)

```go
var (
    listSource    string
    listSince     string
    listJSON      bool
    listPathOnly  bool
)

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List downloaded wallpapers",
    Long:  `List and filter downloaded wallpapers from the database or filesystem.`,
    RunE:  runList,
}

func init() {
    listCmd.Flags().StringVar(&listSource, "source", "", "Filter by source (wallhaven, reddit)")
    listCmd.Flags().StringVar(&listSince, "since", "", "Show only files downloaded since (1d, 7d, 30d)")
    listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
    listCmd.Flags().BoolVar(&listPathOnly, "path-only", false, "Output only file paths (for piping)")
}

func runList(cmd *cobra.Command, args []string) error {
    // Load from SQLite DB or scan filesystem
    // Apply filters
    // Output as table, JSON, or paths
}
```

---

### 2. Export Command (New)

**File:** `cmd/export.go` (new file)

```go
var exportCmd = &cobra.Command{
    Use:   "export",
    Short: "Export wallpaper metadata",
    Long:  `Export wallpaper metadata to JSON for integration with other tools.`,
    RunE:  runExport,
}

func runExport(cmd *cobra.Command, args []string) error {
    // Read from SQLite DB
    // Filter by source, date, etc.
    // Write to JSON file or stdout
}
```

---

### 3. JSON Output Flag (Global)

**File:** `cmd/root.go` + all command files

Add `--json` as persistent flag, and helpers for JSON output.

---

### 4. Notify Flag for Fetch

**File:** `cmd/fetch.go`

```go
fetchCmd.Flags().Bool("notify-app", false, "Notify macOS app of new downloads")
```

---

## What to Tell Users

### For macOS Users Who Want Integration NOW

**Workaround until macOS app is updated:**

1. Download wallpapers: `./wallpaper-cli fetch --limit 10`
2. In macOS WallpaperEngine app:
   - Open Settings → Content Sources
   - Click "Add Folder"
   - Select `~/Pictures/wallpapers/wallhaven/` or `~/Pictures/wallpapers/reddit/`
3. Wallpapers appear in browser

**Limitation:** Source shows as "wallhaven" (lowercase) instead of "Wallhaven" — this requires macOS app update for custom display names.

---

## Status Update for M002

### M002 Progress (Revised)

| Slice | Status | Description |
|-------|--------|-------------|
| S01 | ✅ Complete | Cross-platform wallpaper setting |
| S02 | ✅ Complete | TUI with Bubble Tea |
| S03 | ✅ Complete | Thumbnail integration + fuzzy search |
| **S04** | 📝 **Revised** | CLI-side work possible; macOS app changes external |

**M002 Status:** 75% complete + CLI-side S04 work

**Completion Path:**
1. Complete CLI-side S04 enhancements (list, export, JSON)
2. Coordinate with macOS app maintainer for T02/T03 PRs
3. Mark S04 as "partially complete — external dependency tracked"

---

## Open Questions

1. **macOS App Repository Access:** Do we have write access to the macOS WallpaperEngine repo to submit PRs?

2. **Integration Priority:** Should we implement sidecar metadata (P5) or focus on `export` command (P3)? Sidecars are better for ongoing sync; export is simpler.

3. **Notification Method:** Which notification approach (A/B/C from section 3) should be implemented?

4. **SQLite Sharing:** Should the CLI provide a read-only SQLite connection helper, or stick to JSON export for macOS app integration?

---

## Next Steps Recommendation

### Immediate (This Week)
1. ✅ Confirm this analysis with project stakeholders
2. 📝 Update S04-PLAN.md to reflect external dependency
3. 📝 Create implementation tickets for CLI-side enhancements

### Short-term (Next 2 Weeks)
1. Implement `list` command (P1)
2. Add `--json` output (P2)
3. Create PR for macOS app auto-discovery (coordinate with external repo)

### Medium-term (Next Month)
1. Implement `export` command (P3)
2. Add notification support (P4)
3. Complete E2E testing once macOS app changes land

---

*Analysis complete. Deliverable: This document serves as the research foundation for planning realistic S04 work in the CLI repository.*
