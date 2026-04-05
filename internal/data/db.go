package data

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/wallpaper-cli/internal/model"
	_ "modernc.org/sqlite"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// ImageRecord represents a downloaded image in the database
type ImageRecord struct {
	ID            int64
	Hash          string
	Source        string
	SourceID      string
	URL           string
	LocalPath     string
	Resolution    string
	AspectRatio   string
	Tags          string // JSON array
	DownloadedAt  time.Time
	FileSize      int64
}

// NewDB creates a new database connection
func NewDB(path string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	// Open database
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db := &DB{conn: conn}

	// Create tables
	if err := db.createTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("creating tables: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Exec executes a query that doesn't return rows.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// Query executes a query that returns rows.
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryRow executes a query that returns at most one row.
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// createTables creates the database schema
func (db *DB) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT UNIQUE NOT NULL,
		source TEXT NOT NULL,
		source_id TEXT,
		url TEXT NOT NULL,
		local_path TEXT,
		resolution TEXT,
		aspect_ratio TEXT,
		tags TEXT,
		downloaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		file_size INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_images_hash ON images(hash);
	CREATE INDEX IF NOT EXISTS idx_images_source ON images(source);
	CREATE INDEX IF NOT EXISTS idx_images_source_id ON images(source, source_id);

	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT
	);

	CREATE TABLE IF NOT EXISTS favorites (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		image_hash TEXT UNIQUE NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (image_hash) REFERENCES images(hash)
	);

	CREATE TABLE IF NOT EXISTS ratings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		image_hash TEXT UNIQUE NOT NULL,
		rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
		notes TEXT DEFAULT '',
		rated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (image_hash) REFERENCES images(hash)
	);

	CREATE TABLE IF NOT EXISTS playlists (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS playlist_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		playlist_id TEXT NOT NULL,
		image_hash TEXT NOT NULL,
		position INTEGER NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
		FOREIGN KEY (image_hash) REFERENCES images(hash),
		UNIQUE(playlist_id, image_hash)
	);

	CREATE INDEX IF NOT EXISTS idx_favorites_hash ON favorites(image_hash);
	CREATE INDEX IF NOT EXISTS idx_ratings_hash ON ratings(image_hash);
	CREATE INDEX IF NOT EXISTS idx_playlist_items_playlist ON playlist_items(playlist_id);

	CREATE TABLE IF NOT EXISTS source_tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		category TEXT DEFAULT '',
		category_id INTEGER DEFAULT 0,
		source TEXT NOT NULL,
		UNIQUE(source, name)
	);

	CREATE INDEX IF NOT EXISTS idx_source_tags_name ON source_tags(name);
	CREATE INDEX IF NOT EXISTS idx_source_tags_source ON source_tags(source);
	CREATE INDEX IF NOT EXISTS idx_source_tags_category ON source_tags(category);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// SaveImage saves an image record to the database
func (db *DB) SaveImage(img *ImageRecord) error {
	query := `
	INSERT INTO images (hash, source, source_id, url, local_path, resolution, aspect_ratio, tags, downloaded_at, file_size)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(hash) DO UPDATE SET
		local_path = excluded.local_path,
		resolution = excluded.resolution,
		aspect_ratio = excluded.aspect_ratio,
		tags = excluded.tags,
		file_size = excluded.file_size,
		downloaded_at = excluded.downloaded_at
	`

	_, err := db.conn.Exec(query,
		img.Hash,
		img.Source,
		img.SourceID,
		img.URL,
		img.LocalPath,
		img.Resolution,
		img.AspectRatio,
		img.Tags,
		img.DownloadedAt,
		img.FileSize,
	)

	return err
}

// GetImageByHash retrieves an image by its hash
func (db *DB) GetImageByHash(hash string) (*ImageRecord, error) {
	query := `SELECT id, hash, source, source_id, url, local_path, resolution, aspect_ratio, tags, downloaded_at, file_size
	FROM images WHERE hash = ?`

	row := db.conn.QueryRow(query, hash)
	return scanImage(row)
}

