package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/wallpaper-cli/internal/collections"
	"github.com/user/wallpaper-cli/internal/data"
)

var (
	favPath string
)

var favoriteCmd = &cobra.Command{
	Use:   "favorite [path]",
	Short: "Manage favorite wallpapers",
	Long: `Add or remove wallpapers from favorites.

Favorites are wallpapers you want quick access to. They can be browsed
and used for automatic rotation.`,
	RunE: runFavorite,
}

func init() {
	rootCmd.AddCommand(favoriteCmd)
	favoriteCmd.Flags().StringVarP(&favPath, "path", "p", "", "Wallpaper path (omit to toggle current)")
}

func runFavorite(cmd *cobra.Command, args []string) error {
	// Get database
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".local", "share", "wallpaper-cli", "wallpapers.db")
	db, err := data.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	manager := collections.NewManager(db)

	// Determine target wallpaper
	targetPath := favPath
	if len(args) > 0 {
		targetPath = args[0]
	}

	if targetPath == "" {
		// List favorites
		favs, err := manager.ListFavorites(100)
		if err != nil {
			return fmt.Errorf("listing favorites: %w", err)
		}

		if len(favs) == 0 {
			fmt.Println("No favorites yet. Add one with: wallpaper-cli favorite <path>")
			return nil
		}

		fmt.Printf("⭐ Favorites (%d):\n\n", len(favs))
		for _, f := range favs {
			fmt.Printf("  %s\n", f.ImageHash)
		}
		return nil
	}

	// Get hash for the path (simplified - in real impl, look up hash from path)
	hash := filepath.Base(targetPath)

	// Toggle favorite
	isFav, err := manager.ToggleFavorite(hash)
	if err != nil {
		return fmt.Errorf("toggling favorite: %w", err)
	}

	if isFav {
		fmt.Printf("⭐ Added to favorites: %s\n", targetPath)
	} else {
		fmt.Printf("Removed from favorites: %s\n", targetPath)
	}

	return nil
}

// playlistCmd manages playlists
var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Manage wallpaper playlists",
	Long: `Create and manage themed playlists for wallpaper rotation.

Playlists allow you to organize wallpapers into themed collections
like "cozy", "energetic", "work", etc.`,
}

var (
	playlistName string
	playlistDesc string
)

var playlistCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new playlist",
	Example: `  wallpaper-cli playlist create cozy-winter --description "Warm winter vibes"
  wallpaper-cli playlist create work-focus`,
	RunE: runPlaylistCreate,
}

var playlistListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all playlists",
	RunE:  runPlaylistList,
}

var playlistAddCmd = &cobra.Command{
	Use:     "add [playlist] [path]",
	Short:   "Add wallpaper to playlist",
	Example: `  wallpaper-cli playlist add cozy-winter ~/Pictures/wp/01_abc.jpg`,
	RunE:    runPlaylistAdd,
}

var playlistShowCmd = &cobra.Command{
	Use:   "show [playlist]",
	Short: "Show playlist contents",
	RunE:  runPlaylistShow,
}

func init() {
	rootCmd.AddCommand(playlistCmd)
	playlistCmd.AddCommand(playlistCreateCmd)
	playlistCmd.AddCommand(playlistListCmd)
	playlistCmd.AddCommand(playlistAddCmd)
	playlistCmd.AddCommand(playlistShowCmd)

	playlistCreateCmd.Flags().StringVarP(&playlistDesc, "description", "d", "", "Playlist description")
}

func runPlaylistCreate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("playlist name required")
	}
	name := args[0]

	db, manager, err := getCollectionsManager()
	if err != nil {
		return err
	}
	defer db.Close()

	playlist, err := manager.CreatePlaylist(name, playlistDesc)
	if err != nil {
		return fmt.Errorf("creating playlist: %w", err)
	}

	fmt.Printf("📋 Created playlist: %s (id: %s)\n", playlist.Name, playlist.ID)
	return nil
}

