package wallhaven

import (
	"strings"
)

// SearchBuilder helps construct search queries
type SearchBuilder struct {
	opts *SearchOptions
}

// NewSearchBuilder creates a new search builder with defaults
func NewSearchBuilder() *SearchBuilder {
	return &SearchBuilder{
		opts: DefaultSearchOptions(),
	}
}

// WithQuery sets the search query (tags)
func (b *SearchBuilder) WithQuery(query string) *SearchBuilder {
	b.opts.Query = query
	return b
}

// WithResolution adds a resolution filter
func (b *SearchBuilder) WithResolution(res string) *SearchBuilder {
	b.opts.Resolutions = append(b.opts.Resolutions, res)
	return b
}

// WithAspectRatio adds an aspect ratio filter
func (b *SearchBuilder) WithAspectRatio(ratio string) *SearchBuilder {
	// Convert "16:9" to "16x9" for API
	ratio = strings.ReplaceAll(ratio, ":", "x")
	b.opts.Ratios = append(b.opts.Ratios, ratio)
	return b
}

// WithSorting sets the sorting method
func (b *SearchBuilder) WithSorting(sorting string) *SearchBuilder {
	b.opts.Sorting = sorting
	return b
}

// WithPage sets the page number
func (b *SearchBuilder) WithPage(page int) *SearchBuilder {
	b.opts.Page = page
	return b
}

// WithPurity sets the content purity filter
func (b *SearchBuilder) WithPurity(purity string) *SearchBuilder {
	b.opts.Purity = purity
	return b
}

// WithAnimeOnly restricts to anime category only
func (b *SearchBuilder) WithAnimeOnly() *SearchBuilder {
	b.opts.Categories = "010" // Only anime bit set
	return b
}

// WithRandom uses random sorting for variety
func (b *SearchBuilder) WithRandom() *SearchBuilder {
	b.opts.Sorting = "random"
	return b
}

// WithTopRange sets the time range for top sorting
func (b *SearchBuilder) WithTopRange(topRange string) *SearchBuilder {
	b.opts.TopRange = topRange
	return b
}

// WithOrder sets the sort order
func (b *SearchBuilder) WithOrder(order string) *SearchBuilder {
	b.opts.Order = order
	return b
}

// Build returns the search options
func (b *SearchBuilder) Build() *SearchOptions {
	return b.opts
}

// ResolutionPresets maps common resolution names to API format
var ResolutionPresets = map[string]string{
	"1920x1080": "1920x1080",
	"2560x1440": "2560x1440",
	"3840x2160": "3840x2160",
	"7680x4320": "7680x4320",
}

// AspectRatioPresets maps common aspect ratios to API format
var AspectRatioPresets = map[string]string{
	"16x9":  "16x9",
	"21x9":  "21x9",
	"32x9":  "32x9",
	"4x3":   "4x3",
	"1x1":   "1x1",
	"16x10": "16x10",
}

// ParseTags converts comma-separated tags to query format
func ParseTags(tags string) string {
	// Wallhaven uses spaces for AND, + for OR
	// We use comma as separator, convert to space
	tags = strings.ReplaceAll(tags, ",", " ")
	return strings.TrimSpace(tags)
}
