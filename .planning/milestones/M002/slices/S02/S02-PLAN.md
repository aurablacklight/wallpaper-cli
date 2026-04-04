

---

### T09: macOS Integration Hint
**Estimate:** 30m

**Description:**
Add a creative feature: detect macOS WallpaperEngine app and show a helpful hint in TUI.

**Motivation:**
The CLI TUI and macOS app work great together:
- TUI = Fast terminal browsing with fuzzy search
- macOS App = Live wallpaper rendering (video/web/scene)

This feature bridges both experiences by suggesting the app when relevant.

**Steps:**
1. Create `internal/tui/hint.go` for platform detection
2. Detect macOS: `runtime.GOOS == "darwin"`
3. Check if `/Applications/WallpaperEngine.app` exists
4. Show hint banner: "💡 Tip: Open WallpaperEngine.app for live wallpapers (press 'o')"
5. Add 'o' keybinding to open the app (via `open` command)
6. Make hint dismissible with 'd' key

**UI Mockup:**
```
┌─────────────────────────────────────────────────────┐
│ Browse Wallpapers                     [?q help]     │
├─────────────────────────────────────────────────────┤
│  > 01_anime_sunset_4k.jpg    3840x2160  Wallhaven  │
│    02_city_night_4k.jpg      3840x2160  Reddit     │
│    03_forest_mist_4k.jpg     3840x2160  Wallhaven  │
│                                                     │
│  ────────────────────────────────────────────────  │
│  💡 Tip: Open WallpaperEngine.app for live wp's    │
│         Press 'o' to open  |  'd' to dismiss     │
│  ────────────────────────────────────────────────  │
└─────────────────────────────────────────────────────┘
```

**Files Likely Touched:**
- internal/tui/hint.go
- internal/tui/view.go (add hint to layout)
- internal/tui/keys.go (add 'o' and 'd' keys)

**Expected Output:**
- macOS users see helpful integration hint
- Press 'o' opens WallpaperEngine.app
- Press 'd' dismisses hint for session

**Verification:**
```bash
# On macOS with WallpaperEngine.app installed
./wallpaper-cli browse
# See hint at bottom
# Press 'o' — app should open
```

**Cross-Reference:**
This feature complements the macOS app integration implemented in:
- `macos-wallpaper-app` branch: `feature/cli-integration`
- Commit: Auto-discovery of CLI folders

---

*S02 includes creative macOS integration bridge feature*
