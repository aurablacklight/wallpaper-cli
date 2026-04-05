package danbooru

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
	"github.com/user/wallpaper-cli/internal/model"
	"github.com/user/wallpaper-cli/internal/sources"
	"github.com/user/wallpaper-cli/internal/sources/booru"
)

const (
	BaseURL      = "https://danbooru.donmai.us"
	MaxPerPage   = 200
	MaxAnonTags  = 2
	DefaultRPS   = 6.0
)

type Client struct {
	baseURL     string
	httpClient  *http.Client
	login       string
	apiKey      string
	rateLimiter *sources.RateLimiter
}

func NewClient(login, apiKey string) *Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 10 * time.Second
	retryClient.Logger = nil

	return &Client{
		baseURL:     BaseURL,
		httpClient:  retryClient.StandardClient(),
		login:       login,
		apiKey:      apiKey,
		rateLimiter: sources.NewRateLimiterPerSecond(DefaultRPS),
	}
}

func (c *Client) IsAuthenticated() bool {
	return c.login != "" && c.apiKey != ""
}

func (c *Client) Search(ctx context.Context, tags string, limit, page int) ([]booru.Post, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	u, err := url.Parse(c.baseURL + "/posts.json")
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

	if c.IsAuthenticated() {
		q.Set("login", c.login)
		q.Set("api_key", c.apiKey)
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
		return nil, fmt.Errorf("danbooru request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("danbooru API: %s (status %d)", resp.Status, resp.StatusCode)
	}

	var posts []booru.Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, fmt.Errorf("danbooru decode: %w", err)
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

// ExtractCategorizedTags extracts tags with categories from Danbooru's tag_string_* fields.
func ExtractCategorizedTags(post booru.Post) []model.Tag {
	var tags []model.Tag

	addTags := func(tagStr string, category string, categoryID int) {
		for _, name := range strings.Fields(tagStr) {
			if name != "" {
				tags = append(tags, model.Tag{
					Name:       name,
					Category:   category,
					CategoryID: categoryID,
					Source:     "danbooru",
				})
			}
		}
	}

	addTags(post.TagStringGeneral, "general", int(booru.TagCategoryGeneral))
	addTags(post.TagStringArtist, "artist", int(booru.TagCategoryArtist))
	addTags(post.TagStringCopyright, "copyright", int(booru.TagCategoryCopyright))
	addTags(post.TagStringCharacter, "character", int(booru.TagCategoryCharacter))
	addTags(post.TagStringMeta, "meta", int(booru.TagCategoryMeta))

	return tags
}
