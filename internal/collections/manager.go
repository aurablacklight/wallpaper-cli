package collections

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/user/wallpaper-cli/internal/data"
)

// Manager handles collection operations (favorites, playlists, ratings)
type Manager struct {
	db *data.DB
}

// NewManager creates a new collections manager
func NewManager(db *data.DB) *Manager {
	return &Manager{db: db}
}

// Favorite represents a favorited wallpaper
type Favorite struct {
	ImageHash string    `json:"image_hash"`
	AddedAt   time.Time `json:"added_at"`
}

// Rating represents a user's quality rating
type Rating struct {
	ImageHash string    `json:"image_hash"`
	Rating    int       `json:"rating"` // 1-5
	Notes     string    `json:"notes,omitempty"`
	RatedAt   time.Time `json:"rated_at"`
}

// Playlist represents a themed collection
type Playlist struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	ItemCount   int       `json:"item_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PlaylistItem represents a wallpaper in a playlist
type PlaylistItem struct {
	ID         int       `json:"id"`
	PlaylistID string    `json:"playlist_id"`
	ImageHash  string    `json:"image_hash"`
	Position   int       `json:"position"`
	AddedAt    time.Time `json:"added_at"`
}

// ==================== FAVORITES ====================

// ToggleFavorite adds or removes a wallpaper from favorites
func (m *Manager) ToggleFavorite(imageHash string) (isFavorite bool, err error) {
	// Check if already favorited
	exists, err := m.IsFavorite(imageHash)
	if err != nil {
		return false, err
	}

	if exists {
		// Remove from favorites
		err = m.RemoveFavorite(imageHash)
		return false, err
	}

	// Add to favorites
	err = m.AddFavorite(imageHash)
	return true, err
}

// AddFavorite adds a wallpaper to favorites
func (m *Manager) AddFavorite(imageHash string) error {
	query := `INSERT INTO favorites (image_hash) VALUES (?)`
	_, err := m.db.Exec(query, imageHash)
	return err
}

// IsFavorite checks if a wallpaper is favorited
func (m *Manager) IsFavorite(imageHash string) (bool, error) {
	query := `SELECT 1 FROM favorites WHERE image_hash = ?`
	var exists int
	err := m.db.QueryRow(query, imageHash).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ListFavorites returns all favorited wallpapers
func (m *Manager) ListFavorites(limit int) ([]Favorite, error) {
	query := `SELECT image_hash, added_at FROM favorites ORDER BY added_at DESC LIMIT ?`
	rows, err := m.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []Favorite
	for rows.Next() {
		var f Favorite
		var addedAt string
		if err := rows.Scan(&f.ImageHash, &addedAt); err == nil {
			f.AddedAt, _ = time.Parse("2006-01-02 15:04:05", addedAt)
			favorites = append(favorites, f)
		}
	}

	return favorites, rows.Err()
}

// RemoveFavorite removes a wallpaper from favorites
func (m *Manager) RemoveFavorite(imageHash string) error {
	query := `DELETE FROM favorites WHERE image_hash = ?`
	_, err := m.db.Exec(query, imageHash)
	return err
}

// ==================== RATINGS ====================

// SetRating sets a rating for a wallpaper
func (m *Manager) SetRating(imageHash string, rating int, notes string) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	query := `
		INSERT INTO ratings (image_hash, rating, notes) 
		VALUES (?, ?, ?)
		ON CONFLICT(image_hash) DO UPDATE SET 
			rating = excluded.rating,
			notes = excluded.notes,
			rated_at = excluded.rated_at`

	_, err := m.db.Exec(query, imageHash, rating, notes)
	return err
}

// GetRating gets the rating for a wallpaper
func (m *Manager) GetRating(imageHash string) (*Rating, error) {
	query := `SELECT image_hash, rating, notes, rated_at FROM ratings WHERE image_hash = ?`
	var r Rating
	var ratedAt string

	err := m.db.QueryRow(query, imageHash).Scan(&r.ImageHash, &r.Rating, &r.Notes, &ratedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	r.RatedAt, _ = time.Parse("2006-01-02 15:04:05", ratedAt)
	return &r, nil
}

// ListByMinRating returns wallpapers with minimum rating
func (m *Manager) ListByMinRating(minRating int, limit int) ([]data.ImageRecord, error) {
	query := `
		SELECT i.id, i.hash, i.source, i.source_id, i.url, i.local_path, 
		       i.resolution, i.aspect_ratio, i.tags, i.downloaded_at, i.file_size
		FROM images i
		JOIN ratings r ON i.hash = r.image_hash
		WHERE r.rating >= ?
		ORDER BY r.rating DESC, r.rated_at DESC
		LIMIT ?`

	rows, err := m.db.Query(query, minRating, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []data.ImageRecord
	for rows.Next() {
		var img data.ImageRecord
		var downloadedAt string
		if err := rows.Scan(&img.ID, &img.Hash, &img.Source, &img.SourceID, &img.URL,
			&img.LocalPath, &img.Resolution, &img.AspectRatio, &img.Tags,
			&downloadedAt, &img.FileSize); err == nil {
			img.DownloadedAt, _ = time.Parse("2006-01-02 15:04:05", downloadedAt)
			results = append(results, img)
		}
	}

	return results, rows.Err()
}

// ==================== PLAYLISTS ====================

// CreatePlaylist creates a new playlist
func (m *Manager) CreatePlaylist(name, description string) (*Playlist, error) {
	id := generatePlaylistID(name)
	query := `
		INSERT INTO playlists (id, name, description) 
		VALUES (?, ?, ?)`

	_, err := m.db.Exec(query, id, name, description)
	if err != nil {
		return nil, err
	}

	return &Playlist{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// ListPlaylists returns all playlists
func (m *Manager) ListPlaylists() ([]Playlist, error) {
	query := `
		SELECT p.id, p.name, p.description, p.created_at, p.updated_at,
		       COUNT(pi.id) as item_count
		FROM playlists p
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		GROUP BY p.id
		ORDER BY p.updated_at DESC`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []Playlist
	for rows.Next() {
		var p Playlist
		var createdAt, updatedAt string
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &createdAt, &updatedAt, &p.ItemCount); err == nil {
			p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
			p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
			playlists = append(playlists, p)
		}
	}

	return playlists, rows.Err()
}

// GetPlaylist gets a playlist by ID
func (m *Manager) GetPlaylist(id string) (*Playlist, error) {
	query := `
		SELECT p.id, p.name, p.description, p.created_at, p.updated_at,
		       COUNT(pi.id) as item_count
		FROM playlists p
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		WHERE p.id = ?
		GROUP BY p.id`

	var p Playlist
	var createdAt, updatedAt string
	err := m.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Description,
		&createdAt, &updatedAt, &p.ItemCount)
	if err != nil {
		return nil, err
	}

	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &p, nil
}

// UpdatePlaylist updates a playlist
func (m *Manager) UpdatePlaylist(id string, name, description string) error {
	query := `
		UPDATE playlists 
		SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`
	_, err := m.db.Exec(query, name, description, id)
	return err
}

// DeletePlaylist deletes a playlist
func (m *Manager) DeletePlaylist(id string) error {
	query := `DELETE FROM playlists WHERE id = ?`
	_, err := m.db.Exec(query, id)
	return err
}

// AddToPlaylist adds a wallpaper to a playlist
func (m *Manager) AddToPlaylist(playlistID, imageHash string) error {
	// Get next position
	var position int
	query := `SELECT COALESCE(MAX(position), 0) + 1 FROM playlist_items WHERE playlist_id = ?`
	m.db.QueryRow(query, playlistID).Scan(&position)

	// Insert item
	query = `INSERT INTO playlist_items (playlist_id, image_hash, position) VALUES (?, ?, ?)`
	_, err := m.db.Exec(query, playlistID, imageHash, position)

	// Update playlist timestamp
	if err == nil {
		m.db.Exec(`UPDATE playlists SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, playlistID)
	}

	return err
}

