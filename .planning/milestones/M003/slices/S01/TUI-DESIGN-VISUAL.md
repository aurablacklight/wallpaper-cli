# M003-S01: Split-Pane TUI Design Visual Reference

**Visual mockups of the split-pane TUI layout at different terminal sizes**

---

## Layout A: Wide Terminal (140+ columns)

**Optimal experience — maximum information density**

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                               wallpaper-cli v1.3 • 156 wallpapers                      │
├────────────────────────┬─────────────────────────────────────────────────────────────┤
│                        │                                                             │
│  📋 WALLPAPERS         │  👁️  PREVIEW                                                │
│  ═══════════════       │                                                             │
│                        │  ┌─────────────────────────────────────────────────────┐   │
│  ⭐🖼️  01_abc.jpg     │  │                                                     │   │
│    🖼️  02_def.png     │  │           [200px height thumbnail]                  │   │
│  ⭐🖼️  03_ghi.jpg  ◄──┼──│           3840x2160 • 2.4 MB                        │   │
│  ★★🖼️  04_jkl.png     │  │                                                     │   │
│    🖼️  05_mno.jpg     │  └─────────────────────────────────────────────────────┘   │
│    🖼️  06_pqr.jpg     │                                                             │
│  ⭐🖼️  07_stu.jpg     │  📊 METADATA                                                │
│    🖼️  08_vwx.jpg     │  ═══════════                                                │
│                        │  Source:     wallhaven                                      │
│  📊 Page 2/16          │  Resolution: 3840x2160 (4K)                                 │
│  ⭐ 23 favorites       │  Size:       2.4 MB                                         │
│  📋 3 playlists        │  Tags:       anime, landscape, sunset                         │
│                        │  Downloaded: 2026-04-04                                       │
│  ⌨️  j/k navigate      │  Rating:     ★★★★★ (5/5)                                      │
│     Enter set          │  Favorite:   ⭐ Yes                                           │
│     f favorite         │  Playlists:  cozy, focus                                      │
│     r rate             │                                                             │
│     p playlist         │  ⌨️  ACTIONS                                                  │
│     / search           │  ═══════════                                                  │
│     q quit             │  [Enter] Set as wallpaper     [o] Open in macOS app          │
│                        │  [f]     Toggle favorite      [d] Start daemon             │
│                        │  [r]     Rate 1-5             [?] Help                     │
│                        │  [p]     Add to playlist      [q] Quit                     │
│                        │                                                             │
├────────────────────────┴─────────────────────────────────────────────────────────────┤
│  💡 Tip: Press 'd' to start automatic rotation     │  🔄 Rotation: OFF  │  ⏰ Next: --   │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

**Specs:**
- Left pane: 40% width (56 cols), 64x64 thumbnails
- Right pane: 60% width (84 cols), 200px preview height
- Metadata: Full display with all fields
- Actions: All options visible

---

## Layout B: Standard Terminal (100-140 columns)

**Balanced experience — good for most users**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     wallpaper-cli v1.3 • 156 wallpapers                      │
├────────────────────┬────────────────────────────────────────────────────────┤
│                    │                                                        │
│  📋 WALLPAPERS     │  👁️  PREVIEW                                           │
│  ═══════════════   │                                                        │
│                    │  ┌────────────────────────────────────────────────┐   │
│  🖼️ 01_abc.jpg    │  │                                                │   │
│  🖼️ 02_def.png    │  │      [150px height thumbnail]                  │   │
│  🖼️ 03_ghi.jpg ◄──┼──│      3840x2160 • 2.4 MB                          │   │
│  🖼️ 04_jkl.png    │  │                                                │   │
│  🖼️ 05_mno.jpg    │  └────────────────────────────────────────────────┘   │
│                    │                                                        │
│  📊 Page 2/16      │  📊 METADATA                                           │
│  ⭐ 23 favorites   │  Source: wallhaven    Resolution: 3840x2160            │
│                    │  Size: 2.4 MB           Rating: ★★★★★                  │
│  ⌨️ j/k navigate   │                                                        │
│    Enter set       │  ⌨️  [Enter] Set  [f] Fav  [r] Rate  [p] Playlist    │
│    f favorite      │                                                        │
│    q quit          │                                                        │
│                    │                                                        │
├────────────────────┴────────────────────────────────────────────────────────┤
│  💡 Press 'd' to start rotation  │  🔄 OFF  │  ⏰ Next: --                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Specs:**
- Left pane: 45% width (45 cols), 48x48 thumbnails  
- Right pane: 55% width (55 cols), 150px preview height
- Metadata: Condensed display
- Actions: Inline shortcuts

---

## Layout C: Compact Terminal (80-100 columns)

**Minimum split-pane — still functional**

