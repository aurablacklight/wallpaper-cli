package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/wallpaper-cli/internal/config"
	"github.com/user/wallpaper-cli/internal/platform"
	"github.com/user/wallpaper-cli/internal/thumbs"
)

// LayoutType represents the current responsive layout
type LayoutType int

const (
	LayoutStacked  LayoutType = iota // < 80 cols
	LayoutCompact                    // 80-100 cols
	LayoutStandard                   // 100-140 cols
	LayoutWide                       // > 140 cols
)

// PaneFocus represents which pane has focus
type PaneFocus int

const (
	FocusList PaneFocus = iota
	FocusPreview
	FocusModal
)

// RootModel is the top-level model implementing tree-of-models architecture
type RootModel struct {
	// Terminal dimensions
	width  int
	height int
	layout LayoutType
	focus  PaneFocus

	// Child models (tree structure)
	listModel    *ListPaneModel
	previewModel *PreviewPaneModel
	modalModel   tea.Model // Dynamic modal (rating, playlist selector)
	statusModel  *StatusBarModel

	// State
	showHelp  bool
	showModal bool
	modalType ModalType

	// Dependencies
	cache  *thumbs.Cache
	cfg    *config.Config
	setter platform.Setter

	// macOS hint
	showMacHint              bool
	wallpaperEngineInstalled bool
	hintDismissed            bool
}

// ModalType represents the type of modal to show
type ModalType int

const (
	ModalNone ModalType = iota
	ModalRating
	ModalPlaylist
	ModalPlaylistCreate
)

// NewRootModel creates the new tree-of-models root
func NewRootModel(wallpapers []string, cfg *config.Config, setter platform.Setter) (*RootModel, error) {
	// Initialize thumbnail cache
	cache, err := thumbs.NewCache()
	if err != nil {
		return nil, fmt.Errorf("initializing thumbnail cache: %w", err)
	}

	// Create list pane
	listModel, err := NewListPaneModel(wallpapers, cache)
	if err != nil {
		return nil, err
	}

	// Create preview pane
	previewModel := NewPreviewPaneModel(cache)

	// Create status bar
	statusModel := NewStatusBarModel()

	m := &RootModel{
		listModel:    listModel,
		previewModel: previewModel,
		statusModel:  statusModel,
		cache:        cache,
		cfg:          cfg,
		setter:       setter,
		focus:        FocusList,
		layout:       LayoutStandard,
	}

	// Check for macOS WallpaperEngine
	if runtime.GOOS == "darwin" {
		m.checkWallpaperEngine()
	}

	return m, nil
}

// Init initializes the model
func (m *RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.listModel.Init(),
		m.preloadInitialThumbnails(),
	)
}

// Update handles messages using tree-of-models pattern
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// Handle window resize - update all child models
	case tea.WindowSizeMsg:
		m.handleWindowResize(msg.Width, msg.Height)

		// Propagate to children
		listCmd := m.updateChildModel(m.listModel, msg)
		previewCmd := m.updateChildModel(m.previewModel, msg)
		statusCmd := m.updateChildModel(m.statusModel, msg)

		cmds = append(cmds, listCmd, previewCmd, statusCmd)

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Help toggle (global)
		if msg.String() == "?" && !m.showModal {
			m.showHelp = !m.showHelp
			return m, nil
		}

		// Handle modal first
		if m.showModal && m.modalModel != nil {
			if msg.String() == "esc" || msg.String() == "q" {
				m.closeModal()
				return m, nil
			}

			updatedModal, cmd := m.modalModel.Update(msg)
			m.modalModel = updatedModal
			return m, cmd
		}

		// Handle help overlay
		if m.showHelp {
			if msg.String() != "q" && msg.String() != "esc" {
				m.showHelp = false
				return m, nil
			}
		}

		// Global key handlers
		switch msg.String() {
		case "tab":
			// Cycle focus: List → Preview (modal handled separately)
			m.focus = (m.focus + 1) % 2
			return m, nil

		case "f":
			// Toggle favorite for selected item
			return m, m.toggleFavorite()

		case "r":
			// Open rating modal
			m.openRatingModal()
			return m, nil

		case "p":
			// Open playlist modal
			m.openPlaylistModal()
			return m, nil

		case "P":
			// Open create playlist modal
			m.openCreatePlaylistModal()
			return m, nil

		case "enter":
			// Set wallpaper
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

		case "q", "esc":
			return m, tea.Quit
		}

		// Route to focused child model
		switch m.focus {
		case FocusList:
			updated, cmd := m.listModel.Update(msg)
			m.listModel = updated.(*ListPaneModel)

			// Update preview when selection changes
			if newItem := m.listModel.SelectedItem(); newItem != nil {
				previewCmd := m.previewModel.SetItem(newItem)
				cmds = append(cmds, cmd, previewCmd)
			} else {
				cmds = append(cmds, cmd)
			}

		case FocusPreview:
			updated, cmd := m.previewModel.Update(msg)
			m.previewModel = updated.(*PreviewPaneModel)
			cmds = append(cmds, cmd)
		}

	default:
		// Pass to all children
		listCmd := m.updateChildModel(m.listModel, msg)
		previewCmd := m.updateChildModel(m.previewModel, msg)
		cmds = append(cmds, listCmd, previewCmd)
	}

	return m, tea.Batch(cmds...)
}

