package tui

import (
	"fmt"
)

const (
	DefaultPageSize = 10
)

// Paginator manages loading wallpapers in batches
type Paginator struct {
	allWallpapers []string
	loadedCount   int
	pageSize      int
	hasMore       bool
}

// NewPaginator creates a new paginator
func NewPaginator(wallpapers []string) *Paginator {
	return &Paginator{
		allWallpapers: wallpapers,
		loadedCount:   0,
		pageSize:      DefaultPageSize,
		hasMore:       len(wallpapers) > DefaultPageSize,
	}
}

// GetNextBatch returns the next batch of wallpapers to display
func (p *Paginator) GetNextBatch() []string {
	start := p.loadedCount
	end := start + p.pageSize
	
	if end > len(p.allWallpapers) {
		end = len(p.allWallpapers)
	}
	
	batch := p.allWallpapers[start:end]
	p.loadedCount = end
	p.hasMore = p.loadedCount < len(p.allWallpapers)
	
	return batch
}

// HasMore returns true if there are more wallpapers to load
func (p *Paginator) HasMore() bool {
	return p.hasMore
}

// RemainingCount returns how many more wallpapers are available
func (p *Paginator) RemainingCount() int {
	return len(p.allWallpapers) - p.loadedCount
}

// TotalCount returns total number of wallpapers
func (p *Paginator) TotalCount() int {
	return len(p.allWallpapers)
}

// LoadedCount returns number of loaded wallpapers
func (p *Paginator) LoadedCount() int {
	return p.loadedCount
}

// IsAtEnd returns true if we're at the last loaded item
func (p *Paginator) IsAtEnd(currentIndex int) bool {
	return currentIndex == p.loadedCount-1 && p.hasMore
}

// GetEndMessage returns the message to show at end of list
func (p *Paginator) GetEndMessage() string {
	if p.hasMore {
		return fmt.Sprintf("▼ Press 'n' to load %d more (%d remaining)", 
			min(p.pageSize, p.RemainingCount()), 
			p.RemainingCount())
	}
	return "End of collection"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
