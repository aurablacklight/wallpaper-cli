# S04: macOS App Integration

**Goal:** Enable seamless integration between wallpaper-cli-tool and the macOS WallpaperEngine app, allowing CLI downloads to appear automatically in the native macOS browser.

**Success Criteria:**
- macOS app auto-discovers CLI-downloaded wallpapers
- Wallpapers appear in browser alongside Steam Workshop content
- Source label shows "Wallhaven" or "Reddit" in UI
- No manual configuration required
- Folder structure preserved (wallhaven/, reddit/)

**Status:** 📝 Partial — CLI-side work available, macOS app changes blocked externally

---

## Quick Status

| Component | Status | Location |
|-----------|--------|----------|
| Protocol Documentation | ✅ Complete | This repo |
| CLI Enhancements (list, export, json) | 📝 Ready to implement | This repo |
| macOS App Auto-Discovery | ⏳ Blocked | External repo |
| macOS App Source Labels | ⏳ Blocked | External repo |
| End-to-End Testing | ⏳ Pending | Both repos |

**For detailed analysis:** See [S04-REPOSITORY-BOUNDARY-ANALYSIS.md](./S04-REPOSITORY-BOUNDARY-ANALYSIS.md)

---

## Integration Closure

CLI downloads flow into macOS app via LocalFolderContentSource, creating unified wallpaper library.

## Observability Impact

Content source detection logged, scan results show CLI folder contents.

## Proof Level

L1 - Uses existing LocalFolderContentSource infrastructure

---

## Dependencies

- S01: Cross-platform Wallpaper Setting (foundational v1.2)
- **macOS WallpaperEngine app (external dependency — BLOCKS T02, T03, T04)**

## External Dependency Warning ⚠️

**CRITICAL:** Tasks T02, T03, and T04 require modifications to the **macOS WallpaperEngine app repository**, which is NOT in this codebase.

### What This Means:
- **CLI repository (this repo):** Can complete T01 + CLI-side enhancements
- **macOS app repository (external):** Required for T02, T03, T04

### Repository Boundary Analysis:
See [S04-REPOSITORY-BOUNDARY-ANALYSIS.md](./S04-REPOSITORY-BOUNDARY-ANALYSIS.md) for detailed breakdown of:
- What CAN be implemented in this repo (CLI enhancements)
- What is BLOCKED on external dependency
- Recommended implementation plan

## Revised Task Split

| Task | Repository | Status | Notes |
|------|------------|--------|-------|
| T01 | CLI | ✅ Complete | Protocol documented |
| **T01b** | **CLI** | **📝 Todo** | **Implement `list` command** |
| **T01c** | **CLI** | **📝 Todo** | **Add `--json` output** |
| **T01d** | **CLI** | **📝 Todo** | **Add `export` command** |
| T02 | macOS App | ⏳ External | Auto-discovery |
| T03 | macOS App | ⏳ External | Source labels |
| T04 | Both | ⏳ External | E2E testing pending T02/T03 |

---

## Risk

Low - Leverages existing app architecture

---

## Demo

