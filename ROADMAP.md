# Wallpaper CLI Tool - Product Roadmap

**Current Status:** v1.0 Complete ✅  
**Last Updated:** 2026-04-04

---

## ✅ v1.0 - Core Foundation (COMPLETE)

**M001: Wallpaper CLI Tool**

### Delivered:
- ✅ Go project scaffold with Cobra CLI
- ✅ Config system (JSON persistence)
- ✅ Wallhaven API adapter with pagination
- ✅ Concurrent download manager (5 parallel)
- ✅ Perceptual hash deduplication (pHash + SQLite)
- ✅ Organization modes (source/date/tags)
- ✅ Cross-platform builds (macOS/Linux/Windows)
- ✅ Rate limiting (polite API usage)
- ✅ Input validation with helpful errors

### Metrics:
- Binary size: 11MB (target: <20MB) ✅
- Concurrent downloads: 5 parallel ✅
- Memory: <10MB at idle ✅
- Cross-platform: 5 targets ✅

---

## 🚧 v1.1 - Better UX + More Sources + Sorting (NEXT)

**M002: Enhanced Experience**

### Planned:
1. **Better Progress Bar**
   - Real-time progress bars (percentage)
   - Download speed indicator (MB/s)
   - ETA estimation
   - Visual queue status
   - Library: `schollz/progressbar` or `vbauerster/mpb`

2. **Reddit Source Adapter**
   - r/Animewallpaper integration
   - Reddit JSON API (no OAuth)
   - Rate limiting (60 req/min)
   - Cross-source deduplication
   - Subreddit configuration in config

3. **Popularity & Time Filtering**
   - `--top` / `--most-popular` - Top rated wallpapers
   - `--most-liked` / `--favorites` - Most favorited
   - `--most-viewed` - Most viewed
   - `--latest` / `--recent` - Newest first
   - `--random` - Random selection (default now)
   - Time period filters:
     - `--today` / `--1d` - Last 24 hours
     - `--week` / `--7d` - Last 7 days
     - `--month` / `--30d` - Last 30 days
     - `--year` / `--1y` - Last year
     - `--all-time` - All time (default for top)

### New CLI Examples:
```bash
# Top anime wallpapers this week
./wallpaper-cli fetch --top --week --tags "anime" --limit 10

# Most liked 4K wallpapers of all time
./wallpaper-cli fetch --most-liked --resolution 4k --limit 5

# Most viewed from last month
./wallpaper-cli fetch --most-viewed --month --limit 10

# Latest uploads (newest first)
./wallpaper-cli fetch --latest --limit 10

# Reddit's top posts this week
./wallpaper-cli fetch --source reddit --top --week --limit 10
```

### Estimated: 5-6 hours

---

## 📋 v1.2 - System Integration (FUTURE)

**M003: Desktop Integration**

### Planned:
1. **Auto-Set Wallpaper**
   - macOS: `osascript` or `sqlite` (dock db)
   - Linux: `feh`, `nitrogen`, or `gsettings`
   - Windows: Registry or PowerShell
   - Per-monitor support (stretch)

2. **TUI with FZF**
   - Interactive wallpaper picker
   - Preview thumbnails
   - fuzzy search by tags/source
   - Batch selection
   - Library: `charmbracelet/bubbletea` or `ktr0731/go-fuzzyfinder`

### Estimated: 6-8 hours

---

## 🎯 Backlog (Future Consideration)

| Feature | Version | Priority |
|---------|---------|----------|
| Zerochan source | v1.3 | Low |
| AI auto-tagging | v1.3+ | Low |
| Multi-monitor | v1.3+ | Medium |
| Wallpaper rotation | v1.3+ | Low |
| Web UI | v2.0 | Very Low |

---

## Success Criteria by Version

### v1.0 (COMPLETE)
- [x] Fetch and download wallpapers
- [x] Deduplication works
- [x] Organization works
- [x] Cross-platform builds

### v1.1 (Target)
- [ ] Progress bar shows % and speed
- [ ] Reddit source fetches r/Animewallpaper
- [ ] Both sources work together
- [ ] Config supports multiple sources
- [ ] Sorting by popularity (top, favorites, views)
- [ ] Time period filtering (today, week, month, year)
- [ ] `--latest` for newest uploads

### v1.2 (Target)
- [ ] Auto-set wallpaper on all 3 platforms
- [ ] TUI for interactive selection
- [ ] Fuzzy search working

---

## Version Philosophy

- **v1.0:** It works (core functionality)
- **v1.1:** It feels good + find the best wallpapers (UX + more content + sorting)
- **v1.2:** It integrates (system + interactivity)
- **v1.3+:** It gets fancy (AI, multi-monitor, etc.)

---

*Roadmap aligned with user request: v1.1 = progress bar + Reddit + popularity sorting + time filtering, v1.2 = auto-set + TUI*
