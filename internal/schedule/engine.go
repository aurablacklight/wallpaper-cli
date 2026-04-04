package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/user/wallpaper-cli/internal/collections"
	"github.com/user/wallpaper-cli/internal/platform"
)

// ScheduleType represents the type of schedule
type ScheduleType string

const (
	ScheduleInterval  ScheduleType = "interval"
	ScheduleFixedTime ScheduleType = "fixed_time"
	ScheduleTheme     ScheduleType = "theme"
)

// SourceType represents the wallpaper source for rotation
type SourceType string

const (
	SourceAll       SourceType = "all"
	SourceFavorites SourceType = "favorites"
	SourcePlaylist  SourceType = "playlist"
)

// Schedule defines when and how to rotate wallpapers
type Schedule struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Enabled bool         `json:"enabled"`
	Type    ScheduleType `json:"type"`

	// Interval settings
	Interval time.Duration `json:"interval,omitempty"`

	// Fixed time settings
	FixedTimes []string `json:"fixed_times,omitempty"` // ["08:00", "12:00", "18:00"]

	// Theme settings
	Theme *ThemeConfig `json:"theme,omitempty"`

	// Source settings
	Source     SourceType `json:"source"`
	PlaylistID string     `json:"playlist_id,omitempty"`

	// Shuffle or sequential
	Shuffle bool `json:"shuffle"`

	// State
	LastRun    time.Time `json:"last_run"`
	NextRun    time.Time `json:"next_run"`
	CurrentPos int       `json:"current_pos"`
}

// ThemeConfig for time-based themes
type ThemeConfig struct {
	MorningStart string `json:"morning_start"` // "06:00"
	DayStart     string `json:"day_start"`     // "09:00"
	EveningStart string `json:"evening_start"` // "18:00"
	NightStart   string `json:"night_start"`   // "22:00"

	MorningPlaylist string `json:"morning_playlist,omitempty"`
	DayPlaylist     string `json:"day_playlist,omitempty"`
	EveningPlaylist string `json:"evening_playlist,omitempty"`
	NightPlaylist   string `json:"night_playlist,omitempty"`
}

// Engine manages wallpaper rotation schedules
type Engine struct {
	schedules map[string]*Schedule
	manager   *collections.Manager
	setter    platform.Setter
	state     *StateManager
	cron      *cron.Cron

	running      bool
	tickInterval time.Duration
}

// NewEngine creates a new scheduling engine
func NewEngine(manager *collections.Manager, setter platform.Setter) *Engine {
	e := &Engine{
		schedules:    make(map[string]*Schedule),
		manager:      manager,
		setter:       setter,
		state:        NewStateManager(),
		cron:         cron.New(),
		tickInterval: 1 * time.Minute,
	}

	// Load saved state immediately so CLI commands can access schedules
	e.LoadState()

	return e
}

// Start starts the scheduling engine
func (e *Engine) Start() error {
	// Load saved state
	if err := e.LoadState(); err != nil {
		// Not fatal, start with empty state
	}

	e.running = true

	// Start cron scheduler
	e.cron.Start()

	// Register schedules
	for _, schedule := range e.schedules {
		if schedule.Enabled {
			e.registerSchedule(schedule)
		}
	}

	return nil
}

// Stop stops the scheduling engine
func (e *Engine) Stop() error {
	e.running = false

	// Stop cron
	ctx := e.cron.Stop()
	<-ctx.Done()

	// Save state
	return e.SaveState()
}

// IsRunning returns true if the engine is running
func (e *Engine) IsRunning() bool {
	return e.running
}

// AddSchedule adds a new schedule
func (e *Engine) AddSchedule(s *Schedule) error {
	if s.ID == "" {
		s.ID = generateScheduleID()
	}

	e.schedules[s.ID] = s

	if e.running && s.Enabled {
		e.registerSchedule(s)
	}

	return e.SaveState()
}

// RemoveSchedule removes a schedule
func (e *Engine) RemoveSchedule(id string) error {
	delete(e.schedules, id)
	return e.SaveState()
}

// GetSchedule gets a schedule by ID
func (e *Engine) GetSchedule(id string) (*Schedule, error) {
	s, ok := e.schedules[id]
	if !ok {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}
	return s, nil
}

// ListSchedules returns all schedules
func (e *Engine) ListSchedules() []*Schedule {
	result := make([]*Schedule, 0, len(e.schedules))
	for _, s := range e.schedules {
		result = append(result, s)
	}
	return result
}

// EnableSchedule enables a schedule
func (e *Engine) EnableSchedule(id string) error {
	s, ok := e.schedules[id]
	if !ok {
		return fmt.Errorf("schedule not found: %s", id)
	}

	s.Enabled = true

	if e.running {
		e.registerSchedule(s)
	}

	return e.SaveState()
}

// DisableSchedule disables a schedule
func (e *Engine) DisableSchedule(id string) error {
	s, ok := e.schedules[id]
	if !ok {
		return fmt.Errorf("schedule not found: %s", id)
	}

	s.Enabled = false
	return e.SaveState()
}

