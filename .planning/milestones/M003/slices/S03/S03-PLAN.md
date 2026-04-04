# M003-S03: Scheduling Engine — Time-Based Rotation Logic

**Slice:** S03 of M003  
**Goal:** Implement core rotation scheduling logic, time-based themes, and interval management  
**Estimate:** 5-6 hours  
**Dependencies:** S02 (Collection Management) — uses playlists/favorites as rotation sources

---

## Overview

Build the scheduling engine that powers automatic wallpaper rotation:
- **Interval-based:** Rotate every N minutes/hours
- **Time-based themes:** Morning/day/evening/night wallpapers
- **Event triggers:** Workspace switch, idle detection (future)
- **Source selection:** Rotate through all, favorites, or specific playlists

---

## Core Concepts

### Schedule Types

| Type | Description | Example |
|------|-------------|---------|
| **Interval** | Fixed time between changes | Every 30 minutes |
| **Fixed Time** | Change at specific times | 8:00am, 12:00pm, 6:00pm |
| **Smart Theme** | Dynamic based on time of day | Morning→Day→Evening→Night |
| **Playlist Sequence** | Advance through playlist in order | Next item every hour |

### Schedule Configuration

```go
// internal/schedule/types.go

package schedule

import "time"

// Schedule defines when and how to rotate wallpapers
type Schedule struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Enabled     bool          `json:"enabled"`
    Type        ScheduleType  `json:"type"`
    
    // Interval settings
    Interval    time.Duration `json:"interval,omitempty"`
    
    // Fixed time settings
    FixedTimes  []string      `json:"fixed_times,omitempty"` // ["08:00", "12:00", "18:00"]
    
    // Theme settings
    Theme       *ThemeConfig  `json:"theme,omitempty"`
    
    // Source settings
    Source      SourceType    `json:"source"`
    PlaylistID  string        `json:"playlist_id,omitempty"`
    Favorites   bool          `json:"favorites,omitempty"`
    
    // Shuffle or sequential
    Shuffle     bool          `json:"shuffle"`
    
    // State
    LastRun     time.Time     `json:"last_run"`
    NextRun     time.Time     `json:"next_run"`
    CurrentPos  int           `json:"current_pos"` // Position in playlist or history
}

type ScheduleType string
const (
    ScheduleInterval   ScheduleType = "interval"
    ScheduleFixedTime  ScheduleType = "fixed_time"
    ScheduleTheme      ScheduleType = "theme"
)

type SourceType string
const (
    SourceAll        SourceType = "all"
    SourceFavorites  SourceType = "favorites"
    SourcePlaylist   SourceType = "playlist"
)

// ThemeConfig for time-based themes
type ThemeConfig struct {
    MorningStart string `json:"morning_start"`  // "06:00"
    DayStart     string `json:"day_start"`      // "09:00"
    EveningStart string `json:"evening_start"`  // "18:00"
    NightStart   string `json:"night_start"`    // "22:00"
    
    MorningPlaylist string `json:"morning_playlist,omitempty"`
    DayPlaylist     string `json:"day_playlist,omitempty"`
    EveningPlaylist string `json:"evening_playlist,omitempty"`
    NightPlaylist   string `json:"night_playlist,omitempty"`
}
```

---

## CLI Commands

### schedule — Manage schedules

```bash
# Create interval-based schedule
wallpaper-cli schedule create "work-focus" \
  --every 30m \
  --source favorites \
  --shuffle

# Create theme-based schedule
wallpaper-cli schedule create "day-cycle" \
  --theme \
  --morning 06:00 --morning-playlist "energetic" \
  --day 09:00 --day-playlist "focus" \
  --evening 18:00 --evening-playlist "cozy" \
  --night 22:00 --night-playlist "calm"

# Create fixed-time schedule
wallpaper-cli schedule create "meals" \
  --at "07:00,12:00,19:00" \
  --playlist "food-aesthetic"

# List all schedules
wallpaper-cli schedule list

# Enable/disable schedule
wallpaper-cli schedule enable "work-focus"
wallpaper-cli schedule disable "work-focus"

# Delete schedule
wallpaper-cli schedule delete "work-focus"

# Show schedule details
wallpaper-cli schedule show "work-focus"

# Preview what would be set (dry run)
wallpaper-cli schedule preview "work-focus"

# Trigger immediate rotation
wallpaper-cli schedule trigger "work-focus"
```

### Quick Scheduling (Convenience)

```bash
# Quick interval (creates default schedule)
wallpaper-cli schedule --every 30m --random

# Quick theme (creates default theme schedule)
wallpaper-cli schedule --theme daily

# Stop all rotation
wallpaper-cli schedule --stop
```

