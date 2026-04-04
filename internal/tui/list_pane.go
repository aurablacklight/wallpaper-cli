package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/wallpaper-cli/internal/thumbs"
)

// thumbnailMsg is sent when a thumbnail is generated
type thumbnailMsg struct {
	index     int
	thumbnail string
}

// ListPaneModel manages the left pane wallpaper list
type ListPaneModel struct {
	list          list.Model
	delegate      *ThumbnailDelegate
	cache         *thumbs.Cache
	imageMethod   ImageMethod
	layout        LayoutType
	width         int
	height        int
	thumbnailSize int

	// Data
	items         []ListItem
	filteredItems []ListItem
	paginator     *Paginator
	allPaths      []string

	// Search
	searchMode  bool
	searchQuery string
	searcher    *FuzzySearcher

	// Collections (M003)
	favorites map[string]bool
	ratings   map[string]int
}

// ListItem represents a wallpaper in the list
type ListItem struct {
	Path       string
	Name       string
	Source     string
	Thumbnail  string
	IsFavorite bool
	Rating     int // 0-5
	HasError   bool
}

// Title returns the item title for the list
func (w ListItem) Title() string {
	return w.Name
}

// Description returns the item description
func (w ListItem) Description() string {
	if w.HasError {
		return "⚠️ Error loading thumbnail"
	}

	// Build description with metadata
	var parts []string
	if w.IsFavorite {
		parts = append(parts, "⭐")
	}
	if w.Rating > 0 {
		parts = append(parts, strings.Repeat("★", w.Rating))
	}
	parts = append(parts, fmt.Sprintf("Source: %s", w.Source))

	return strings.Join(parts, " ")
}

// FilterValue returns the value to filter on
func (w ListItem) FilterValue() string {
	return w.Name
}

// NewListPaneModel creates a new list pane
func NewListPaneModel(wallpaperPaths []string, cache *thumbs.Cache) (*ListPaneModel, error) {
	imageMethod := detectImageMethod()

	// Setup pagination
	paginator := NewPaginator(wallpaperPaths)
	firstBatch := paginator.GetNextBatch()

	// Create items
	items := make([]list.Item, len(firstBatch))
	listItems := make([]ListItem, len(firstBatch))

	for i, path := range firstBatch {
		item := createListItem(path, cache)
		items[i] = item
		listItems[i] = item
	}

	// Setup list with custom delegate
	delegate := NewThumbnailDelegate(imageMethod, cache)
	l := list.New(items, delegate, 0, 0)
	l.Title = fmt.Sprintf("📁 Wallpapers (%d/%d)", len(items), len(wallpaperPaths))
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false) // We use custom fuzzy search

	// Style the title - NO background color (transparent)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6366F1")).
		Padding(0, 1)

	// Setup fuzzy searcher
	searchableItems := make([]FuzzySearchable, len(listItems))
	for i := range listItems {
		searchableItems[i] = listItems[i]
	}
	searcher := NewFuzzySearcher(searchableItems)

	return &ListPaneModel{
		list:          l,
		delegate:      &delegate,
		cache:         cache,
		imageMethod:   imageMethod,
		layout:        LayoutStandard,
		thumbnailSize: 64,
		items:         listItems,
		filteredItems: listItems,
		paginator:     paginator,
		allPaths:      wallpaperPaths,
		searcher:      searcher,
		favorites:     make(map[string]bool),
		ratings:       make(map[string]int),
	}, nil
}

// Init initializes the model
func (m *ListPaneModel) Init() tea.Cmd {
	return m.preloadThumbnails()
}

// Update handles messages
func (m *ListPaneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search mode
		if m.searchMode {
			switch keypress := msg.String(); keypress {
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.filteredItems = convertSearchResults(m.searcher.Search(""))
				m.refreshListItems()
				return m, nil

			case "enter":
				m.searchMode = false
				m.filteredItems = convertSearchResults(m.searcher.Search(m.searchQuery))
				m.refreshListItems()
				return m, nil

			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filteredItems = convertSearchResults(m.searcher.Search(m.searchQuery))
					m.refreshListItems()
				}
				return m, nil

			default:
				// Add character to search
				if len(keypress) == 1 && keypress[0] >= 32 && keypress[0] <= 126 {
					m.searchQuery += keypress
					m.filteredItems = convertSearchResults(m.searcher.Search(m.searchQuery))
					m.refreshListItems()
				}
				return m, nil
			}
		}

		// Normal mode
		switch msg.String() {
		case "/":
			m.searchMode = true
			m.searchQuery = ""
			return m, nil

		case "n":
			// Load next batch
			if m.paginator != nil && m.paginator.HasMore() {
				return m, m.loadNextBatch()
			}
		}

	case thumbnailMsg:
		m.updateItemThumbnail(msg.index, msg.thumbnail)
		return m, nil

	case setWallpaperMsg:
		// Wallpaper was set, could update UI here
		return m, nil
	}

	// Update list
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the list pane
func (m *ListPaneModel) View() string {
	// Update delegate with current thumbnail size
	m.delegate.SetThumbnailSize(m.thumbnailSize)

	// Style based on layout - transparent background
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height)

	if m.layout != LayoutStacked {
		// Add right border for split-pane
		style = style.Border(lipgloss.RoundedBorder()).BorderRight(true)
	}

	content := m.list.View()

	// Add search indicator if in search mode
	if m.searchMode {
		searchIndicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6366F1")).
			Bold(true).
			Render(fmt.Sprintf("🔍 Search: %s_ | Enter: apply | Esc: cancel", m.searchQuery))
		content = searchIndicator + "\n" + content
	}

	return style.Render(content)
}

