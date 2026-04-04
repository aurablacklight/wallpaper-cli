# S03: Thumbnail Integration & Fuzzy Search

**Goal:** Add terminal thumbnail rendering to TUI list items, plus fuzzy search capability for wallpaper filtering.

**Success Criteria:**
- ✅ Thumbnails display inline in TUI list (not separate view)
- ✅ Works on Kitty, iTerm2, SIXEL terminals
- ✅ Graceful fallback to ASCII art on unsupported terminals
- ✅ `/` activates fuzzy search in TUI
- ✅ Search filters wallpapers in real-time (<100ms)
- ✅ Selecting a wallpaper sets it immediately
- ✅ Search can be cancelled with ESC

---

## TOP PRIORITY: Terminal Thumbnail Integration

This is the **#1 priority** for Phase 3, requested by user after Phase 2 UAT.

---

## Integration Closure

Fuzzy search works in TUI with thumbnail previews, selection triggers wallpaper set.

## Observability Impact

Search query metrics, selection tracking, thumbnail generation metrics.

## Proof Level

L1 - Terminal image protocols are well-documented
L1 - Standard fuzzy matching algorithm

---

## Tasks

| ID | Title | Est. | Input | Output | Priority |
|----|-------|------|-------|--------|----------|
| T00 | **Terminal Thumbnail Integration** ⭐ | 2h | internal/tui/model.go | Custom list delegate with images | **TOP** |
| T01 | Integrate fuzzy search library | 30m | - | internal/tui/fuzzy.go | Normal |
| T02 | Search mode UI | 30m | fuzzy.go | Search bar component | Normal |
| T03 | Real-time filtering | 45m | search.go | Filtered list updates | Normal |
| T04 | Selection to set integration | 30m | S01 set code | Set on selection | Normal |
| T05 | Keyboard shortcuts refinement | 30m | All code | Polish keybindings | Normal |
| T06 | Integration testing | 30m | All code | Test results | Normal |

---

## Dependencies

- S01: Cross-Platform Wallpaper Setting (provides set functionality)
- S02: TUI with Bubble Tea (provides UI framework)
- go-termimg library (terminal image rendering)

---

## Risk

**Thumbnail Integration:** Medium - Complex Bubble Tea list customization required
**Fuzzy Search:** Low - Fuzzy matching is well-understood, libraries available

---

## Demo

```bash
./wallpaper-cli browse
# See thumbnails inline in list (not separate view)
# Press / to search
# Type "anime 4k" -> filters list
# Press Enter on selection -> wallpaper set
```

---

## Plans

### T00: Terminal Thumbnail Integration ⭐ TOP PRIORITY
**Estimate:** 2h

**Description:**
Display wallpaper thumbnails **inline in the list**, not as a separate view. This requires creating a custom Bubble Tea list item delegate that renders images alongside text.

**Terminal Support Matrix:**
| Protocol | Quality | Terminals | Fallback |
|----------|---------|-----------|----------|
| Kitty Graphics | ⭐⭐⭐ Excellent | Kitty, Ghostty | Auto |
| iTerm2 Inline | ⭐⭐⭐ Excellent | iTerm2 | Auto |
| SIXEL | ⭐⭐ Good | mlterm, yaft, xterm | Auto |
| Half-blocks | ⭐ Low | All terminals | Universal |

**Implementation Approach:**
1. Create custom `list.ItemDelegate` that renders thumbnails
2. Use `blacktop/go-termimg` for image rendering
3. Detect terminal capability at runtime
4. Generate thumbnails on-demand with caching
5. Layout: `[Thumbnail] Filename | Source`

**UI Mockup:**
```
┌────────────────────────────────────────┐
│ Browse Wallpapers (10/47)    [? help]  │
├────────────────────────────────────────┤
│                                        │
│ ┌────┐ 01_anime_sunset.jpg  WH 3840x2160│
│ │🖼️ │                                    │
│ └────┘ 02_city_night.png    RD 3840x2160│
│                                        │
│ ┌────┐▶03_forest_mist.jpg   WH 3840x2160│  <- Selected
│ │🖼️ │                                    │
│ └────┘ ...                              │
│                                        │
├────────────────────────────────────────┤
│ 📷 03_forest_mist.jpg | Enter: set    │
└────────────────────────────────────────┘
```

**Technical Details:**
- Thumbnail size: 64x64 or 128x128 (fits in list item height)
- Position: Left side of list item
- Text: Filename + Source to the right
- Selected item: Highlight thumbnail + text
- Non-selected: Dim thumbnail + text

**Files Likely Touched:**
- `internal/tui/delegate.go` - Custom item delegate (NEW)
- `internal/tui/model.go` - Integrate custom delegate
- `internal/thumbs/thumbs.go` - Ensure 128x128 size available
- `internal/tui/image.go` - Terminal image rendering wrapper

**Expected Output:**
- Thumbnails render inline in list
- Works on supported terminals
- Graceful fallback on others
- No separate thumbnail view needed

**Verification:**
```bash
./wallpaper-cli browse
# Should see thumbnails left of each filename
# Navigate up/down - selection highlight moves
# Thumbnails should be ~128x128
# On unsupported terminals, shows placeholder or ASCII
```

---

### T01: Integrate Fuzzy Search Library
**Estimate:** 30m

**Description:**
Add fuzzy string matching library for search functionality.

**Library Options:**
1. **sahilm/fuzzy** - Popular, simple API
2. **ktr0731/go-fuzzyfinder** - Full TUI component (might conflict with Bubble Tea)
3. **Custom implementation** - Levenshtein distance