// RemoveFromPlaylist removes a wallpaper from a playlist
func (m *Manager) RemoveFromPlaylist(playlistID, imageHash string) error {
	query := `DELETE FROM playlist_items WHERE playlist_id = ? AND image_hash = ?`
	_, err := m.db.Exec(query, playlistID, imageHash)
	return err
}

// ListPlaylistItems returns all items in a playlist
func (m *Manager) ListPlaylistItems(playlistID string) ([]PlaylistItem, error) {
	query := `
		SELECT id, playlist_id, image_hash, position, added_at
		FROM playlist_items
		WHERE playlist_id = ?
		ORDER BY position`

	rows, err := m.db.Query(query, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PlaylistItem
	for rows.Next() {
		var item PlaylistItem
		var addedAt string
		if err := rows.Scan(&item.ID, &item.PlaylistID, &item.ImageHash, &item.Position, &addedAt); err == nil {
			item.AddedAt, _ = time.Parse("2006-01-02 15:04:05", addedAt)
			items = append(items, item)
		}
	}

	return items, rows.Err()
}

// ReorderPlaylist reorders items in a playlist
func (m *Manager) ReorderPlaylist(playlistID string, newOrder []string) error {
	// Update positions
	for i, hash := range newOrder {
		query := `UPDATE playlist_items SET position = ? WHERE playlist_id = ? AND image_hash = ?`
		if _, err := m.db.Exec(query, i+1, playlistID, hash); err != nil {
			return err
		}
	}
	return nil
}

// GetNextInPlaylist gets the next wallpaper in sequence
func (m *Manager) GetNextInPlaylist(playlistID string, currentHash string) (string, error) {
	// Get current position
	var currentPos int
	query := `SELECT position FROM playlist_items WHERE playlist_id = ? AND image_hash = ?`
	err := m.db.QueryRow(query, playlistID, currentHash).Scan(&currentPos)
	if err != nil {
		return "", err
	}

	// Get next item
	var nextHash string
	query = `SELECT image_hash FROM playlist_items WHERE playlist_id = ? AND position > ? ORDER BY position LIMIT 1`
	err = m.db.QueryRow(query, playlistID, currentPos).Scan(&nextHash)
	if err == sql.ErrNoRows {
		// Wrap around to first item
		query = `SELECT image_hash FROM playlist_items WHERE playlist_id = ? ORDER BY position LIMIT 1`
		err = m.db.QueryRow(query, playlistID).Scan(&nextHash)
	}

	return nextHash, err
}

// ==================== STATS ====================

// GetCollectionStats returns statistics about the collection
func (m *Manager) GetCollectionStats() (favCount, ratedCount, playlistCount int, err error) {
	// Favorite count
	m.db.QueryRow(`SELECT COUNT(*) FROM favorites`).Scan(&favCount)

	// Rated count
	m.db.QueryRow(`SELECT COUNT(*) FROM ratings`).Scan(&ratedCount)

	// Playlist count
	m.db.QueryRow(`SELECT COUNT(*) FROM playlists`).Scan(&playlistCount)

	return favCount, ratedCount, playlistCount, nil
}

// Helper function
func generatePlaylistID(name string) string {
	// Simple ID generation: lowercase + timestamp
	id := fmt.Sprintf("%s_%d", name, time.Now().Unix())
	return id
}
