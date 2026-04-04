# M003: Smart Automation & Collection Curation

**Working Title:** "Set It and Forget It"  
**Version:** v1.3.0  
**Status:** Planning Phase  
**Scope:** Combined Automation (Concept 1) + Collection Management (Concept 4) + TUI Overhaul

---

## Vision

Transform wallpaper-cli from a manual tool into an intelligent, automated wallpaper companion that:
1. **Automatically rotates** wallpapers based on time, events, or user patterns
2. **Helps curate** collections with favorites, playlists, and ratings
3. **Provides a rich TUI** with split-pane browsing, better thumbnails, and scaling

**User Experience:**
> "I run `wallpaper-cli daemon start` once, and my wallpapers change beautifully throughout the day. I favorite the ones I love, create themed playlists for work/focus/relaxation, and the TUI lets me browse my curated collection in a rich, visual interface."

---

## Combined Scope

### From Concept 1 (Automation)
- ⏰ Time-based rotation (every N minutes/hours)
- 🌅 Smart themes (morning/day/evening/night)
- 🔄 Background daemon/service (cron, Task Scheduler, launchd)
- 💾 Resume support (remember position after reboot)

### From Concept 4 (Collection Management)
- ⭐ Favorites system with quick access
- 📋 Playlists for themed rotation
- ⭐ Ratings (1-5 stars) and custom notes
- 🔍 Search history
- 📦 Import/export with metadata

### New: TUI Overhaul
- 📐 Split-pane layout (list left, preview right)
- 🔍 Better thumbnail scaling (adaptive sizing)
- 📱 Responsive design (handles terminal resize)
- ⌨️  Enhanced keybindings (vim-style navigation)
- 🎨 Improved visual polish

---

## Definition of Done

- [ ] `schedule` command creates time-based rotation
- [ ] `daemon` runs as background service on all platforms
- [ ] `favorite` system with quick favorite toggle
- [ ] `playlist` management with themed collections
- [ ] `rate` command for 1-5 star ratings
- [ ] TUI shows split-pane layout with live preview
- [ ] Thumbnails scale appropriately to terminal size
- [ ] All features work together (daemon rotates through playlists)
- [ ] State persists across reboots
- [ ] Binary size remains under 25MB

---

## Success Criteria

1. **Automation:** Daemon runs reliably for 7+ days without intervention
2. **Scheduling:** Rotation works on all 3 platforms (cron, Task Scheduler, launchd)
3. **Favorites:** Users can mark/unmark favorites in TUI with single keypress
4. **Playlists:** At least 3 playlists can be created and rotated through
5. **TUI:** Split-pane renders correctly on terminals 80x24 and larger
6. **Thumbnails:** Scale appropriately (64x64 to 128x128 based on terminal width)
7. **Integration:** Daemon can rotate through favorites-only or specific playlists

---

## Out of Scope (Explicitly)

- ❌ Multi-monitor support (deferred to future milestone)
- ❌ AI-powered organization (deferred, too complex for this scope)
- ❌ Cloud sync (deferred, distributed systems complexity)
- ❌ Smart/auto-tagging (deferred, requires ML models)

---

## Slices Overview

| ID | Title | Goal | Focus | Est. Effort |
|----|-------|------|-------|-------------|
| **S01** | **TUI Overhaul: Split-Pane & Scaling** | Redesign TUI with split layout, better thumbnails | Frontend | 6-8h |
| **S02** | **Collection Management** | Favorites, playlists, ratings, metadata | Data Layer | 6-8h |
| **S03** | **Scheduling Engine** | Core rotation logic, time-based themes | Core Logic | 5-6h |
| **S04** | **Daemon & Platform Services** | Cross-platform background services | Platform | 8-10h |

**Total Estimated Effort:** 25-32 hours (manageable for single milestone)

---

## Slice Dependencies

