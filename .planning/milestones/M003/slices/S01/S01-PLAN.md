# M003-S01: TUI Overhaul — Split-Pane Layout & Better Thumbnails

**Slice:** S01 of M003  
**Goal:** Redesign TUI with split-pane layout, adaptive thumbnail scaling, and responsive design  
**Estimate:** 6-8 hours  
**Dependencies:** None (builds on existing M002 TUI)  
**Best Practices:** See [TUI Best Practices](../../TUI-BEST-PRACTICES.md) (EXA research)

---

## Current State (M002)

**Layout:** Single-pane list with inline 64x64 thumbnails
```
┌────────────────────────────────────┐
│ wallpaper-cli browse               │
│                                    │
│ 🖼️  01_abc.jpg    anime, landscape │
│ 🖼️  02_def.png    anime, city     │
│ 🖼️  03_ghi.jpg    anime, sunset   │
│ ...                                │
│                                    │
│ n: next  |  ?: help  |  q: quit    │
└────────────────────────────────────┘
```

**Issues to Solve:**
- No preview pane — users can't see full wallpaper before setting
- Fixed thumbnail size (64x64) — doesn't adapt to terminal size
- Single-pane limits information density
- No metadata display without pressing keys

---

## Target State (M003)

**Layout:** Split-pane with list left, preview right
```
┌─────────────────────────────────────────────────────────────┐
│                    wallpaper-cli v1.3                       │
├──────────────────────┬──────────────────────────────────────┤
│                      │                                      │
│  📋 WALLPAPERS       │  👁️  PREVIEW                         │
│  ─────────────────── │                                      │
│  🖼️  01_abc.jpg     │  ┌──────────────────────────────┐   │
│  🖼️  02_def.png     │  │                              │   │
│  🖼️  03_ghi.jpg  ◄──┼──│    [Large Thumbnail]         │   │
│  🖼️  04_jkl.png     │  │    3840x2160 • 2.4MB         │   │
│  🖼️  05_mno.jpg     │  │                              │   │
│                      │  └──────────────────────────────┘   │
│  [j/k navigate]      │                                      │
│  [Enter: set]        │  📊 METADATA                         │
│  [f: favorite]       │  ─────────────────                   │
│  [r: rate]           │  Source: wallhaven                   │
│  [p: playlist]       │  Resolution: 3840x2160               │
│  [q: quit]           │  Size: 2.4 MB                          │
│                      │  Tags: anime, landscape              │
│  📊 Page 1/16        │  Downloaded: 2026-04-04              │
│  ⭐ 23 favorites      │                                      │
│                      │  ⌨️  ACTIONS                          │
│                      │  [Enter] Set as wallpaper            │
│                      │  [f] Toggle favorite                 │
│                      │  [r] Rate 1-5                        │
│                      │  [p] Add to playlist                 │
│                      │                                      │
├──────────────────────┴──────────────────────────────────────┤
│  💡 Tip: Press 'd' to start rotation  │  🔄 Rotation: OFF    │
└─────────────────────────────────────────────────────────────┘
```

---

## Technical Design

### Layout Components

```go
// Model structure
type Model struct {
    // Layout
    terminalWidth  int
    terminalHeight int
    showSplitPane  bool  // false if terminal < 80 cols
    leftPaneWidth  int   // 40-50% of width
    rightPaneWidth int   // 50-60% of width
    
    // Left pane (list)
    list            list.Model
    listThumbnailSize int  // 32, 48, or 64 based on width
    
    // Right pane (preview)
    selectedItem    *WallpaperItem
    previewImage    termimg.Image
    previewHeight   int  // Dynamic based on terminal height
    
    // State
    showHelp        bool
    searchMode      bool
    currentPlaylist string
    onlyFavorites   bool
}
```

### Responsive Breakpoints

| Terminal Size | Layout Mode | Split Ratio | List Thumb | Preview Height |
|---------------|-------------|-------------|------------|----------------|
| < 80x24 | Stacked (mobile) | N/A (single) | 32x32 | 80px |
| 80-100x24 | Compact split | 50/50 | 48x48 | 100px |
| 100-140x30 | Standard split | 45/55 | 48x48 | 150px |
| > 140x35 | Wide split | 40/60 | 64x64 | 200px |

### Split Pane Implementation

```go
func (m Model) View() string {
    if !m.showSplitPane {
        // Mobile/small terminal: stacked layout
        return m.stackedView()
    }
    
    // Split pane layout
    left := m.leftPaneView()
    right := m.rightPaneView()
    
    return lipgloss.JoinHorizontal(
        lipgloss.Top,
        left,
        right,
    )
}

func (m Model) leftPaneView() string {
    // Wallpaper list with thumbnails
    return lipgloss.NewStyle().
        Width(m.leftPaneWidth).
        Height(m.terminalHeight - 2). // Leave room for status bar
        Border(lipgloss.RoundedBorder()).
        BorderRight(true).
        Render(m.list.View())
}

func (m Model) rightPaneView() string {
    // Preview pane with metadata
    preview := m.renderPreview()
    metadata := m.renderMetadata()
    actions := m.renderActions()
    
    content := lipgloss.JoinVertical(
        lipgloss.Left,
        preview,
        metadata,
        actions,
    )
    
    return lipgloss.NewStyle().
        Width(m.rightPaneWidth).
        Height(m.terminalHeight - 2).
        Padding(1).
        Render(content)
}
```

### Adaptive Thumbnails

