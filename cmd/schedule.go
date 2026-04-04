package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/wallpaper-cli/internal/collections"
	"github.com/user/wallpaper-cli/internal/data"
	"github.com/user/wallpaper-cli/internal/platform"
	"github.com/user/wallpaper-cli/internal/schedule"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage wallpaper rotation schedules",
	Long: `Create and manage automatic wallpaper rotation schedules.

Schedule types:
  • Interval: Rotate every N minutes/hours
  • Fixed: Rotate at specific times (e.g., 8:00, 12:00, 18:00)
  • Theme: Different wallpapers for morning/day/evening/night`,
}

var (
	schedName     string
	schedType     string
	schedInterval string
	schedTimes    []string
	schedSource   string
	schedPlaylist string
	schedShuffle  bool
)

var scheduleCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new schedule",
	Example: `  # Rotate every 30 minutes from favorites
  wallpaper-cli schedule create "work-mode" --interval 30m --source favorites

  # Rotate at specific times
  wallpaper-cli schedule create "meal-times" --times "07:00,12:00,19:00" --source all

  # Theme-based rotation
  wallpaper-cli schedule create "day-cycle" --theme --morning 06:00 --day 09:00 --evening 18:00`,
	RunE: runScheduleCreate,
}

var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all schedules",
	RunE:  runScheduleList,
}

var scheduleEnableCmd = &cobra.Command{
	Use:   "enable [schedule-id]",
	Short: "Enable a schedule",
	RunE:  runScheduleEnable,
}

var scheduleDisableCmd = &cobra.Command{
	Use:   "disable [schedule-id]",
	Short: "Disable a schedule",
	RunE:  runScheduleDisable,
}

var scheduleDeleteCmd = &cobra.Command{
	Use:   "delete [schedule-id]",
	Short: "Delete a schedule",
	RunE:  runScheduleDelete,
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.AddCommand(scheduleCreateCmd)
	scheduleCmd.AddCommand(scheduleListCmd)
	scheduleCmd.AddCommand(scheduleEnableCmd)
	scheduleCmd.AddCommand(scheduleDisableCmd)
	scheduleCmd.AddCommand(scheduleDeleteCmd)

	scheduleCreateCmd.Flags().StringVar(&schedType, "type", "interval", "Schedule type: interval, fixed, theme")
	scheduleCreateCmd.Flags().StringVar(&schedInterval, "interval", "30m", "Rotation interval (e.g., 30m, 1h)")
	scheduleCreateCmd.Flags().StringSliceVar(&schedTimes, "times", nil, "Fixed times (e.g., 08:00,12:00,18:00)")
	scheduleCreateCmd.Flags().StringVar(&schedSource, "source", "favorites", "Source: all, favorites, playlist")
	scheduleCreateCmd.Flags().StringVar(&schedPlaylist, "playlist", "", "Playlist ID (if source=playlist)")
	scheduleCreateCmd.Flags().BoolVar(&schedShuffle, "shuffle", true, "Shuffle wallpapers (vs sequential)")

	// Theme flags
	scheduleCreateCmd.Flags().String("morning", "06:00", "Morning start time")
	scheduleCreateCmd.Flags().String("day", "09:00", "Day start time")
	scheduleCreateCmd.Flags().String("evening", "18:00", "Evening start time")
	scheduleCreateCmd.Flags().String("night", "22:00", "Night start time")
}

func runScheduleCreate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("schedule name required")
	}
	name := args[0]

	// Get collections manager and setter
	engine, err := getScheduleEngine()
	if err != nil {
		return err
	}

	// Create schedule based on type
	s := &schedule.Schedule{
		Name:    name,
		Enabled: true,
		Source:  schedule.SourceType(schedSource),
		Shuffle: schedShuffle,
	}

	switch schedType {
	case "interval":
		s.Type = schedule.ScheduleInterval
		duration, err := time.ParseDuration(schedInterval)
		if err != nil {
			return fmt.Errorf("invalid interval: %w", err)
		}
		s.Interval = duration

	case "fixed":
		s.Type = schedule.ScheduleFixedTime
		s.FixedTimes = schedTimes

	case "theme":
		s.Type = schedule.ScheduleTheme
		s.Theme = &schedule.ThemeConfig{
			MorningStart: cmd.Flag("morning").Value.String(),
			DayStart:     cmd.Flag("day").Value.String(),
			EveningStart: cmd.Flag("evening").Value.String(),
			NightStart:   cmd.Flag("night").Value.String(),
		}

		// For theme, default to favorites
		if schedSource == "all" {
			s.Source = schedule.SourceFavorites
		}
	}

	// Set playlist if specified
	if schedPlaylist != "" {
		s.PlaylistID = schedPlaylist
		s.Source = schedule.SourcePlaylist
	}

	if err := engine.AddSchedule(s); err != nil {
		return fmt.Errorf("creating schedule: %w", err)
	}

	fmt.Printf("📅 Created schedule: %s (id: %s)\n", s.Name, s.ID)
	fmt.Printf("   Type: %s | Source: %s | Shuffle: %v\n", s.Type, s.Source, s.Shuffle)

	return nil
}

func runScheduleList(cmd *cobra.Command, args []string) error {
	engine, err := getScheduleEngine()
	if err != nil {
		return err
	}

	schedules := engine.ListSchedules()
	if len(schedules) == 0 {
		fmt.Println("No schedules. Create one with: wallpaper-cli schedule create <name>")
		return nil
	}

	fmt.Printf("📅 Schedules (%d):\n\n", len(schedules))
	for _, s := range schedules {
		status := "❌ Disabled"
		if s.Enabled {
			status = "✅ Enabled"
		}

		details := ""
		switch s.Type {
		case schedule.ScheduleInterval:
			details = fmt.Sprintf("every %s", s.Interval)
		case schedule.ScheduleFixedTime:
			details = fmt.Sprintf("at %v", s.FixedTimes)
		case schedule.ScheduleTheme:
			details = "theme-based"
		}

		fmt.Printf("  %s: %s (%s) [%s]\n", s.ID[:8], s.Name, details, status)
	}

	return nil
}

func runScheduleEnable(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("schedule ID required")
	}

	engine, err := getScheduleEngine()
	if err != nil {
		return err
	}

	if err := engine.EnableSchedule(args[0]); err != nil {
		return err
	}

	fmt.Printf("✅ Enabled schedule: %s\n", args[0])
	return nil
}

func runScheduleDisable(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("schedule ID required")
	}

	engine, err := getScheduleEngine()
	if err != nil {
		return err
	}

	if err := engine.DisableSchedule(args[0]); err != nil {
		return err
	}

	fmt.Printf("❌ Disabled schedule: %s\n", args[0])
	return nil
}

func runScheduleDelete(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("schedule ID required")
	}

	engine, err := getScheduleEngine()
	if err != nil {
		return err
	}

	if err := engine.RemoveSchedule(args[0]); err != nil {
		return err
	}

	fmt.Printf("🗑️  Deleted schedule: %s\n", args[0])
	return nil
}

// getScheduleEngine creates a schedule engine with all dependencies
func getScheduleEngine() (*schedule.Engine, error) {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".local", "share", "wallpaper-cli", "wallpapers.db")
	db, err := data.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	// Don't close db here - it might be needed by the engine's collections manager
	// The caller should handle cleanup or we need to refactor to share db connections
	_ = db // Keep db reference for potential future use

	manager := collections.NewManager(db)

	setter, err := platform.Get()
	if err != nil {
		return nil, fmt.Errorf("getting platform setter: %w", err)
	}

	engine := schedule.NewEngine(manager, setter)
	return engine, nil
}
