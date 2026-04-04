# Wallpaper CLI Tool - Product Roadmap

**Current Status:** v1.1 Complete ✅  
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

## ✅ v1.1 - Better UX + More Sources + Sorting (COMPLETE)

**M002: Enhanced Experience**

### Delivered:
1. **Progress Bar with schollz/progressbar/v3**
   - ✅ Real-time progress bars (percentage)
   - ✅ Download speed indicator (MB/s)
   - ✅ Visual queue status
   - ✅ Completion summary

2. **Reddit Source Adapter**
   - ✅ r/Animewallpaper integration
   - ✅ Reddit JSON API (no OAuth)
   - ✅ Multi-subreddit support (configurable)
   - ✅ Cross-source deduplication
   - ✅ Direct image URL extraction

3. **Popularity & Time Filtering**
   - ✅ `--sort favorites` - Most favorited
   - ✅ `--sort views` - Most viewed
   - ✅ `--sort top` - Top rated + time range
   - ✅ `--sort hot` - Hot posts (Reddit)
   - ✅ `--sort new` - Newest first
   - ✅ `--latest` - Shorthand for newest
   - ✅ `--popular` - Shorthand for top
   - ✅ Time period filters:
     - `--day` / `--today` - Last 24 hours
     - `--week` / `--7d` - Last 7 days
     - `--month` / `--30d` - Last 30 days
     - `--year` / `--1y` - Last year
     - `--all-time` - All time

4. **Multi-Source Support**
   - ✅ `--source wallhaven` - Wallhaven only
   - ✅ `--source reddit` - Reddit only
   - ✅ `--source all` - Both sources combined

### Verified CLI Examples:
```bash
# Top anime wallpapers this week
./wallpaper-cli fetch --sort top --week --tags "anime" --limit 10

# Most favorited 4K wallpapers of all time
./wallpaper-cli fetch --favorites --all-time --resolution 4k --limit 5

# Most viewed from last month
./wallpaper-cli fetch --sort views --month --limit 10

# Latest uploads (newest first)
./wallpaper-cli fetch --latest --limit 10

# Reddit's hot posts
./wallpaper-cli fetch --source reddit --sort hot --limit 10

# Multi-source fetch
./wallpaper-cli fetch --source all --limit 10
```

### Metrics:
- Features delivered: 100% ✅
- Smoke tests passed: 10/10 ✅
- Binary size: 11MB (still <20MB) ✅

---

## 🚧 v1.2 - System Integration (NEXT)

**M003: Desktop Integration**

### Planned:
1. **Auto-Set Wallpaper**
   - macOS: `osascript` or `sqlite` (dock db)
   - Linux: `feh`, `nitrogen`, or `gsettings`
   - Windows: Registry or PowerShell
   - Per-monitor support (stretch)

2. **TUI with Interactive Picker**
   - Browse downloaded wallpapers
   - Preview thumbnails
   - Fuzzy search by tags/filename
   - Batch selection for setting
   - Library: `charmbracelet/bubbletea` or `charmbracelet/lipgloss`

3. **Wallpaper Set Command**
   - `wallpaper-cli set <path>` - Set specific wallpaper
   - `wallpaper-cli set --random` - Set random from collection
   - `wallpaper-cli set --latest` - Set most recent download

### Estimated: 6-8 hours

---

## 📋 v1.3+ - Future Consideration

| Feature | Version | Priority |
|---------|---------|----------|
| Zerochan source | v1.3 | Low |
| AI auto-tagging | v1.3+ | Low |
| Multi-monitor | v1.3+ | Medium |
| Wallpaper rotation | v1.3+ | Low |
| Web UI | v2.0 | Very Low |

---

## Success Criteria by Version

### v1.0 (COMPLETE) ✅
- [x] Fetch and download wallpapers
- [x] Deduplication works
- [x] Organization works
- [x] Cross-platform builds

### v1.1 (COMPLETE) ✅
- [x] Progress bar shows % and speed
- [x] Reddit source fetches r/Animewallpaper
- [x] Both sources work together (`--source all`)
- [x] Config supports multiple sources
- [x] Sorting by popularity (top, favorites, views)
- [x] Time period filtering (day, week, month, year)
- [x] `--latest` for newest uploads

### v1.2 (TARGET)
- [ ] Auto-set wallpaper on all 3 platforms
- [ ] TUI for interactive selection
- [ ] `set` command for wallpaper management
- [ ] **macOS Integration:** CLI downloads appear in WallpaperEngine app

---

## Version Philosophy

- **v1.0:** It works (core functionality) ✅
- **v1.1:** It feels good + find the best wallpapers (UX + more content + sorting) ✅
- **v1.2:** It integrates (system + interactivity)
- **v1.3+:** It gets fancy (AI, multi-monitor, etc.)

---

*Roadmap updated: v1.0 and v1.1 verified complete via smoke tests*
