package reddit

import "strings"

// Wallpaper represents a wallpaper from Reddit
type Wallpaper struct {
	ID          string
	Title       string
	URL         string // Direct image URL
	Permalink   string // Reddit post URL
	Score       int    // Upvotes
	Ups         int    // Upvotes (raw)
	NumComments int
	Resolution  string
	Thumbnail   string
}

// ToPosts converts Reddit Posts to Wallpapers
func ToPosts(posts []Post) []Wallpaper {
	var wallpapers []Wallpaper
	for _, post := range posts {
		url := GetDirectImageURL(post)
		if url == "" {
			continue // Skip posts without direct image URLs
		}

		res := GetResolution(post)

		wallpapers = append(wallpapers, Wallpaper{
			ID:          post.ID,
			Title:       post.Title,
			URL:         url,
			Permalink:   "https://reddit.com" + post.Permalink,
			Score:       post.Score,
			Ups:         post.Ups,
			NumComments: post.NumComments,
			Resolution:  res,
			Thumbnail:   post.Thumbnail,
		})
	}
	return wallpapers
}

// ParseSubreddits parses comma-separated subreddit names
func ParseSubreddits(input string) []string {
	if input == "" {
		return []string{"Animewallpaper"}
	}

	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			// Remove r/ prefix if present
			s = strings.TrimPrefix(s, "r/")
			result = append(result, s)
		}
	}
	return result
}