---

## Scheduler Engine

### Core Interface

```go
// internal/schedule/engine.go

package schedule

import (
    "time"
    "github.com/user/wallpaper-cli/internal/collections"
    "github.com/user/wallpaper-cli/internal/platform"
)

type Engine struct {
    schedules  map[string]*Schedule
    manager    *collections.Manager
    setter     platform.Setter
    state      *StateManager
    
    // Runtime
    running    bool
    stopChan   chan bool
    tickInterval time.Duration
}

func NewEngine(manager *collections.Manager, setter platform.Setter) *Engine {
    return &Engine{
        schedules: make(map[string]*Schedule),
        manager:     manager,
        setter:      setter,
        state:       NewStateManager(),
        tickInterval: 1 * time.Minute, // Check every minute
    }
}

// Lifecycle
func (e *Engine) Start() error
func (e *Engine) Stop() error
func (e *Engine) IsRunning() bool

// Schedule Management
func (e *Engine) AddSchedule(s *Schedule) error
func (e *Engine) RemoveSchedule(id string) error
func (e *Engine) GetSchedule(id string) (*Schedule, error)
func (e *Engine) ListSchedules() []*Schedule
func (e *Engine) EnableSchedule(id string) error
func (e *Engine) DisableSchedule(id string) error

// Execution
func (e *Engine) Tick() // Check and execute due schedules
func (e *Engine) ExecuteSchedule(id string) error // Manual trigger
```

### Schedule Evaluation

```go
func (e *Engine) shouldExecute(s *Schedule) bool {
    if !s.Enabled {
        return false
    }
    
    now := time.Now()
    
    switch s.Type {
    case ScheduleInterval:
        return now.After(s.NextRun)
        
    case ScheduleFixedTime:
        // Check if any fixed time matches current minute
        currentTime := now.Format("15:04")
        for _, ft := range s.FixedTimes {
            if ft == currentTime && now.Sub(s.LastRun) > 1*time.Minute {
                return true
            }
        }
        return false
        
    case ScheduleTheme:
        return e.shouldExecuteTheme(s, now)
    }
    
    return false
}

func (e *Engine) shouldExecuteTheme(s *Schedule, now time.Time) bool {
    if s.Theme == nil {
        return false
    }
    
    currentTime := now.Format("15:04")
    themeTimes := []string{
        s.Theme.MorningStart,
        s.Theme.DayStart,
        s.Theme.EveningStart,
        s.Theme.NightStart,
    }
    
    for _, tt := range themeTimes {
        if currentTime == tt && now.Sub(s.LastRun) > 1*time.Minute {
            return true
        }
    }
    
    return false
}
```

### Wallpaper Selection

```go
func (e *Engine) selectWallpaper(s *Schedule) (string, error) {
    var candidates []string
    
    // Get candidates based on source
    switch s.Source {
    case SourceFavorites:
        favs, _ := e.manager.ListFavorites(1000)
        for _, f := range favs {
            candidates = append(candidates, f.ImageHash)
        }
        
    case SourcePlaylist:
        items, _ := e.manager.ListPlaylistItems(s.PlaylistID)
        for _, item := range items {
            candidates = append(candidates, item.ImageHash)
        }
        
    case SourceAll:
        // Get all wallpapers from database or filesystem
        candidates = e.getAllWallpapers()
    }
    
    if len(candidates) == 0 {
        return "", fmt.Errorf("no wallpapers found for source: %s", s.Source)
    }
    
    // Select based on mode
    if s.Shuffle {
        return candidates[rand.Intn(len(candidates))], nil
    } else {
        // Sequential
        s.CurrentPos = (s.CurrentPos + 1) % len(candidates)
        return candidates[s.CurrentPos], nil
    }
}

func (e *Engine) getCurrentThemePlaylist(s *Schedule, now time.Time) string {
    if s.Theme == nil {
        return ""
    }
    
    currentTime := now.Format("15:04")
    
    // Determine current time period
    switch {
    case currentTime >= s.Theme.NightStart:
        return s.Theme.NightPlaylist
    case currentTime >= s.Theme.EveningStart:
        return s.Theme.EveningPlaylist
    case currentTime >= s.Theme.DayStart:
        return s.Theme.DayPlaylist
    case currentTime >= s.Theme.MorningStart:
        return s.Theme.MorningPlaylist
    default:
        // Before morning start = night
        return s.Theme.NightPlaylist
    }
}
```

---

## State Management

### Persistent State