```
S01 (TUI Overhaul) ────┐
                       ├──► S02 (Collection Mgmt) ────┐
S03 (Scheduling) ─────┘                               ├──► S04 (Daemon)
                                                        │
                                       (Integration) ──┘
```

**Parallel Work Possible:**
- S01 (TUI) and S03 (Scheduling) can be developed in parallel
- S02 (Collections) depends on both S01 and S03
- S04 (Daemon) must integrate everything at the end

---

## Key Design Decisions

### Split-Pane TUI Layout

```
┌─────────────────────────────────────────────────────────────┐
│                    wallpaper-cli v1.3                       │
├──────────────────────┬──────────────────────────────────────┤
│                      │                                      │
│  📋 WALLPAPERS       │  👁️  PREVIEW                         │
│  ─────────────────── │                                      │
│  🖼️  01_abc.jpg     │  ┌──────────────────────────────┐   │
│  🖼️  02_def.png     │  │                              │   │
│  🖼️  03_ghi.jpg  ◄──┼──│    [Large Thumbnail]         │   │
│  🖼️  04_jkl.png     │  │    3840x2160 • 2.4MB         │   │
│  🖼️  05_mno.jpg     │  │                              │   │
│                      │  │    ⭐ Favorite  [♥]          │   │
│  [j/k navigate]      │  │    ⭐⭐⭐⭐⭐ Rating [1-5]    │   │
│  [Enter: set]        │  │    📋 In Playlists: cozy     │   │
│  [f: favorite]       │  │                              │   │
│  [r: rate]           │  └──────────────────────────────┘   │
│  [p: playlist]       │                                      │
│  [/: search]         │  ℹ️  Metadata                      │
│  [q: quit]           │  Source: wallhaven                 │
│                      │  Tags: anime, landscape            │
│  📊 Stats: 156 total │  Downloaded: 2026-04-04            │
│  ⭐ 23 favorites      │                                      │
│                      │                                      │
├──────────────────────┴──────────────────────────────────────┤
│  💡 Tip: Press 'd' to start daemon  │  🔄 Rotation: OFF    │
└─────────────────────────────────────────────────────────────┘
```

**Layout Features:**
- **Left pane (40% width):** Scrollable wallpaper list with small thumbnails (32x32)
- **Right pane (60% width):** Large preview (up to 200px height), metadata, actions
- **Bottom bar:** Status, tips, current rotation state
- **Responsive:** Adjusts split ratio based on terminal width

### Thumbnail Scaling Strategy

| Terminal Width | List Thumbnail | Preview Size | Layout |
|----------------|----------------|--------------|--------|
| < 80 cols | 32x32 | 100px height | Stacked (no split) |
| 80-120 cols | 48x48 | 150px height | Split 50/50 |
| > 120 cols | 64x64 | 200px height | Split 40/60 |

### Daemon Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     DAEMON ARCHITECTURE                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    │
│   │   Config    │    │  Scheduler  │    │   State     │    │
│   │   Store     │◄──►│   Engine    │◄──►│   Manager   │    │
│   │             │    │             │    │             │    │
│   │ • Schedule  │    │ • Interval  │    │ • Position  │    │
│   │ • Playlist  │    │ • Themes    │    │ • History   │    │
│   │ • Favorites │    │ • Events    │    │ • Last set  │    │
│   └─────────────┘    └──────┬──────┘    └─────────────┘    │
│                              │                               │
│                              ▼                               │
│   ┌─────────────────────────────────────────────────────┐    │
│   │              PLATFORM ABSTRACTION                  │    │
│   │  ┌─────────┐  ┌─────────────┐  ┌─────────────────┐  │    │
│   │  │  macOS  │  │    Linux    │  │     Windows     │  │    │
│   │  │ launchd │  │    cron     │  │ Task Scheduler  │  │    │
│   │  └─────────┘  │   (systemd) │  │                 │  │    │
│   │             └─────────────┘  └─────────────────┘  │    │
│   └─────────────────────────────────────────────────────┘    │
│                              │                               │
│                              ▼                               │
│   ┌─────────────────────────────────────────────────────┐    │
│   │              WALLPAPER SETTER                       │    │
│   │         (Existing M002 platform code)               │    │
│   └─────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Cross-platform daemon complexity | High | High | Abstract platform layer, extensive testing per OS |
| TUI split-pane rendering bugs | Medium | Medium | Use proven Bubble Tea patterns, incremental rollout |
| Database migrations for collections | Medium | Medium | Versioned migrations, backup before upgrade |
| Binary size increase | Low | Low | Monitor dependencies, feature flags for heavy features |
| State corruption on crash | Medium | High | Atomic writes, periodic backups, recovery mode |

