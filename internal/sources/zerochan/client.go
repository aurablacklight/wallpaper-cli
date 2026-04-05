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

	"github.com/user/wallpaper-cli/internal/sources"
)

const (
	BaseURL    = "https://www.zerochan.net"
	MaxPerPage = 250
	DefaultRPS = 1.0 // 60 req/min = 1 req/s
)

// ListEntry is an item from the paginated search response (no full URL).
type ListEntry struct {
	ID        int      `json:"id"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	Thumbnail string   `json:"thumbnail"`
	Tag       string   `json:"tag"`     // primary tag
	Tags      []string `json:"tags"`
	Source    string   `json:"source"`  // pixiv/other source URL
	MD5       string   `json:"md5"`
}

// DetailEntry is a single-entry response with full-resolution URLs.
type DetailEntry struct {
	ID      int      `json:"id"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	Size    int      `json:"size"`
	Full    string   `json:"full"`    // full-resolution URL
	Large   string   `json:"large"`
	Medium  string   `json:"medium"`
	Small   string   `json:"small"`
	Hash    string   `json:"hash"`
	Primary string   `json:"primary"` // primary tag
	Tags    []string `json:"tags"`
	Source  string   `json:"source"`
}

// SearchResponse is the envelope for paginated Zerochan results.
type SearchResponse struct {
	Items []ListEntry `json:"items"`
}

type Client struct {
	baseURL     string
	httpClient  *http.Client
	username    string
	cookies     string // raw cookie header value
	rateLimiter *sources.RateLimiter
}

func NewClient(username string, cookies string) *Client {
	return &Client{
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			// Don't follow redirects automatically — we need to preserve ?json
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		username:    username,
		cookies:     cookies,
		rateLimiter: sources.NewRateLimiterPerSecond(DefaultRPS),
	}
}

func (c *Client) userAgent() string {
	return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"
}

// Search queries the Zerochan JSON API for a tag.
func (c *Client) Search(ctx context.Context, tag string, limit, page int) ([]ListEntry, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	path := "/"
	if tag != "" {
		path += url.PathEscape(tag)
	}

	u := c.baseURL + path + "?json"
	if limit > 0 {
		if limit > MaxPerPage {
			limit = MaxPerPage
		}
		u += "&l=" + strconv.Itoa(limit)
	}
	if page > 1 {
		u += "&p=" + strconv.Itoa(page)
	}

	return c.doSearch(ctx, u)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent())
	if c.cookies != "" {
		req.Header.Set("Cookie", c.cookies)
	}
}

func (c *Client) doSearch(ctx context.Context, rawURL string) ([]ListEntry, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zerochan request: %w", err)
	}
	defer resp.Body.Close()

	// 404 = empty results (not an error)
	if resp.StatusCode == http.StatusNotFound {
		return []ListEntry{}, nil
	}

	// Handle redirects (tag aliases) — preserve ?json
	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		loc := resp.Header.Get("Location")
		if loc != "" {
			if !strings.HasPrefix(loc, "http") {
				loc = c.baseURL + loc
			}
			if !strings.Contains(loc, "json") {
				if strings.Contains(loc, "?") {
					loc += "&json"
				} else {
					loc += "?json"
				}
			}
			if err := c.rateLimiter.Wait(ctx); err != nil {
				return nil, err
			}
			return c.doSearch(ctx, loc)
		}
	}

	// 503 = JS anti-bot challenge
	if resp.StatusCode == http.StatusServiceUnavailable {
		return nil, fmt.Errorf("zerochan returned 503 (anti-bot challenge) — set session cookies in config (from browser login) to bypass")
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

// GetDetail fetches a single entry by ID for the full-resolution URL.
func (c *Client) GetDetail(ctx context.Context, id int) (*DetailEntry, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s/%d?json", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zerochan detail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zerochan detail: %s (status %d)", resp.Status, resp.StatusCode)
	}

	var entry DetailEntry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("zerochan detail decode: %w", err)
	}

	return &entry, nil
}

func (c *Client) PaginatedSearch(ctx context.Context, tag string, limit int) ([]ListEntry, error) {
	var all []ListEntry
	page := 1
	perPage := MaxPerPage
	if limit < perPage {
		perPage = limit
	}

	for len(all) < limit {
		entries, err := c.Search(ctx, tag, perPage, page)
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
