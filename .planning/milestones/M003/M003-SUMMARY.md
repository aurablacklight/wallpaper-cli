# M003: Smart Automation & Collection Curation — Summary

**Version:** v1.3.0  
**Status:** Planning Complete — Ready for Execution  
**Estimated Effort:** 25-32 hours  
**Slices:** 4 (S01-S04)

---

## Vision

Transform wallpaper-cli from a manual tool into an intelligent, automated wallpaper companion that:
1. **Automatically rotates** wallpapers based on time, events, or user patterns
2. **Helps curate** collections with favorites, playlists, and ratings  
3. **Provides a rich TUI** with split-pane browsing and better thumbnails

**Key User Experience:**
> "I run `wallpaper-cli daemon start` once, and my wallpapers change beautifully throughout the day. I favorite the ones I love, create themed playlists for work/focus/relaxation, and the TUI lets me browse my curated collection in a rich, visual interface."

---

## What's Included

### ✅ Concept 1: Automation & Scheduling
- Time-based rotation (every N minutes/hours)
- Smart themes (morning/day/evening/night with different playlists)
- Background daemon with platform-native service integration
- Resume support (remember position after reboot)

### ✅ Concept 4: Collection Management  
- Favorites system (⭐ quick toggle)
- Playlists for themed rotation ("cozy", "focus", "energetic")
- Ratings 1-5 stars with optional notes
- Search history
- Import/export with metadata

### ✅ Bonus: TUI Overhaul (Split-Pane)
- Split-pane layout: list left, preview right
- Adaptive thumbnail scaling (32/48/64px based on terminal)
- Large preview pane (80-200px based on terminal height)
- Responsive design (handles terminal resize)
- Enhanced keybindings (vim-style: j/k/h/l, f/r/p for actions)

### ❌ Out of Scope (Explicitly Excluded)
- Multi-monitor support (deferred to M004)
- AI-powered organization (deferred, too complex)
- Cloud sync (deferred, distributed systems complexity)

---

## Slices Breakdown

| Slice | Title | Focus | Est. Hours | Key Deliverables |
|-------|-------|-------|------------|------------------|
| **S01** | TUI Overhaul | Frontend | 6-8h | Split-pane layout, adaptive thumbnails, responsive design |
| **S02** | Collection Management | Data Layer | 6-8h | Favorites, playlists, ratings, database schema |
| **S03** | Scheduling Engine | Core Logic | 5-6h | Interval/ theme/ fixed-time scheduling, rotation logic |
| **S04** | Daemon & Platform Services | Platform | 8-10h | launchd/systemd/Task Scheduler integration, background service |

**Total:** 25-32 hours (manageable single milestone)

---

## New Commands

### Schedule Management
```bash
wallpaper-cli schedule create "work-focus" --every 30m --source favorites
wallpaper-cli schedule create "day-cycle" --theme --morning 06:00 --evening 18:00
wallpaper-cli schedule enable "work-focus"
wallpaper-cli schedule list
wallpaper-cli schedule delete "work-focus"
```

### Collection Management
```bash
wallpaper-cli favorite toggle <path>
wallpaper-cli favorite list
wallpaper-cli playlist create "cozy-winter"
wallpaper-cli playlist add "cozy-winter" <path>
wallpaper-cli rate <path> 5 --notes "Great colors"
wallpaper-cli set --favorite --random
wallpaper-cli set --playlist "cozy-winter"
```

### Daemon Control
```bash
wallpaper-cli daemon start
wallpaper-cli daemon stop
wallpaper-cli daemon status
wallpaper-cli daemon install  # Platform service registration
wallpaper-cli daemon logs --follow
```

---

## Technical Highlights

### Split-Pane TUI
```
┌─────────────────────────────────────────────────────────────┐
│  📋 WALLPAPERS       │  👁️  PREVIEW + METADATA              │
│  🖼️  01_abc.jpg     │  ┌──────────────────────────────┐   │
│  🖼️  02_def.png  ◄──┼──│    [Large Thumbnail]         │   │
│  🖼️  03_ghi.jpg     │  │    3840x2160 • 2.4MB         │   │
│                      │  └──────────────────────────────┘   │
│  [j/k navigate]      │  ⭐ [f] Toggle favorite            │
│  [f favorite]        │  ★★★★★ [r] Rate                  │
│  [r rate]            │  📋 [p] Add to playlist          │
└──────────────────────┴──────────────────────────────────────┘
```

### Database Additions
```sql
-- Favorites, ratings, playlists (6 new tables)
-- Full migration plan in S02-PLAN.md
```

### Platform Services
- **macOS:** launchd plist, `launchctl` integration
- **Linux:** systemd user service (preferred), cron fallback
- **Windows:** Task Scheduler, `schtasks.exe` integration

---

## Key Design Decisions

1. **Split-pane TUI:** 40/60 ratio on wide terminals, stacked on narrow
2. **Adaptive thumbnails:** Scale 32→48→64px based on terminal width  
3. **Sequential + Shuffle:** Both rotation modes supported
4. **Theme scheduling:** Morning/day/evening/night with different playlists
5. **User-level daemon:** No root/admin required
6. **State persistence:** JSON state file for resume after crash/reboot

---

## Risk Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Cross-platform daemon complexity | High | Abstract platform layer, test per OS |
| TUI rendering bugs | Medium | Proven Bubble Tea patterns, incremental rollout |
| Database migrations | Medium | Versioned migrations, backup before upgrade |
| Binary size increase | Low | Monitor deps, current: 18MB, target: <25MB |

---

## Success Criteria

- [ ] Split-pane TUI renders on terminals ≥ 80 columns
- [ ] Favorites persist and show ⭐ in list
- [ ] Playlists can be created and rotated through
- [ ] Ratings are 1-5 with optional notes
- [ ] Daemon runs on all 3 platforms
- [ ] Rotation works: interval, theme, fixed-time
- [ ] State persists across daemon restarts
- [ ] Service auto-starts on login
- [ ] Binary size remains under 25MB

---

## Post-M003 Vision

After M003, wallpaper-cli becomes a "smart wallpaper companion":
- Set up once, runs automatically
- Collections curated with favorites and playlists
- Rich TUI for browsing and management

**Future M004 possibilities:**
- Multi-monitor support (finally!)
- Weather-aware wallpaper selection
- Activity-based rotation (work vs. break)
- Basic smart suggestions based on favorites
- Time-of-day automatic themes

---

## Ready to Execute

**Next Step:** Begin slice execution in order:
1. **S01:** TUI Overhaul (establishes new interface)
2. **S02:** Collection Management (adds data layer)
3. **S03:** Scheduling Engine (adds rotation logic)
4. **S04:** Daemon & Platform Services (brings it all together)

**Can parallelize:** S02 and S03 can be developed in parallel after S01

---

*M003 combines automation and curation — set it, forget it, but always have control when you want it.*
