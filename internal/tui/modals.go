package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RatingModal allows users to rate a wallpaper 1-5 stars
type RatingModal struct {
	wallpaperName string
	currentRating int
	selected      int
}

// NewRatingModal creates a new rating modal
func NewRatingModal(name string, currentRating int) tea.Model {
	return &RatingModal{
		wallpaperName: name,
		currentRating: currentRating,
		selected:      currentRating,
	}
}

// Init initializes the modal
func (m *RatingModal) Init() tea.Cmd {
	return nil
}

// Update handles input
func (m *RatingModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4", "5":
			// Set rating
			m.selected = int(msg.String()[0] - '0')
			return m, func() tea.Msg {
				return modalResultMsg{Type: "rating", Value: m.selected}
			}

		case "enter":
			// Confirm selection
			if m.selected > 0 {
				return m, func() tea.Msg {
					return modalResultMsg{Type: "rating", Value: m.selected}
				}
			}

		case "esc", "q":
			// Cancel
			return m, func() tea.Msg {
				return modalResultMsg{Type: "rating", Value: 0}
			}
		}
	}
	return m, nil
}

// View renders the modal
func (m *RatingModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6366F1")).
		Padding(2).
		Width(50)

	title := lipgloss.NewStyle().Bold(true).Render("Rate Wallpaper")
	name := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(m.wallpaperName)

	// Render stars
	stars := ""
	for i := 1; i <= 5; i++ {
		if i <= m.selected {
			stars += "★ "
		} else {
			stars += "☆ "
		}
	}
	starStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render(stars)

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("[1-5] Select | [Enter] Confirm | [Esc] Cancel")

	content := title + "\n\n" + name + "\n\n" + starStyle + "\n\n" + help
	return style.Render(content)
}

// PlaylistModal allows users to select a playlist to add to
type PlaylistModal struct {
	wallpaperName string
	playlists     []string
	selected      int
}

// NewPlaylistModal creates a playlist selection modal
func NewPlaylistModal(name string) tea.Model {
	// For now, return empty playlists - will be populated from collections manager
	return &PlaylistModal{
		wallpaperName: name,
		playlists:     []string{"favorites", "cozy", "energetic"}, // Placeholder
		selected:      0,
	}
}

// Init initializes the modal
func (m *PlaylistModal) Init() tea.Cmd {
	return nil
}

// Update handles input
func (m *PlaylistModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.selected < len(m.playlists)-1 {
				m.selected++
			}

		case "k", "up":
			if m.selected > 0 {
				m.selected--
			}

		case "enter":
			if len(m.playlists) > 0 {
				return m, func() tea.Msg {
					return modalResultMsg{Type: "playlist", Text: m.playlists[m.selected]}
				}
			}

		case "esc", "q":
			return m, func() tea.Msg {
				return modalResultMsg{Type: "playlist", Text: ""}
			}
		}
	}
	return m, nil
}

// View renders the modal
func (m *PlaylistModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6366F1")).
		Padding(2).
		Width(40)

	title := lipgloss.NewStyle().Bold(true).Render("Add to Playlist")

	var listContent string
	for i, playlist := range m.playlists {
		marker := "  "
		if i == m.selected {
			marker = "▸ "
		}
		listContent += marker + playlist + "\n"
	}

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("[j/k] Navigate | [Enter] Select | [Esc] Cancel")

	content := title + "\n\n" + listContent + "\n" + help
	return style.Render(content)
}

// CreatePlaylistModal allows users to create a new playlist
type CreatePlaylistModal struct {
	name   string
	cursor int
}

// NewCreatePlaylistModal creates a new playlist creation modal
func NewCreatePlaylistModal() tea.Model {
	return &CreatePlaylistModal{
		name:   "",
		cursor: 0,
	}
}

// Init initializes the modal
func (m *CreatePlaylistModal) Init() tea.Cmd {
	return nil
}

// Update handles input
func (m *CreatePlaylistModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.name != "" {
				return m, func() tea.Msg {
					return modalResultMsg{Type: "playlist_create", Text: m.name}
				}
			}

		case "esc", "q":
			return m, func() tea.Msg {
				return modalResultMsg{Type: "playlist_create", Text: ""}
			}

		case "backspace":
			if len(m.name) > 0 {
				m.name = m.name[:len(m.name)-1]
			}

		default:
			if len(msg.String()) == 1 && msg.String()[0] >= 32 {
				m.name += msg.String()
			}
		}
	}
	return m, nil
}

// View renders the modal
func (m *CreatePlaylistModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6366F1")).
		Padding(2).
		Width(50)

	title := lipgloss.NewStyle().Bold(true).Render("Create New Playlist")

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(40).
		Padding(0, 1)

	input := inputStyle.Render(m.name + "_")

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("[Enter] Create | [Esc] Cancel")

	content := title + "\n\n" + input + "\n\n" + help
	return style.Render(content)
}
