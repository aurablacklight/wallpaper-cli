# M003 TUI Best Practices Research

**Research Source:** EXA search + expert blogs + Bubble Tea documentation  
**Date:** 2026-04-04  
**Purpose:** Apply best TUI patterns to M003 split-pane design

---

## Key Research Sources

1. **"Tips for building Bubble Tea programs"** by Louis Garman (PUG author)
   - https://leg100.github.io/en/posts/building-bubbletea-programs/
   
2. **Bubble Tea Multi-View Application Example** (DeepWiki)
   - https://deepwiki.com/charmbracelet/bubbletea/6.2-multi-view-application-example
   
3. **go-termimg library** by blacktop
   - https://github.com/blacktop/go-termimg
   - Modern terminal image rendering with multiple protocol support

4. **Bubble Tea Layout Handling Discussions** (GitHub)
   - https://github.com/charmbracelet/bubbletea/discussions/307

---

## 🎯 Critical Best Practices for M003

### 1. Model Architecture: Tree of Models

**Best Practice:** Build a tree of models for non-trivial TUIs (from PUG author)

```go
// DON'T: Single massive model
type Model struct {
    // 50+ fields for everything
}

// DO: Tree structure with focused models
type RootModel struct {
    width, height int
    
    // Child models
    listModel    ListModel      // Left pane
    previewModel PreviewModel   // Right pane
    modalModel   *ModalModel    // Optional modal (dynamically created)
    
    // State
    currentView ViewType
    showModal   bool
}

type ListModel struct {
    list         list.Model
    items        []WallpaperItem
    thumbnailSize int
}

type PreviewModel struct {
    selectedItem  *WallpaperItem
    previewImage  *termimg.ImageWidget
    showMetadata  bool
}
```

**Benefits:**
- Each model handles its own domain
- Easier to test individual components
- Can create models dynamically (e.g., modals)
- Message routing is cleaner

**For M003:** Create separate models for:
- Left pane (list with thumbnails)
- Right pane (preview, metadata, actions)
- Modals (rating selector, playlist selector)

---

### 2. Layout Arithmetic: Use lipgloss.Height/Width

**Best Practice:** Never hardcode dimensions — use lipgloss to measure (from PUG author)

```go
// DON'T: Hardcoded arithmetic (breaks when adding borders)
contentHeight := m.height - 1 - 1  // header + footer

// DO: Use lipgloss to measure actual rendered heights
header := lipgloss.NewStyle().
    Border(lipgloss.NormalBorder(), false, false, true, false).
    Render("Header")

footer := lipgloss.NewStyle().
    Render("Footer")

// Calculate remaining space dynamically
contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
```

**M003 Application:**
```go
func (m RootModel) View() string {
    left := m.leftPane.View()
    right := m.rightPane.View()
    
    // Use actual rendered heights for joining
    return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m LeftPaneModel) View() string {
    // Calculate thumbnail size based on actual available width
    availableWidth := m.width - 4 // padding + borders
    m.thumbnailSize = calculateThumbSize(availableWidth)
    
    return lipgloss.NewStyle().
        Width(m.width).
        Height(m.height).
        Render(m.list.View())
}
```

---

### 3. State-Driven View Selection

**Best Practice:** Use enum/boolean to switch between views (from Bubble Tea examples)

```go
type ViewType int

const (
    ViewList ViewType = iota
    ViewPreview
    ViewModal
)

type Model struct {
    currentView ViewType
    
    // View-specific state
    listModel    ListModel
    previewModel PreviewModel
    modalModel   ModalModel
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Global keys first
    if msg, ok := msg.(tea.KeyMsg); ok {
        if msg.String() == "q" || msg.String() == "esc" {
            return m, tea.Quit
        }
    }
    
    // Delegate to view-specific update
    switch m.currentView {
    case ViewList:
        return m.updateList(msg)
    case ViewPreview:
        return m.updatePreview(msg)
    case ViewModal:
        return m.updateModal(msg)
    }
    
    return m, nil
}

func (m Model) View() string {
    switch m.currentView {
    case ViewList:
        return m.listModel.View()
    case ViewPreview:
        return m.previewModel.View()
    case ViewModal:
        // Overlay modal on current view
        return m.renderModalOverlay()
    }
    return ""
}
```

**M003 Application:**
```go
type PaneFocus int
const (
    FocusList PaneFocus = iota
    FocusPreview
    FocusModal
)

// Tab to switch panes
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            m.focus = (m.focus + 1) % 3
            return m, nil
        }
    }
    
    // Route to focused pane
    switch m.focus {
    case FocusList:
        return m.listModel.Update(msg)
    case FocusPreview:
        return m.previewModel.Update(msg)
    }
    return m, nil
}
```