```go
// internal/schedule/state.go

type StateManager struct {
    statePath string
}

func NewStateManager() *StateManager {
    home, _ := os.UserHomeDir()
    return &StateManager{
        statePath: filepath.Join(home, ".local", "share", "wallpaper-cli", "schedule-state.json"),
    }
}

type ScheduleState struct {
    Schedules  map[string]*Schedule `json:"schedules"`
    LastRun    time.Time           `json:"last_run"`
    CurrentWallpaper string         `json:"current_wallpaper"`
}

func (sm *StateManager) Load() (*ScheduleState, error)
func (sm *StateManager) Save(state *ScheduleState) error
```

### State Structure (JSON)

```json
{
  "schedules": {
    "work-focus": {
      "id": "work-focus",
      "name": "Work Focus",
      "enabled": true,
      "type": "interval",
      "interval": "30m",
      "source": "favorites",
      "shuffle": true,
      "last_run": "2026-04-04T14:30:00Z",
      "next_run": "2026-04-04T15:00:00Z",
      "current_pos": 3
    },
    "day-cycle": {
      "id": "day-cycle",
      "name": "Day Cycle",
      "enabled": true,
      "type": "theme",
      "theme": {
        "morning_start": "06:00",
        "day_start": "09:00",
        "evening_start": "18:00",
        "night_start": "22:00",
        "morning_playlist": "energetic",
        "day_playlist": "focus",
        "evening_playlist": "cozy",
        "night_playlist": "calm"
      },
      "last_run": "2026-04-04T08:00:00Z"
    }
  },
  "last_run": "2026-04-04T14:30:00Z",
  "current_wallpaper": "/Users/derek/Pictures/wallpapers/wallhaven/01_abc.jpg"
}
```

---

## Tasks

| ID | Title | Est. | Details |
|----|-------|------|---------|
| T01 | Define schedule types and config | 0.5h | Types, validation, defaults |
| T02 | Implement schedule evaluation | 1h | Interval, fixed-time, theme logic |
| T03 | Build wallpaper selection | 1h | Source filtering, shuffle, sequential |
| T04 | Create schedule manager | 1h | CRUD operations, persistence |
| T05 | Create `schedule` CLI command | 1h | All subcommands |
| T06 | Theme-based scheduling | 1h | Time period detection, playlist switching |
| T07 | State persistence | 0.5h | JSON state, resume support |

**Total: 6 hours**

---

## Integration Points

### With S02 (Collections)
- Uses favorites and playlists as rotation sources
- Sequential mode tracks position within playlist
- Integrates with rating system (could filter by min rating)

### With S04 (Daemon)
- Engine runs inside daemon process
- Daemon calls `engine.Tick()` periodically
- State saved after each execution for crash recovery

### With S01 (TUI)
- TUI shows active schedules in status bar
- Can enable/disable schedules from TUI
- Shows countdown to next rotation

---

## Testing

**Unit Tests:**
- [ ] Interval calculation (30m → correct next run time)
- [ ] Fixed time matching (08:00 triggers at 08:00)
- [ ] Theme detection (06:15 = morning, 20:30 = evening)
- [ ] Wallpaper selection (shuffle vs sequential)
- [ ] Playlist switching on theme change

**Integration Tests:**
- [ ] Create and execute interval schedule
- [ ] Create and execute theme schedule
- [ ] State persistence (save/load)
- [ ] Multiple schedules (one interval, one theme)
- [ ] Empty source handling (graceful error)

---

## Success Criteria

- [ ] Interval schedules execute at correct intervals
- [ ] Fixed-time schedules trigger at specified times
- [ ] Theme schedules switch playlists at theme boundaries
- [ ] Sequential mode advances position correctly
- [ ] Shuffle mode randomly selects (with no repeats in short term)
- [ ] State persists across engine restarts
- [ ] Can have multiple schedules active simultaneously
- [ ] CLI commands work: create, list, enable, disable, delete
- [ ] Dry-run mode shows what would be set without setting

---

## Pre-defined Schedules (Built-in)

```go
// Quick-start schedules users can activate
var BuiltInSchedules = map[string]*Schedule{
    "focus": {
        Name:     "Focus Mode",
        Type:     ScheduleInterval,
        Interval: 30 * time.Minute,
        Source:   SourceFavorites,
        Shuffle:  true,
    },
    "daylight": {
        Name:  "Daylight Cycle",
        Type:  ScheduleTheme,
        Theme: &ThemeConfig{
            MorningStart: "06:00",
            DayStart:     "09:00",
            EveningStart: "18:00",
            NightStart:   "22:00",
        },
    },
}
```

Usage:
```bash
wallpaper-cli schedule enable focus
wallpaper-cli schedule enable daylight
```

---

*The scheduling engine is the brain of M003 — it decides when to rotate, what to show, and how to transition between wallpapers.*