**Decision:** Use sahilm/fuzzy for matching logic, integrate into Bubble Tea TUI.

**Steps:**
1. `go get github.com/sahilm/fuzzy`
2. Create `internal/tui/fuzzy.go`
3. Wrap fuzzy matcher for wallpaper items
4. Match against: filename, tags, source, resolution

**Files Likely Touched:**
- internal/tui/fuzzy.go
- go.mod

**Expected Output:**
- Fuzzy matching available
- Can match patterns against wallpapers

**Verification:**
```go
// Unit test
matches := fuzzy.Find("anime", items)
// matches should contain anime-tagged wallpapers
```

---

### T02: Search Mode UI
**Estimate:** 30m

**Description:**
Create the search input component in TUI.

**Steps:**
1. Create `internal/tui/search.go`
2. Add SearchMode state to Model
3. Search bar at top of screen
4. Text input component (Bubble Tea has this)
5. Show "Search: [input]" with cursor

**UI Layout:**
```
┌─────────────────────────────┐
│ Search: anime 4k _          │  <- Input line
├─────────────────────────────┤
│ [1] anime_sunset_4k.jpg     │  <- Filtered results
│ [2] anime_night_4k.png      │
│ [3] ...                     │
├─────────────────────────────┤
│ ↑↓ navigate | Enter set |   │  <- Help
│ ESC cancel | q quit         │
└─────────────────────────────┘
```

**Files Likely Touched:**
- internal/tui/search.go
- internal/tui/model.go (add search state)

**Expected Output:**
- Press `/` to enter search mode
- Search bar appears
- ESC exits search mode

**Verification:**
```bash
./wallpaper-cli browse
# Press /
# Type "anime"
# Press ESC to cancel
```

---

### T03: Real-Time Filtering
**Estimate:** 45m

**Description:**
Filter wallpaper list as user types search query.

**Steps:**
1. On each keystroke in search mode:
   - Run fuzzy match on all wallpapers
   - Update visible list with matches
   - Preserve original list (for clearing search)
2. Sort matches by score (best first)
3. Highlight matched portions
4. Handle no matches state

**Performance Targets:**
- <100ms for 1000 wallpapers
- Use debouncing (wait for typing pause)

**Files Likely Touched:**
- internal/tui/update.go
- internal/tui/fuzzy.go

**Expected Output:**
- Typing filters list in real-time
- Good performance
- Clear when no matches

**Verification:**
```bash
./wallpaper-cli browse
# Press /
# Type "night" -> list filters to night wallpapers
```

---

### T04: Selection to Set Integration
**Estimate:** 30m

**Description:**
Wire up Enter key to set the selected wallpaper.

**Steps:**
1. On Enter press:
   - Get selected wallpaper path
   - Call platform.SetWallpaper() from S01
   - Show success message
   - Optionally exit TUI or stay open
2. Handle set errors gracefully
3. Show "Setting..." spinner during operation

**Files Likely Touched:**
- internal/tui/set.go (new file)
- internal/tui/update.go

**Expected Output:**
- Enter sets wallpaper
- Success message shown
- Desktop background changes

**Verification:**
```bash
./wallpaper-cli browse
# Navigate to wallpaper
# Press Enter
# See "Wallpaper set!" message
# Desktop background changes
```

---

### T05: Keyboard Shortcuts Refinement
**Estimate:** 30m

**Description:**
Polish keybindings for intuitive UX.

**Final Key Map:**
| Key | Normal Mode | Search Mode |
|-----|-------------|-------------|
| ↑ / k | Previous item | - |
| ↓ / j | Next item | - |
| Enter | Set wallpaper | Apply search / Set first match |
| / | Enter search mode | - |
| ESC | Quit | Exit search mode |
| q | Quit | - |
| Ctrl+C | Force quit | Force quit |
| ? | Show help | - |

**Steps:**
1. Update KeyMap in keys.go
2. Handle all edge cases
3. Help screen with all shortcuts
4. Consistent behavior across modes

**Files Likely Touched:**
- internal/tui/keys.go
- internal/tui/view.go (help display)

**Expected Output:**
- Intuitive keybindings
- Help accessible
- No conflicts

**Verification:**
```bash
./wallpaper-cli browse
# Press ? for help
# Test each keybinding
```

---

### T06: Integration Testing
**Estimate:** 30m

**Description:**
End-to-end testing of complete v1.2 workflow.

**Test Scenarios:**
1. Launch browse, navigate, set wallpaper
2. Launch browse, search "anime", set filtered result
3. Launch browse, search (no matches), clear, set
4. Set via CLI arg: `wallpaper-cli set <path>`
5. Set random: `wallpaper-cli set --random`
6. Set latest: `wallpaper-cli set --latest`

**Files Likely Touched:**
- Test notes only

**Expected Output:**
- All scenarios pass
- No crashes
- Desktop wallpaper changes correctly

**Verification:**
```bash
# Full workflow test
./wallpaper-cli fetch --limit 5 --output ~/test-tui
./wallpaper-cli browse
# Navigate, search, set
# Verify desktop changed
```

---

## Slice Completion Criteria

- [ ] All tasks complete
- [ ] Fuzzy search <100ms for 1000 items
- [ ] Set on Enter works on all platforms
- [ ] Help screen shows all keybindings
- [ ] No flickering or rendering artifacts
- [ ] Integration tests pass
