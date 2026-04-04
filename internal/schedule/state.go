package schedule

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// StateManager handles persistence of schedule state
type StateManager struct {
	statePath string
}

// EngineState represents the persisted state of the scheduling engine
type EngineState struct {
	Schedules        []*Schedule `json:"schedules"`
	LastRun          int64       `json:"last_run"` // Unix timestamp
	CurrentWallpaper string      `json:"current_wallpaper"`
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	home, _ := os.UserHomeDir()
	stateDir := filepath.Join(home, ".local", "share", "wallpaper-cli")
	os.MkdirAll(stateDir, 0755)

	return &StateManager{
		statePath: filepath.Join(stateDir, "schedule-state.json"),
	}
}

// Load loads the engine state from disk
func (sm *StateManager) Load() (*EngineState, error) {
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty state
			return &EngineState{
				Schedules: []*Schedule{},
			}, nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var state EngineState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
	}

	return &state, nil
}

// Save saves the engine state to disk
func (sm *StateManager) Save(state *EngineState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(sm.statePath, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}
