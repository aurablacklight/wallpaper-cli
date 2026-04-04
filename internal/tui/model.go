package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/wallpaper-cli/internal/config"
	"github.com/user/wallpaper-cli/internal/platform"
	"github.com/user/wallpaper-cli/internal/thumbs"
)

// ImageMethod represents the terminal image rendering method
type ImageMethod int

const (
	MethodUnknown ImageMethod = iota
	MethodKitty
	MethodIterm2
	MethodSixel
	MethodHalfBlocks
	MethodNone
)

// WallpaperItem represents a wallpaper in the list
type WallpaperItem struct {
	Path      string
	Name      string
	Source    string
	Thumbnail string
	HasError  bool
}

// Title returns the item title for the list
func (w WallpaperItem) Title() string {
	return w.Name
}

// Description returns the item description
func (w WallpaperItem) Description() string {
	if w.HasError {
		return "⚠️ Error loading thumbnail"
	}
	source := w.Source
	if source == "" {
		source = "unknown"
	}
	return fmt.Sprintf("Source: %s", source)
}

// FilterValue returns the value to filter on
func (w WallpaperItem) FilterValue() string {
	return w.Name
}

// Model is the Bubble Tea model for the wallpaper browser
type Model struct {
	list         list.Model
	cache        *thumbs.Cache
	imageMethod  ImageMethod
	showHelp     bool
	err          error
	msg          string
	width        int
	height       int
	
	// macOS WallpaperEngine hint
	showMacHint      bool
	wallpaperEngineInstalled bool
	hintDismissed    bool
	
	// Dependencies
	cfg          *config.Config
	setter       platform.Setter
}

// NewModel creates a new TUI model
func NewModel(wallpapers []string, cfg *config.Config, setter platform.Setter) (*Model, error) {
	// Initialize thumbnail cache
	cache, err := thumbs.NewCache()
	if err != nil {
		return nil, fmt.Errorf("initializing thumbnail cache: %w", err)
	}

	// Detect image rendering method
	imageMethod := detectImageMethod()

	// Create list items
	items := make([]list.Item, len(wallpapers))
	for i, path := range wallpapers {
		items[i] = createWallpaperItem(path, cache)
	}

	// Setup list
	delegate := newItemDelegate(imageMethod, cache)
	l := list.New(items, delegate, 0, 0)
	l.Title = "Browse Wallpapers"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false) // Fuzzy search deferred to Phase 03
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	// Setup help
	l.Help.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	l.Help.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	m := &Model{
		list:        l,
		cache:       cache,
		imageMethod: imageMethod,
		cfg:         cfg,
		setter:      setter,
	}

	// Check for macOS WallpaperEngine
	if runtime.GOOS == "darwin" {
		m.checkWallpaperEngine()
	}

	return m, nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	// Pre-generate thumbnails in background
	return m.preloadThumbnails()
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-3) // Reserve space for status bar

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "esc":
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "enter":
			return m, m.setSelectedWallpaper()

		case "o":
			if m.showMacHint && !m.hintDismissed {
				return m, m.openWallpaperEngine()
			}

		case "d":
			if m.showMacHint {
				m.hintDismissed = true
				return m, nil
			}
		}

	case thumbnailMsg:
		// Thumbnail generated, refresh item
		m.updateItemThumbnail(msg.index, msg.thumbnail)
		return m, nil

	case setWallpaperMsg:
		if msg.err != nil {
			m.err = msg.err
			m.msg = fmt.Sprintf("❌ Error: %v", msg.err)
		} else {
			m.msg = fmt.Sprintf("✅ Wallpaper set: %s", msg.name)
			// Save to config
			m.cfg.AddWallpaper(msg.path, "manual")
			m.cfg.Save(config.GetConfigPath())
		}
		return m, nil

	case openAppMsg:
		if msg.err != nil {
			m.msg = fmt.Sprintf("⚠️ Could not open app: %v", msg.err)
		} else {
			m.msg = "🖼️  Opening WallpaperEngine.app..."
		}
		return m, nil
	}

	// Update list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI
func (m *Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	var b strings.Builder

	// Main list view
	b.WriteString(m.list.View())

	// Status bar
	statusBar := m.renderStatusBar()
	b.WriteString("\n")
	b.WriteString(statusBar)

	return b.String()
}

// renderStatusBar renders the bottom status bar
func (m *Model) renderStatusBar() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1).
		Width(m.width)

	var parts []string

	// Selection info
	if m.list.SelectedItem() != nil {
		selected := m.list.SelectedItem().(WallpaperItem)
		parts = append(parts, fmt.Sprintf("📷 %s", selected.Name))
	}

	// Message or help hint
	if m.msg != "" {
		parts = append(parts, m.msg)
	} else {
		parts = append(parts, "Enter: set | ?: help | q: quit")
	}

	// macOS hint
	if m.showMacHint && !m.hintDismissed && m.wallpaperEngineInstalled {
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Background(lipgloss.Color("#333333"))
		parts = append(parts, hintStyle.Render("💡 o: WallpaperEngine | d: dismiss"))
	}

	return style.Render(strings.Join(parts, " | "))
}