// updateChildModel updates a child model and returns its command
func (m *RootModel) updateChildModel(child tea.Model, msg tea.Msg) tea.Cmd {
	if child == nil {
		return nil
	}
	updated, cmd := child.Update(msg)

	// Update the reference based on type
	switch updated.(type) {
	case *ListPaneModel:
		if m.listModel != nil {
			*m.listModel = *(updated.(*ListPaneModel))
		}
	case *PreviewPaneModel:
		if m.previewModel != nil {
			*m.previewModel = *(updated.(*PreviewPaneModel))
		}
	case *StatusBarModel:
		if m.statusModel != nil {
			*m.statusModel = *(updated.(*StatusBarModel))
		}
	}

	return cmd
}

// View renders the split-pane layout
func (m *RootModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	// Calculate dimensions based on layout
	leftWidth, rightWidth := m.calculatePaneWidths()
	height := m.height - 2 // Reserve space for status bar

	// Render children with proper dimensions
	m.listModel.SetDimensions(leftWidth, height)
	m.previewModel.SetDimensions(rightWidth, height)

	// Get rendered views
	leftView := m.listModel.View()
	rightView := m.previewModel.View()

	// Join horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)

	// Add status bar
	statusBar := m.statusModel.View()

	return content + "\n" + statusBar
}

// calculatePaneWidths determines left/right pane widths based on layout
func (m *RootModel) calculatePaneWidths() (int, int) {
	switch m.layout {
	case LayoutStacked:
		// No split - full width for both (stacked vertically conceptually)
		return m.width, m.width
	case LayoutCompact:
		return int(float64(m.width) * 0.5), int(float64(m.width) * 0.5)
	case LayoutStandard:
		return int(float64(m.width) * 0.45), int(float64(m.width) * 0.55)
	case LayoutWide:
		return int(float64(m.width) * 0.4), int(float64(m.width) * 0.6)
	default:
		return m.width / 2, m.width / 2
	}
}

// handleWindowResize updates layout based on terminal size
func (m *RootModel) handleWindowResize(width, height int) {
	m.width = width
	m.height = height

	// Determine layout type
	switch {
	case width < 80 || height < 24:
		m.layout = LayoutStacked
	case width < 100:
		m.layout = LayoutCompact
	case width < 140:
		m.layout = LayoutStandard
	default:
		m.layout = LayoutWide
	}

	// Update children with responsive settings
	m.listModel.SetLayout(m.layout)
	m.previewModel.SetLayout(m.layout)
}

// openRatingModal creates a rating selector modal
func (m *RootModel) openRatingModal() {
	selected := m.listModel.SelectedItem()
	if selected == nil {
		return
	}

	m.modalType = ModalRating
	m.modalModel = NewRatingModal(selected.Name, 0) // 0 = no current rating
	m.showModal = true
	m.focus = FocusModal
}

