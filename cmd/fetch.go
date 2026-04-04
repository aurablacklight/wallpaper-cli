package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/user/wallpaper-cli/internal/config"
	"github.com/user/wallpaper-cli/internal/data"
	"github.com/user/wallpaper-cli/internal/dedup"
	"github.com/user/wallpaper-cli/internal/download"
	"github.com/user/wallpaper-cli/internal/sources/reddit"
	"github.com/user/wallpaper-cli/internal/sources/wallhaven"
	"github.com/user/wallpaper-cli/internal/utils"
	"github.com/user/wallpaper-cli/internal/validate"
)

var (
	// Basic flags
	source       string
	resolution   string
	aspectRatio  string
	tags         string
	limit        int
	output       string
	organizeBy   string
	format       string
	dedupFlag    bool
	concurrent   int
	dryRun       bool
	animeOnly    bool

	// Sorting flags (v1.1)
	sortBy    string
	timeRange string
	latest    bool
	popular   bool
	favorites bool
	mostViewed bool
)

var validSources = []string{"wallhaven", "reddit", "all"}
var validOrganizeBy = []string{"source", "tags", "date"}
var validFormats = []string{"webp", "jpg", "png", "original"}
var validSortBy = []string{"random", "top", "hot", "new", "favorites", "views"}
var validTimeRanges = []string{"day", "week", "month", "year", "all"}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download wallpapers from sources",
	Long: `Fetch wallpapers from configured sources with filtering and organization.

Examples:
  # Fetch anime wallpapers in 4K
  wallpaper-cli fetch --resolution 4k --limit 10

  # Fetch top wallpapers this week
  wallpaper-cli fetch --top --week --limit 10

  # Fetch from Reddit (hot posts)
  wallpaper-cli fetch --source reddit --hot --limit 10

  # Fetch most favorited of all time
  wallpaper-cli fetch --favorites --all-time --limit 5`,
	RunE: runFetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Basic flags
	fetchCmd.Flags().StringVar(&source, "source", "wallhaven", "Image source (wallhaven, reddit, all)")
	fetchCmd.Flags().StringVar(&resolution, "resolution", "", "Target resolution (1080p, 1440p, 4k, 8k, or WxH)")
	fetchCmd.Flags().StringVar(&aspectRatio, "aspect-ratio", "", "Filter by aspect ratio (16:9, 21:9, 32:9)")
	fetchCmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags to search")
	fetchCmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of images to download")
	fetchCmd.Flags().StringVar(&output, "output", "", "Output directory (default: ~/Pictures/wallpapers/)")
	fetchCmd.Flags().StringVar(&organizeBy, "organize-by", "source", "Organization method (source, tags, date)")
	fetchCmd.Flags().StringVar(&format, "format", "original", "Preferred format (webp, jpg, png, original)")
	fetchCmd.Flags().BoolVar(&dedupFlag, "dedup", true, "Enable deduplication")
	fetchCmd.Flags().IntVar(&concurrent, "concurrent", 5, "Number of concurrent downloads")
	fetchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be downloaded without downloading")
	fetchCmd.Flags().BoolVar(&animeOnly, "anime", false, "Search for anime wallpapers only")

	// Sorting flags (v1.1)
	fetchCmd.Flags().StringVar(&sortBy, "sort", "random", "Sort by: random, top, hot, new, favorites, views")
	fetchCmd.Flags().StringVar(&timeRange, "time", "", "Time range: day, week, month, year, all")
	fetchCmd.Flags().BoolVar(&latest, "latest", false, "Get latest/newest uploads")
	fetchCmd.Flags().BoolVar(&popular, "popular", false, "Get most popular (alias for --top)")
	fetchCmd.Flags().BoolVar(&favorites, "favorites", false, "Get most favorited")
	fetchCmd.Flags().BoolVar(&mostViewed, "views", false, "Get most viewed")

	// Shorthand flags for time ranges
	fetchCmd.Flags().Bool("day", false, "Last 24 hours (shorthand for --time day)")
	fetchCmd.Flags().Bool("week", false, "Last 7 days (shorthand for --time week)")
	fetchCmd.Flags().Bool("month", false, "Last 30 days (shorthand for --time month)")
	fetchCmd.Flags().Bool("year", false, "Last year (shorthand for --time year)")
	fetchCmd.Flags().Bool("all-time", false, "All time (shorthand for --time all)")

	viper.BindPFlag("source", fetchCmd.Flags().Lookup("source"))
	viper.BindPFlag("resolution", fetchCmd.Flags().Lookup("resolution"))
	viper.BindPFlag("output", fetchCmd.Flags().Lookup("output"))
}

