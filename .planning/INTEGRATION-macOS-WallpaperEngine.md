# Integration Plan: wallpaper-cli-tool ↔ macOS WallpaperEngine

**Date:** 2026-04-04  
**Status:** Planning Phase  
**Scope:** Bi-directional integration between CLI downloader and native macOS app

---

## Executive Summary

The **wallpaper-cli-tool** (Go CLI) and **macOS WallpaperEngine** (Swift app) are complementary projects that can form a complete wallpaper ecosystem:

- **CLI Tool** = Content acquisition (download from Wallhaven/Reddit, deduplicate, organize)
- **macOS App** = Content rendering (live wallpapers: video, web, scene, image)

**Primary Integration Strategy:** CLI downloads → App displays via LocalFolderContentSource

---

## Project Comparison

| Aspect | wallpaper-cli-tool | macOS WallpaperEngine |
|--------|-------------------|----------------------|
| **Language** | Go | Swift |
| **Platform** | Cross-platform (macOS/Linux/Win) | macOS-only |
| **Primary Function** | Download & organize | Render & display |
| **Wallpaper Types** | Static images (JPG/PNG/WebP) | Video, Web, Scene, Image |
| **Sources** | Wallhaven API, Reddit JSON | Steam Workshop, Local Folders |
| **UI** | CLI + TUI (planned) | Native macOS GUI |
| **Deduplication** | pHash + SQLite | None (relies on file paths) |
| **Live Wallpapers** | No | Yes (video/web/scene) |

---

## Integration Opportunities (Ranked by Value)

### 1. CLI → App: Content Source Integration (HIGH VALUE)

**Concept:** CLI downloads wallpapers to a folder, macOS app automatically picks them up as a content source.

**How It Works:**
1. CLI downloads wallpapers to `~/Pictures/wallpapers/wallhaven/` and `~/Pictures/wallpapers/reddit/`
2. macOS app adds these as `LocalFolderContentSource` instances
3. App scans and displays them alongside Workshop content
4. User sees unified library in browser UI

**Value Proposition:**
- CLI's powerful download capabilities (concurrent, filtered, deduplicated)
- App's superior rendering (live wallpapers, video, web, scene)
- Unified browsing experience
- macOS users get best of both worlds

**Implementation Complexity:** LOW
- macOS app already supports LocalFolderContentSource
- Just needs to auto-add CLI output folders on launch
- Or expose "Add Folder" UI that points at CLI output

---

### 2. Shared Library: SQLite Sync (MEDIUM VALUE)

**Concept:** Both tools share a SQLite database for wallpaper metadata.

**How It Works:**
1. CLI writes download metadata to shared DB
2. App reads from shared DB for rich metadata display
3. App can show: tags, source URL, download date, resolution
4. Deduplication works across both tools

**Value Proposition:**
- Rich metadata in app (not just filenames)
- Cross-tool deduplication
- Download history visible in app
- Consistent organization

**Implementation Complexity:** MEDIUM
- Need to define shared schema
- macOS app currently doesn't use SQLite (only file scanning)
- Would require new Data layer in Swift app
- File-based DB path: `~/.local/share/wallpaper-cli/wallpapers.db`

---

### 3. App → CLI: Download Delegation (MEDIUM VALUE)

**Concept:** macOS app has "Get More Wallpapers" button that invokes CLI.

**How It Works:**
1. User clicks "Get More" in macOS app
2. App launches CLI subprocess with parameters
3. CLI downloads to shared folder
4. App rescans folder and shows new items

**Value Proposition:**
- One-click content acquisition from native UI
- No terminal required for casual users
- Leverages CLI's powerful filtering

**Implementation Complexity:** MEDIUM
- Need to bundle CLI binary with app
- Or expect CLI installed separately
- IPC coordination (wait for download complete)
- UI for CLI output/errors

---

### 4. App → CLI: Set Wallpaper via App (LOW VALUE)

**Concept:** CLI's `set` command activates wallpaper through macOS app instead of osascript.

**How It Works:**
1. CLI has `--via-app` flag
2. Sends XPC/AppleEvent to running macOS app
3. App activates specified wallpaper
4. Gets live wallpaper benefits (video/web/scene)

**Value Proposition:**
- CLI users can trigger live wallpapers
- Consistent rendering through app

**Implementation Complexity:** HIGH
- Requires XPC service or AppleEvent handling
- App needs to expose control API
- Overlap with planned v1.2 TUI features

---

## Recommended Integration: Approach #1 (Content Source)

### Phase 1: Auto-Discovery (Minimal Change)

**macOS App Changes:**
```swift
// In AppDelegate.restoreLocalFolders() or new method
func addCLIFolderSources(to library: WallpaperLibrary) {
    let cliOutputPath = FileManager.default.homeDirectoryForCurrentUser
        .appendingPathComponent("Pictures/wallpapers")
    
    // Add wallhaven subfolder if exists
    let wallhavenPath = cliOutputPath.appendingPathComponent("wallhaven")
    if FileManager.default.fileExists(atPath: wallhavenPath.path) {
        library.addSource(LocalFolderContentSource(folderURL: wallhavenPath))
    }
    
    // Add reddit subfolder if exists
    let redditPath = cliOutputPath.appendingPathComponent("reddit")
    if FileManager.default.fileExists(atPath: redditPath.path) {
        library.addSource(LocalFolderContentSource(folderURL: redditPath))
    }
}
```