---

### 4. Keep Event Loop Fast

**Best Practice:** Offload expensive operations to tea.Cmd (from PUG author)

```go
// DON'T: Block the event loop
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            // BLOCKING: Loads and resizes image synchronously
            img := loadAndResizeHugeImage(m.selected)
            m.preview = img
            return m, nil
        }
    }
    return m, nil
}

// DO: Use async commands
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            // NON-BLOCKING: Returns immediately, processes in background
            return m, func() tea.Msg {
                img := loadAndResizeHugeImage(m.selected)
                return ImageLoadedMsg{image: img}
            }
        }
        
    case ImageLoadedMsg:
        // Handle the loaded image when it's ready
        m.preview = msg.image
        return m, nil
    }
    return m, nil
}
```

**M003 Application:**
```go
// For thumbnail generation (can be slow)
case tea.KeyMsg:
    if msg.String() == "j" || msg.String() == "k" {
        // Update selection immediately (fast)
        m.listModel.MoveCursor(msg.String())
        
        // Load preview asynchronously (may be slow)
        return m, func() tea.Msg {
            item := m.listModel.SelectedItem()
            preview := loadPreview(item.Path, m.previewWidth)
            return PreviewReadyMsg{item: item, preview: preview}
        }
    }

case PreviewReadyMsg:
    // Update preview when ready
    m.previewModel.SetImage(msg.preview)
    return m, nil
```

---

### 5. Responsive Thumbnail Sizing

**Best Practice:** Use breakpoints based on actual terminal measurements

```go
type ThumbnailConfig struct {
    Size        int  // 32, 48, or 64
    PreviewHeight int // 80, 100, 150, or 200
    ShowMetadata  bool
    Layout        LayoutType
}

type LayoutType int
const (
    LayoutStacked LayoutType = iota
    LayoutCompact
    LayoutStandard
    LayoutWide
)

func calculateLayout(width, height int) ThumbnailConfig {
    switch {
    case width < 80 || height < 24:
        return ThumbnailConfig{
            Size:         32,
            PreviewHeight: 80,
            ShowMetadata:  false,
            Layout:        LayoutStacked,
        }
    case width < 100:
        return ThumbnailConfig{
            Size:         48,
            PreviewHeight: 100,
            ShowMetadata:  true,
            Layout:        LayoutCompact,
        }
    case width < 140:
        return ThumbnailConfig{
            Size:         48,
            PreviewHeight: 150,
            ShowMetadata:  true,
            Layout:        LayoutStandard,
        }
    default:
        return ThumbnailConfig{
            Size:         64,
            PreviewHeight: 200,
            ShowMetadata:  true,
            Layout:        LayoutWide,
        }
    }
}
```

**M003 Implementation:**
```go
func (m RootModel) handleWindowSize(msg tea.WindowSizeMsg) (RootModel, tea.Cmd) {
    m.width, m.height = msg.Width, msg.Height
    
    config := calculateLayout(msg.Width, msg.Height)
    
    // Update child models with new dimensions
    m.listModel.SetThumbnailSize(config.Size)
    m.listModel.SetDimensions(config.LeftWidth(), m.height-2)
    
    m.previewModel.SetPreviewHeight(config.PreviewHeight)
    m.previewModel.SetDimensions(config.RightWidth(), m.height-2)
    
    // Toggle between split and stacked
    m.showSplitPane = config.Layout != LayoutStacked
    
    return m, nil
}
```

---

### 6. Terminal Image Protocol Best Practices

**From go-termimg research:**

**Protocol Priority:**
1. **Kitty** — Fastest, virtual images, z-index support
2. **iTerm2** — Good performance, macOS native
3. **SIXEL** — High quality, slower (~90ms), limited terminal support
4. **Halfblocks** — Universal fallback, fastest (~800µs), lowest quality

**Auto-Detection Strategy:**
```go
// go-termimg approach
protocol := termimg.DetectProtocol()

// Check specific support
if termimg.KittySupported() {
    // Use Kitty with advanced features
} else if termimg.ITerm2Supported() {
    // Use iTerm2
} else if termimg.SixelSupported() {
    // Use SIXEL with dithering
} else {
    // Fall back to halfblocks
}
```

**M003 Integration:**
```go
type PreviewModel struct {
    widget *termimg.ImageWidget
    currentPath string
}

func (m *PreviewModel) SetImage(path string, width, height int) error {
    // Use go-termimg's TUI widget
    m.widget = termimg.NewImageWidgetFromFile(path)
    m.widget.SetSize(width, height)
    m.widget.SetProtocol(termimg.Auto) // Auto-detect best
    m.widget.SetScale(termimg.ScaleFit) // Fit within bounds
    
    return nil
}

func (m PreviewModel) View() string {
    rendered, _ := m.widget.Render()
    return rendered
}
```

