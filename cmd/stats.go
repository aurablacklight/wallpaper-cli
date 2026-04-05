package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/wallpaper-cli/internal/utils"
)

var statsJSON bool

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show wallpaper collection statistics",
	Long: `Display statistics about your wallpaper collection including:
- Total count
- Breakdown by source
- Recent download activity
- Storage usage
- Resolution distribution`,
	RunE: runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().BoolVar(&statsJSON, "json", false, "Output as JSON")
}

type StatsInfo struct {
	TotalCount   int
	BySource     map[string]int
	TotalSize    int64
	ByResolution map[string]int
	RecentCount  int // Last 7 days
	OldestFile   time.Time
	NewestFile   time.Time
}

func runStats(cmd *cobra.Command, args []string) error {
	// Always use filesystem scan for full stats (size, resolution)
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	outputDir := filepath.Join(home, "Pictures", "wallpapers")
	return showFilesystemStats(outputDir)
}

func showFilesystemStats(basePath string) error {
	stats := StatsInfo{
		BySource:     make(map[string]int),
		ByResolution: make(map[string]int),
	}

	sources := []string{"wallhaven", "reddit"}

	for _, source := range sources {
		srcPath := filepath.Join(basePath, source)
		entries, err := os.ReadDir(srcPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			stats.TotalCount++
			stats.BySource[source]++
			stats.TotalSize += info.Size()

			// Check if recent (last 7 days)
			if time.Since(info.ModTime()) < 7*24*time.Hour {
				stats.RecentCount++
			}

			// Track newest/oldest
			if stats.NewestFile.IsZero() || info.ModTime().After(stats.NewestFile) {
				stats.NewestFile = info.ModTime()
			}
			if stats.OldestFile.IsZero() || info.ModTime().Before(stats.OldestFile) {
				stats.OldestFile = info.ModTime()
			}

			// Extract resolution from filename
			_, resolution := utils.ParseWallpaperFilename(entry.Name())
			if resolution != "" {
				stats.ByResolution[resolution]++
			}
		}
	}

	if statsJSON {
		return printStatsJSON(stats)
	}
	printStats(stats)
	return nil
}

func printStats(stats StatsInfo) {
	fmt.Println("📊 Wallpaper Collection Statistics")
	fmt.Println("====================================")
	fmt.Printf("\n🖼️  Total Wallpapers: %d\n", stats.TotalCount)

	if len(stats.BySource) > 0 {
		fmt.Println("\n📁 By Source:")
		sources := make([]string, 0, len(stats.BySource))
		for source := range stats.BySource {
			sources = append(sources, source)
		}
		sort.Strings(sources)

		for _, source := range sources {
			count := stats.BySource[source]
			fmt.Printf("   %-12s %d\n", source+":", count)
		}
	}

	if stats.TotalSize > 0 {
		fmt.Printf("\n💾 Storage Used: %s\n", formatBytes(stats.TotalSize))
	}

	if stats.RecentCount > 0 {
		fmt.Printf("\n🆕 Downloads (7d): %d\n", stats.RecentCount)
	}

	if !stats.NewestFile.IsZero() {
		fmt.Printf("\n🕐 Latest Download: %s\n", stats.NewestFile.Format("2006-01-02"))
	}

	if len(stats.ByResolution) > 0 {
		fmt.Println("\n📐 By Resolution:")
		resolutions := make([]string, 0, len(stats.ByResolution))
		for res := range stats.ByResolution {
			resolutions = append(resolutions, res)
		}
		sort.Strings(resolutions)

		for _, res := range resolutions {
			count := stats.ByResolution[res]
			fmt.Printf("   %-15s %d\n", res+":", count)
		}
	}
}

func printStatsJSON(stats StatsInfo) error {
	type JSONStats struct {
		TotalCount   int            `json:"total_count"`
		TotalSize    int64          `json:"total_size_bytes"`
		BySource     map[string]int `json:"by_source"`
		ByResolution map[string]int `json:"by_resolution"`
		RecentCount  int            `json:"recent_count_7d"`
		OldestFile   string         `json:"oldest_file,omitempty"`
		NewestFile   string         `json:"newest_file,omitempty"`
	}

	out := JSONStats{
		TotalCount:   stats.TotalCount,
		TotalSize:    stats.TotalSize,
		BySource:     stats.BySource,
		ByResolution: stats.ByResolution,
		RecentCount:  stats.RecentCount,
	}
	if !stats.OldestFile.IsZero() {
		out.OldestFile = stats.OldestFile.UTC().Format(time.RFC3339)
	}
	if !stats.NewestFile.IsZero() {
		out.NewestFile = stats.NewestFile.UTC().Format(time.RFC3339)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
