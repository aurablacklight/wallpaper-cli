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
			return nil, fmt.Errorf("zerochan requires a username in config (sources.zerochan.username) — see https://www.zerochan.net/api")
		}

		// Resolve cookies: cookies_file (Netscape format) takes priority, then inline cookies string
		cookies := cfg["cookies"]
		if cookiesFile := cfg["cookies_file"]; cookiesFile != "" {
			parsed, err := ParseCookiesFile(cookiesFile, "www.zerochan.net")
			if err != nil {
				return nil, fmt.Errorf("zerochan cookies_file: %w", err)
			}
			cookies = parsed
		}

		return NewAdapter(username, cookies), nil
	})
}

type Adapter struct {
	client *Client
}

func NewAdapter(username, cookies string) *Adapter {
	return &Adapter{client: NewClient(username, cookies)}
}

func (a *Adapter) Name() string { return "zerochan" }

func (a *Adapter) Search(ctx context.Context, params *sources.SearchParams) (*sources.SearchResult, error) {
	tag := strings.ReplaceAll(params.Tags, ",", " ")
	tag = strings.TrimSpace(tag)

	// Zerochan uses single-tag URL path — take the first tag
	if parts := strings.Fields(tag); len(parts) > 0 {
		tag = parts[0]
	}

	// Phase 1: List search (returns IDs + thumbnails, no full URLs)
	entries, err := a.client.PaginatedSearch(ctx, tag, params.Limit)
	if err != nil {
		return nil, err
	}

	result := &sources.SearchResult{
		Wallpapers: make([]sources.ResultWallpaper, 0, len(entries)),
	}

	seen := make(map[string]bool)

	// Phase 2: Fetch full details for each entry (2-call pattern)
	for _, e := range entries {
		detail, err := a.client.GetDetail(ctx, e.ID)
		if err != nil {
			// Fall back to thumbnail if detail fetch fails
			w := sources.ResultWallpaper{
				ID:         fmt.Sprintf("%d", e.ID),
				SourceName: "zerochan",
				SourceID:   fmt.Sprintf("%d", e.ID),
				URL:        e.Thumbnail,
				Resolution: fmt.Sprintf("%dx%d", e.Width, e.Height),
				Format:     "jpg",
			}
			result.Wallpapers = append(result.Wallpapers, w)
			continue
		}

		fileURL := detail.Full
		if fileURL == "" {
			fileURL = detail.Large
		}
		if fileURL == "" {
			continue
		}

		w := sources.ResultWallpaper{
			ID:         fmt.Sprintf("%d", detail.ID),
			SourceName: "zerochan",
			SourceID:   fmt.Sprintf("%d", detail.ID),
			URL:        fileURL,
			Resolution: fmt.Sprintf("%dx%d", detail.Width, detail.Height),
			FileSize:   int64(detail.Size),
			Format:     extensionFromURL(fileURL),
		}

		// Collect tags from detail
		if detail.Primary != "" {
			w.Tags = append(w.Tags, detail.Primary)
		}
		w.Tags = append(w.Tags, detail.Tags...)

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
		MaxTags:             1,
		SupportedSorting:    []string{"latest"},
		SupportsTimePeriod:  false,
		RequiresAuth:        true,
		AuthOptional:        false,
		AuthFields:          []string{"username", "cookies"},
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
