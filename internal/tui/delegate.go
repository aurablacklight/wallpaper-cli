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
		thumbnailSize: thumbs.ThumbnailSizeSmall, // 128x128
		styles: DelegateStyles{
			NormalTitle: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				PaddingLeft(2),
			NormalDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				PaddingLeft(2),
			SelectedTitle: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				PaddingLeft(2),
			SelectedDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				PaddingLeft(2),
			ThumbnailNormal: lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder()).
				PaddingRight(1),
			ThumbnailSelected: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				PaddingRight(1),
			PlaceholderStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Background(lipgloss.Color("#333333")).
				Width(10).
				Align(lipgloss.Center),
		},
	}
}

// Height returns the height of the list item
func (d ThumbnailDelegate) Height() int {
	return 5 // 3 lines for text + padding for thumbnail
}

// Spacing returns the spacing between items
func (d ThumbnailDelegate) Spacing() int {
	return 1
}

// Update is called when the item is updated
func (d ThumbnailDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders the list item with thumbnail
func (d ThumbnailDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(WallpaperItem)
	if !ok {
		return
	}

	// Determine if this item is selected
	isSelected := index == m.Index()

	// Get or generate thumbnail
	thumbPath := d.getThumbnail(item.Path)

	// Render thumbnail
	var thumbStr string
	if thumbPath != "" {
		thumbStr = d.renderThumbnailImage(thumbPath, isSelected)
	} else {
		thumbStr = d.renderThumbnailPlaceholder(isSelected)
	}

	// Prepare text content
	title := item.Title()
	desc := item.Description()

	// Apply styles based on selection
	var titleStr, descStr string
	if isSelected {
		titleStr = d.styles.SelectedTitle.Render(title)
		descStr = d.styles.SelectedDesc.Render(desc)
		// Add selection indicator
		titleStr = "▸ " + titleStr
	} else {
		titleStr = d.styles.NormalTitle.Render("  " + title)
		descStr = d.styles.NormalDesc.Render("  " + desc)
	}

	// Combine thumbnail and text side by side
	// Thumbnail is 128px wide, so we need to position it properly
	lines := strings.Split(thumbStr, "\n")
	textLines := []string{titleStr, descStr, ""}
	
	var output strings.Builder
	maxLines := max(len(lines), len(textLines))
	
	for i := 0; i < maxLines; i++ {
		var line strings.Builder
		
		// Add thumbnail line if available
		if i < len(lines) {
			line.WriteString(lines[i])
		} else {
			// Padding for alignment
			line.WriteString(strings.Repeat(" ", 12))
		}
		
		// Add text line if available
		if i < len(textLines) {
			line.WriteString(" ")
			line.WriteString(textLines[i])
		}
		
		output.WriteString(line.String())
		if i < maxLines-1 {
			output.WriteString("\n")
		}
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

// renderThumbnailImage renders an actual thumbnail image
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

	// Set dimensions (in characters for terminal)
	// 128px roughly equals 16 characters wide, 8 lines tall
	widget.SetSize(16, 8)

	rendered, err := widget.Render()
	if err != nil {
		// Failed to render, use placeholder
		return d.renderThumbnailPlaceholder(isSelected)
	}

	return rendered
}

// renderThumbnailPlaceholder renders a placeholder when image can't be displayed
func (d ThumbnailDelegate) renderThumbnailPlaceholder(isSelected bool) string {
	placeholder := "[📷 128px]"
	
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
