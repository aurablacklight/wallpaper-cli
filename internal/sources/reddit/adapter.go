package reddit

import (
	"context"

	"github.com/user/wallpaper-cli/internal/model"
	"github.com/user/wallpaper-cli/internal/sources"
)

// Adapter wraps the Reddit Client as a sources.Source.
type Adapter struct {
	client     *Client
	subreddits []string
}

func init() {
	sources.Register("reddit", func(cfg map[string]string) (sources.Source, error) {
		subs := ParseSubreddits(cfg["subreddits"])
		return &Adapter{client: NewClient(), subreddits: subs}, nil
	})
}

func (a *Adapter) Name() string { return "reddit" }

func (a *Adapter) Search(ctx context.Context, params *sources.SearchParams) (*sources.SearchResult, error) {
	sortOpt := SortHot
	switch params.Sorting {
	case "top", "popular":
		sortOpt = SortTop
	case "new", "latest":
		sortOpt = SortNew
	case "hot":
		sortOpt = SortHot
	}

	var allPosts []Wallpaper
	for _, sub := range a.subreddits {
		opts := &SearchOptions{
			Subreddit: sub,
			Sort:      sortOpt,
			Time:      TimePeriod(mapRedditTime(params.TimePeriod)),
			Limit:     params.Limit,
		}

		posts, err := a.client.Search(ctx, opts)
		if err != nil {
			continue
		}
		allPosts = append(allPosts, ToPosts(posts)...)
	}

	if len(allPosts) > params.Limit {
		allPosts = allPosts[:params.Limit]
	}

	result := &sources.SearchResult{
		Wallpapers: make([]sources.ResultWallpaper, 0, len(allPosts)),
		Tags:       []model.Tag{},
	}

	for _, p := range allPosts {
		rw := sources.ResultWallpaper{
			ID:         p.ID,
			SourceName: "reddit",
			SourceID:   p.ID,
			URL:        p.URL,
			Title:      p.Title,
			Resolution: p.Resolution,
			Format:     extensionFromURL(p.URL),
		}
		result.Wallpapers = append(result.Wallpapers, rw)
	}

	return result, nil
}

func (a *Adapter) Capabilities() *sources.Capabilities {
	return &sources.Capabilities{
		Name:               "reddit",
		SupportsResolution: false,
		SupportsAspectRatio: false,
		SupportsTags:       false,
		MaxTags:            0,
		SupportedSorting:   []string{"hot", "new", "top"},
		SupportsTimePeriod: true,
		RequiresAuth:       false,
		RateLimit:          "60 requests per minute",
	}
}

func mapRedditTime(tp string) string {
	switch tp {
	case "day", "1d":
		return "day"
	case "week", "7d":
		return "week"
	case "month", "30d":
		return "month"
	case "year", "1y":
		return "year"
	case "all":
		return "all"
	default:
		return "week"
	}
}

func extensionFromURL(url string) string {
	switch {
	case len(url) > 4 && url[len(url)-4:] == ".png":
		return "png"
	case len(url) > 5 && url[len(url)-5:] == ".webp":
		return "webp"
	default:
		return "jpg"
	}
}
