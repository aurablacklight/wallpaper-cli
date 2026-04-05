package sources

import (
	"context"

	"github.com/user/wallpaper-cli/internal/model"
)

// Source is the common interface for all wallpaper sources.
type Source interface {
	Name() string
	Search(ctx context.Context, params *SearchParams) (*SearchResult, error)
	Capabilities() *Capabilities
}

// SearchParams holds parameters common to all source searches.
type SearchParams struct {
	Tags        string // raw tag string (comma-separated or space-separated depending on source)
	Resolution  string // normalized, e.g. "3840x2160"
	AspectRatio string // normalized, e.g. "16x9"
	Sorting     string // top, hot, new, random, favorites, views, latest
	TimePeriod  string // 1d, 7d, 30d, 1y, all
	Limit       int
	Page        int
	Purity      string // sfw, sketchy, nsfw
	AnimeOnly   bool
}

// SearchResult holds the results of a source search.
type SearchResult struct {
	Wallpapers []ResultWallpaper
	Tags       []model.Tag
	Page       int
	TotalPages int
	Total      int
}

// ResultWallpaper is a wallpaper returned from a source search.
type ResultWallpaper struct {
	ID          string
	SourceName  string
	SourceID    string
	URL         string // direct download URL
	Title       string
	Resolution  string
	AspectRatio string
	Tags        []string
	FileSize    int64
	Format      string // jpg, png, webp
	Purity      string
	Category    string
}

// Capabilities describes what a source supports.
type Capabilities struct {
	Name               string   `json:"name"`
	SupportsResolution bool     `json:"supports_resolution"`
	SupportsAspectRatio bool    `json:"supports_aspect_ratio"`
	SupportsTags       bool     `json:"supports_tags"`
	MaxTags            int      `json:"max_tags"`             // 0 = unlimited
	SupportedSorting   []string `json:"supported_sorting"`
	SupportsTimePeriod bool     `json:"supports_time_period"`
	RequiresAuth       bool     `json:"requires_auth"`
	AuthOptional       bool     `json:"auth_optional"`        // auth improves results but isn't required
	AuthFields         []string `json:"auth_fields,omitempty"`
	RateLimit          string   `json:"rate_limit"`           // human-readable description
}
