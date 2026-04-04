package tui

import (
	"fmt"

	"github.com/blacktop/go-termimg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/wallpaper-cli/internal/thumbs"
)

// PreviewPaneModel manages the right pane preview and metadata
type PreviewPaneModel struct {
	cache         *thumbs.Cache
	currentItem   *ListItem
	layout        LayoutType
	width         int
	height        int
	previewHeight int
	imageMethod   ImageMethod

	// Image loading state
	loading     bool
	loadingPath string
	widget      *termimg.ImageWidget
}

// previewReadyMsg is sent when the preview image is loaded
type previewReadyMsg struct {
	path   string
	widget *termimg.ImageWidget
}

// NewPreviewPaneModel creates a new preview pane
func NewPreviewPaneModel(cache *thumbs.Cache) *PreviewPaneModel {
	return &PreviewPaneModel{
		cache:         cache,
		layout:        LayoutStandard,
		previewHeight: 150,
		imageMethod:   detectImageMethod(),
	}
}

// Init initializes the model
func (m *PreviewPaneModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *PreviewPaneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle preview-specific keys
		switch msg.String() {
		case "f":
			// Toggle favorite (delegated to root)
			return m, nil
		case "r":
			// Rate (delegated to root)
			return m, nil
		}

	case previewReadyMsg:
		// Preview image loaded
		if msg.path == m.loadingPath {
			m.loading = false
			m.widget = msg.widget
		}
		return m, nil
	}

	return m, nil
}

// View renders the preview pane
func (m *PreviewPaneModel) View() string {
	if m.currentItem == nil {
		return m.renderEmpty()
	}

	// Calculate available space
	metadataHeight := 8
	actionsHeight := 6
	availableHeight := m.height - metadataHeight - actionsHeight - 4 // padding

	// Render sections
	preview := m.renderPreview(availableHeight)
	metadata := m.renderMetadata()
	actions := m.renderActions()

	// Join vertically
	content := lipgloss.JoinVertical(lipgloss.Left,
		preview,
		"",
		metadata,
		"",
		actions,
	)

	// Style the pane - transparent background, no wallpaper bleeding
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1)

	if m.layout != LayoutStacked {
		style = style.Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4B5563"))
	}

	return style.Render(content)
}

// SetDimensions sets the pane dimensions
func (m *PreviewPaneModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height

	// Calculate preview height based on available space
	m.calculatePreviewHeight()
}

// SetLayout updates the responsive layout
func (m *PreviewPaneModel) SetLayout(layout LayoutType) {
	m.layout = layout
	m.calculatePreviewHeight()
}

// calculatePreviewHeight determines optimal preview height
func (m *PreviewPaneModel) calculatePreviewHeight() {
	// Reserve space for metadata (8) and actions (6) plus padding
	reserved := 8 + 6 + 4
	available := m.height - reserved

	switch m.layout {
	case LayoutStacked:
		m.previewHeight = min(80, available)
	case LayoutCompact:
		m.previewHeight = min(100, available)
	case LayoutStandard:
		m.previewHeight = min(150, available)
	case LayoutWide:
		m.previewHeight = min(200, available)
	}
}

// SetItem sets the current item to preview
func (m *PreviewPaneModel) SetItem(item *ListItem) tea.Cmd {
	if item == nil {
		return nil
	}

	// Check if it's the same item (avoid reload)
	if m.currentItem != nil && item.Path == m.currentItem.Path {
		return nil
	}

	m.currentItem = item
	m.loading = true
	m.loadingPath = item.Path
	m.widget = nil // Clear old widget

	// Load preview asynchronously
	return func() tea.Msg {
		// Get thumbnail for preview
		thumbPath, exists := m.cache.GetWithSize(item.Path, m.previewHeight)
		if !exists {
			var err error
			thumbPath, err = m.cache.GenerateWithSize(item.Path, m.previewHeight)
			if err != nil {
				return previewReadyMsg{path: item.Path, widget: nil}
			}
		}

		// Try to create image widget
		if thumbPath != "" {
			widget, err := termimg.NewImageWidgetFromFile(thumbPath)
			if err == nil {
				// Calculate dimensions to fit in preview area
				maxWidth := m.width - 6 // Account for padding/borders
				maxHeight := m.previewHeight

				// Set size
				widget.SetSize(maxWidth, maxHeight)

				return previewReadyMsg{path: item.Path, widget: widget}
			}
		}

		return previewReadyMsg{path: item.Path, widget: nil}
	}
}

