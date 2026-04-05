package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/user/wallpaper-cli/internal/config"
	"github.com/user/wallpaper-cli/internal/data"
	"github.com/user/wallpaper-cli/internal/dedup"
	"github.com/user/wallpaper-cli/internal/download"
	"github.com/user/wallpaper-cli/internal/output"
	"github.com/user/wallpaper-cli/internal/sources"
	"github.com/user/wallpaper-cli/internal/utils"
	"github.com/user/wallpaper-cli/internal/validate"

	// Register source adapters
	_ "github.com/user/wallpaper-cli/internal/sources/danbooru"
	_ "github.com/user/wallpaper-cli/internal/sources/konachan"
	_ "github.com/user/wallpaper-cli/internal/sources/reddit"
	_ "github.com/user/wallpaper-cli/internal/sources/wallhaven"
	_ "github.com/user/wallpaper-cli/internal/sources/zerochan"
)

var (
	// Basic flags
	source      string
	resolution  string
	aspectRatio string
	tags        string
	limit       int
	outputDir   string
	organizeBy  string
	format      string
	dedupFlag   bool
	concurrent  int
	dryRun      bool
	animeOnly   bool
	jsonOutput  bool

	// Sorting flags (v1.1)
	sortBy     string
	timeRange  string
	latest     bool
	popular    bool
	favorites  bool
	mostViewed bool
)

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
  wallpaper-cli fetch --favorites --all-time --limit 5

  # Fetch from all sources with JSON output
  wallpaper-cli fetch --source all --json`,
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
	fetchCmd.Flags().StringVar(&outputDir, "output", "", "Output directory (default: ~/Pictures/wallpapers/)")
	fetchCmd.Flags().StringVar(&organizeBy, "organize-by", "source", "Organization method (source, tags, date)")
	fetchCmd.Flags().StringVar(&format, "format", "original", "Preferred format (webp, jpg, png, original)")
	fetchCmd.Flags().BoolVar(&dedupFlag, "dedup", true, "Enable deduplication")
	fetchCmd.Flags().IntVar(&concurrent, "concurrent", 5, "Number of concurrent downloads")
	fetchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be downloaded without downloading")
	fetchCmd.Flags().BoolVar(&animeOnly, "anime", false, "Search for anime wallpapers only")
	fetchCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output structured JSON events to stdout")

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
	if outputDir == "" {
		outputDir = cfg.OutputDirectory
	}

	// Expand output path
	outputDir, err = utils.ExpandPath(outputDir)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Normalize resolution
	normalizedRes := validate.NormalizeResolution(resolution)
	normalizedRatio := validate.NormalizeAspectRatio(aspectRatio)

	// Determine sorting
	sorting := getSorting(cmd)
	timePeriod := getTimePeriod(cmd)

	// Select emitter
	var emitter output.Emitter
	if jsonOutput {
		emitter = output.NewJSONEmitter()
	} else {
		emitter = output.NewTextEmitter()
	}
	defer emitter.Close()

	if dryRun {
		fmt.Println("DRY RUN: Would fetch wallpapers with the following settings:")
		fmt.Printf("  Source: %s\n", source)
		fmt.Printf("  Resolution: %s (normalized: %s)\n", resolution, normalizedRes)
		fmt.Printf("  Aspect Ratio: %s (normalized: %s)\n", aspectRatio, normalizedRatio)
		fmt.Printf("  Tags: %s\n", tags)
		fmt.Printf("  Sort: %s\n", sorting)
		fmt.Printf("  Time Period: %s\n", timePeriod)
		fmt.Printf("  Limit: %d\n", limit)
		fmt.Printf("  Output: %s\n", outputDir)
		fmt.Printf("  Organize by: %s\n", organizeBy)
		fmt.Printf("  Format: %s\n", format)
		fmt.Printf("  Deduplication: %v\n", dedupFlag)
		fmt.Printf("  Concurrent: %d\n", concurrent)
		return nil
	}

	// Open database
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

	// Build search params
	params := &sources.SearchParams{
		Tags:        tags,
		Resolution:  normalizedRes,
		AspectRatio: normalizedRatio,
		Sorting:     sorting,
		TimePeriod:  timePeriod,
		Limit:       limit,
		AnimeOnly:   animeOnly,
	}

	// Resolve which sources to query
	sourceNames := resolveSourceNames(source)

	// Emit capabilities at stream start (JSON mode only)
	if jsonOutput {
		var caps []interface{}
		for _, name := range sourceNames {
			srcCfg := buildSourceConfig(cfg, name)
			if src, err := sources.Get(name, srcCfg); err == nil {
				caps = append(caps, src.Capabilities())
			}
		}
		emitter.Emit(output.NewEvent("capabilities", "", output.CapabilitiesData{Sources: caps}))
	}

	// Parallel fetch when multiple sources
	if len(sourceNames) > 1 {
		return fetchParallel(sourceNames, cfg, params, outputDir, db, checker, emitter)
	}

	// Single source — sequential
	name := sourceNames[0]
	return fetchSingle(name, cfg, params, outputDir, db, checker, emitter)
}

func fetchSingle(name string, cfg *config.Config, params *sources.SearchParams, outputPath string, db *data.DB, checker *dedup.Checker, emitter output.Emitter) error {
	srcCfg := buildSourceConfig(cfg, name)
	src, err := sources.Get(name, srcCfg)
	if err != nil {
		emitter.Emit(output.NewErrorEvent(name, err))
		return nil
	}

	emitter.Emit(output.NewEvent("search_started", name, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	result, err := src.Search(ctx, params)
	cancel()
	if err != nil {
		emitter.Emit(output.NewErrorEvent(name, err))
		return nil
	}

	emitter.Emit(output.NewEvent("search_complete", name, output.SearchCompleteData{Count: len(result.Wallpapers)}))

	if len(result.Wallpapers) == 0 {
		return nil
	}

	if db != nil && len(result.Tags) > 0 {
		_ = db.SaveTags(result.Tags)
	}

	if err := downloadResults(context.Background(), result, name, outputPath, db, checker, emitter); err != nil {
		emitter.Emit(output.NewErrorEvent(name, err))
	}
	return nil
}

func fetchParallel(sourceNames []string, cfg *config.Config, params *sources.SearchParams, outputPath string, db *data.DB, checker *dedup.Checker, emitter output.Emitter) error {
	type searchResult struct {
		name   string
		result *sources.SearchResult
		err    error
	}

	// Phase 1: Search all sources in parallel
	results := make(chan searchResult, len(sourceNames))
	var wg sync.WaitGroup

	for _, name := range sourceNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			srcCfg := buildSourceConfig(cfg, name)
			src, err := sources.Get(name, srcCfg)
			if err != nil {
				results <- searchResult{name: name, err: err}
				return
			}

			emitter.Emit(output.NewEvent("search_started", name, nil))

			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			result, err := src.Search(ctx, params)
			cancel()

			results <- searchResult{name: name, result: result, err: err}
		}(name)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Phase 2: Collect results, emit partial results on error
	var allResults []searchResult
	for r := range results {
		if r.err != nil {
			emitter.Emit(output.NewErrorEvent(r.name, r.err))
			continue
		}
		emitter.Emit(output.NewEvent("search_complete", r.name, output.SearchCompleteData{Count: len(r.result.Wallpapers)}))
		allResults = append(allResults, r)
	}

	// Phase 3: Download from each successful source
	for _, r := range allResults {
		if len(r.result.Wallpapers) == 0 {
			continue
		}

		if db != nil && len(r.result.Tags) > 0 {
			_ = db.SaveTags(r.result.Tags)
		}

		if err := downloadResults(context.Background(), r.result, r.name, outputPath, db, checker, emitter); err != nil {
			emitter.Emit(output.NewErrorEvent(r.name, err))
		}
	}

	return nil
}

func resolveSourceNames(source string) []string {
	if source == "all" {
		return sources.List()
	}
	return []string{source}
}

func buildSourceConfig(cfg *config.Config, name string) map[string]string {
	m := make(map[string]string)
	if sc, ok := cfg.Sources[name]; ok {
		if sc.APIKey != "" {
			m["api_key"] = sc.APIKey
		}
		if sc.Login != "" {
			m["login"] = sc.Login
		}
		if sc.Username != "" {
			m["username"] = sc.Username
		}
		if sc.Cookies != "" {
			m["cookies"] = sc.Cookies
		}
		if sc.CookiesFile != "" {
			m["cookies_file"] = sc.CookiesFile
		}
		if len(sc.Subreddits) > 0 {
			m["subreddits"] = strings.Join(sc.Subreddits, ",")
		}
	}
	return m
}

func downloadResults(ctx context.Context, result *sources.SearchResult, sourceName, outputPath string, db *data.DB, checker *dedup.Checker, emitter output.Emitter) error {
	jobs := make([]download.DownloadJob, 0, len(result.Wallpapers))
	for i, w := range result.Wallpapers {
		filename := generateFilename(w, i+1)
		subdir := getSubdir(sourceName, w, organizeBy)
		fullPath := filepath.Join(outputPath, subdir, filename)

		jobs = append(jobs, download.DownloadJob{
			URL:      w.URL,
			Filename: fullPath,
			Index:    i,
			Total:    len(result.Wallpapers),
		})
	}

	if len(jobs) == 0 {
		return nil
	}

	emitter.Emit(output.NewEvent("download_started", sourceName, output.DownloadStartedData{Total: len(jobs)}))

	opts := &download.Options{
		Concurrency: concurrent,
		OutputDir:   outputPath,
		Timeout:     60 * time.Second,
		EnableDedup: dedupFlag,
	}

	dlCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var results []download.DownloadResult
	if jsonOutput {
		// In JSON mode, use silent progress (no progress bar on stderr)
		manager := download.NewManager(opts, nil, checker)
		results = manager.DownloadBatch(dlCtx, jobs)

		completed, skipped, errors := manager.Stats()
		emitter.Emit(output.NewEvent("download_complete", sourceName, output.DownloadCompleteData{
			Total:     len(results),
			Completed: completed,
			Skipped:   skipped,
			Errors:    errors,
		}))
	} else {
		// In text mode, use progress bar
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
		manager := download.NewManagerWithBar(opts, bar, checker)
		results = manager.DownloadBatch(dlCtx, jobs)
		bar.Finish()

		completed, skipped, errors := manager.Stats()
		emitter.Emit(output.NewEvent("download_complete", sourceName, output.DownloadCompleteData{
			Total:     len(results),
			Completed: completed,
			Skipped:   skipped,
			Errors:    errors,
		}))
	}

	// Save to database
	if db != nil {
		for i, res := range results {
			if res.Error != nil || res.Skipped {
				continue
			}

			hash, err := dedup.ComputeHash(res.Path)
			if err != nil {
				continue
			}
			info, err := os.Stat(res.Path)
			if err != nil {
				continue
			}

			w := result.Wallpapers[i]
			record := &data.ImageRecord{
				Hash:         hash,
				Source:       sourceName,
				SourceID:     w.SourceID,
				URL:          res.URL,
				LocalPath:    res.Path,
				Resolution:   w.Resolution,
				AspectRatio:  w.AspectRatio,
				DownloadedAt: time.Now(),
				FileSize:     info.Size(),
			}
			_ = db.SaveImage(record)

			// Save source URL as extended attribute
			switch sourceName {
			case "wallhaven":
				_ = utils.SetWallhavenURL(res.Path, fmt.Sprintf("https://wallhaven.cc/w/%s", w.SourceID))
			case "reddit":
				// Reddit permalink would need to be in Extra data; skip for now
			}
		}
	}

	return nil
}

func generateFilename(w sources.ResultWallpaper, rank int) string {
	ext := "." + w.Format
	if ext == "." {
		ext = ".jpg"
	}
	if format != "" && format != "original" {
		ext = "." + format
	}

	title := sanitizeFilename(w.Title)
	if len(title) > 30 {
		title = title[:30]
	}

	if title != "" && w.Resolution != "" {
		return fmt.Sprintf("%02d_%s_%s_%s%s", rank, w.SourceID, w.Resolution, title, ext)
	}
	if w.Resolution != "" && w.Resolution != "unknown" {
		return fmt.Sprintf("%02d_%s_%s%s", rank, w.SourceID, w.Resolution, ext)
	}
	return fmt.Sprintf("%02d_%s%s", rank, w.SourceID, ext)
}

func getSubdir(sourceName string, w sources.ResultWallpaper, organize string) string {
	switch organize {
	case "source":
		return sourceName
	case "date":
		return time.Now().Format("2006/01")
	case "tags":
		if len(w.Tags) > 0 {
			return sanitizeFilename(w.Tags[0])
		}
		return "unsorted"
	default:
		return sourceName
	}
}

func getSorting(cmd *cobra.Command) string {
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
	// Validate source - check registry or "all"
	if source != "all" && !sources.IsRegistered(source) {
		return fmt.Errorf("invalid source: %s (registered: %s)", source, strings.Join(sources.List(), ", "))
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
	valid := false
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
