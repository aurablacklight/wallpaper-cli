package model

// Tag represents a tag harvested from a wallpaper source.
type Tag struct {
	Name       string
	Category   string // general, character, copyright, artist, meta, or ""
	CategoryID int    // numeric ID from source (e.g., Danbooru: 0=general, 1=artist, 3=copyright, 4=character, 5=meta)
	Source     string // wallhaven, danbooru, konachan, zerochan, reddit
}
