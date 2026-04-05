package konachan

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/wallpaper-cli/internal/model"
	"github.com/user/wallpaper-cli/internal/sources"
)

func init() {
	sources.Register("konachan", func(cfg map[string]string) (sources.Source, error) {
		return NewAdapter(), nil
	})
}

type Adapter struct {
	client *Client
}

func NewAdapter() *Adapter {
	return &Adapter{client: NewClient()}
}

func (a *Adapter) Name() string { return "konachan" }

func (a *Adapter) Search(ctx context.Context, params *sources.SearchParams) (*sources.SearchResult, error) {
	tagQuery := strings.ReplaceAll(params.Tags, ",", " ")
	tagQuery = strings.TrimSpace(tagQuery)

	// Add sorting
	switch params.Sorting {
	case "top", "popular":
		tagQuery += " order:score"
	case "favorites":
		tagQuery += " order:favcount"
	case "latest", "new":
		tagQuery += " order:id_desc"
	}

	// Add rating filter
	if params.Purity == "" || params.Purity == "sfw" {
		tagQuery += " rating:safe"
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
			SourceName:  "konachan",
			SourceID:    fmt.Sprintf("%d", p.ID),
			URL:         fileURL,
			Resolution:  fmt.Sprintf("%dx%d", p.GetWidth(), p.GetHeight()),
			FileSize:    int64(p.FileSize),
			Format:      p.FileExt,
			Purity:      mapRating(p.Rating),
			Tags:        strings.Fields(p.GetTags()),
		}
		result.Wallpapers = append(result.Wallpapers, w)

		// Konachan provides a flat tag list without categories
		for _, name := range strings.Fields(p.GetTags()) {
			if !seen[name] {
				seen[name] = true
				result.Tags = append(result.Tags, model.Tag{
					Name:   name,
					Source: "konachan",
				})
			}
		}
	}

	return result, nil
}

func (a *Adapter) Capabilities() *sources.Capabilities {
	return &sources.Capabilities{
		Name:                "konachan",
		SupportsResolution:  false,
		SupportsAspectRatio: false,
		SupportsTags:        true,
		MaxTags:             0,
		SupportedSorting:    []string{"top", "favorites", "latest"},
		SupportsTimePeriod:  false,
		RequiresAuth:        false,
		RateLimit:           "conservative 0.5 req/s (HTTP 421 on throttle)",
	}
}

func mapRating(r string) string {
	switch r {
	case "s":
		return "sfw"
	case "q":
		return "sketchy"
	case "e":
		return "nsfw"
	default:
		return "sfw"
	}
}
