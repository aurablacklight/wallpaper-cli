package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/wallpaper-cli/internal/config"
	"github.com/user/wallpaper-cli/internal/platform"
	"github.com/user/wallpaper-cli/internal/utils"
)

var (
	randomFlag  bool
	latestFlag  bool
	currentFlag bool
)

var setCmd = &cobra.Command{
	Use:   "set [path]",
	Short: "Set desktop wallpaper",
	Long: `Set the desktop wallpaper to a specific image, random image from your collection,
or the most recently downloaded image.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle --current flag (D-17)
		if currentFlag {
			cfg, err := config.Load(config.GetConfigPath())
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if cfg.CurrentWallpaper == "" {
				fmt.Println("No wallpaper currently set")
				return nil
			}
			fmt.Println(cfg.CurrentWallpaper)
			return nil
		}

		var path string

		// Handle --random flag (D-15)
		if randomFlag {
			cfg, err := config.Load(config.GetConfigPath())
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			path, err = utils.GetRandomWallpaper(cfg.OutputDirectory)
			if err != nil {
				return fmt.Errorf("failed to find random wallpaper: %w", err)
			}
			fmt.Printf("Selected random wallpaper: %s\n", path)
		} else if latestFlag {
			// Handle --latest flag (D-16)
			cfg, err := config.Load(config.GetConfigPath())
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			path, err = utils.GetLatestWallpaper(cfg.OutputDirectory)
			if err != nil {
				return fmt.Errorf("failed to find latest wallpaper: %w", err)
			}
			fmt.Printf("Selected latest wallpaper: %s\n", path)
		} else {
			// Handle set <path> (D-14)
			if len(args) == 0 {
				return fmt.Errorf("path required (or use --random or --latest)")
			}
			path = args[0]
		}

		// D-13: Validate file exists
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("wallpaper file not found: %w", err)
		}

		// D-13: Validate it's an image
		if !utils.IsImageFile(path) {
			return fmt.Errorf("file is not a supported image format (supported: jpg, png, gif, bmp, webp)")
		}

		// Get platform setter
		setter, err := platform.Get()
		if err != nil {
			return fmt.Errorf("platform not supported: %w", err)
		}

		// Set wallpaper
		if err := setter.SetWallpaper(path); err != nil {
			return fmt.Errorf("failed to set wallpaper: %w", err)
		}

		// Load config for updating
		cfg, err := config.Load(config.GetConfigPath())
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Determine source for history (D-09)
		source := "manual"
		if randomFlag {
			source = "random"
		} else if latestFlag {
			source = "latest"
		}

		// Add to history and update current (D-08, D-09)
		cfg.AddWallpaper(path, source)

		// Save config
		if err := cfg.Save(config.GetConfigPath()); err != nil {
			// Log warning but don't fail - wallpaper was set successfully
			fmt.Fprintf(os.Stderr, "Warning: failed to save wallpaper to config: %v\n", err)
		}

		fmt.Printf("Wallpaper set successfully on %s\n", setter.Platform())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.Flags().BoolVar(&randomFlag, "random", false, "Set a random wallpaper from your collection")
	setCmd.Flags().BoolVar(&latestFlag, "latest", false, "Set the most recently downloaded wallpaper")
	setCmd.Flags().BoolVar(&currentFlag, "current", false, "Show the currently set wallpaper path")
}
