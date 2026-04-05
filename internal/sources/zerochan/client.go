package zerochan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/user/wallpaper-cli/internal/sources"
)

const (
	BaseURL    = "https://www.zerochan.net"
	MaxPerPage = 250
	DefaultRPS = 1.0 // 60 req/min = 1 req/s
)

// Entry represents a single image from the Zerochan JSON API.
type Entry struct {
	ID      int    `json:"id"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Size    int    `json:"size"`
	Primary string `json:"primary"` // primary tag
	Src     string `json:"src"`     // may be thumbnail or full URL
	Full    string `json:"full"`    // full-resolution URL (may be empty in list)
	Tags    []string `json:"tags"`
}

// SearchResponse is the envelope for paginated Zerochan results.
type SearchResponse struct {
	Items []Entry `json:"items"`
}

type Client struct {
	baseURL     string
	httpClient  *http.Client
	username    string
	rateLimiter *sources.RateLimiter
}

func NewClient(username string) *Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 15 * time.Second
	retryClient.Logger = nil

	return &Client{
		baseURL:     BaseURL,
		httpClient:  retryClient.StandardClient(),
		username:    username,
		rateLimiter: sources.NewRateLimiterPerSecond(DefaultRPS),
	}
}

func (c *Client) userAgent() string {
	if c.username != "" {
		return fmt.Sprintf("wallpaper-cli - %s", c.username)
	}
	return "wallpaper-cli/1.3"
}

// Search queries Zerochan's JSON API.
// Zerochan uses tag-in-URL-path: /{Tag}?json&l=N&p=page
func (c *Client) Search(ctx context.Context, tag string, limit, page int, strict bool) ([]Entry, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Build URL
	path := "/"
	if tag != "" {
		path += url.PathEscape(tag)
	}

	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("json", "")
	if limit > 0 {
		if limit > MaxPerPage {
			limit = MaxPerPage
		}
		q.Set("l", strconv.Itoa(limit))
	}
	if page > 1 {
		q.Set("p", strconv.Itoa(page))
	}
	if strict {
		q.Set("strict", "")
	}

	u.RawQuery = q.Encode()
	// Zerochan expects ?json (no value), fix the encoding
	u.RawQuery = strings.ReplaceAll(u.RawQuery, "json=", "json")
	u.RawQuery = strings.ReplaceAll(u.RawQuery, "strict=", "strict")

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zerochan request: %w", err)
	}
	defer resp.Body.Close()

	// Zerochan returns 404 for empty results — this is NOT an error
	if resp.StatusCode == http.StatusNotFound {
		return []Entry{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zerochan API: %s (status %d)", resp.Status, resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("zerochan decode: %w", err)
	}

	return searchResp.Items, nil
}

// GetEntry fetches a single entry by ID for the full-resolution URL.
func (c *Client) GetEntry(ctx context.Context, id int) (*Entry, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s/%d?json", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zerochan entry request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zerochan entry: %s (status %d)", resp.Status, resp.StatusCode)
	}

	var entry Entry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("zerochan entry decode: %w", err)
	}

	return &entry, nil
}

func (c *Client) PaginatedSearch(ctx context.Context, tag string, limit int, strict bool) ([]Entry, error) {
	var all []Entry
	page := 1
	perPage := MaxPerPage
	if limit < perPage {
		perPage = limit
	}

	for len(all) < limit {
		entries, err := c.Search(ctx, tag, perPage, page, strict)
		if err != nil {
			return all, err
		}
		if len(entries) == 0 {
			break
		}

		remaining := limit - len(all)
		if remaining >= len(entries) {
			all = append(all, entries...)
		} else {
			all = append(all, entries[:remaining]...)
		}
		page++
	}

	return all, nil
}