---

## Technical Stack (Additions to M002)

**New Dependencies:**
- `github.com/robfig/cron/v3` — Cron expression parsing (scheduling)
- `github.com/charmbracelet/lipgloss` v0.9+ — Enhanced styling (already present)
- `golang.org/x/sys` — Service control (Windows services, launchd)

**No AI/ML Dependencies** — Keeping binary size reasonable

**Database Schema Additions:**
```sql
-- Favorites
CREATE TABLE favorites (image_hash TEXT PRIMARY KEY, added_at DATETIME);

-- Ratings
CREATE TABLE ratings (image_hash TEXT PRIMARY KEY, rating INTEGER, notes TEXT);

-- Playlists
CREATE TABLE playlists (id TEXT PRIMARY KEY, name TEXT, created_at DATETIME);
CREATE TABLE playlist_items (playlist_id TEXT, image_hash TEXT, position INTEGER);

-- Schedule State
CREATE TABLE schedule_state (key TEXT PRIMARY KEY, value TEXT);
```

---

## Integration Points

**Daemon + Collections:**
```bash
# Rotate only through favorites
wallpaper-cli daemon start --source favorites

# Rotate through specific playlist
wallpaper-cli daemon start --playlist "cozy-winter"

# Mix: 70% favorites, 30% random
wallpaper-cli daemon start --mix favorites:70,random:30
```

**TUI + Collections:**
```
In TUI:
  [f] — Toggle favorite (star appears instantly)
  [r] — Rate 1-5 (opens rating selector)
  [p] — Add to playlist (opens playlist selector)
  [P] — Create new playlist from selection
```

---

## Technical Resources

**TUI Best Practices Research:**
- [TUI Best Practices](./TUI-BEST-PRACTICES.md) — Research from EXA search + expert sources
- [Split-Pane Visual Design](./slices/S01/TUI-DESIGN-VISUAL.md) — Mockups for all terminal sizes

**Key Research Sources:**
1. **PUG Author's Tips** (Louis Garman) — Tree of models, layout arithmetic, event loop performance
2. **Bubble Tea Multi-View Example** — State-driven view selection patterns
3. **go-termimg Library** (blacktop) — Modern terminal image protocol support

**Applied Best Practices:**
- ✅ Tree of models architecture (Root → List + Preview + Modals)
- ✅ lipgloss.Height/Width for layout calculations (not hardcoded)
- ✅ Async image loading via tea.Cmd (non-blocking)
- ✅ Responsive breakpoints (32/48/64px thumbnails)
- ✅ go-termimg with auto-protocol detection (Kitty → iTerm2 → SIXEL → Halfblocks)

---

## Post-M003 Vision

After M003, wallpaper-cli becomes a complete "smart wallpaper companion":
- Users set it up once and it runs automatically
- Collections are curated with favorites and playlists
- Rich TUI makes browsing delightful

**Future M004 possibilities:**
- Time-based themes (morning energy → evening calm)
- Weather-aware wallpaper selection
- Activity-based rotation (work vs. break wallpapers)
- Basic smart suggestions based on favorites

---

*M003 combines the best of automation and curation — set it, forget it, but always have control when you want it.*
