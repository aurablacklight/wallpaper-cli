package booru

// Post is the shared response shape for Danbooru/Moebooru-compatible APIs.
type Post struct {
	ID            int    `json:"id"`
	TagString     string `json:"tag_string"`      // Danbooru: space-separated
	Tags          string `json:"tags"`             // Moebooru (Konachan): space-separated
	FileURL       string `json:"file_url"`
	LargeFileURL  string `json:"large_file_url"`   // Danbooru only
	PreviewURL    string `json:"preview_file_url"`
	SampleURL     string `json:"sample_url"`       // Moebooru only
	ImageWidth    int    `json:"image_width"`
	ImageHeight   int    `json:"image_height"`
	Width         int    `json:"width"`            // Moebooru alias
	Height        int    `json:"height"`           // Moebooru alias
	FileSize      int    `json:"file_size"`
	FileExt       string `json:"file_ext"`
	Source        string `json:"source"`
	Rating        string `json:"rating"`           // s, q, e
	Score         int    `json:"score"`
	FavCount      int    `json:"fav_count"`        // Danbooru
	TagStringArtist    string `json:"tag_string_artist"`     // Danbooru
	TagStringCharacter string `json:"tag_string_character"`  // Danbooru
	TagStringCopyright string `json:"tag_string_copyright"`  // Danbooru
	TagStringGeneral   string `json:"tag_string_general"`    // Danbooru
	TagStringMeta      string `json:"tag_string_meta"`       // Danbooru
	MD5           string      `json:"md5"`
	CreatedAt     interface{} `json:"created_at"` // string on Danbooru, number on Moebooru
}

// GetWidth returns the width, supporting both Danbooru and Moebooru field names.
func (p Post) GetWidth() int {
	if p.ImageWidth > 0 {
		return p.ImageWidth
	}
	return p.Width
}

// GetHeight returns the height, supporting both Danbooru and Moebooru field names.
func (p Post) GetHeight() int {
	if p.ImageHeight > 0 {
		return p.ImageHeight
	}
	return p.Height
}

// GetFileURL returns the best available file URL.
func (p Post) GetFileURL() string {
	if p.FileURL != "" {
		return p.FileURL
	}
	if p.LargeFileURL != "" {
		return p.LargeFileURL
	}
	return p.SampleURL
}

// GetTags returns the tag string, supporting both Danbooru and Moebooru.
func (p Post) GetTags() string {
	if p.TagString != "" {
		return p.TagString
	}
	return p.Tags
}

// TagCategory represents a Danbooru tag category.
type TagCategory int

const (
	TagCategoryGeneral   TagCategory = 0
	TagCategoryArtist    TagCategory = 1
	TagCategoryCopyright TagCategory = 3
	TagCategoryCharacter TagCategory = 4
	TagCategoryMeta      TagCategory = 5
)

// CategoryName returns the string name for a tag category ID.
func CategoryName(id int) string {
	switch TagCategory(id) {
	case TagCategoryGeneral:
		return "general"
	case TagCategoryArtist:
		return "artist"
	case TagCategoryCopyright:
		return "copyright"
	case TagCategoryCharacter:
		return "character"
	case TagCategoryMeta:
		return "meta"
	default:
		return ""
	}
}