// openPlaylistModal creates a playlist selector modal
func (m *RootModel) openPlaylistModal() {
	selected := m.listModel.SelectedItem()
	if selected == nil {
		return
	}

	m.modalType = ModalPlaylist
	m.modalModel = NewPlaylistModal(selected.Name)
	m.showModal = true
	m.focus = FocusModal
}

// openCreatePlaylistModal creates a new playlist modal
func (m *RootModel) openCreatePlaylistModal() {
	m.modalType = ModalPlaylistCreate
	m.modalModel = NewCreatePlaylistModal()
	m.showModal = true
	m.focus = FocusModal
}

// closeModal closes the current modal
func (m *RootModel) closeModal() {
	m.showModal = false
	m.modalModel = nil
	m.modalType = ModalNone
	m.focus = FocusList // Return focus to list
}

// handleModalResult processes modal completion
func (m *RootModel) handleModalResult(result modalResultMsg) {
	switch result.Type {
	case "rating":
		// Update rating for selected item
		if item := m.listModel.SelectedItem(); item != nil {
			m.listModel.SetItemRating(item.Path, result.Value)
		}

	case "playlist":
		// Add to playlist
		if item := m.listModel.SelectedItem(); item != nil {
			m.statusModel.SetMessage(fmt.Sprintf("Added to playlist: %s", result.Text))
		}

	case "playlist_create":
		// Create new playlist
		m.statusModel.SetMessage(fmt.Sprintf("Created playlist: %s", result.Text))
	}
}

// toggleFavorite toggles favorite status for selected item
func (m *RootModel) toggleFavorite() tea.Cmd {
	item := m.listModel.SelectedItem()
	if item == nil {
		return nil
	}

	return func() tea.Msg {
		isFav := m.listModel.ToggleFavorite(item.Path)
		return favoriteToggleMsg{path: item.Path, isFavorite: isFav}
	}
}

// setSelectedWallpaper sets the currently selected wallpaper
func (m *RootModel) setSelectedWallpaper() tea.Cmd {
	item := m.listModel.SelectedItem()
	if item == nil {
		return nil
	}

	return func() tea.Msg {
		err := m.setter.SetWallpaper(item.Path)
		if err == nil {
			m.cfg.AddWallpaper(item.Path, "manual")
			m.cfg.Save(config.GetConfigPath())
		}
		return setWallpaperMsg{path: item.Path, name: item.Name, err: err}
	}
}

// openWallpaperEngine opens the WallpaperEngine app on macOS
func (m *RootModel) openWallpaperEngine() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("open", "-a", "WallpaperEngine")
		err := cmd.Run()
		return openAppMsg{err: err}
	}
}

// checkWallpaperEngine checks if WallpaperEngine.app is installed
func (m *RootModel) checkWallpaperEngine() {
	if runtime.GOOS != "darwin" {
		return
	}

	appPath := "/Applications/WallpaperEngine.app"
	if _, err := os.Stat(appPath); err == nil {
		m.wallpaperEngineInstalled = true
		m.showMacHint = true
	}
}

// preloadInitialThumbnails preloads first batch
func (m *RootModel) preloadInitialThumbnails() tea.Cmd {
	return m.listModel.Init()
}

// renderHelp renders the help overlay
func (m *RootModel) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(2).
		Width(70)

	help := `
Keyboard Shortcuts:

  Navigation:
    j / ↓          Navigate down in list
    k / ↑          Navigate up in list
    Tab            Switch focus (list ↔ preview)
    Enter          Set selected wallpaper

  Collections:
    f              Toggle favorite ⭐
    r              Rate wallpaper (1-5 stars)
    p              Add to playlist
    P              Create new playlist

  Search & Control:
    /              Enter search mode
    n              Load next 10 wallpapers
    d              Start/stop daemon rotation

  General:
    ?              Toggle help (this screen)
    q / Esc / Ctrl+C   Quit

Press any key to close help...
`

	return helpStyle.Render(help)
}

// Message types
type modalResultMsg struct {
	Type  string // "rating", "playlist", "playlist_create"
	Value int
	Text  string
}

type favoriteToggleMsg struct {
	path       string
	isFavorite bool
}

type setWallpaperMsg struct {
	path string
	name string
	err  error
}

type openAppMsg struct {
	err error
}