```
┌──────────────────────────────────────────────────────────────┐
│              wallpaper-cli v1.3 • 156 wallpapers            │
├──────────────────┬─────────────────────────────────────────┤
│                  │                                         │
│  📋 WALLPAPERS   │  👁️  PREVIEW                            │
│                  │                                         │
│  🖼 01_abc.jpg   │  ┌─────────────────────────────────┐    │
│  🖼 02_def.png   │  │   [100px height thumbnail]      │    │
│  🖼 03_ghi.jpg◄──┼──│   3840x2160 • 2.4 MB              │    │
│  🖼 04_jkl.png   │  └─────────────────────────────────┘    │
│                  │                                         │
│  📊 2/16         │  Source: wallhaven 3840x2160            │
│  ⭐ 23 favs      │  ★★★★★ [f]av [r]ate [p]laylist          │
│                  │                                         │
│  ⌨️ j/k nav      │                                         │
│    Ent set       │                                         │
│    q quit        │                                         │
│                  │                                         │
├──────────────────┴─────────────────────────────────────────┤
│  🔄 OFF │ ⏰ -- │ [?] Help                                 │
└──────────────────────────────────────────────────────────────┘
```

**Specs:**
- Left pane: 50% width (40 cols), 48x48 thumbnails (compact)
- Right pane: 50% width (40 cols), 100px preview height
- Metadata: Minimal display
- Actions: Single-letter shortcuts only

---

## Layout D: Narrow Terminal (< 80 columns)

**Stacked fallback — no split, single column**

```
┌─────────────────────────────────┐
│ wallpaper-cli v1.3 • 156 total │
├─────────────────────────────────┤
│                                 │
│ 📋 WALLPAPERS                   │
│                                 │
│ 👁️  PREVIEW OF SELECTED:        │
│ ┌───────────────────────────┐  │
│ │  [80px thumbnail]         │  │
│ │  3840x2160 • 2.4 MB       │  │
│ └───────────────────────────┘  │
│                                 │
│ 🖼 01_abc.jpg (wallhaven)      │
│ 🖼 02_def.png (wallhaven) ⭐    │
│ 🖼 03_ghi.jpg (wallhaven) ★★★  │
│ 🖼 04_jkl.png (reddit)          │
│ 🖼 05_mno.jpg (wallhaven) ⭐   │
│                                 │
│ 📊 2/16 • ⭐ 23 • 📋 3          │
│                                 │
│ ⌨️  [Ent]set [f]av [r]te [q]uit│
│                                 │
├─────────────────────────────────┤
│ 🔄 OFF │ [?]Help [d]Daemon     │
└─────────────────────────────────┘
```

**Specs:**
- No split — stacked layout
- Thumbnails: 32x32 (small)
- Preview: 80px height (inline above list)
- Actions: Minimal shortcuts
- Full functionality preserved

---

## Responsive Breakpoint Summary

| Terminal Size | Layout | Left % | List Thumb | Preview Height | Actions |
|---------------|--------|--------|------------|----------------|---------|
| > 140 cols | Wide split | 40/60 | 64x64 | 200px | Full panel |
| 100-140 cols | Standard split | 45/55 | 48x48 | 150px | Inline |
| 80-100 cols | Compact split | 50/50 | 48x48 | 100px | Minimal |
| < 80 cols | Stacked | N/A | 32x32 | 80px | Abbreviated |

---

## Key Interactions

### Keyboard Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Navigate down |
| `k` / `↑` | Navigate up |
| `h` / `←` | Previous page |
| `l` / `→` | Next page |
| `Enter` | Set wallpaper |
| `f` | Toggle favorite ⭐ |
| `r` | Open rating selector |
| `p` | Add to playlist |
| `P` | Create playlist |
| `o` | Open in macOS app |
| `d` | Start/stop daemon |
| `/` | Enter search mode |
| `?` | Toggle help |
| `q` / `Esc` | Quit |

### Rating Selector Modal

```
┌─────────────────────┐
│  Rate this wallpaper│
│                     │
│   ★ ★ ★ ★ ★        │
│   1 2 3 4 5         │
│                     │
│   [1-5] Select      │
│   [Enter] Confirm     │
│   [Esc] Cancel      │
│                     │
│   Notes: [________] │
└─────────────────────┘
```

### Playlist Selector Modal

```
┌─────────────────────┐
│  Select Playlist      │
│                     │
│  ▸ cozy (12 items)  │
│    focus (8 items)  │
│    energetic (5)    │
│    winter (15 items)│
│                     │
│  [n] New playlist   │
│  [Enter] Select     │
│  [Esc] Cancel       │
└─────────────────────┘
```

---

## Visual Polish

### Colors (lipgloss)

```go
var (
    // Header
    headerStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#7C3AED")).
        Foreground(lipgloss.Color("#FFFFFF")).
        Bold(true).
        Padding(0, 1)
    
    // Pane borders
    leftPaneStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#6366F1"))
    
    rightPaneStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#8B5CF6"))
    
    // Selection
    selectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#4F46E5")).
        Foreground(lipgloss.Color("#FFFFFF"))
    
    // Status bar
    statusBarStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#1F2937")).
        Foreground(lipgloss.Color("#9CA3AF"))
    
    // Favorites/Ratings
    favoriteStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FBBF24")) // Yellow star
    
    ratingStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#F59E0B")) // Orange star
)
```

### Animations

- **Smooth scrolling:** List scrolls smoothly when navigating
- **Fade transitions:** Preview pane fades when changing selection
- **Pulse indicator:** "Rotation: ON" pulses gently when daemon is active

---

*This visual reference ensures consistent implementation across all terminal sizes.*