// SetDimensions sets the pane dimensions
func (m *ListPaneModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-3) // Reserve space for status
}

// SetLayout updates the responsive layout
func (m *ListPaneModel) SetLayout(layout LayoutType) {
	m.layout = layout

	// Update thumbnail size based on layout
	switch layout {
	case LayoutStacked:
		m.thumbnailSize = 32
	case LayoutCompact:
		m.thumbnailSize = 48
	case LayoutStandard:
		m.thumbnailSize = 48
	case LayoutWide:
		m.thumbnailSize = 64
	}
}

// SelectedItem returns the currently selected item
func (m *ListPaneModel) SelectedItem() *ListItem {
	if m.list.SelectedItem() == nil {
		return nil
	}

	item := m.list.SelectedItem().(ListItem)
	return &item
}

// ToggleFavorite toggles favorite status for an item
func (m *ListPaneModel) ToggleFavorite(path string) bool {
	current := m.favorites[path]
	m.favorites[path] = !current

	// Update the item
	for i, item := range m.items {
		if item.Path == path {
			m.items[i].IsFavorite = !current
			break
		}
	}

	m.refreshListItems()
	return !current
}

// SetItemRating sets the rating for an item
func (m *ListPaneModel) SetItemRating(path string, rating int) {
	m.ratings[path] = rating

	// Update the item
	for i, item := range m.items {
		if item.Path == path {
			m.items[i].Rating = rating
			break
		}
	}

	m.refreshListItems()
}

// refreshListItems updates the list with current items
func (m *ListPaneModel) refreshListItems() {
	// Convert items to list items
	items := make([]list.Item, len(m.filteredItems))
	for i, item := range m.filteredItems {
		// Update favorite and rating from our maps
		if fav, ok := m.favorites[item.Path]; ok {
			item.IsFavorite = fav
		}
		if rating, ok := m.ratings[item.Path]; ok {
			item.Rating = rating
		}
		items[i] = item
	}

	m.list.SetItems(items)

	// Update title
	if m.searchMode {
		m.list.Title = fmt.Sprintf("🔍 Search Results (%d matches)", len(m.filteredItems))
	} else {
		m.list.Title = fmt.Sprintf("📁 Wallpapers (%d/%d)",
			m.paginator.LoadedCount(), m.paginator.TotalCount())
	}
}

// preloadThumbnails pre-generates thumbnails
func (m *ListPaneModel) preloadThumbnails() tea.Cmd {
	return func() tea.Msg {
		items := m.list.Items()
		count := 20
		if len(items) < count {
			count = len(items)
		}

		for i := 0; i < count; i++ {
			item := items[i].(ListItem)
			if thumb, err := m.cache.Generate(item.Path); err == nil {
				return thumbnailMsg{index: i, thumbnail: thumb}
			}
		}
		return nil
	}
}

// updateItemThumbnail updates an item with its thumbnail
func (m *ListPaneModel) updateItemThumbnail(index int, thumbnail string) {
	items := m.list.Items()
	if index >= len(items) {
		return
	}

	item := items[index].(ListItem)
	item.Thumbnail = thumbnail

	// Update in our tracking
	for i, it := range m.items {
		if it.Path == item.Path {
			m.items[i].Thumbnail = thumbnail
			break
		}
	}
}

// loadNextBatch loads more wallpapers for pagination
func (m *ListPaneModel) loadNextBatch() tea.Cmd {
	return func() tea.Msg {
		if m.paginator == nil || !m.paginator.HasMore() {
			return nil
		}

		batch := m.paginator.GetNextBatch()

		for _, path := range batch {
			item := createListItem(path, m.cache)
			m.items = append(m.items, item)
			m.list.InsertItem(len(m.list.Items()), item)
		}

		// Update fuzzy searcher
		m.searcher.UpdateItems(convertToSearchable(m.items))

		// Refresh title
		m.list.Title = fmt.Sprintf("📁 Wallpapers (%d/%d)",
			m.paginator.LoadedCount(), m.paginator.TotalCount())

		return nil
	}
}

// createListItem creates a list item from a path
func createListItem(path string, cache *thumbs.Cache) ListItem {
	name := filepath.Base(path)

	source := "unknown"
	if strings.Contains(path, "wallhaven") {
		source = "wallhaven"
	} else if strings.Contains(path, "reddit") {
		source = "reddit"
	}

	thumbnail, _ := cache.Get(path)

	return ListItem{
		Path:      path,
		Name:      name,
		Source:    source,
		Thumbnail: thumbnail,
	}
}

// convertSearchResults converts []FuzzySearchable to []ListItem
func convertSearchResults(results []FuzzySearchable) []ListItem {
	items := make([]ListItem, len(results))
	for i, r := range results {
		if item, ok := r.(ListItem); ok {
			items[i] = item
		}
	}
	return items
}

// convertToSearchable converts []ListItem to []FuzzySearchable
func convertToSearchable(items []ListItem) []FuzzySearchable {
	result := make([]FuzzySearchable, len(items))
	for i := range items {
		result[i] = items[i]
	}
	return result
}