```go
func (m *Model) calculateThumbnailSize() int {
    switch {
    case m.terminalWidth < 80:
        return 32
    case m.terminalWidth < 120:
        return 48
    default:
        return 64
    }
}

func (m *Model) calculatePreviewHeight() int {
    // Reserve space for metadata and actions
    minMetadataHeight := 8
    available := m.terminalHeight - 4 - minMetadataHeight // -4 for borders/padding
    
    switch {
    case available < 100:
        return 80
    case available < 150:
        return 100
    case available < 200:
        return 150
    default:
        return 200
    }
}
```

### Enhanced List Delegate

```go
type SplitPaneDelegate struct {
    thumbnailSize int
    showMetadata  bool
}

func (d SplitPaneDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    wallpaper := item.(WallpaperItem)
    
    // Render thumbnail
    thumb := d.renderThumbnail(wallpaper, d.thumbnailSize)
    
    // Render text info
    title := truncate(wallpaper.Filename, m.Width()-d.thumbnailSize-4)
    desc := fmt.Sprintf("%s • %s", wallpaper.Source, wallpaper.Resolution)
    
    // Favorite indicator
    fav := ""
    if wallpaper.IsFavorite {
        fav = "⭐ "
    }
    
    // Rating indicator
    rating := strings.Repeat("★", wallpaper.Rating)
    
    fmt.Fprintf(w, "%s %s%s\n%s%s", 
        thumb, fav, title, 
        strings.Repeat(" ", d.thumbnailSize+2), desc)
    
    if rating != "" {
        fmt.Fprintf(w, " %s", rating)
    }
}
```

---

## Keybindings (Updated)

| Key | Action | Context |
|-----|--------|---------|
| `j` / `↓` | Navigate down | Global |
| `k` / `↑` | Navigate up | Global |
| `h` / `←` | Previous playlist | Global |
| `l` / `→` | Next playlist | Global |
| `Enter` | Set wallpaper | Global |
| `f` | Toggle favorite | Global |
| `r` | Open rating selector | Global |
| `p` | Add to playlist | Global |
| `P` | Create new playlist | Global |
| `o` | Open in macOS app | Global |
| `/` | Enter search mode | Global |
| `n` | Next page | Global |
| `N` | Previous page | Global |
| `Tab` | Focus right pane (if split) | Global |
| `Esc` | Cancel / Quit | Global |
| `q` | Quit | Global |
| `?` | Toggle help | Global |

---

## File Structure

```
internal/tui/
├── model.go           # Updated Model struct
├── splitpane.go       # NEW: Split-pane layout logic
├── responsive.go      # NEW: Responsive breakpoint handling
├── delegate.go        # Updated ThumbnailDelegate
├── preview.go         # NEW: Preview pane rendering
├── metadata.go        # NEW: Metadata display component
├── actions.go         # NEW: Actions panel component
├── fuzzy.go           # Existing (unchanged)
├── pagination.go      # Existing (minor updates)
└── keybindings.go     # NEW: Centralized keybinding definitions
```

---

## Tasks

| ID | Title | Est. | Details | Best Practice |
|----|-------|------|---------|---------------|
| T01 | Create responsive layout engine | 1.5h | Breakpoint detection, pane size calculation | Use lipgloss.Width/Height |
| T02 | Implement split-pane container | 1.5h | Left/right pane rendering with lipgloss | Tree of models pattern |
| T03 | Build adaptive thumbnail sizing | 1h | 32/48/64px based on terminal width | Responsive breakpoints |
| T04 | Create preview pane component | 1.5h | Large thumbnail, metadata display | Async image loading |
| T05 | Add actions panel | 1h | Favorite, rate, playlist buttons | State-driven modals |
| T06 | Implement stacked fallback | 1h | Mobile/small terminal layout | Progressive enhancement |
| T07 | Enhanced keybindings | 0.5h | New keys (f, r, p, P, h, l) | Vim-style navigation |
| T08 | Update help overlay | 0.5h | Document new layout and keys | Contextual help |
| T09 | Terminal resize handling | 0.5h | Dynamic layout updates | WindowSizeMsg handling |
| T10 | Performance optimization | 1h | Async preview loading, caching | Non-blocking tea.Cmd |

**Total: 8 hours**

---

## Success Criteria

- [ ] Split-pane renders on terminals ≥ 80 columns
- [ ] Stacked fallback works on terminals < 80 columns  
- [ ] Thumbnails scale: 32px (<80), 48px (80-120), 64px (>120)
- [ ] Preview scales: 80px-200px based on terminal height
- [ ] All M002 functionality preserved (pagination, fuzzy search, etc.)
- [ ] New keybindings work: f (favorite), r (rate), p (playlist)
- [ ] Smooth terminal resize handling
- [ ] No flickering or layout glitches

---

## Testing Checklist

**Terminal Sizes:**
- [ ] 60x20 (stacked mode)
- [ ] 80x24 (compact split)
- [ ] 100x30 (standard split)
- [ ] 150x40 (wide split)

**Features:**
- [ ] Navigation (j/k, arrows)
- [ ] Wallpaper setting (Enter)
- [ ] Favorites toggle (f)
- [ ] Rating selector (r)
- [ ] Playlist add (p)
- [ ] Search (/)
- [ ] Resize handling
- [ ] Help overlay (?)

---

## Integration with Other Slices

**S02 (Collections):**
- List delegate displays favorites ⭐ and ratings ★★★★★
- Preview pane shows favorite toggle and rating selector
- Actions panel provides quick access to collection features

**S03 (Scheduling):**
- Status bar shows current rotation state (ON/OFF)
- Tip bar suggests daemon commands
- Can start/stop rotation from TUI

**S04 (Daemon):**
- TUI can control daemon (start/stop/status)
- Shows current schedule in status bar

---

*Split-pane TUI provides a rich, modern browsing experience while maintaining the CLI's efficiency and keyboard-centric workflow.*
