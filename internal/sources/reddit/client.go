package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://www.reddit.com/r/"
	DefaultTimeout = 30 * time.Second
)

// Client is a Reddit API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new Reddit API client
func NewClient() *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		userAgent: "wallpaper-cli/1.0 (by /u/wallpaper-cli)",
	}
}

// Sorting options
type SortOption string

const (
	SortHot     SortOption = "hot"
	SortNew     SortOption = "new"
	SortTop     SortOption = "top"
	SortRising  SortOption = "rising"
)

// Time period for top sorting
type TimePeriod string

const (
	TimeHour   TimePeriod = "hour"
	TimeDay    TimePeriod = "day"
	TimeWeek   TimePeriod = "week"
	TimeMonth  TimePeriod = "month"
	TimeYear   TimePeriod = "year"
	TimeAll    TimePeriod = "all"
)

// SearchOptions contains parameters for search requests
type SearchOptions struct {
	Subreddit  string     // e.g., "Animewallpaper"
	Sort       SortOption // hot, new, top, rising
	Time       TimePeriod // hour, day, week, month, year, all
	Limit      int        // Max results (max 100 for Reddit)
	After      string     // Pagination token
}

// DefaultSearchOptions returns default search options
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		Subreddit: "Animewallpaper",
		Sort:      SortHot,
		Time:      TimeDay,
		Limit:     25,
	}
}

// SearchResponse represents Reddit's JSON response
type SearchResponse struct {
	Kind string `json:"kind"`
	Data struct {
		Children []struct {
			Kind string `json:"kind"`
			Data Post   `json:"data"`
		} `json:"children"`
		After    string `json:"after"`
		Before   string `json:"before"`
		Dist     int    `json:"dist"`
	} `json:"data"`
}

// Post represents a Reddit post
type Post struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Permalink   string `json:"permalink"`
	Score       int    `json:"score"`
	Ups         int    `json:"ups"`
	NumComments int    `json:"num_comments"`
	Created     float64 `json:"created_utc"`
	IsVideo     bool   `json:"is_video"`
	IsSelf      bool   `json:"is_self"`
	Over18      bool   `json:"over_18"`
	Thumbnail   string `json:"thumbnail"`
	Preview     *struct {
		Images []struct {
			Source struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"source"`
		} `json:"images"`
	} `json:"preview,omitempty"`
}

// IsImageURL checks if URL is a direct image link
func IsImageURL(url string) bool {
	lower := strings.ToLower(url)
	return strings.HasSuffix(lower, ".jpg") ||
		strings.HasSuffix(lower, ".jpeg") ||
		strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".webp") ||
		strings.HasSuffix(lower, ".gif")
}

// Search fetches posts from a subreddit
func (c *Client) Search(ctx context.Context, opts *SearchOptions) ([]Post, error) {
	if opts == nil {
		opts = DefaultSearchOptions()
	}

	// Build URL
	url := fmt.Sprintf("%s%s/%s.json?limit=%d", 
		c.baseURL, 
		opts.Subreddit, 
		opts.Sort,
		opts.Limit)

	// Add time parameter for top sorting
	if opts.Sort == SortTop && opts.Time != "" {
		url = fmt.Sprintf("%s&t=%s", url, opts.Time)
	}

	// Add after for pagination
	if opts.After != "" {
		url = fmt.Sprintf("%s&after=%s", url, opts.After)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited by Reddit (429)")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Reddit API error: %s (status %d)", resp.Status, resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Extract posts
	var posts []Post
	for _, child := range result.Data.Children {
		if child.Data.ID != "" && !child.Data.IsSelf && !child.Data.IsVideo {
			posts = append(posts, child.Data)
		}
	}

	return posts, nil
}

// GetDirectImageURL tries to get a direct image URL from a post
func GetDirectImageURL(post Post) string {
	// First, check if URL is already a direct image
	if IsImageURL(post.URL) {
		return strings.ReplaceAll(post.URL, "&amp;", "&")
	}

	// Try to get from preview
	if post.Preview != nil && len(post.Preview.Images) > 0 {
		src := post.Preview.Images[0].Source.URL
		if src != "" {
			return strings.ReplaceAll(src, "&amp;", "&")
		}
	}

	return ""
}

// GetResolution tries to extract resolution from post
func GetResolution(post Post) string {
	if post.Preview != nil && len(post.Preview.Images) > 0 {
		src := post.Preview.Images[0].Source
		if src.Width > 0 && src.Height > 0 {
			return fmt.Sprintf("%dx%d", src.Width, src.Height)
		}
	}
	return "unknown"
}