// renderEmpty renders the empty state
func (m *PreviewPaneModel) renderEmpty() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))
	// Note: No background color - inherits from container

	return style.Render("No wallpaper selected\n\nNavigate with j/k or ↑/↓")
}

// renderPreview renders the image preview section
func (m *PreviewPaneModel) renderPreview(availableHeight int) string {
	if m.currentItem == nil {
		return ""
	}

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6366F1")).
		MarginBottom(1)

	title := titleStyle.Render("👁️  Preview")

	// Calculate dimensions
	maxWidth := m.width - 6
	maxHeight := min(m.previewHeight, availableHeight)

	// Render image content
	var imageContent string
	if m.loading {
		imageContent = lipgloss.NewStyle().
			Width(maxWidth).
			Height(maxHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4B5563")).
			Foreground(lipgloss.Color("#888888")).
			Render(fmt.Sprintf("Loading %s...", truncateString(m.currentItem.Name, 30)))
	} else if m.widget != nil {
		// Actually render the image using go-termimg
		rendered, err := m.widget.Render()
		if err != nil {
			// Terminal doesn't support images or render failed - show info panel
			imageContent = m.renderInfoPanel(maxWidth, maxHeight)
		} else {
			// Wrap the rendered image
			imageContent = rendered
		}
	} else {
		// Image loading in progress - show filename
		imageContent = lipgloss.NewStyle().
			Width(maxWidth).
			Height(maxHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4B5563")).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#888888")).
			Render(fmt.Sprintf("Loading %s...", truncateString(m.currentItem.Name, 30)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, imageContent)
}

// renderInfoPanel shows wallpaper info when terminal doesn't support image rendering
func (m *PreviewPaneModel) renderInfoPanel(width, height int) string {
	infoStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4B5563")).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#888888"))

	return infoStyle.Render(
		fmt.Sprintf("🖼️  %s\n%d x %d",
			truncateString(m.currentItem.Name, 30),
			width, height),
	)
}

// renderMetadata renders the metadata section
func (m *PreviewPaneModel) renderMetadata() string {
	if m.currentItem == nil {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6366F1"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB"))

	title := titleStyle.Render("📊 Metadata")

	// Build metadata rows
	rows := []string{title}

	// Source
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		labelStyle.Render("Source: "),
		valueStyle.Render(m.currentItem.Source),
	))

	// Filename
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		labelStyle.Render("File: "),
		valueStyle.Render(truncateString(m.currentItem.Name, m.width-12)),
	))

	// Favorite
	favStr := "No"
	if m.currentItem.IsFavorite {
		favStr = "⭐ Yes"
	}
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		labelStyle.Render("Favorite: "),
		valueStyle.Render(favStr),
	))

	// Rating
	if m.currentItem.Rating > 0 {
		ratingStr := renderStars(m.currentItem.Rating)
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Render("Rating: "),
			valueStyle.Render(ratingStr),
		))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderActions renders the actions section
func (m *PreviewPaneModel) renderActions() string {
	if m.currentItem == nil {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6366F1")).
		MarginBottom(1)

	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981"))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	title := titleStyle.Render("⌨️  Actions")

	actions := []string{
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[Enter] "), actionStyle.Render("Set as wallpaper")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[f] "), actionStyle.Render("Toggle favorite")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[r] "), actionStyle.Render("Rate 1-5 stars")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[p] "), actionStyle.Render("Add to playlist")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[P] "), actionStyle.Render("Create playlist")),
	}

	return lipgloss.JoinVertical(lipgloss.Left, append([]string{title}, actions...)...)
}

// Helper functions

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func renderStars(rating int) string {
	if rating <= 0 || rating > 5 {
		return ""
	}
	filled := "★"
	empty := "☆"

	result := ""
	for i := 0; i < 5; i++ {
		if i < rating {
			result += filled
		} else {
			result += empty
		}
	}
	return result
}