**CLI Changes:** NONE

**User Workflow:**
1. User runs CLI to download: `./wallpaper-cli fetch --limit 10`
2. User launches macOS app (or app is already running)
3. App auto-discovers CLI downloads
4. User sees new wallpapers in browser alongside Workshop content

---

### Phase 2: Enhanced Metadata (Future)

**If Shared Library (#2) is implemented:**
- App can display: source URL, tags, resolution, aspect ratio
- Better search/filtering in app UI
- Deduplication aware of CLI downloads

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         USER WORKFLOW                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   ┌──────────────────┐         ┌──────────────────┐                    │
│   │   Terminal       │         │   macOS Desktop  │                    │
│   │                  │         │                  │                    │
│   │  $ wallpaper-cli │         │  ┌────────────┐  │                    │
│   │    fetch --top   │────────►│  │ Menu Bar   │  │                    │
│   │    --limit 20    │         │  │   App      │  │                    │
│   └──────────────────┘         │  └────────────┘  │                    │
│            │                   │       │          │                    │
│            ▼                   │       ▼          │                    │
│   ┌──────────────────┐         │  ┌────────────┐  │                    │
│   │  ~/Pictures/     │         │  │  Browser   │◄─┼────┐              │
│   │  wallpapers/     │────────►│  │   Window   │  │    │              │
│   │                  │         │  └────────────┘  │    │              │
│   │  wallhaven/     │         │       │          │    │              │
│   │  reddit/        │         │       ▼          │    │              │
│   └──────────────────┘         │  ┌────────────┐  │    │              │
│                                 │  │  Desktop   │  │    │              │
│                                 │  │  Window    │◄─┼────┘              │
│                                 │  │ (live wp)  │  │                   │
│                                 │  └────────────┘  │                   │
│                                 │                  │                   │
│                                 └──────────────────┘                   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                     DATA FLOW (Content Source)                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   ┌──────────────────┐         ┌──────────────────┐                    │
│   │  wallpaper-cli   │         │  WallpaperEngine │                    │
│   │    (Go CLI)      │         │    (Swift App)   │                    │
│   │                  │         │                  │                    │
│   │ • Wallhaven API  │         │ • LocalFolder    │◄─────────────────┤
│   │ • Reddit JSON    │         │   ContentSource  │                  │
│   │ • pHash dedup    │         │ • WorkshopSource │                  │
│   │ • Concurrent DL  │         │ • Desktop render │                  │
│   │ • SQLite DB      │         │ • Video/Web/Scene│                  │
│   └──────────────────┘         └──────────────────┘                  │
│            │                              ▲                            │
│            │         Shared Folder         │                            │
│            └──────────────────────────────┘                            │
│                      ~/Pictures/wallpapers/                             │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Implementation Plan

### M002-INT: Integration Slice for macOS App

**Goal:** Auto-discover and display wallpapers downloaded by CLI tool

**Success Criteria:**
- [ ] macOS app auto-adds CLI output folder as content source
- [ ] CLI wallpapers appear in browser alongside Workshop items
- [ ] No manual configuration required
- [ ] Works with existing folder structure (wallhaven/, reddit/)

**Tasks:**

| ID | Title | Est. | File |
|----|-------|------|------|
| T01 | Detect CLI output folder | 30m | AppDelegate.swift |
| T02 | Auto-add as LocalFolderContentSource | 30m | AppDelegate.swift |
| T03 | Handle subfolder scanning | 45m | LocalFolderContentSource.swift |
| T04 | Visual indicator for CLI source | 30m | BrowserView.swift |
| T05 | Test integration | 30m | Manual testing |

**Total:** ~2.5 hours

---

## Benefits Summary

| For CLI Users | For macOS App Users |
|--------------|-------------------|
| Live wallpaper rendering | Powerful batch downloading |
| Video/web/scene support | Deduplication & organization |
| Native macOS UI | Cross-platform workflow |
| Unified library | Reddit + Wallhaven sources |

---

## Next Steps

1. **Approve approach** — Confirm Content Source integration is desired
2. **Create M002-INT slice** — Add to wallpaper-cli-tool v1.2 planning
3. **Implement in macOS app** — Add auto-discovery in AppDelegate
4. **Test workflow** — End-to-end: CLI download → App display
5. **Document** — Update both project READMEs with integration guide

---

## Appendix: Folder Structure

**CLI Output (existing):**
```
~/Pictures/wallpapers/
├── wallhaven/
│   ├── 01_abc_3840x2160.jpg
│   ├── 02_def_1920x1080.png
│   └── ...
└── reddit/
    ├── 01_ghi_7680x4320.jpg
    └── ...
```

**macOS App Content Sources (after integration):**
- Workshop (default): `~/Library/Application Support/Steam/...`
- CLI Wallhaven: `~/Pictures/wallpapers/wallhaven` ← NEW
- CLI Reddit: `~/Pictures/wallpapers/reddit` ← NEW
- User folders: (manually added)

---

*Integration plan created from analysis of both codebases*
