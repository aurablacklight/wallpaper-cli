# Phase 03: Thumbnail Integration & Fuzzy Search - Context

**Gathered:** 2026-04-04
**Status:** Ready for planning
**Priority:** ⭐ Thumbnail integration is TOP PRIORITY

<domain>
## Phase Boundary

Display wallpaper thumbnails **inline in the TUI list** (not as separate view), enabling users to visually browse their collection. Additionally, add fuzzy search with `/` key to filter wallpapers by filename, source, or tags.

**What's New in Phase 3:**
- Thumbnails render left of each list item
- Terminal auto-detection for best image protocol
- Fuzzy search for quick filtering

**Out of Scope:**
- Separate thumbnail preview pane (not inline)
- Full-size image preview
- External image viewer integration

</domain>

<decisions>
## Implementation Decisions

### ⭐ TOP PRIORITY: Terminal Thumbnail Integration

**D-00:** Thumbnails must display **inline in the list**, not as separate view
- Left side of each list item: `[Thumbnail] Filename Source`
- Size: 128x128 pixels (configurable)
- Rationale: User explicitly requested this as priority after Phase 2 UAT

**D-01:** Use custom Bubble Tea `list.ItemDelegate` for thumbnail rendering
- Standard Bubble Tea list doesn't support images
- Must create custom delegate that combines image + text
- Render method generates image string alongside text

**D-02:** Terminal protocol priority (auto-detect at runtime):
1. Kitty Graphics Protocol (best quality)
2. iTerm2 Inline Images (macOS best)
3. SIXEL (good, widely supported)
4. Half-blocks/ASCII (universal fallback)

**D-03:** Use `blacktop/go-termimg` library for all terminal image rendering
- Already added as dependency in Phase 2
- Supports all protocols with auto-detection
- Handles encoding and terminal escape sequences

**D-04:** Thumbnail size: 128x128 for list display
- Fits within 3-line list item height
- Update `internal/thumbs/` to generate 128x128 size
- Cache alongside 256x256 thumbnails

**D-05:** Inline layout structure:
```
┌──────────────────────────────────────────────┐
│ [128x128] filename.jpg         Source: WH    │
│  thumb    resolution: 3840x2160               │
│           tags: anime, landscape             │
└──────────────────────────────────────────────┘
```

**D-06:** Graceful degradation:
- Detect terminal capability at startup
- Show placeholder `[📷]` if protocol unsupported
- Never break TUI if image rendering fails
- Error logged, not displayed to user

### Fuzzy Search (Secondary Priority)

**D-07:** Use `sahilm/fuzzy` library for matching
- Popular, well-tested Go fuzzy matching
- Match against: filename, source, resolution, tags

**D-08:** Activate with `/` key
- Press `/` to enter search mode
- Type to filter in real-time
- ESC exits search mode
- Real-time filtering with debouncing (<100ms)

**D-09:** Search filters visible list only
- Does not paginate through unloaded items
- Search works on currently loaded batch

**D-10:** Selection behavior:
- Enter on item sets wallpaper immediately
- First Enter in search mode sets best match
- Success message shown, TUI stays open

### Integration

**D-11:** Combine thumbnails + fuzzy search
- Thumbnails visible even during search
- Search filters which thumbnails shown
- Same Enter behavior in both modes

**D-12:** Performance targets:
- Thumbnail render: <50ms per image
- Search filter: <100ms for 1000 items
- Total TUI startup: <2s with thumbnails

**D-13:** Memory management:
- Don't hold all thumbnails in memory
- Render on-demand, cache aggressively
- Unload off-screen thumbnails

### OpenCode's Discretion

- Exact spacing between thumbnail and text
- Color scheme for selected vs non-selected items
- Whether to dim thumbnails for non-selected items
- Error handling when thumbnail generation fails
- Debounce timing for search (suggest 150ms)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase Specification
- `.planning/milestones/M002/slices/S03/S03-PLAN.md` — Full task breakdown with T00 as top priority
- `.planning/milestones/M002/M002-ROADMAP.md` — Milestone context

### Prior Phase Artifacts
- `.planning/phases/02-tui-bubble-tea/02-CONTEXT.md` — TUI foundation
- `internal/tui/model.go` — Current TUI implementation
- `internal/thumbs/thumbs.go` — Thumbnail caching (needs 128x128 addition)

### Libraries
- `blacktop/go-termimg` — Terminal image rendering
- `sahilm/fuzzy` — Fuzzy string matching
- `charmbracelet/bubbles/list` — List component (will customize)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **internal/tui/model.go** — Bubble Tea model with list
  - Has `list.Model` with default delegate
  - Need to replace with custom thumbnail delegate
  
- **internal/thumbs/thumbs.go** — Thumbnail cache
  - Currently generates 256x256 only
  - Need to add 128x128 generation for list view
  