// FindSimilarImages finds images with similar hashes
func (db *DB) FindSimilarImages(hash string, threshold int) ([]ImageRecord, error) {
	// For now, exact match only
	// Hamming distance search would require custom SQL
	query := `SELECT id, hash, source, source_id, url, local_path, resolution, aspect_ratio, tags, downloaded_at, file_size
	FROM images WHERE hash = ?`

	rows, err := db.conn.Query(query, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ImageRecord
	for rows.Next() {
		img, err := scanImage(rows)
		if err != nil {
			continue
		}
		results = append(results, *img)
	}

	return results, rows.Err()
}

// ImageExists checks if an image exists by hash
func (db *DB) ImageExists(hash string) (bool, error) {
	query := `SELECT 1 FROM images WHERE hash = ?`
	var exists int
	err := db.conn.QueryRow(query, hash).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetStats returns database statistics
func (db *DB) GetStats() (total int, bySource map[string]int, err error) {
	// Total count
	err = db.conn.QueryRow("SELECT COUNT(*) FROM images").Scan(&total)
	if err != nil {
		return 0, nil, err
	}

	// Count by source
	rows, err := db.conn.Query("SELECT source, COUNT(*) FROM images GROUP BY source")
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	bySource = make(map[string]int)
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err == nil {
			bySource[source] = count
		}
	}

	return total, bySource, rows.Err()
}

// ListImages returns images with optional filtering by source and minimum date.
func (db *DB) ListImages(source string, since time.Time) ([]ImageRecord, error) {
	query := `SELECT id, hash, source, source_id, url, local_path, resolution, aspect_ratio, tags, downloaded_at, file_size
		FROM images WHERE 1=1`
	var args []interface{}

	if source != "" {
		query += ` AND source = ?`
		args = append(args, source)
	}
	if !since.IsZero() {
		query += ` AND downloaded_at >= ?`
		args = append(args, since.Format("2006-01-02 15:04:05"))
	}

	query += ` ORDER BY downloaded_at DESC`

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ImageRecord
	for rows.Next() {
		img, err := scanImage(rows)
		if err != nil {
			continue
		}
		results = append(results, *img)
	}
	return results, rows.Err()
}

// rowScanner interface for scanning rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanImage scans an image from a row
func scanImage(row rowScanner) (*ImageRecord, error) {
	var img ImageRecord
	var downloadedAt sql.NullString

	err := row.Scan(
		&img.ID,
		&img.Hash,
		&img.Source,
		&img.SourceID,
		&img.URL,
		&img.LocalPath,
		&img.Resolution,
		&img.AspectRatio,
		&img.Tags,
		&downloadedAt,
		&img.FileSize,
	)
	if err != nil {
		return nil, err
	}

	if downloadedAt.Valid {
		img.DownloadedAt, _ = time.Parse("2006-01-02 15:04:05", downloadedAt.String)
	}

	return &img, nil
}

// SaveTags upserts a batch of tags into source_tags.
func (db *DB) SaveTags(tags []model.Tag) error {
	if len(tags) == 0 {
		return nil
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO source_tags (name, category, category_id, source)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(source, name) DO UPDATE SET
			category = excluded.category,
			category_id = excluded.category_id`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, t := range tags {
		if _, err := stmt.Exec(t.Name, t.Category, t.CategoryID, t.Source); err != nil {
			tx.Rollback()
			return fmt.Errorf("exec tag %q: %w", t.Name, err)
		}
	}
	return tx.Commit()
}

// GetTags returns all tags for a source, or all sources if source is empty.
func (db *DB) GetTags(source string) ([]model.Tag, error) {
	query := `SELECT name, category, category_id, source FROM source_tags`
	var args []interface{}
	if source != "" {
		query += ` WHERE source = ?`
		args = append(args, source)
	}
	query += ` ORDER BY name`

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.Name, &t.Category, &t.CategoryID, &t.Source); err != nil {
			continue
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// SearchTags finds tags matching a prefix query.
func (db *DB) SearchTags(query string) ([]model.Tag, error) {
	rows, err := db.conn.Query(
		`SELECT name, category, category_id, source FROM source_tags WHERE name LIKE ? ORDER BY name LIMIT 50`,
		query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.Name, &t.Category, &t.CategoryID, &t.Source); err != nil {
			continue
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
