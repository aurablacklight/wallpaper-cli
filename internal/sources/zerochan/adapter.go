package zerochan

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/wallpaper-cli/internal/model"
	"github.com/user/wallpaper-cli/internal/sources"
)

func init() {
	sources.Register("zerochan", func(cfg map[string]string) (sources.Source, error) {
		username := cfg["username"]
		if username == "" {
			return nil, fmt.Errorf("zerochan requires a username in config (sources.zerochan.username) for the User-Agent header — see https://www.zerochan.net/api")
		}
		return NewAdapter(username), nil
	})
}

type Adapter struct {
	client *Client
}

func NewAdapter(username string) *Adapter {
	return &Adapter{client: NewClient(username)}
}

func (a *Adapter) Name() string { return "zerochan" }

func (a *Adapter) Search(ctx context.Context, params *sources.SearchParams) (*sources.SearchResult, error) {
	// Zerochan uses tag-in-URL-path, take the first tag
	tag := strings.ReplaceAll(params.Tags, ",", " ")
	tag = strings.TrimSpace(tag)

	// Zerochan supports single-tag or comma-separated in URL path
	// For simplicity, use the first tag
	if parts := strings.Fields(tag); len(parts) > 0 {
		tag = parts[0]
	}

	strict := false
	// Could be passed via params in future

	entries, err := a.client.PaginatedSearch(ctx, tag, params.Limit, strict)
	if err != nil {
		return nil, err
	}

	result := &sources.SearchResult{
		Wallpapers: make([]sources.ResultWallpaper, 0, len(entries)),
	}

	seen := make(map[string]bool)
	for _, e := range entries {
		// Determine the best URL
		fileURL := e.Full
		if fileURL == "" {
			fileURL = e.Src
		}
		if fileURL == "" {
			continue
		}

		w := sources.ResultWallpaper{
			ID:         fmt.Sprintf("%d", e.ID),
			SourceName: "zerochan",
			SourceID:   fmt.Sprintf("%d", e.ID),
			URL:        fileURL,
			Resolution: fmt.Sprintf("%dx%d", e.Width, e.Height),
			FileSize:   int64(e.Size),
			Format:     extensionFromURL(fileURL),
		}

		// Collect tags
		if e.Primary != "" {
			w.Tags = append(w.Tags, e.Primary)
		}
		w.Tags = append(w.Tags, e.Tags...)

		result.Wallpapers = append(result.Wallpapers, w)

		// Harvest tags
		for _, name := range w.Tags {
			if !seen[name] && name != "" {
				seen[name] = true
				result.Tags = append(result.Tags, model.Tag{
					Name:   name,
					Source: "zerochan",
				})
			}
		}
	}

	return result, nil
}

func (a *Adapter) Capabilities() *sources.Capabilities {
	return &sources.Capabilities{
		Name:                "zerochan",
		SupportsResolution:  false,
		SupportsAspectRatio: false,
		SupportsTags:        true,
		MaxTags:             1, // URL-path based, effectively one primary tag
		SupportedSorting:    []string{"latest"},
		SupportsTimePeriod:  false,
		RequiresAuth:        false,
		AuthOptional:        false,
		AuthFields:          []string{"username"},
		RateLimit:           "60 requests per minute",
	}
}

func extensionFromURL(u string) string {
	lower := strings.ToLower(u)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "png"
	case strings.HasSuffix(lower, ".webp"):
		return "webp"
	default:
		return "jpg"
	}
}