func runFetch(cmd *cobra.Command, args []string) error {
	// Load config
	cfgPath := config.GetConfigPath()
	if cfgFile != "" {
		cfgPath = cfgFile
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate inputs
	if err := validateInputs(); err != nil {
		return err
	}

	// Apply config defaults if flags not set
	if resolution == "" {
		resolution = cfg.DefaultResolution
	}
	if output == "" {
		output = cfg.OutputDirectory
	}

	// Expand output path
	output, err = utils.ExpandPath(output)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Normalize resolution
	normalizedRes := validate.NormalizeResolution(resolution)
	normalizedRatio := validate.NormalizeAspectRatio(aspectRatio)

	// Determine sorting
	sorting := getSorting(cmd)
	timePeriod := getTimePeriod(cmd)

	if dryRun {
		fmt.Println("DRY RUN: Would fetch wallpapers with the following settings:")
		fmt.Printf("  Source: %s\n", source)
		fmt.Printf("  Resolution: %s (normalized: %s)\n", resolution, normalizedRes)
		fmt.Printf("  Aspect Ratio: %s (normalized: %s)\n", aspectRatio, normalizedRatio)
		fmt.Printf("  Tags: %s\n", tags)
		fmt.Printf("  Sort: %s\n", sorting)
		fmt.Printf("  Time Period: %s\n", timePeriod)
		fmt.Printf("  Limit: %d\n", limit)
		fmt.Printf("  Output: %s\n", output)
		fmt.Printf("  Organize by: %s\n", organizeBy)
		fmt.Printf("  Format: %s\n", format)
		fmt.Printf("  Deduplication: %v\n", dedupFlag)
		fmt.Printf("  Concurrent: %d\n", concurrent)
		return nil
	}

	// Open database for deduplication
	var db *data.DB
	var checker *dedup.Checker
	if dedupFlag {
		dbPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "wallpaper-cli", "wallpapers.db")
		db, err = data.NewDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()
		checker = dedup.NewChecker(db, cfg.DedupThreshold)
	}

	// Execute fetch based on source
	switch source {
	case "wallhaven":
		return fetchFromWallhaven(cfg, normalizedRes, normalizedRatio, output, db, checker, sorting, timePeriod)
	case "reddit":
		return fetchFromReddit(cfg, output, db, checker, sorting, timePeriod)
	case "all":
		return fetchFromAll(cfg, normalizedRes, normalizedRatio, output, db, checker, sorting, timePeriod)
	default:
		return fmt.Errorf("unknown source: %s", source)
	}
}

func getSorting(cmd *cobra.Command) string {
	// Check shorthand flags first
	if latest {
		return "latest"
	}
	if popular || sortBy == "top" {
		return "top"
	}
	if favorites {
		return "favorites"
	}
	if mostViewed {
		return "views"
	}
	if cmd.Flags().Changed("sort") {
		return sortBy
	}
	return "random"
}

func getTimePeriod(cmd *cobra.Command) string {
	// Check shorthand time flags
	if cmd.Flags().Lookup("day").Value.String() == "true" {
		return "1d"
	}
	if cmd.Flags().Lookup("week").Value.String() == "true" {
		return "7d"
	}
	if cmd.Flags().Lookup("month").Value.String() == "true" {
		return "30d"
	}
	if cmd.Flags().Lookup("year").Value.String() == "true" {
		return "1y"
	}
	if cmd.Flags().Lookup("all-time").Value.String() == "true" {
		return "all"
	}
	if timeRange != "" {
		return timeRange
	}
	return ""
}

