package tui

import (
	"github.com/sahilm/fuzzy"
)

// FuzzySearcher provides fuzzy matching for wallpaper items
type FuzzySearcher struct {
	items []WallpaperItem
}

// NewFuzzySearcher creates a new fuzzy searcher
func NewFuzzySearcher(items []WallpaperItem) *FuzzySearcher {
	return &FuzzySearcher{
		items: items,
	}
}

// Search performs fuzzy search and returns matched items sorted by score
func (f *FuzzySearcher) Search(query string) []WallpaperItem {
	if query == "" {
		return f.items
	}

	// Create source strings for matching
	// Combine filename, source, and path for comprehensive matching
	sources := make([]string, len(f.items))
	for i, item := range f.items {
		sources[i] = item.Name + " " + item.Source + " " + item.Path
	}

	// Perform fuzzy match
	matches := fuzzy.Find(query, sources)

	// Convert matches back to items
	result := make([]WallpaperItem, len(matches))
	for i, match := range matches {
		result[i] = f.items[match.Index]
	}

	return result
}

// UpdateItems updates the searchable item list
func (f *FuzzySearcher) UpdateItems(items []WallpaperItem) {
	f.items = items
}

// Highlights returns the matched character positions for highlighting
func (f *FuzzySearcher) Highlights(query string, itemIndex int) []int {
	if query == "" {
		return nil
	}

	source := f.items[itemIndex].Name + " " + f.items[itemIndex].Source
	matches := fuzzy.Find(query, []string{source})

	if len(matches) > 0 {
		return matches[0].MatchedIndexes
	}
	return nil
}