```bash
# Terminal: Download wallpapers
./wallpaper-cli fetch --favorites --all-time --limit 10

# macOS App: Auto-discover
# Launch WallpaperEngine app
# Browser shows new "Wallhaven" source with 10 wallpapers
```

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched | Repository |
|----|-------|------|-------|--------|---------------|------------|
| T01 | Document integration protocol | 30m | INTEGRATION-*.md | Protocol doc | This file | ✅ CLI |
| T01b | Implement `list` command | 2h | cmd/list.go | Working list cmd | cmd/list.go | ✅ CLI |
| T01c | Add `--json` output flag | 2h | All cmd files | JSON output | cmd/*.go | ✅ CLI |
| T01d | Add `export` command | 3h | data/db.go | Export command | cmd/export.go | ✅ CLI |
| T02 | Add auto-discovery to macOS app | 45m | AppDelegate.swift | Auto-add source | **macOS App** | ❌ External |
| T03 | Source label enhancement | 30m | LocalFolderContentSource | Source name in UI | **macOS App** | ❌ External |
| T04 | End-to-end testing | 30m | Both projects | Test results | - | ⏳ Pending T02/T03 |

---

## Plan

### T01: Document Integration Protocol
**Estimate:** 30m

**Description:**
Create formal specification for how CLI and macOS app integrate.

**Integration Protocol:**
1. **Output Location:** CLI writes to `~/Pictures/wallpapers/`
2. **Subfolders:**
   - `~/Pictures/wallpapers/wallhaven/` — Wallhaven downloads
   - `~/Pictures/wallpapers/reddit/` — Reddit downloads
   - (Future: `~/Pictures/wallpapers/zerochan/`)
3. **File Naming:** `{index}_{id}_{resolution}.{ext}`
   - Example: `01_8g5dp1_3840x2160.jpg`
4. **Metadata:** Stored in CLI's SQLite DB at `~/.local/share/wallpaper-cli/wallpapers.db`
5. **App Discovery:** macOS app scans subfolders on launch, adds as LocalFolderContentSource

**Files Likely Touched:**
- This plan document
- INTEGRATION-macOS-WallpaperEngine.md

**Expected Output:**
- Clear protocol specification
- PR-ready implementation guide

**Verification:**
```bash
# Review document
cat .planning/INTEGRATION-macOS-WallpaperEngine.md
```

---

### T02: Add Auto-Discovery to macOS App
**Estimate:** 45m

**⚠️ EXTERNAL DEPENDENCY:** This task requires modifying the **macOS WallpaperEngine app repository**, not this CLI repository. See [S04-REPOSITORY-BOUNDARY-ANALYSIS.md](./S04-REPOSITORY-BOUNDARY-ANALYSIS.md).

**Description:**
Modify macOS app's AppDelegate to auto-add CLI folders on launch.

**Implementation:**
```swift
// In AppDelegate.swift, add to applicationDidFinishLaunching or restoreLocalFolders()

private func addCLIContentSources(to library: WallpaperLibrary) {
    let cliBasePath = FileManager.default.homeDirectoryForCurrentUser
        .appendingPathComponent("Pictures/wallpapers")
    
    // Define CLI subfolder sources
    let cliSources = [
        ("wallhaven", "Wallhaven"),
        ("reddit", "Reddit")
    ]
    
    for (folderName, displayName) in cliSources {
        let folderPath = cliBasePath.appendingPathComponent(folderName)
        guard FileManager.default.fileExists(atPath: folderPath.path) else {
            continue
        }
        
        // Create source with custom display name
        let source = LocalFolderContentSource(
            folderURL: folderPath,
            id: "cli-\(folderName)",
            displayName: displayName
        )
        library.addSource(source)
    }
}
```

**Note:** Requires modification to LocalFolderContentSource to accept custom displayName in init.

**Files Likely Touched:**
- macOS: WallpaperEngine/App/AppDelegate.swift
- macOS: WallpaperEngine/Library/LocalFolderContentSource.swift

**Expected Output:**
- PR submitted to macOS app repo
- Auto-discovery working

**Verification:**
```bash
# In macOS app project
swift build
# Test: Launch app with CLI downloads present
# Verify they appear in browser
```

---

### T03: Source Label Enhancement
**Estimate:** 30m

**⚠️ EXTERNAL DEPENDENCY:** This task requires modifying the **macOS WallpaperEngine app repository**, not this CLI repository. See [S04-REPOSITORY-BOUNDARY-ANALYSIS.md](./S04-REPOSITORY-BOUNDARY-ANALYSIS.md).

**Description:**
Ensure CLI sources show proper labels in browser UI.

**Current State:**
- LocalFolderContentSource uses folder name as displayName
- Shows "wallhaven" instead of "Wallhaven"
- No indication it's from CLI vs manually added

**Enhancement:**
```swift
// Option 1: Custom init with display name
init(folderURL: URL, id: String = UUID().uuidString, displayName: String? = nil) {
    self.id = id
    self.folderURL = folderURL
    self.displayName = displayName ?? folderURL.lastPathComponent
}

// Option 2: Metadata in source field
// For CLI sources, set source = "CLI: Wallhaven"
```

**UI Display:**
- Source filter shows "Wallhaven" (not "wallhaven")
- Tooltip or subtitle shows path

**Files Likely Touched:**
- macOS: WallpaperEngine/Library/LocalFolderContentSource.swift
- macOS: WallpaperEngine/UI/BrowserView.swift (if source display logic exists)

**Expected Output:**
- Clean source labels in UI
- Professional presentation

**Verification:**
```swift
// Test source label
let source = LocalFolderContentSource(folderURL: url, displayName: "Wallhaven")
assert(source.displayName == "Wallhaven")
```

---

### T04: End-to-End Testing
**Estimate:** 30m

**Description:**
Test complete integration workflow.

**Test Scenario:**
1. Clean slate: Remove existing CLI downloads
2. CLI: Download 5 Wallhaven wallpapers
3. CLI: Download 3 Reddit wallpapers
4. macOS App: Launch
5. Verify: 8 new wallpapers appear in browser
6. Verify: Source filter shows "Wallhaven" (5) and "Reddit" (3)
7. Verify: Clicking wallpaper shows preview
8. Verify: Setting wallpaper works

**Files Likely Touched:**
- Test notes only

**Expected Output:**
- All tests pass
- Integration verified

**Verification:**
```bash
# Terminal
rm -rf ~/Pictures/wallpapers
./wallpaper-cli fetch --limit 5 --output ~/Pictures/wallpapers
./wallpaper-cli fetch --source reddit --limit 3 --output ~/Pictures/wallpapers

# macOS App
# Launch, verify 8 wallpapers appear
```

---

## Post-Integration Workflow

**User Story:**
1. User discovers anime wallpaper on Reddit
2. Opens terminal: `wallpaper-cli fetch --source reddit --sort hot --limit 10`
3. Downloads complete, saved to `~/Pictures/wallpapers/reddit/`
4. User opens macOS WallpaperEngine app (or clicks menu bar icon)
5. Browser shows new "Reddit" source with 10 wallpapers
6. User clicks one, sees live preview
7. Double-click to set as desktop wallpaper
8. User enjoys live/rendered wallpaper with native performance

**Value:**
- CLI's download power + App's rendering quality
- Best of both worlds
- No manual file management

---

## Future Enhancements (Post-S04)

### Phase 2: Rich Metadata
- Share CLI's SQLite DB with macOS app
- Show tags, source URL, resolution in app UI
- Enable tag-based filtering in app

### Phase 3: Direct Integration
- macOS app bundles CLI binary
- "Download More" button in app triggers CLI
- Progress displayed in native UI

### Phase 4: Cross-Platform Sync
- iCloud/other sync for shared library
- Windows/Linux users use CLI
- macOS users get native app experience
- Shared deduplication across all platforms

---

*S04 integrates two complementary projects into unified wallpaper ecosystem*