- **internal/tui/pagination.go** — Pagination logic
  - Thumbnails should work with paginated lists

### Customization Points

**Current list delegate (default):**
```go
delegate := list.NewDefaultDelegate()
// Just text, no images
```

**Required custom delegate:**
```go
type thumbnailDelegate struct {
    imageMethod ImageMethod
    cache       *thumbs.Cache
}

func (d thumbnailDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    // 1. Get thumbnail path
    // 2. Render image using go-termimg
    // 3. Render text alongside
    // 4. Handle selection highlighting
}
```

### Integration Points

**New files needed:**
- `internal/tui/delegate.go` — Custom item delegate with thumbnails
- `internal/tui/fuzzy.go` — Fuzzy search integration
- `internal/tui/search.go` — Search mode UI

**Modified files:**
- `internal/thumbs/thumbs.go` — Add 128x128 generation
- `internal/tui/model.go` — Replace default delegate with custom

</code_context>

<specifics>
## Specific Ideas

### Custom Delegate Implementation

```go
// internal/tui/delegate.go

type ThumbnailDelegate struct {
    imageMethod ImageMethod
    cache       *thumbs.Cache
    styles      delegateStyles
}

type delegateStyles struct {
    normalItem    lipgloss.Style
    selectedItem  lipgloss.Style
    thumbnailSize int
}

func (d ThumbnailDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
    item := listItem.(WallpaperItem)
    
    // Get or generate thumbnail
    thumbPath, _ := d.cache.Get(item.Path)
    if thumbPath == "" {
        // Generate on-demand
        thumbPath, _ = d.cache.Generate(item.Path)
    }
    
    // Render image
    var imgStr string
    if thumbPath != "" {
        imgStr = renderTerminalImage(thumbPath, d.styles.thumbnailSize)
    } else {
        imgStr = "[📷]"
    }
    
    // Render text
    text := fmt.Sprintf("%s\n%s", item.Name, item.Description())
    
    // Combine based on selection state
    if index == m.Index() {
        // Selected styling
        fmt.Fprint(w, d.styles.selectedItem.Render(imgStr + " " + text))
    } else {
        // Normal styling
        fmt.Fprint(w, d.styles.normalItem.Render(imgStr + " " + text))
    }
}

func renderTerminalImage(path string, size int) string {
    // Use go-termimg
    img := termimg.New(path)
    img.SetMaxWidth(size)
    img.SetMaxHeight(size)
    
    rendered, err := img.Render()
    if err != nil {
        return "[📷]"
    }
    return rendered
}
```

### Thumbnail Cache Update

```go
// Add to internal/thumbs/thumbs.go

const (
    ThumbnailWidthSmall  = 128  // For list view
    ThumbnailWidthLarge  = 256  // For detailed view (future)
)

func (c *Cache) GenerateSize(imagePath string, size int) (string, error) {
    // Similar to Generate() but with configurable size
    // Cache key includes size: "128_<hash>.jpg"
}
```

### Terminal Detection

```go
// Enhanced detectImageMethod()

func detectImageMethod() ImageMethod {
    // Check $TERM and $TERM_PROGRAM
    term := os.Getenv("TERM")
    termProgram := os.Getenv("TERM_PROGRAM")
    
    // Kitty detection
    if strings.Contains(term, "kitty") || os.Getenv("KITTY_WINDOW_ID") != "" {
        return MethodKitty
    }
    
    // iTerm2 detection  
    if termProgram == "iTerm.app" || strings.Contains(term, "iTerm") {
        return MethodIterm2
    }
    
    // Sixel detection
    if hasSixelSupport() {
        return MethodSixel
    }
    
    // Default
    return MethodHalfBlocks
}
```

### Fuzzy Search Integration

```go
// internal/tui/fuzzy.go

import "github.com/sahilm/fuzzy"

type FuzzySearcher struct {
    items []WallpaperItem
}

func (f *FuzzySearcher) Search(query string) []WallpaperItem {
    if query == "" {
        return f.items
    }
    
    // Match against multiple fields
    sources := make([]string, len(f.items))
    for i, item := range f.items {
        sources[i] = item.Name + " " + item.Source
    }
    
    matches := fuzzy.Find(query, sources)
    
    // Return matched items sorted by score
    result := make([]WallpaperItem, len(matches))
    for i, match := range matches {
        result[i] = f.items[match.Index]
    }
    
    return result
}
```

</specifics>

<deferred>
## Deferred Ideas

- Full-size image preview (separate pane)
- External image viewer launch (e.g., `open` on macOS)
- Image metadata display (EXIF, file size)
- Sort by image dimensions
- Filter by resolution in UI
- Multi-select for batch operations
- Animated thumbnails (for video wallpapers)

</deferred>

---

*Phase: 03-thumbnails-fuzzy*  
*Priority: ⭐ Thumbnail integration is TOP priority*  
*Context gathered: 2026-04-04*
