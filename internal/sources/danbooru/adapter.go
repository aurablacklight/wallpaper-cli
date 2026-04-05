package danbooru

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/wallpaper-cli/internal/sources"
)

func init() {
	sources.Register("danbooru", func(cfg map[string]string) (sources.Source, error) {
		return NewAdapter(cfg["login"], cfg["api_key"]), nil
	})
}

type Adapter struct {
	client *Client
}

func NewAdapter(login, apiKey string) *Adapter {
	return &Adapter{client: NewClient(login, apiKey)}
}

func (a *Adapter) Name() string { return "danbooru" }

func (a *Adapter) Search(ctx context.Context, params *sources.SearchParams) (*sources.SearchResult, error) {
	// Build tag query
	tagQuery := strings.ReplaceAll(params.Tags, ",", " ")
	tagQuery = strings.TrimSpace(tagQuery)

	// Check anonymous tag limit
	tagCount := len(strings.Fields(tagQuery))
	if tagCount > MaxAnonTags && !a.client.IsAuthenticated() {
		return nil, fmt.Errorf("danbooru limits anonymous searches to %d tags (you used %d) — set danbooru login and api_key in config to bypass this limit", MaxAnonTags, tagCount)
	}

	// Add sorting
	switch params.Sorting {
	case "top", "popular":
		tagQuery += " order:score"
	case "favorites":
		tagQuery += " order:favcount"
	case "latest", "new":
		tagQuery += " order:id_desc"
	case "random":
		tagQuery += " order:random"
	}

	// Add rating filter (safe by default)
	if params.Purity == "" || params.Purity == "sfw" {
		tagQuery += " rating:general"
	}

	tagQuery = strings.TrimSpace(tagQuery)

	posts, err := a.client.PaginatedSearch(ctx, tagQuery, params.Limit)
	if err != nil {
		return nil, err
	}

	result := &sources.SearchResult{
		Wallpapers: make([]sources.ResultWallpaper, 0, len(posts)),
	}

	seen := make(map[string]bool)
	for _, p := range posts {
		fileURL := p.GetFileURL()
		if fileURL == "" {
			continue
		}

		w := sources.ResultWallpaper{
			ID:          fmt.Sprintf("%d", p.ID),
			SourceName:  "danbooru",
			SourceID:    fmt.Sprintf("%d", p.ID),
			URL:         fileURL,
			Resolution:  fmt.Sprintf("%dx%d", p.GetWidth(), p.GetHeight()),
			FileSize:    int64(p.FileSize),
			Format:      p.FileExt,
			Purity:      mapRating(p.Rating),
			Tags:        strings.Fields(p.GetTags()),
		}
		result.Wallpapers = append(result.Wallpapers, w)

		// Extract categorized tags
		for _, t := range ExtractCategorizedTags(p) {
			if !seen[t.Name] {
				seen[t.Name] = true
				result.Tags = append(result.Tags, t)
			}
		}
	}

	return result, nil
}

func (a *Adapter) Capabilities() *sources.Capabilities {
	maxTags := MaxAnonTags
	if a.client.IsAuthenticated() {
		maxTags = 0
	}
	return &sources.Capabilities{
		Name:                "danbooru",
		SupportsResolution:  false,
		SupportsAspectRatio: false,
		SupportsTags:        true,
		MaxTags:             maxTags,
		SupportedSorting:    []string{"top", "favorites", "latest", "random"},
		SupportsTimePeriod:  false,
		RequiresAuth:        false,
		AuthOptional:        true,
		AuthFields:          []string{"login", "api_key"},
		RateLimit:           "10 requests per second",
	}
}

func mapRating(r string) string {
	switch r {
	case "g", "s":
		return "sfw"
	case "q":
		return "sketchy"
	case "e":
		return "nsfw"
	default:
		return "sfw"
	}
}