**Performance Considerations:**
- Cache rendered images (don't re-render every frame)
- Use thumbnail cache for list view (pre-generated small thumbs)
- For large preview, generate once on selection change
- Kitty virtual images are most efficient for TUI (don't redraw)

---

### 7. Split Pane Implementation Pattern

**Best Practice from PUG:** Use lipgloss.JoinHorizontal with measured widths

```go
type SplitPaneModel struct {
    leftWidth  int
    rightWidth int
    height     int
}

func (m SplitPaneModel) View() string {
    left := m.left.View()
    right := m.right.View()
    
    // Join horizontally with top alignment
    return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m *SplitPaneModel) SetDimensions(width, height int) {
    m.height = height - 2 // Reserve space for status bar
    
    // 40/60 split
    m.leftWidth = int(float64(width) * 0.4)
    m.rightWidth = width - m.leftWidth - 1 // -1 for border
}

// Child model styling
func (m LeftPaneModel) View() string {
    return lipgloss.NewStyle().
        Width(m.width).
        Height(m.height).
        Border(lipgloss.RoundedBorder()).
        BorderRight(true).
        Render(m.list.View())
}

func (m RightPaneModel) View() string {
    content := lipgloss.JoinVertical(lipgloss.Left,
        m.preview.View(),
        m.metadata.View(),
        m.actions.View(),
    )
    
    return lipgloss.NewStyle().
        Width(m.width).
        Height(m.height).
        Border(lipgloss.RoundedBorder()).
        Padding(1).
        Render(content)
}
```

---

### 8. Modal Overlay Pattern

**Best Practice:** Render modal on top of existing view with dimming

```go
func (m RootModel) View() string {
    // Base view
    base := m.splitPane.View()
    
    if !m.showModal {
        return base
    }
    
    // Render modal centered on top
    modal := m.modal.View()
    
    // Option 1: Simple overlay (modal obscures background)
    return m.centerModal(modal, base)
    
    // Option 2: Dimmed background (if supported)
    // dimmed := lipgloss.NewStyle().
    //     Faint(true).
    //     Render(base)
    // return m.overlay(modal, dimmed)
}

func (m RootModel) centerModal(modal, background string) string {
    modalWidth := lipgloss.Width(modal)
    modalHeight := lipgloss.Height(modal)
    
    // Center horizontally and vertically
    leftPadding := (m.width - modalWidth) / 2
    topPadding := (m.height - modalHeight) / 2
    
    // Use Place for precise positioning
    return lipgloss.Place(
        m.width, m.height,
        lipgloss.Center, lipgloss.Center,
        modal,
        lipgloss.WithWhitespaceChars(" "),
    )
}
```

---

### 9. Testing with teatest

**Best Practice:** End-to-end testing of TUI interactions

```go
func TestTUI(t *testing.T) {
    m := NewModel()
    tm := teatest.NewTestModel(t, m)
    
    // Wait for initial render
    waitForString(t, tm, "WALLPAPERS")
    
    // Navigate down
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})
    waitForString(t, tm, "▼") // Selection indicator
    
    // Toggle favorite
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
    waitForString(t, tm, "⭐")
    
    // Quit
    tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
    tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
```

---

## Summary: M003 TUI Design Recommendations

### Architecture
- ✅ Use tree of models (Root → List + Preview + Modals)
- ✅ State-driven view selection with enum
- ✅ Delegate updates to child models

### Layout
- ✅ Use lipgloss.Height/Width for calculations
- ✅ Responsive breakpoints: <80, 80-100, 100-140, >140 cols
- ✅ Dynamic thumbnail sizing: 32→48→64px
- ✅ Split-pane with lipgloss.JoinHorizontal

### Performance
- ✅ Keep event loop fast — offload to tea.Cmd
- ✅ Async image loading for preview
- ✅ Cache thumbnails for list view
- ✅ Use go-termimg with auto-detection

### User Experience
- ✅ Modal overlays for rating/playlist selection
- ✅ Vim-style navigation (j/k/h/l)
- ✅ Clear visual feedback (⭐ ★ indicators)
- ✅ Help overlay with ? key

### Testing
- ✅ Use teatest for E2E tests
- ✅ Use VHS for demos/documentation
- ✅ Test multiple terminal sizes

---

**Next Step:** Apply these best practices when implementing S01 TUI Overhaul
