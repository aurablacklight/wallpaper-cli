package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBarModel manages the bottom status bar
type StatusBarModel struct {
	width        int
	message      string
	rotationOn   bool
	nextRotation string
	favCount     int
	totalCount   int
}

// NewStatusBarModel creates a new status bar
func NewStatusBarModel() *StatusBarModel {
	return &StatusBarModel{
		width:        80,
		message:      "Welcome! Press ? for help",
		rotationOn:   false,
		nextRotation: "--",
		favCount:     0,
		totalCount:   0,
	}
}

// Init initializes the model
func (m *StatusBarModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *StatusBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case setWallpaperMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("❌ Error: %v", msg.err)
		} else {
			m.message = fmt.Sprintf("✅ Set: %s", truncateFilename(msg.name, 30))
		}

	case favoriteToggleMsg:
		if msg.isFavorite {
			m.favCount++
			m.message = "⭐ Added to favorites"
		} else {
			m.favCount--
			if m.favCount < 0 {
				m.favCount = 0
			}
			m.message = "Removed from favorites"
		}

	case openAppMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("⚠️ Could not open app: %v", msg.err)
		} else {
			m.message = "🖼️  Opening WallpaperEngine.app..."
		}
	}

	return m, nil
}

// View renders the status bar
func (m *StatusBarModel) View() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1).
		Width(m.width)

	var parts []string

	// Left: Message or tip
	msgStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF"))
	parts = append(parts, msgStyle.Render(m.message))

	// Center: Stats
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))
	stats := fmt.Sprintf("⭐ %d favorites | 📁 %d total", m.favCount, m.totalCount)
	parts = append(parts, statsStyle.Render(stats))

	// Right: Rotation status
	rotationStyle := lipgloss.NewStyle()
	if m.rotationOn {
		rotationStyle = rotationStyle.Foreground(lipgloss.Color("#10B981"))
		parts = append(parts, rotationStyle.Render(fmt.Sprintf("🔄 ON | Next: %s", m.nextRotation)))
	} else {
		rotationStyle = rotationStyle.Foreground(lipgloss.Color("#6B7280"))
		parts = append(parts, rotationStyle.Render("🔄 OFF"))
	}

	return style.Render(strings.Join(parts, " | "))
}

// SetMessage sets the status message
func (m *StatusBarModel) SetMessage(msg string) {
	m.message = msg
}

// SetStats updates the wallpaper stats
func (m *StatusBarModel) SetStats(fav, total int) {
	m.favCount = fav
	m.totalCount = total
}

// SetRotation updates rotation status
func (m *StatusBarModel) SetRotation(on bool, next string) {
	m.rotationOn = on
	m.nextRotation = next
}

// Helper function
func truncateFilename(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-3] + "..."
}