// renderHelp renders the help overlay
func (m *Model) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(2).
		Width(60)

	help := `
Keyboard Shortcuts:

  ↑ / ↓          Navigate up/down
  Enter          Set selected wallpaper
  q / Esc        Quit

  ?              Toggle this help

macOS Integration:
  o              Open WallpaperEngine.app (if installed)
  d              Dismiss WallpaperEngine hint

Image Rendering:
  Current method: ` + methodName(m.imageMethod) + `

Press any key to close help...
`

	return helpStyle.Render(help)
}

// setSelectedWallpaper sets the currently selected wallpaper
func (m *Model) setSelectedWallpaper() tea.Cmd {
	if m.list.SelectedItem() == nil {
		return nil
	}

	item := m.list.SelectedItem().(WallpaperItem)

	return func() tea.Msg {
		err := m.setter.SetWallpaper(item.Path)
		return setWallpaperMsg{
			path: item.Path,
			name: item.Name,
			err:  err,
		}
	}
}

// openWallpaperEngine opens the WallpaperEngine app on macOS
func (m *Model) openWallpaperEngine() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("open", "-a", "WallpaperEngine")
		err := cmd.Run()
		return openAppMsg{err: err}
	}
}

// preloadThumbnails pre-generates thumbnails in background
func (m *Model) preloadThumbnails() tea.Cmd {
	return func() tea.Msg {
		// Start with first 20 items
		items := m.list.Items()
		count := 20
		if len(items) < count {
			count = len(items)
		}

		for i := 0; i < count; i++ {
			item := items[i].(WallpaperItem)
			if thumb, err := m.cache.Generate(item.Path); err == nil {
				return thumbnailMsg{index: i, thumbnail: thumb}
			}
		}
		return nil
	}
}

// updateItemThumbnail updates an item with its generated thumbnail
func (m *Model) updateItemThumbnail(index int, thumbnail string) {
	items := m.list.Items()
	if index >= len(items) {
		return
	}

	item := items[index].(WallpaperItem)
	item.Thumbnail = thumbnail

	// This is a bit hacky - Bubble Tea list doesn't have a direct update method
	// In production, we'd use a custom list implementation
}

// checkWallpaperEngine checks if WallpaperEngine.app is installed
func (m *Model) checkWallpaperEngine() {
	if runtime.GOOS != "darwin" {
		return
	}

	appPath := "/Applications/WallpaperEngine.app"
	if _, err := os.Stat(appPath); err == nil {
		m.wallpaperEngineInstalled = true
		m.showMacHint = true
	}
}

// createWallpaperItem creates a list item from a wallpaper path
func createWallpaperItem(path string, cache *thumbs.Cache) WallpaperItem {
	name := filepath.Base(path)
	
	// Determine source from path
	source := "unknown"
	if strings.Contains(path, "wallhaven") {
		source = "wallhaven"
	} else if strings.Contains(path, "reddit") {
		source = "reddit"
	}

	// Check for cached thumbnail
	thumbnail, _ := cache.Get(path)

	return WallpaperItem{
		Path:      path,
		Name:      name,
		Source:    source,
		Thumbnail: thumbnail,
	}
}

// detectImageMethod detects the best available image rendering method
func detectImageMethod() ImageMethod {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	// Check for Kitty
	if term == "xterm-kitty" || os.Getenv("KITTY_WINDOW_ID") != "" {
		return MethodKitty
	}

	// Check for iTerm2
	if termProgram == "iTerm.app" {
		return MethodIterm2
	}

	// Check for SIXEL support
	if hasSixelSupport() {
		return MethodSixel
	}

	// Default to half-blocks (universal)
	return MethodHalfBlocks
}

// hasSixelSupport checks if terminal supports SIXEL
func hasSixelSupport() bool {
	// Check common SIXEL-capable terminals
	term := os.Getenv("TERM")
	sixelTerms := []string{"mlterm", "yaft", "xterm-256color"}
	for _, t := range sixelTerms {
		if strings.Contains(term, t) {
			return true
		}
	}
	return false
}

// methodName returns the display name for an image method
func methodName(method ImageMethod) string {
	switch method {
	case MethodKitty:
		return "Kitty Graphics Protocol"
	case MethodIterm2:
		return "iTerm2 Inline Images"
	case MethodSixel:
		return "SIXEL"
	case MethodHalfBlocks:
		return "Half-Blocks (ASCII)"
	default:
		return "None"
	}
}

// newItemDelegate creates a custom list item delegate
func newItemDelegate(method ImageMethod, cache *thumbs.Cache) list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	
	// Customize styles
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4"))

	return delegate
}

// Message types
type thumbnailMsg struct {
	index     int
	thumbnail string
}

type setWallpaperMsg struct {
	path string
	name string
	err  error
}
type openAppMsg struct {
	err error
}