func fetchFromWallhaven(cfg *config.Config, resolution, ratio, output string, db *data.DB, checker *dedup.Checker, sorting, timePeriod string) error {
	fmt.Printf("Fetching from Wallhaven (sort: %s)...\n", sorting)

	client := wallhaven.NewClient()

	// Build search query
	builder := wallhaven.NewSearchBuilder()
	if animeOnly {
		builder.WithAnimeOnly()
	}
	if tags != "" {
		builder.WithQuery(wallhaven.ParseTags(tags))
	}
	if resolution != "" {
		builder.WithResolution(resolution)
	}
	if ratio != "" {
		builder.WithAspectRatio(ratio)
	}

	// Apply sorting
	switch sorting {
	case "top", "popular":
		builder.WithSorting("toplist")
		if timePeriod != "" {
			builder.WithTopRange(mapTimeToWallhavenRange(timePeriod))
		}
	case "favorites":
		builder.WithSorting("favorites")
	case "views":
		builder.WithSorting("views")
	case "latest", "new":
		builder.WithSorting("date_added")
		builder.WithOrder("desc")
	case "random":
		builder.WithRandom()
	default:
		builder.WithRandom()
	}

	opts := builder.Build()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Fetch wallpapers with pagination
	wallpapers, err := client.PaginatedSearch(ctx, opts, limit)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	fmt.Printf("Found %d wallpapers\n", len(wallpapers))

	if len(wallpapers) == 0 {
		fmt.Println("No wallpapers found matching criteria.")
		return nil
	}

	return downloadWallpapers(wallpapers, output, db, checker, "wallhaven")
}

func fetchFromReddit(cfg *config.Config, output string, db *data.DB, checker *dedup.Checker, sorting, timePeriod string) error {
	fmt.Printf("Fetching from Reddit (sort: %s)...\n", sorting)

	client := reddit.NewClient()

	// Determine sort option
	sortOpt := reddit.SortHot
	switch sorting {
	case "top", "popular":
		sortOpt = reddit.SortTop
	case "new", "latest":
		sortOpt = reddit.SortNew
	case "hot":
		sortOpt = reddit.SortHot
	}

	// Get subreddits from config
	subreddits := cfg.Sources["reddit"].Subreddits
	if len(subreddits) == 0 {
		subreddits = []string{"Animewallpaper"}
	}

	var allPosts []reddit.Wallpaper

	for _, sub := range subreddits {
		opts := &reddit.SearchOptions{
			Subreddit: sub,
			Sort:      sortOpt,
			Time:      reddit.TimePeriod(mapTimeToRedditRange(timePeriod)),
			Limit:     limit,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		posts, err := client.Search(ctx, opts)
		cancel()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Reddit search failed for r/%s: %v\n", sub, err)
			continue
		}

		wallpapers := reddit.ToPosts(posts)
		allPosts = append(allPosts, wallpapers...)
	}

	// Limit results
	if len(allPosts) > limit {
		allPosts = allPosts[:limit]
	}

	fmt.Printf("Found %d wallpapers from Reddit\n", len(allPosts))

	if len(allPosts) == 0 {
		fmt.Println("No wallpapers found from Reddit.")
		return nil
	}

	// Convert to common format for download
	return downloadRedditPosts(allPosts, output, db, checker)
}

func fetchFromAll(cfg *config.Config, resolution, ratio, output string, db *data.DB, checker *dedup.Checker, sorting, timePeriod string) error {
	// Fetch from both sources
	fmt.Println("Fetching from all sources...")

	// Wallhaven first
	fmt.Println("\n[Wallhaven]")
	if err := fetchFromWallhaven(cfg, resolution, ratio, output, db, checker, sorting, timePeriod); err != nil {
		fmt.Fprintf(os.Stderr, "Wallhaven error: %v\n", err)
	}

	// Then Reddit
	fmt.Println("\n[Reddit]")
	if err := fetchFromReddit(cfg, output, db, checker, sorting, timePeriod); err != nil {
		fmt.Fprintf(os.Stderr, "Reddit error: %v\n", err)
	}

	return nil
}

func mapTimeToWallhavenRange(timeRange string) string {
	switch timeRange {
	case "day", "1d":
		return "1d"
	case "week", "7d":
		return "1w"
	case "month", "30d":
		return "1M"
	case "year", "1y":
		return "1y"
	case "all":
		return "1y"
	default:
		return "1M"
	}
}

func mapTimeToRedditRange(timeRange string) string {
	switch timeRange {
	case "day", "1d":
		return "day"
	case "week", "7d":
		return "week"
	case "month", "30d":
		return "month"
	case "year", "1y":
		return "year"
	case "all":
		return "all"
	default:
		return "week"
	}
}

