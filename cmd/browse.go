package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/user/wallpaper-cli/internal/config"
	"github.com/user/wallpaper-cli/internal/platform"
	"github.com/user/wallpaper-cli/internal/tui"
	"github.com/user/wallpaper-cli/internal/utils"
)

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse wallpapers in interactive TUI",
	Long: `Browse your wallpaper collection in an interactive terminal UI.
	
Navigate with arrow keys, press Enter to set a wallpaper, or press 'q' to quit.
	
Features:
  • Thumbnail previews (when terminal supports it)
  • Navigate with ↑/↓ or j/k
  • Press Enter to set wallpaper immediately
  • Press '?' for help
  • Press 'q' or Esc to quit
	
On macOS with WallpaperEngine.app installed:
  • Press 'o' to open WallpaperEngine
  • Press 'd' to dismiss the hint`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load(config.GetConfigPath())
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Find all wallpapers
		wallpapers, err := utils.FindWallpapers(cfg.OutputDirectory)
		if err != nil {
			return fmt.Errorf("failed to scan wallpapers: %w", err)
		}

		if len(wallpapers) == 0 {
			return fmt.Errorf("no wallpapers found in %s. Run 'wallpaper-cli fetch' first", cfg.OutputDirectory)
		}

		// Get platform setter
		setter, err := platform.Get()
		if err != nil {
			return fmt.Errorf("platform not supported: %w", err)
		}

		// Create TUI model
		model, err := tui.NewModel(wallpapers, cfg, setter)
		if err != nil {
			return fmt.Errorf("failed to create TUI: %w", err)
		}

		// Run Bubble Tea program
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(browseCmd)
}
