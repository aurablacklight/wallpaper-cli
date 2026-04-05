package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/wallpaper-cli/internal/data"
)

var (
	listSource   string
	listSince    string
	listJSON     bool
	listPathOnly bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List downloaded wallpapers",
	Long:  `List and filter downloaded wallpapers from the database.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listSource, "source", "", "Filter by source (wallhaven, reddit)")
	listCmd.Flags().StringVar(&listSince, "since", "", "Show only files downloaded since (1d, 7d, 30d)")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().BoolVar(&listPathOnly, "path-only", false, "Output only file paths (for piping)")
}

func runList(cmd *cobra.Command, args []string) error {
	// Get database path
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	dbPath := filepath.Join(home, ".local", "share", "wallpaper-cli", "wallpapers.db")

	// Open database
	db, err := data.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	// Build query
	images, err := queryImages(db, listSource, listSince)
	if err != nil {
		return fmt.Errorf("querying images: %w", err)
	}

	// Output based on flags
	if listPathOnly {
		for _, img := range images {
			if img.LocalPath != "" {
				fmt.Println(img.LocalPath)
			}
		}
		return nil
	}

	if listJSON {
		return outputJSON(images)
	}

	return outputTable(images)
}

func queryImages(db *data.DB, source, since string) ([]data.ImageRecord, error) {
	return db.ListImages(source, parseSinceDuration(since))
}

// parseSinceDuration converts a duration string like "1d", "7d", "30d" to
// the corresponding cutoff time. Returns zero time if empty or invalid.
func parseSinceDuration(since string) time.Time {
	if since == "" {
		return time.Time{}
	}
	days := 0
	switch since {
	case "1d":
		days = 1
	case "7d":
		days = 7
	case "30d":
		days = 30
	default:
		if strings.HasSuffix(since, "d") {
			fmt.Sscanf(since, "%dd", &days)
		}
	}
	if days > 0 {
		return time.Now().AddDate(0, 0, -days)
	}
	return time.Time{}
}

func outputJSON(images []data.ImageRecord) error {
	type JSONImage struct {
		Source       string    `json:"source"`
		LocalPath    string    `json:"local_path"`
		DownloadedAt time.Time `json:"downloaded_at"`
	}

	var output []JSONImage
	for _, img := range images {
		output = append(output, JSONImage{
			Source:       img.Source,
			LocalPath:    img.LocalPath,
			DownloadedAt: img.DownloadedAt,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputTable(images []data.ImageRecord) error {
	if len(images) == 0 {
		fmt.Println("No wallpapers found.")
		return nil
	}

	fmt.Printf("Found %d wallpaper(s):\n\n", len(images))
	fmt.Printf("%-12s %-30s %s\n", "Source", "Filename", "Downloaded")
	fmt.Println(strings.Repeat("-", 80))

	for _, img := range images {
		filename := filepath.Base(img.LocalPath)
		if len(filename) > 28 {
			filename = filename[:25] + "..."
		}

		timeStr := img.DownloadedAt.Format("2006-01-02")
		if img.DownloadedAt.IsZero() {
			timeStr = "unknown"
		}

		fmt.Printf("%-12s %-30s %s\n", img.Source, filename, timeStr)
	}

	return nil
}
