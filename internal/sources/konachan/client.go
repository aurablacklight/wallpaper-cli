package konachan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/user/wallpaper-cli/internal/sources"
	"github.com/user/wallpaper-cli/internal/sources/booru"
)

const (
	BaseURL    = "https://konachan.com"
	MaxPerPage = 100
	DefaultRPS = 0.5 // Conservative: 1 request per 2 seconds
)

type Client struct {
	baseURL     string
	httpClient  *http.Client
	rateLimiter *sources.RateLimiter
}

func NewClient() *Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 15 * time.Second
	retryClient.Logger = nil

	// Treat HTTP 421 as retryable (Konachan throttle)
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}
		if resp.StatusCode == 421 || resp.StatusCode == 429 || resp.StatusCode == 503 {
			return true, nil
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	return &Client{
		baseURL:     BaseURL,
		httpClient:  retryClient.StandardClient(),
		rateLimiter: sources.NewRateLimiterPerSecond(DefaultRPS),
	}
}

func (c *Client) Search(ctx context.Context, tags string, limit, page int) ([]booru.Post, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	u, err := url.Parse(c.baseURL + "/post.json")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	if tags != "" {
		q.Set("tags", tags)
	}
	if limit > 0 {
		if limit > MaxPerPage {
			limit = MaxPerPage
		}
		q.Set("limit", strconv.Itoa(limit))
	}
	if page > 1 {
		q.Set("page", strconv.Itoa(page))
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "wallpaper-cli/1.3")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("konachan request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("konachan API: %s (status %d)", resp.Status, resp.StatusCode)
	}

	var posts []booru.Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, fmt.Errorf("konachan decode: %w", err)
	}

	return posts, nil
}

func (c *Client) PaginatedSearch(ctx context.Context, tags string, limit int) ([]booru.Post, error) {
	var all []booru.Post
	page := 1
	perPage := MaxPerPage
	if limit < perPage {
		perPage = limit
	}

	for len(all) < limit {
		posts, err := c.Search(ctx, tags, perPage, page)
		if err != nil {
			return all, err
		}
		if len(posts) == 0 {
			break
		}

		remaining := limit - len(all)
		if remaining >= len(posts) {
			all = append(all, posts...)
		} else {
			all = append(all, posts[:remaining]...)
		}
		page++
	}

	return all, nil
}
