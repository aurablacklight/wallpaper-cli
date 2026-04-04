package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blacktop/go-termimg"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/wallpaper-cli/internal/thumbs"
)

// ThumbnailDelegate is a custom list item delegate that renders thumbnails
type ThumbnailDelegate struct {
	imageMethod ImageMethod
	cache       *thumbs.Cache
	styles      DelegateStyles
	thumbnailSize int
}

// DelegateStyles holds the styles for the delegate
type DelegateStyles struct {
	NormalTitle       lipgloss.Style
	NormalDesc        lipgloss.Style
	SelectedTitle     lipgloss.Style
	SelectedDesc      lipgloss.Style
	ThumbnailNormal   lipgloss.Style
	ThumbnailSelected lipgloss.Style
	PlaceholderStyle  lipgloss.Style
}

// NewThumbnailDelegate creates a new thumbnail delegate
func NewThumbnailDelegate(method ImageMethod, cache *thumbs.Cache) ThumbnailDelegate {
	return ThumbnailDelegate{
		imageMethod:   method,
		cache:         cache,
		thumbnailSize: thumbs.ThumbnailSizeCompact, // 64x64 compact
		styles: DelegateStyles{
			NormalTitle: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				PaddingLeft(1),
			NormalDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				PaddingLeft(1).
				Italic(true),
			SelectedTitle: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				PaddingLeft(1),
			SelectedDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				PaddingLeft(1).
				Italic(true),
			ThumbnailNormal: lipgloss.NewStyle().
				PaddingRight(1),
			ThumbnailSelected: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				PaddingRight(1),
			PlaceholderStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Width(6).
				Align(lipgloss.Center),
		},
	}
}

// Height returns the height of the list item (compact single line)
func (d ThumbnailDelegate) Height() int {
	return 3 // Compact: thumbnail (2-3 lines) + minimal text
}

// Spacing returns the spacing between items
func (d ThumbnailDelegate) Spacing() int {
	return 0 // No extra spacing between items
}

// Update is called when the item is updated
func (d ThumbnailDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders the list item with thumbnail (compact layout)
func (d ThumbnailDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(WallpaperItem)
	if !ok {
		return
	}

	// Determine if this item is selected
	isSelected := index == m.Index()

	// Get or generate thumbnail
	thumbPath := d.getThumbnail(item.Path)

	// Render thumbnail (compact 64x64)
	var thumbStr string
	if thumbPath != "" {
		thumbStr = d.renderThumbnailImage(thumbPath, isSelected)
	} else {
		thumbStr = d.renderThumbnailPlaceholder(isSelected)
	}

	// Prepare compact text content
	var output strings.Builder
	
	// Selection indicator
	if isSelected {
		output.WriteString("▸ ")
	} else {
		output.WriteString("  ")
	}
	
	// Add thumbnail (first line only for compactness)
	thumbLines := strings.Split(thumbStr, "\n")
	if len(thumbLines) > 0 {
		output.WriteString(thumbLines[0])
		output.WriteString(" ")
	}
	
	// Compact single-line text: filename | source
	line := fmt.Sprintf("%s | %s", item.Name, item.Source)
	
	if isSelected {
		output.WriteString(d.styles.SelectedTitle.Render(line))
	} else {
		output.WriteString(d.styles.NormalTitle.Render(line))
	}

	fmt.Fprint(w, output.String())
}

// getThumbnail gets or generates a thumbnail for an image
func (d ThumbnailDelegate) getThumbnail(imagePath string) string {
	// Check if already cached
	thumbPath, exists := d.cache.GetWithSize(imagePath, d.thumbnailSize)
	if exists {
		return thumbPath
	}

	// Generate thumbnail on-demand (this might be slow, so we do it async in real implementation)
	// For now, generate synchronously
	thumbPath, err := d.cache.GenerateWithSize(imagePath, d.thumbnailSize)
	if err != nil {
		return ""
	}

	return thumbPath
}

// renderThumbnailImage renders an actual thumbnail image (compact)
func (d ThumbnailDelegate) renderThumbnailImage(thumbPath string, isSelected bool) string {
	// Check if terminal supports image rendering
	if d.imageMethod == MethodNone || d.imageMethod == MethodUnknown {
		return d.renderThumbnailPlaceholder(isSelected)
	}

	// Try to render with go-termimg using ImageWidget
	widget, err := termimg.NewImageWidgetFromFile(thumbPath)
	if err != nil {
		return d.renderThumbnailPlaceholder(isSelected)
	}

	// Set compact dimensions (64x64px = ~8 chars wide, 4 lines tall)
	widget.SetSize(8, 4)

	rendered, err := widget.Render()
	if err != nil {
		// Failed to render, use placeholder
		return d.renderThumbnailPlaceholder(isSelected)
	}

	return rendered
}

// renderThumbnailPlaceholder renders a compact placeholder
func (d ThumbnailDelegate) renderThumbnailPlaceholder(isSelected bool) string {
	placeholder := "[📷]"
	
	if isSelected {
		return d.styles.ThumbnailSelected.Render(placeholder)
	}
	return d.styles.ThumbnailNormal.Render(placeholder)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Helper to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
