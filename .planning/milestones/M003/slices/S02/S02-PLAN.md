# M003-S02: Collection Management — Favorites, Playlists, Ratings

**Slice:** S02 of M003  
**Goal:** Implement favorites, playlists, ratings, and metadata management  
**Estimate:** 6-8 hours  
**Dependencies:** S01 (TUI Overhaul) — uses new UI components

---

## Overview

Build the data layer and CLI commands for collection curation:
- **Favorites:** Star wallpapers for quick access
- **Playlists:** Themed collections ("cozy", "work", "anime landscapes")
- **Ratings:** 1-5 star quality ratings with notes
- **Metadata:** Enhanced wallpaper tracking

---

## Database Schema Additions

```sql
-- Favorites: Quick bookmark system
CREATE TABLE favorites (
    image_hash TEXT PRIMARY KEY,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Ratings: Quality scores 1-5
CREATE TABLE ratings (
    image_hash TEXT PRIMARY KEY,
    rating INTEGER CHECK(rating >= 1 AND rating <= 5),
    notes TEXT,
    rated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Playlists: Themed collections
CREATE TABLE playlists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Playlist items: Many-to-many relationship
CREATE TABLE playlist_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    playlist_id TEXT NOT NULL,
    image_hash TEXT NOT NULL,
    position INTEGER NOT NULL,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
    FOREIGN KEY (image_hash) REFERENCES images(hash) ON DELETE CASCADE,
    UNIQUE(playlist_id, image_hash)
);

-- Search history: Remember common queries
CREATE TABLE search_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query TEXT NOT NULL,
    filters TEXT, -- JSON of applied filters
    result_count INTEGER,
    searched_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_favorites_added ON favorites(added_at);
CREATE INDEX idx_ratings_rating ON ratings(rating);
CREATE INDEX idx_playlist_items_playlist ON playlist_items(playlist_id);
CREATE INDEX idx_playlist_items_position ON playlist_items(playlist_id, position);
CREATE INDEX idx_search_history_time ON search_history(searched_at);
```

---

## CLI Commands

### favorite — Manage favorites

```bash
# Toggle favorite status
wallpaper-cli favorite toggle <path-or-id>

# List all favorites
wallpaper-cli favorite list [--json] [--limit 10]

# Check if wallpaper is favorited
wallpaper-cli favorite check <path-or-id>

# Remove from favorites
wallpaper-cli favorite remove <path-or-id>

# Set from favorites (random)
wallpaper-cli set --favorite --random
```

### playlist — Manage playlists

```bash
# Create new playlist
wallpaper-cli playlist create "cozy-winter" --description "Warm winter vibes"

# List all playlists
wallpaper-cli playlist list

# Add wallpaper to playlist
wallpaper-cli playlist add "cozy-winter" <path-or-id>

# Remove from playlist
wallpaper-cli playlist remove "cozy-winter" <path-or-id>

# Show playlist contents
wallpaper-cli playlist show "cozy-winter"

# Delete playlist
wallpaper-cli playlist delete "cozy-winter"

# Rename playlist
wallpaper-cli playlist rename "cozy-winter" "winter-warmth"

# Set from playlist (random)
wallpaper-cli set --playlist "cozy-winter" --random

# Set next in playlist sequence
wallpaper-cli set --playlist "cozy-winter" --next
```

### rate — Rate wallpapers

```bash
# Rate a wallpaper
wallpaper-cli rate <path-or-id> 5

# Rate with notes
wallpaper-cli rate <path-or-id> 4 --notes "Great colors, slightly blurry"

# Show rating
wallpaper-cli rate show <path-or-id>

# List top-rated
wallpaper-cli list --min-rating 4 --sort rating

# Unset rating
wallpaper-cli rate unset <path-or-id>
```

---

## Data Layer Implementation

### Types

```go
// internal/collections/types.go

package collections

import "time"

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
    ID        int       `json:"id"`
    PlaylistID string   `json:"playlist_id"`
    ImageHash string   `json:"image_hash"`
    Position  int      `json:"position"`
    AddedAt   time.Time `json:"added_at"`
    
    // Joined data from images table
    Source     string `json:"source"`
    Resolution string `json:"resolution"`
    LocalPath  string `json:"local_path"`
}
```

### Manager Interface

```go
// internal/collections/manager.go

package collections

import "github.com/user/wallpaper-cli/internal/data"

type Manager struct {
    db *data.DB
}

func NewManager(db *data.DB) *Manager {
    return &Manager{db: db}
}

// Favorites
func (m *Manager) ToggleFavorite(imageHash string) (isFavorite bool, err error)
func (m *Manager) IsFavorite(imageHash string) (bool, error)
func (m *Manager) ListFavorites(limit int) ([]Favorite, error)
func (m *Manager) RemoveFavorite(imageHash string) error

// Ratings
func (m *Manager) SetRating(imageHash string, rating int, notes string) error
func (m *Manager) GetRating(imageHash string) (*Rating, error)
func (m *Manager) ListByMinRating(minRating int, limit int) ([]data.ImageRecord, error)

// Playlists
func (m *Manager) CreatePlaylist(name, description string) (*Playlist, error)
func (m *Manager) ListPlaylists() ([]Playlist, error)
func (m *Manager) GetPlaylist(id string) (*Playlist, error)
func (m *Manager) UpdatePlaylist(id string, name, description string) error
func (m *Manager) DeletePlaylist(id string) error
func (m *Manager) AddToPlaylist(playlistID, imageHash string) error
func (m *Manager) RemoveFromPlaylist(playlistID, imageHash string) error
func (m *Manager) ListPlaylistItems(playlistID string) ([]PlaylistItem, error)
func (m *Manager) ReorderPlaylist(playlistID string, newOrder []string) error
func (m *Manager) GetNextInPlaylist(playlistID string, currentHash string) (string, error)

// Stats
func (m *Manager) GetCollectionStats() (*Stats, error)
```