func downloadWallpapers(wallpapers []wallhaven.Wallpaper, output string, db *data.DB, checker *dedup.Checker, sourceName string) error {
	// Create download jobs
	jobs := make([]download.DownloadJob, 0, len(wallpapers))
	for i, w := range wallpapers {
		filename := generateWallhavenFilename(w, i+1)  // +1 for 1-based ranking

		subdir := ""
		switch organizeBy {
		case "source":
			subdir = sourceName
		case "date":
			subdir = time.Now().Format("2006/01")
		case "tags":
			if len(w.Tags) > 0 {
				subdir = sanitizeFilename(w.Tags[0].Name)
			} else {
				subdir = "unsorted"
			}
		}

		fullPath := filepath.Join(output, subdir, filename)

		jobs = append(jobs, download.DownloadJob{
			URL:      w.Path,
			Filename: fullPath,
			Index:    i,
			Total:    len(wallpapers),
		})
	}

	// Download with progress bar
	fmt.Printf("\nDownloading %d wallpapers...\n", len(jobs))

	opts := &download.Options{
		Concurrency: concurrent,
		OutputDir:   output,
		Timeout:     60 * time.Second,
		EnableDedup: dedupFlag,
	}

	// Create progress bar
	bar := progressbar.NewOptions(len(jobs),
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	manager := download.NewManagerWithBar(opts, bar, checker)
	results := manager.DownloadBatch(ctx, jobs)

	// Force finish progress bar
	bar.Finish()

	// Save to database and metadata
	for i, result := range results {
		if result.Error != nil || result.Skipped {
			continue
		}

		// Save to database (if enabled)
		if db != nil {
			hash, err := dedup.ComputeHash(result.Path)
			if err != nil {
				continue
			}

			info, err := os.Stat(result.Path)
			if err != nil {
				continue
			}

			record := &data.ImageRecord{
				Hash:         hash,
				Source:       sourceName,
				SourceID:     wallpapers[i].ID,
				URL:          result.URL,
				LocalPath:    result.Path,
				Resolution:   wallpapers[i].Resolution,
				AspectRatio:  wallpapers[i].Ratio,
				DownloadedAt: time.Now(),
				FileSize:     info.Size(),
			}

			db.SaveImage(record)
		}

		// Save Wallhaven source URL as file metadata (always)
		wallhavenURL := fmt.Sprintf("https://wallhaven.cc/w/%s", wallpapers[i].ID)
		_ = utils.SetWallhavenURL(result.Path, wallhavenURL)
	}

	// Print summary
	completed, skipped, errors := manager.Stats()
	fmt.Fprintf(os.Stderr, "\n✓ Downloaded: %d/%d", completed, len(results))
	if skipped > 0 {
		fmt.Fprintf(os.Stderr, " | Skipped: %d", skipped)
	}
	if errors > 0 {
		fmt.Fprintf(os.Stderr, " | Failed: %d", errors)
	}
	fmt.Fprintln(os.Stderr)

	return nil
}

func downloadRedditPosts(posts []reddit.Wallpaper, output string, db *data.DB, checker *dedup.Checker) error {
	// Create download jobs
	jobs := make([]download.DownloadJob, 0, len(posts))
	for i, p := range posts {
		filename := generateRedditFilename(p, i+1)  // +1 for 1-based ranking

		subdir := ""
		switch organizeBy {
		case "source":
			subdir = "reddit"
		case "date":
			subdir = time.Now().Format("2006/01")
		case "tags":
			subdir = "unsorted"
		}

		fullPath := filepath.Join(output, subdir, filename)

		jobs = append(jobs, download.DownloadJob{
			URL:      p.URL,
			Filename: fullPath,
			Index:    i,
			Total:    len(posts),
		})
	}

	if len(jobs) == 0 {
		return nil
	}

	// Download with progress bar (Reddit)
	fmt.Fprintf(os.Stderr, "\nDownloading %d wallpapers from Reddit...\n\n", len(jobs))

	opts := &download.Options{
		Concurrency: concurrent,
		OutputDir:   output,
		Timeout:     60 * time.Second,
		EnableDedup: dedupFlag,
	}

	// Create progress bar
	bar := progressbar.NewOptions(len(jobs),
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	manager := download.NewManagerWithBar(opts, bar, checker)
	results := manager.DownloadBatch(ctx, jobs)

	// Force finish progress bar
	bar.Finish()

	// Save to database and metadata
	for i, result := range results {
		if result.Error != nil || result.Skipped {
			continue
		}

		// Save to database (if enabled)
		if db != nil {
			hash, err := dedup.ComputeHash(result.Path)
			if err != nil {
				continue
			}

			info, err := os.Stat(result.Path)
			if err != nil {
				continue
			}

			record := &data.ImageRecord{
				Hash:         hash,
				Source:       "reddit",
				SourceID:     posts[i].ID,
				URL:          result.URL,
				LocalPath:    result.Path,
				Resolution:   posts[i].Resolution,
				AspectRatio:  "",
				DownloadedAt: time.Now(),
				FileSize:     info.Size(),
			}

			db.SaveImage(record)
		}

		// Save Reddit permalink as file metadata (always, if available)
		if posts[i].Permalink != "" {
			err := utils.SetRedditURL(result.Path, posts[i].Permalink)
			if err != nil {
				// Silently ignore errors - metadata is optional
				_ = err
			}
		}
	}

	// Print summary
	completed, skipped, errors := manager.Stats()
	fmt.Fprintf(os.Stderr, "\n✓ Downloaded: %d/%d", completed, len(results))
	if skipped > 0 {
		fmt.Fprintf(os.Stderr, " | Skipped: %d", skipped)
	}
	if errors > 0 {
		fmt.Fprintf(os.Stderr, " | Failed: %d", errors)
	}
	fmt.Fprintln(os.Stderr)

	return nil
}

func generateWallhavenFilename(w wallhaven.Wallpaper, rank int) string {
	ext := ".jpg"
	if strings.Contains(w.Path, ".png") {
		ext = ".png"
	} else if strings.Contains(w.Path, ".webp") {
		ext = ".webp"
	}

	if format != "" && format != "original" {
		ext = "." + format
	}

	return fmt.Sprintf("%02d_%s_%s%s", rank, w.ID, w.Resolution, ext)
}

func generateRedditFilename(p reddit.Wallpaper, rank int) string {
	ext := ".jpg"
	if strings.Contains(p.URL, ".png") {
		ext = ".png"
	} else if strings.Contains(p.URL, ".webp") {
		ext = ".webp"
	}

	if format != "" && format != "original" {
		ext = "." + format
	}

	// Use title or ID in filename
	title := sanitizeFilename(p.Title)
	if len(title) > 30 {
		title = title[:30]
	}

	// Include resolution if available
	if p.Resolution != "" && p.Resolution != "unknown" {
		return fmt.Sprintf("%02d_%s_%s_%s%s", rank, p.ID, p.Resolution, title, ext)
	}

	return fmt.Sprintf("%02d_%s_%s%s", rank, p.ID, title, ext)
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "-",
	)
	return replacer.Replace(name)
}

func validateInputs() error {
	// Validate source
	valid := false
	for _, s := range validSources {
		if source == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid source: %s (expected: wallhaven, reddit, all)", source)
	}

	// Validate resolution
	if err := validate.ValidateResolution(resolution); err != nil {
		return err
	}

	// Validate aspect ratio
	if err := validate.ValidateAspectRatio(aspectRatio); err != nil {
		return err
	}

	// Validate organize-by
	valid = false
	for _, o := range validOrganizeBy {
		if organizeBy == o {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid organize-by: %s (expected: source, tags, date)", organizeBy)
	}

	// Validate format
	valid = false
	for _, f := range validFormats {
		if format == f {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid format: %s (expected: webp, jpg, png, original)", format)
	}

	// Validate limit
	if limit < 1 || limit > 1000 {
		return fmt.Errorf("invalid limit: %d (expected: 1-1000)", limit)
	}

	// Validate concurrent
	if concurrent < 1 || concurrent > 50 {
		return fmt.Errorf("invalid concurrent: %d (expected: 1-50)", concurrent)
	}

	// Validate sort-by
	if sortBy != "" {
		valid = false
		for _, s := range validSortBy {
			if sortBy == s {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid sort: %s (expected: %s)", sortBy, strings.Join(validSortBy, ", "))
		}
	}

	// Validate time range
	if timeRange != "" {
		valid = false
		for _, t := range validTimeRanges {
			if timeRange == t {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid time range: %s (expected: %s)", timeRange, strings.Join(validTimeRanges, ", "))
		}
	}

	return nil
}
