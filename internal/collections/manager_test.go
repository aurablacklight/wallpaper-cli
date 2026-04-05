package collections

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/user/wallpaper-cli/internal/data"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	db, err := data.NewDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Seed an image so foreign keys are satisfied
	db.SaveImage(&data.ImageRecord{
		Hash: "img1", Source: "test", URL: "https://example.com/1.jpg", DownloadedAt: time.Now(),
	})
	db.SaveImage(&data.ImageRecord{
		Hash: "img2", Source: "test", URL: "https://example.com/2.jpg", DownloadedAt: time.Now(),
	})

	return NewManager(db)
}

func TestFavorites_AddAndRemove(t *testing.T) {
	m := newTestManager(t)

	if err := m.AddFavorite("img1"); err != nil {
		t.Fatalf("AddFavorite: %v", err)
	}

	ok, err := m.IsFavorite("img1")
	if err != nil {
		t.Fatalf("IsFavorite: %v", err)
	}
	if !ok {
		t.Error("IsFavorite returned false after add")
	}

	favs, err := m.ListFavorites(10)
	if err != nil {
		t.Fatalf("ListFavorites: %v", err)
	}
	if len(favs) != 1 {
		t.Errorf("favorites count = %d, want 1", len(favs))
	}

	if err := m.RemoveFavorite("img1"); err != nil {
		t.Fatalf("RemoveFavorite: %v", err)
	}

	ok, _ = m.IsFavorite("img1")
	if ok {
		t.Error("IsFavorite returned true after remove")
	}
}

func TestFavorites_Toggle(t *testing.T) {
	m := newTestManager(t)

	// Toggle on
	isFav, err := m.ToggleFavorite("img1")
	if err != nil {
		t.Fatalf("ToggleFavorite on: %v", err)
	}
	if !isFav {
		t.Error("expected isFavorite=true after first toggle")
	}

	// Toggle off
	isFav, err = m.ToggleFavorite("img1")
	if err != nil {
		t.Fatalf("ToggleFavorite off: %v", err)
	}
	if isFav {
		t.Error("expected isFavorite=false after second toggle")
	}
}

func TestRatings_SetAndGet(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetRating("img1", 4, "good stuff"); err != nil {
		t.Fatalf("SetRating: %v", err)
	}

	r, err := m.GetRating("img1")
	if err != nil {
		t.Fatalf("GetRating: %v", err)
	}
	if r == nil {
		t.Fatal("GetRating returned nil")
	}
	if r.Rating != 4 {
		t.Errorf("rating = %d, want 4", r.Rating)
	}
	if r.Notes != "good stuff" {
		t.Errorf("notes = %q, want 'good stuff'", r.Notes)
	}
}

func TestRatings_Upsert(t *testing.T) {
	m := newTestManager(t)

	m.SetRating("img1", 3, "ok")
	m.SetRating("img1", 5, "amazing")

	r, _ := m.GetRating("img1")
	if r.Rating != 5 {
		t.Errorf("rating after upsert = %d, want 5", r.Rating)
	}
}

func TestRatings_InvalidRange(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetRating("img1", 0, ""); err == nil {
		t.Error("expected error for rating 0")
	}
	if err := m.SetRating("img1", 6, ""); err == nil {
		t.Error("expected error for rating 6")
	}
}

func TestRatings_GetNonExistent(t *testing.T) {
	m := newTestManager(t)

	r, err := m.GetRating("nonexistent")
	if err != nil {
		t.Fatalf("GetRating: %v", err)
	}
	if r != nil {
		t.Error("expected nil for unrated image")
	}
}

func TestPlaylists_CRUD(t *testing.T) {
	m := newTestManager(t)

	// Create
	p, err := m.CreatePlaylist("chill", "relaxing wallpapers")
	if err != nil {
		t.Fatalf("CreatePlaylist: %v", err)
	}
	if p.Name != "chill" {
		t.Errorf("name = %q, want chill", p.Name)
	}

	// List
	playlists, err := m.ListPlaylists()
	if err != nil {
		t.Fatalf("ListPlaylists: %v", err)
	}
	if len(playlists) != 1 {
		t.Fatalf("playlist count = %d, want 1", len(playlists))
	}

	// Get
	got, err := m.GetPlaylist(p.ID)
	if err != nil {
		t.Fatalf("GetPlaylist: %v", err)
	}
	if got.Description != "relaxing wallpapers" {
		t.Errorf("description = %q", got.Description)
	}

	// Update
	if err := m.UpdatePlaylist(p.ID, "chill vibes", "updated desc"); err != nil {
		t.Fatalf("UpdatePlaylist: %v", err)
	}
	got, _ = m.GetPlaylist(p.ID)
	if got.Name != "chill vibes" {
		t.Errorf("name after update = %q", got.Name)
	}

	// Delete
	if err := m.DeletePlaylist(p.ID); err != nil {
		t.Fatalf("DeletePlaylist: %v", err)
	}
	playlists, _ = m.ListPlaylists()
	if len(playlists) != 0 {
		t.Errorf("playlist count after delete = %d", len(playlists))
	}
}

func TestPlaylistItems_AddAndList(t *testing.T) {
	m := newTestManager(t)

	p, _ := m.CreatePlaylist("test", "")

	if err := m.AddToPlaylist(p.ID, "img1"); err != nil {
		t.Fatalf("AddToPlaylist: %v", err)
	}
	if err := m.AddToPlaylist(p.ID, "img2"); err != nil {
		t.Fatalf("AddToPlaylist: %v", err)
	}

	items, err := m.ListPlaylistItems(p.ID)
	if err != nil {
		t.Fatalf("ListPlaylistItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("item count = %d, want 2", len(items))
	}
	if items[0].Position != 1 || items[1].Position != 2 {
		t.Errorf("positions = [%d, %d], want [1, 2]", items[0].Position, items[1].Position)
	}
}

func TestPlaylistItems_RemoveAndNext(t *testing.T) {
	m := newTestManager(t)

	p, _ := m.CreatePlaylist("test", "")
	m.AddToPlaylist(p.ID, "img1")
	m.AddToPlaylist(p.ID, "img2")

	// GetNext wraps around
	next, err := m.GetNextInPlaylist(p.ID, "img2")
	if err != nil {
		t.Fatalf("GetNextInPlaylist: %v", err)
	}
	if next != "img1" {
		t.Errorf("next = %q, want img1 (wrap around)", next)
	}

	// Remove
	if err := m.RemoveFromPlaylist(p.ID, "img1"); err != nil {
		t.Fatalf("RemoveFromPlaylist: %v", err)
	}
	items, _ := m.ListPlaylistItems(p.ID)
	if len(items) != 1 {
		t.Errorf("item count after remove = %d, want 1", len(items))
	}
}

func TestGetCollectionStats(t *testing.T) {
	m := newTestManager(t)

	m.AddFavorite("img1")
	m.SetRating("img1", 5, "")
	m.CreatePlaylist("test", "")

	favs, rated, playlists, err := m.GetCollectionStats()
	if err != nil {
		t.Fatalf("GetCollectionStats: %v", err)
	}
	if favs != 1 {
		t.Errorf("favs = %d, want 1", favs)
	}
	if rated != 1 {
		t.Errorf("rated = %d, want 1", rated)
	}
	if playlists != 1 {
		t.Errorf("playlists = %d, want 1", playlists)
	}
}
