package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/wallpaper-cli/internal/data"
)

var (
	exportFormat string
	exportOutput string
	exportSource string
	exportSince  string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export wallpaper metadata",
	Long: `Export wallpaper metadata to JSON for integration with other tools.

This command exports metadata about downloaded wallpapers, useful for
integrating with the macOS WallpaperEngine app or other external tools.

Examples:
  # Export all wallpapers to stdout
  wallpaper-cli export

  # Export to a file
  wallpaper-cli export --output ~/Pictures/wallpapers/metadata.json

  # Export only Wallhaven wallpapers from the last 7 days
  wallpaper-cli export --source wallhaven --since 7d

  # Export with specific format (currently only json supported)
  wallpaper-cli export --format json`,
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "Export format (json)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	exportCmd.Flags().StringVar(&exportSource, "source", "", "Filter by source (wallhaven, reddit)")
	exportCmd.Flags().StringVar(&exportSince, "since", "", "Export only files downloaded since (1d, 7d, 30d)")
}

type ExportRecord struct {
	ID           string   `json:"id"`
	Source       string   `json:"source"`
	LocalPath    string   `json:"local_path"`
	URL          string   `json:"url,omitempty"`
	Resolution   string   `json:"resolution,omitempty"`
	AspectRatio  string   `json:"aspect_ratio,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	DownloadedAt string   `json:"downloaded_at"`
	FileSize     int64    `json:"file_size,omitempty"`
}

type ExportData struct {
	Version     string         `json:"version"`
	GeneratedAt string         `json:"generated_at"`
	CLIVersion  string         `json:"cli_version"`
	Count       int            `json:"count"`
	Wallpapers  []ExportRecord `json:"wallpapers"`
}

func runExport(cmd *cobra.Command, args []string) error {
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

	// Query images from database
	images, err := queryImagesForExport(db, exportSource, exportSince)
	if err != nil {
		return fmt.Errorf("querying images: %w", err)
	}

	// Build export data
	export := ExportData{
		Version:     "1.0",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		CLIVersion:  "1.2.0",
		Count:       len(images),
		Wallpapers:  images,
	}

	// Output
	return writeExport(export)
}

func queryImagesForExport(db *data.DB, source, since string) ([]ExportRecord, error) {
	images, err := db.ListImages(source, parseSinceDuration(since))
	if err != nil {
		return nil, err
	}

	var records []ExportRecord
	for _, img := range images {
		var tags []string
		if img.Tags != "" {
			json.Unmarshal([]byte(img.Tags), &tags)
		}
		records = append(records, ExportRecord{
			ID:           img.Hash,
			Source:       img.Source,
			LocalPath:    img.LocalPath,
			URL:          img.URL,
			Resolution:   img.Resolution,
			AspectRatio:  img.AspectRatio,
			Tags:         tags,
			DownloadedAt: img.DownloadedAt.Format(time.RFC3339),
			FileSize:     img.FileSize,
		})
	}

	return records, nil
}

func writeExport(data ExportData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, jsonData, 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Printf("Exported %d wallpaper(s) to %s\n", data.Count, exportOutput)
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}
