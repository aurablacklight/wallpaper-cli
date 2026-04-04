package wallhaven

// SearchResponse represents the Wallhaven API search response
type SearchResponse struct {
	Data  []Wallpaper `json:"data"`
	Meta  Meta        `json:"meta"`
}

// Wallpaper represents a single wallpaper from Wallhaven
type Wallpaper struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	ShortURL    string   `json:"short_url"`
	Views       int      `json:"views"`
	Favorites   int      `json:"favorites"`
	Source      string   `json:"source"`
	Purity      string   `json:"purity"` // sfw, sketchy, nsfw
	Category    string   `json:"category"` // general, anime, people
	DimensionX  int      `json:"dimension_x"`
	DimensionY  int      `json:"dimension_y"`
	Resolution  string   `json:"resolution"` // "3840x2160"
	Ratio       string   `json:"ratio"`      // "16x9"
	FileSize    int64    `json:"file_size"`
	FileType    string   `json:"file_type"`
	CreatedAt   string   `json:"created_at"`
	Colors      []string `json:"colors"`
	Path        string   `json:"path"` // Direct image URL
	Thumbs      Thumbs   `json:"thumbs"`
	Tags        []Tag    `json:"tags"`
}

// Thumbs represents thumbnail URLs
type Thumbs struct {
	Large    string `json:"large"`
	Original string `json:"original"`
	Small    string `json:"small"`
}

// Tag represents a wallpaper tag
type Tag struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	CategoryID int   `json:"category_id"`
	Category  string `json:"category"`
	Purity    string `json:"purity"`
	CreatedAt string `json:"created_at"`
}

// Meta contains pagination info
type Meta struct {
	CurrentPage int         `json:"current_page"`
	LastPage    int         `json:"last_page"`
	PerPage     int         `json:"per_page"`
	Total       int         `json:"total"`
	Query       interface{} `json:"query,omitempty"` // Can be string or object
	Seed        string      `json:"seed,omitempty"`  // For random sorting
}

// Query represents the search query details
type Query struct {
	Tag      string `json:"tag,omitempty"`
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Key      string `json:"key,omitempty"`
}

// ToInternal converts Wallhaven Wallpaper to internal model
func (w Wallpaper) ToInternal() map[string]interface{} {
	return map[string]interface{}{
		"id":           w.ID,
		"source":       "wallhaven",
		"source_id":    w.ID,
		"url":          w.Path,
		"resolution":   w.Resolution,
		"aspect_ratio": w.Ratio,
		"file_size":    w.FileSize,
		"format":       w.FileType,
		"tags":         w.Tags,
		"purity":       w.Purity,
		"category":     w.Category,
	}
}
