package wallhaven

import (
	"context"
	"fmt"
)

// PaginatedSearch performs a search that automatically handles pagination
// to fetch more than 24 results (the API's max per page)
func (c *Client) PaginatedSearch(ctx context.Context, opts *SearchOptions, limit int) ([]Wallpaper, error) {
	if limit <= 0 {
		return nil, nil
	}

	var allWallpapers []Wallpaper
	page := 1
	maxPerPage := 24

	for len(allWallpapers) < limit {
		opts.Page = page
		opts.PerPage = maxPerPage

		resp, err := c.Search(ctx, opts)
		if err != nil {
			return allWallpapers, fmt.Errorf("search page %d failed: %w", page, err)
		}

		// If no results, we're done
		if len(resp.Data) == 0 {
			break
		}

		// Add results up to the limit
		remaining := limit - len(allWallpapers)
		if remaining >= len(resp.Data) {
			allWallpapers = append(allWallpapers, resp.Data...)
		} else {
			allWallpapers = append(allWallpapers, resp.Data[:remaining]...)
		}

		// Check if we've reached the last page
		if page >= resp.Meta.LastPage {
			break
		}

		page++
	}

	return allWallpapers, nil
}