---

## TUI Integration

### List Enhancements

```go
// Show favorite and rating in list items
func renderListItem(wallpaper data.ImageRecord, collections *collections.Manager) string {
    fav := ""
    if collections.IsFavorite(wallpaper.Hash) {
        fav = "⭐"
    }
    
    rating := ""
    if r, _ := collections.GetRating(wallpaper.Hash); r != nil {
        rating = strings.Repeat("★", r.Rating)
    }
    
    return fmt.Sprintf("%s %s %s", fav, wallpaper.Filename, rating)
}
```

### Preview Pane Actions

```go
// Right pane shows interactive actions
func renderActions(m Model) string {
    actions := []string{
        "[Enter] Set as wallpaper",
        "[f] Toggle favorite",
        "[r] Rate 1-5",
        "[p] Add to playlist",
        "[P] Create playlist",
    }
    
    if m.selectedItem.IsFavorite {
        actions[1] = "[f] ★ Unfavorite"
    }
    
    return lipgloss.JoinVertical(lipgloss.Left, actions...)
}

// Rating selector overlay
func renderRatingSelector(currentRating int) string {
    stars := []string{"☆", "☆", "☆", "☆", "☆"}
    for i := 0; i < currentRating; i++ {
        stars[i] = "★"
    }
    return fmt.Sprintf("Rate: %s (1-5)", strings.Join(stars, " "))
}
```

### Playlist Selector Modal

```go
// Modal overlay for playlist selection
type PlaylistSelector struct {
    playlists []collections.Playlist
    selected  int
}

func (ps PlaylistSelector) View() string {
    var items []string
    items = append(items, "Select Playlist:")
    items = append(items, "")
    
    for i, p := range ps.playlists {
        marker := "  "
        if i == ps.selected {
            marker := "▸ "
        }
        items = append(items, fmt.Sprintf("%s%s (%d items)", marker, p.Name, p.ItemCount))
    }
    
    items = append(items, "")
    items = append(items, "[n] New playlist  [Enter] Select  [Esc] Cancel")
    
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Padding(1).
        Render(strings.Join(items, "\n"))
}
```

---

## Tasks

| ID | Title | Est. | Details |
|----|-------|------|---------|
| T01 | Create database migrations | 1h | Add favorites, ratings, playlists tables |
| T02 | Implement favorites manager | 1h | Toggle, check, list, remove |
| T03 | Implement ratings manager | 1h | Set, get, list by rating |
| T04 | Implement playlist manager | 2h | CRUD, items, reorder, next |
| T05 | Create `favorite` command | 0.5h | CLI commands |
| T06 | Create `playlist` command | 1h | CLI commands |
| T07 | Create `rate` command | 0.5h | CLI commands |
| T08 | Integrate with TUI | 1h | List display, actions, modals |
| T09 | Add `set` flags integration | 0.5h | --favorite, --playlist flags |

**Total: 7.5 hours**

---

## Integration Points

### With S01 (TUI)
- List items show ⭐ and ★ rating indicators
- Preview pane has interactive favorite/rate/playlist buttons
- Modal dialogs for playlist selection and rating

### With S03 (Scheduling)
- Daemon can rotate through favorites: `--source favorites`
- Daemon can rotate through playlist: `--playlist "cozy"`
- Schedule can specify playlist for time-based themes

### With S04 (Daemon)
- State includes current playlist position
- Resume from last position in playlist after reboot
- Favorites filter for curated rotation

---

## Testing

**Unit Tests:**
- [ ] Toggle favorite (adds, removes, re-adds)
- [ ] Set rating (valid 1-5, invalid rejected)
- [ ] Playlist CRUD (create, read, update, delete)
- [ ] Playlist item order (add, remove, reorder)
- [ ] Next in playlist (sequential order)

**Integration Tests:**
- [ ] CLI favorite commands
- [ ] CLI playlist commands
- [ ] CLI rate commands
- [ ] TUI favorite toggle
- [ ] TUI rating selector
- [ ] TUI playlist modal

---

## Success Criteria

- [ ] Database migrations run successfully
- [ ] Favorites persist across sessions
- [ ] Playlists can be created, populated, and deleted
- [ ] Ratings are 1-5 integers with optional notes
- [ ] TUI shows ⭐ and ★ indicators in list
- [ ] TUI preview pane has favorite/rate/playlist actions
- [ ] `set --favorite` works
- [ ] `set --playlist <name>` works
- [ ] Playlist rotation tracks position correctly

---

*Collection management transforms the CLI from a download tool into a curation platform — users can build meaningful, organized wallpaper collections.*
