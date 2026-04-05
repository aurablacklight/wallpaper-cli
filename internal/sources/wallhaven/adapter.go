package wallhaven

import (
	"context"
	"strings"

	"github.com/user/wallpaper-cli/internal/model"
	"github.com/user/wallpaper-cli/internal/sources"
)

// Adapter wraps the Wallhaven Client as a sources.Source.
type Adapter struct {
	client *Client
}

func init() {
	sources.Register("wallhaven", func(cfg map[string]string) (sources.Source, error) {
		var opts []ClientOption
		if key := cfg["api_key"]; key != "" {
			opts = append(opts, WithAPIKey(key))
		}
		return &Adapter{client: NewClient(opts...)}, nil
	})
}

func (a *Adapter) Name() string { return "wallhaven" }

func (a *Adapter) Search(ctx context.Context, params *sources.SearchParams) (*sources.SearchResult, error) {
	builder := NewSearchBuilder()

	if params.Tags != "" {
		builder.WithQuery(ParseTags(params.Tags))
	}
	if params.Resolution != "" {
		builder.WithResolution(params.Resolution)
	}
	if params.AspectRatio != "" {
		builder.WithAspectRatio(params.AspectRatio)
	}
	if params.AnimeOnly {
		builder.WithAnimeOnly()
	}

	switch params.Sorting {
	case "top", "popular":
		builder.WithSorting("toplist")
		if params.TimePeriod != "" {
			builder.WithTopRange(mapTimePeriod(params.TimePeriod))
		}
	case "favorites":
		builder.WithSorting("favorites")
	case "views":
		builder.WithSorting("views")
	case "latest", "new":
		builder.WithSorting("date_added")
		builder.WithOrder("desc")
	case "random":
		builder.WithRandom()
	default:
		builder.WithRandom()
	}

	opts := builder.Build()

	wallpapers, err := a.client.PaginatedSearch(ctx, opts, params.Limit)
	if err != nil {
		return nil, err
	}

	result := &sources.SearchResult{
		Wallpapers: make([]sources.ResultWallpaper, 0, len(wallpapers)),
		Tags:       make([]model.Tag, 0),
	}

	seen := make(map[string]bool)
	for _, w := range wallpapers {
		rw := sources.ResultWallpaper{
			ID:          w.ID,
			SourceName:  "wallhaven",
			SourceID:    w.ID,
			URL:         w.Path,
			Resolution:  w.Resolution,
			AspectRatio: w.Ratio,
			FileSize:    w.FileSize,
			Format:      extensionFromPath(w.Path),
			Purity:      w.Purity,
			Category:    w.Category,
		}
		for _, t := range w.Tags {
			rw.Tags = append(rw.Tags, t.Name)
			if !seen[t.Name] {
				seen[t.Name] = true
				result.Tags = append(result.Tags, model.Tag{
					Name:       t.Name,
					Category:   t.Category,
					CategoryID: t.CategoryID,
					Source:     "wallhaven",
				})
			}
		}
		result.Wallpapers = append(result.Wallpapers, rw)
	}

	return result, nil
}

func (a *Adapter) Capabilities() *sources.Capabilities {
	return &sources.Capabilities{
		Name:                "wallhaven",
		SupportsResolution:  true,
		SupportsAspectRatio: true,
		SupportsTags:        true,
		MaxTags:             0,
		SupportedSorting:    []string{"random", "top", "latest", "favorites", "views"},
		SupportsTimePeriod:  true,
		RequiresAuth:        false,
		AuthOptional:        true,
		AuthFields:          []string{"api_key"},
		RateLimit:           "45 requests per 15 minutes",
	}
}

func mapTimePeriod(tp string) string {
	switch tp {
	case "day", "1d":
		return "1d"
	case "week", "7d":
		return "1w"
	case "month", "30d":
		return "1M"
	case "year", "1y":
		return "1y"
	case "all":
		return "1y"
	default:
		return "1M"
	}
}

func extensionFromPath(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "png"
	case strings.HasSuffix(lower, ".webp"):
		return "webp"
	default:
		return "jpg"
	}
}