// registerSchedule registers a schedule with cron
func (e *Engine) registerSchedule(s *Schedule) {
	switch s.Type {
	case ScheduleInterval:
		// Register interval-based schedule
		e.cron.Schedule(cron.Every(s.Interval), cron.FuncJob(func() {
			e.executeSchedule(s)
		}))

	case ScheduleFixedTime:
		// Register each fixed time
		for _, ft := range s.FixedTimes {
			// Parse time (e.g., "08:00")
			parts := splitTime(ft)
			cronExpr := fmt.Sprintf("%s %s * * *", parts[1], parts[0])
			e.cron.AddFunc(cronExpr, func() {
				e.executeSchedule(s)
			})
		}

	case ScheduleTheme:
		// Register theme boundaries
		if s.Theme != nil {
			times := []string{
				s.Theme.MorningStart,
				s.Theme.DayStart,
				s.Theme.EveningStart,
				s.Theme.NightStart,
			}
			for _, t := range times {
				if t != "" {
					parts := splitTime(t)
					cronExpr := fmt.Sprintf("%s %s * * *", parts[1], parts[0])
					e.cron.AddFunc(cronExpr, func() {
						e.executeSchedule(s)
					})
				}
			}
		}
	}
}

// executeSchedule executes a schedule
func (e *Engine) executeSchedule(s *Schedule) {
	if !s.Enabled {
		return
	}

	// Get wallpaper to set
	path, err := e.selectWallpaper(s)
	if err != nil {
		return
	}

	// Set wallpaper
	e.setter.SetWallpaper(path)

	// Update state
	s.LastRun = time.Now()
	s.CurrentPos++
	e.SaveState()
}

// selectWallpaper selects a wallpaper based on schedule
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
		if s.PlaylistID != "" {
			items, _ := e.manager.ListPlaylistItems(s.PlaylistID)
			for _, item := range items {
				candidates = append(candidates, item.ImageHash)
			}
		}

	case SourceAll:
		// Would need to get all wallpapers from database
		// For now, use favorites as fallback
		favs, _ := e.manager.ListFavorites(1000)
		for _, f := range favs {
			candidates = append(candidates, f.ImageHash)
		}
	}

	// Handle theme-based source
	if s.Type == ScheduleTheme && s.Theme != nil {
		playlistID := e.getCurrentThemePlaylist(s.Theme)
		if playlistID != "" {
			items, _ := e.manager.ListPlaylistItems(playlistID)
			candidates = []string{}
			for _, item := range items {
				candidates = append(candidates, item.ImageHash)
			}
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no wallpapers available")
	}

	// Select based on mode
	if s.Shuffle {
		// Random selection
		idx := time.Now().UnixNano() % int64(len(candidates))
		return candidates[idx], nil
	} else {
		// Sequential
		pos := s.CurrentPos % len(candidates)
		return candidates[pos], nil
	}
}

// getCurrentThemePlaylist returns the playlist for the current time period
func (e *Engine) getCurrentThemePlaylist(theme *ThemeConfig) string {
	now := time.Now()
	hour := now.Hour()
	minute := now.Minute()
	currentTime := hour*60 + minute

	// Parse theme times to minutes
	morning := parseTimeToMinutes(theme.MorningStart)
	day := parseTimeToMinutes(theme.DayStart)
	evening := parseTimeToMinutes(theme.EveningStart)
	night := parseTimeToMinutes(theme.NightStart)

	// Determine current period
	switch {
	case night > 0 && currentTime >= night:
		return theme.NightPlaylist
	case evening > 0 && currentTime >= evening:
		return theme.EveningPlaylist
	case day > 0 && currentTime >= day:
		return theme.DayPlaylist
	case morning > 0 && currentTime >= morning:
		return theme.MorningPlaylist
	default:
		// Before morning start = night
		return theme.NightPlaylist
	}
}

// LoadState loads engine state from disk
func (e *Engine) LoadState() error {
	state, err := e.state.Load()
	if err != nil {
		return err
	}

	for _, s := range state.Schedules {
		e.schedules[s.ID] = s
	}

	return nil
}

// SaveState saves engine state to disk
func (e *Engine) SaveState() error {
	state := &EngineState{
		Schedules: make([]*Schedule, 0, len(e.schedules)),
	}

	for _, s := range e.schedules {
		state.Schedules = append(state.Schedules, s)
	}

	return e.state.Save(state)
}

// Helper functions

func generateScheduleID() string {
	return fmt.Sprintf("schedule_%d", time.Now().Unix())
}

func splitTime(t string) []string {
	// Split "08:00" into ["08", "00"]
	parts := make([]string, 2)
	if len(t) >= 5 {
		parts[0] = t[0:2]
		parts[1] = t[3:5]
	}
	return parts
}

func parseTimeToMinutes(t string) int {
	if t == "" {
		return -1
	}
	parts := splitTime(t)
	hour := 0
	minute := 0
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)
	return hour*60 + minute
}
