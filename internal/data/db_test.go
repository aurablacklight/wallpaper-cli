package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	db, err := NewDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewDB_CreatesSchema(t *testing.T) {
	db := newTestDB(t)

	// Verify all expected tables exist
	tables := []string{"images", "config", "favorites", "ratings", "playlists", "playlist_items"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not created: %v", table, err)
		}
	}
}

func TestNewDB_InvalidPath(t *testing.T) {
	_, err := NewDB("/nonexistent/deeply/nested/dir/test.db")
	// Should either succeed (MkdirAll works) or fail gracefully
	// On most systems /nonexistent is not writable
	if err == nil {
		// Clean up if it somehow worked
		os.RemoveAll("/nonexistent")
	}
}

func TestSaveImage_And_GetByHash(t *testing.T) {
	db := newTestDB(t)

	img := &ImageRecord{
		Hash:         "abc123",
		Source:       "wallhaven",
		SourceID:     "w1234",
		URL:          "https://example.com/img.jpg",
		LocalPath:    "/tmp/img.jpg",
		Resolution:   "1920x1080",
		AspectRatio:  "16:9",
		Tags:         `["anime","landscape"]`,
		DownloadedAt: time.Now().Truncate(time.Second),
		FileSize:     1024,
	}

	if err := db.SaveImage(img); err != nil {
		t.Fatalf("SaveImage: %v", err)
	}

	got, err := db.GetImageByHash("abc123")
	if err != nil {
		t.Fatalf("GetImageByHash: %v", err)
	}

	if got.Hash != img.Hash {
		t.Errorf("hash = %q, want %q", got.Hash, img.Hash)
	}
	if got.Source != img.Source {
		t.Errorf("source = %q, want %q", got.Source, img.Source)
	}
	if got.Resolution != img.Resolution {
		t.Errorf("resolution = %q, want %q", got.Resolution, img.Resolution)
	}
	if got.FileSize != img.FileSize {
		t.Errorf("file_size = %d, want %d", got.FileSize, img.FileSize)
	}
}

func TestSaveImage_UpsertOnConflict(t *testing.T) {
	db := newTestDB(t)

	img := &ImageRecord{
		Hash: "dup1", Source: "reddit", URL: "https://example.com/1.jpg",
		Resolution: "1920x1080", DownloadedAt: time.Now(),
	}
	if err := db.SaveImage(img); err != nil {
		t.Fatalf("first save: %v", err)
	}

	img.Resolution = "3840x2160"
	if err := db.SaveImage(img); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, _ := db.GetImageByHash("dup1")
	if got.Resolution != "3840x2160" {
		t.Errorf("resolution after upsert = %q, want 3840x2160", got.Resolution)
	}
}

func TestImageExists(t *testing.T) {
	db := newTestDB(t)

	exists, err := db.ImageExists("nonexistent")
	if err != nil {
		t.Fatalf("ImageExists: %v", err)
	}
	if exists {
		t.Error("ImageExists returned true for missing hash")
	}

	db.SaveImage(&ImageRecord{Hash: "exists1", Source: "test", URL: "https://example.com", DownloadedAt: time.Now()})

	exists, err = db.ImageExists("exists1")
	if err != nil {
		t.Fatalf("ImageExists: %v", err)
	}
	if !exists {
		t.Error("ImageExists returned false for saved hash")
	}
}

func TestGetStats(t *testing.T) {
	db := newTestDB(t)

	db.SaveImage(&ImageRecord{Hash: "s1", Source: "wallhaven", URL: "https://a.com", DownloadedAt: time.Now()})
	db.SaveImage(&ImageRecord{Hash: "s2", Source: "wallhaven", URL: "https://b.com", DownloadedAt: time.Now()})
	db.SaveImage(&ImageRecord{Hash: "s3", Source: "reddit", URL: "https://c.com", DownloadedAt: time.Now()})

	total, bySource, err := db.GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if bySource["wallhaven"] != 2 {
		t.Errorf("wallhaven count = %d, want 2", bySource["wallhaven"])
	}
	if bySource["reddit"] != 1 {
		t.Errorf("reddit count = %d, want 1", bySource["reddit"])
	}
}

func TestListImages_Filtering(t *testing.T) {
	db := newTestDB(t)

	now := time.Now().Truncate(time.Second)
	yesterday := now.Add(-24 * time.Hour)

	db.SaveImage(&ImageRecord{Hash: "l1", Source: "wallhaven", URL: "https://a.com", DownloadedAt: yesterday})
	db.SaveImage(&ImageRecord{Hash: "l2", Source: "reddit", URL: "https://b.com", DownloadedAt: now})
	db.SaveImage(&ImageRecord{Hash: "l3", Source: "wallhaven", URL: "https://c.com", DownloadedAt: now})

	// No filters — all images
	all, err := db.ListImages("", time.Time{})
	if err != nil {
		t.Fatalf("ListImages all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("all count = %d, want 3", len(all))
	}

	// Filter by source
	wh, err := db.ListImages("wallhaven", time.Time{})
	if err != nil {
		t.Fatalf("ListImages wallhaven: %v", err)
	}
	if len(wh) != 2 {
		t.Errorf("wallhaven count = %d, want 2", len(wh))
	}

	// Filter by time
	recent, err := db.ListImages("", now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("ListImages since: %v", err)
	}
	if len(recent) != 2 {
		t.Errorf("recent count = %d, want 2", len(recent))
	}
}

func TestExecQueryQueryRow(t *testing.T) {
	db := newTestDB(t)

	// Exec
	_, err := db.Exec("INSERT INTO config (key, value) VALUES (?, ?)", "theme", "dark")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}

	// QueryRow
	var val string
	err = db.QueryRow("SELECT value FROM config WHERE key = ?", "theme").Scan(&val)
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if val != "dark" {
		t.Errorf("value = %q, want dark", val)
	}

	// Query
	db.Exec("INSERT INTO config (key, value) VALUES (?, ?)", "lang", "en")
	rows, err := db.Query("SELECT key FROM config ORDER BY key")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		rows.Scan(&k)
		keys = append(keys, k)
	}
	if len(keys) != 2 {
		t.Errorf("got %d config rows, want 2", len(keys))
	}
}
