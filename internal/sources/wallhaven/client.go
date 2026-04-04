package wallhaven

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultBaseURL    = "https://wallhaven.cc/api/v1/"
	DefaultTimeout    = 30 * time.Second
	DefaultRateLimit  = 2 * time.Second // 1 request per 2 seconds = 30 requests per minute (well under 45/15min limit)
)

// Client is a Wallhaven API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	rateLimiter *time.Ticker
	lastRequest time.Time
	mu         sync.Mutex
}

// ClientOption configures the Client
type ClientOption func(*Client)

// WithAPIKey sets the API key
func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// NewClient creates a new Wallhaven API client
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		rateLimiter: time.NewTicker(DefaultRateLimit),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SearchOptions contains parameters for search requests
type SearchOptions struct {
	Query       string   // Search query (tags)
	Resolutions []string // Exact resolutions (e.g., "3840x2160")
	Ratios      []string // Aspect ratios (e.g., "16x9")
	Sorting     string   // relevance, random, date_added, views, favorites, toplist
	Order       string   // desc, asc
	Page        int      // Page number
	PerPage     int      // Results per page (max 24 for API)
	Purity      string   // sfw, sketchy, nsfw (requires API key for nsfw)
	Categories  string   // 100/101/111 (general/anime/people bits)
	TopRange    string   // 1d, 3d, 1w, 1M, 3M, 6M, 1y for toplist sorting
	Seed        string   // Seed for random sorting (for consistent pagination)
}

// DefaultSearchOptions returns default search options
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		Sorting:    "relevance",
		Order:      "desc",
		Page:       1,
		PerPage:    24,
		Purity:     "sfw",
		Categories: "111", // All categories
	}
}

// Search performs a search request
func (c *Client) Search(ctx context.Context, opts *SearchOptions) (*SearchResponse, error) {
	// Rate limiting - be polite to the API
	c.waitForRateLimit()
	
	if opts == nil {
		opts = DefaultSearchOptions()
	}

	u, err := url.Parse(c.baseURL + "search")
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	q := u.Query()

	// Add query parameters
	if opts.Query != "" {
		q.Set("q", opts.Query)
	}
	if len(opts.Resolutions) > 0 {
		q.Set("resolutions", joinStrings(opts.Resolutions, ","))
	}
	if len(opts.Ratios) > 0 {
		q.Set("ratios", joinStrings(opts.Ratios, ","))
	}
	if opts.Sorting != "" {
		q.Set("sorting", opts.Sorting)
	}
	if opts.Order != "" {
		q.Set("order", opts.Order)
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Purity != "" {
		q.Set("purity", opts.Purity)
	}
	if opts.Categories != "" {
		q.Set("categories", opts.Categories)
	}
	if opts.TopRange != "" {
		q.Set("topRange", opts.TopRange)
	}
	if opts.Seed != "" {
		q.Set("seed", opts.Seed)
	}
	if c.apiKey != "" {
		q.Set("apikey", c.apiKey)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "wallpaper-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s (status %d)", resp.Status, resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// GetWallpaper fetches a specific wallpaper by ID
func (c *Client) GetWallpaper(ctx context.Context, id string) (*Wallpaper, error) {
	// Rate limiting
	c.waitForRateLimit()
	
	u, err := url.Parse(c.baseURL + "w/" + id)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	q := u.Query()
	if c.apiKey != "" {
		q.Set("apikey", c.apiKey)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result struct {
		Data Wallpaper `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result.Data, nil
}

// waitForRateLimit ensures we don't exceed API rate limits
// Wallhaven allows 45 requests per 15 minutes for anonymous users
// We use a conservative 1 request per 2 seconds = 30 requests per minute
func (c *Client) waitForRateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	elapsed := time.Since(c.lastRequest)
	if elapsed < DefaultRateLimit {
		time.Sleep(DefaultRateLimit - elapsed)
	}
	c.lastRequest = time.Now()
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