func runPlaylistList(cmd *cobra.Command, args []string) error {
	db, manager, err := getCollectionsManager()
	if err != nil {
		return err
	}
	defer db.Close()

	playlists, err := manager.ListPlaylists()
	if err != nil {
		return fmt.Errorf("listing playlists: %w", err)
	}

	if len(playlists) == 0 {
		fmt.Println("No playlists yet. Create one with: wallpaper-cli playlist create <name>")
		return nil
	}

	fmt.Printf("📋 Playlists (%d):\n\n", len(playlists))
	for _, p := range playlists {
		desc := ""
		if p.Description != "" {
			desc = fmt.Sprintf(" - %s", p.Description)
		}
		fmt.Printf("  %s (%d items)%s\n", p.Name, p.ItemCount, desc)
	}
	return nil
}

func runPlaylistAdd(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: playlist add [playlist] [path]")
	}
	playlistName := args[0]
	path := args[1]

	db, manager, err := getCollectionsManager()
	if err != nil {
		return err
	}
	defer db.Close()

	// Find playlist by name (simplified - look up by name prefix)
	playlists, _ := manager.ListPlaylists()
	var targetPlaylist *collections.Playlist
	for _, p := range playlists {
		if strings.EqualFold(p.Name, playlistName) || strings.HasPrefix(p.ID, playlistName) {
			targetPlaylist = &p
			break
		}
	}

	if targetPlaylist == nil {
		return fmt.Errorf("playlist not found: %s", playlistName)
	}

	// Get hash from path
	hash := filepath.Base(path)

	err = manager.AddToPlaylist(targetPlaylist.ID, hash)
	if err != nil {
		return fmt.Errorf("adding to playlist: %w", err)
	}

	fmt.Printf("✅ Added to %s: %s\n", targetPlaylist.Name, path)
	return nil
}

func runPlaylistShow(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("playlist name or ID required")
	}
	playlistName := args[0]

	db, manager, err := getCollectionsManager()
	if err != nil {
		return err
	}
	defer db.Close()

	// Find playlist
	playlists, _ := manager.ListPlaylists()
	var targetPlaylist *collections.Playlist
	for _, p := range playlists {
		if strings.EqualFold(p.Name, playlistName) || strings.HasPrefix(p.ID, playlistName) {
			targetPlaylist = &p
			break
		}
	}

	if targetPlaylist == nil {
		return fmt.Errorf("playlist not found: %s", playlistName)
	}

	items, err := manager.ListPlaylistItems(targetPlaylist.ID)
	if err != nil {
		return fmt.Errorf("listing playlist items: %w", err)
	}

	fmt.Printf("📋 %s (%d items):\n\n", targetPlaylist.Name, len(items))
	for i, item := range items {
		fmt.Printf("  %d. %s\n", i+1, item.ImageHash)
	}
	return nil
}

// rateCmd manages ratings
var rateCmd = &cobra.Command{
	Use:   "rate [path] [rating]",
	Short: "Rate wallpapers 1-5 stars",
	Long: `Rate wallpapers to help curate your collection.

Ratings are used for filtering and can be used by the scheduler
to prefer higher-rated wallpapers.`,
	Example: `  wallpaper-cli rate ~/Pictures/wp/01_abc.jpg 5
  wallpaper-cli rate ~/Pictures/wp/02_def.jpg 4`,
	RunE: runRate,
}

var rateNotes string

func init() {
	rootCmd.AddCommand(rateCmd)
	rateCmd.Flags().StringVarP(&rateNotes, "notes", "n", "", "Optional notes about the wallpaper")
}

func runRate(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: rate [path] [1-5]")
	}
	path := args[0]
	rating := 0
	fmt.Sscanf(args[1], "%d", &rating)

	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be 1-5")
	}

	db, manager, err := getCollectionsManager()
	if err != nil {
		return err
	}
	defer db.Close()

	// Get hash from path
	hash := filepath.Base(path)

	err = manager.SetRating(hash, rating, rateNotes)
	if err != nil {
		return fmt.Errorf("setting rating: %w", err)
	}

	stars := strings.Repeat("★", rating) + strings.Repeat("☆", 5-rating)
	fmt.Printf("Rated %s: %s\n", stars, path)
	if rateNotes != "" {
		fmt.Printf("Notes: %s\n", rateNotes)
	}
	return nil
}

// getCollectionsManager creates a database connection and collections manager
func getCollectionsManager() (*data.DB, *collections.Manager, error) {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".local", "share", "wallpaper-cli", "wallpapers.db")
	db, err := data.NewDB(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("opening database: %w", err)
	}

	manager := collections.NewManager(db)
	return db, manager, nil
}
