# M003: Milestone Concept Brainstorm

**Date:** 2026-04-04  
**Context:** Post-M002 v1.2.0 — Desktop Integration Complete

---

## Current State Recap

**M002 Delivered:**
- Cross-platform wallpaper setting (`set` command with --random, --latest)
- Interactive TUI browser (`browse` with thumbnails, fuzzy search, pagination)
- Collection management (`list`, `export`, `stats`)
- macOS app integration (CLI-side complete, Swift PRs pending)

**User Workflow Now:**
1. `fetch` — Download from Wallhaven/Reddit
2. `browse` — Explore collection with TUI
3. `set` — Set wallpapers directly from CLI

**M003 Opportunity:** Build on this foundation with advanced automation, organization, and system integration features.

---

## Concept 1: Wallpaper Automation & Scheduling

**Working Title:** "Smart Wallpaper Rotation"
**Version:** v1.3

### Vision
Transform the CLI from manual wallpaper setting to automated, intelligent wallpaper management. Set it and forget it — wallpapers change automatically based on time, events, or user patterns.

### Core Features

| Feature | Description | Value |
|---------|-------------|-------|
| Time-based Rotation | Change wallpaper every N minutes/hours | Hands-free operation |
| Smart Scheduling | Different wallpapers for morning/day/evening/night | Context-aware ambiance |
| Cron Integration | Native cron job generation | Unix-native scheduling |
| Event Triggers | Change on workspace switch, idle detection, etc. | Responsive environment |
| Resume Support | Remember position after reboot | Seamless experience |

### CLI Commands
```bash
wallpaper-cli schedule --every 30m --random          # Rotate every 30 min
wallpaper-cli schedule --theme daytime --from 8am   # Day theme
wallpaper-cli schedule --theme nighttime --from 8pm  # Night theme
wallpaper-cli schedule --disable                     # Stop rotation
wallpaper-cli daemon start                           # Background service
```

### Technical Approach
- **Daemon Mode:** Background process (cross-platform service)
- **Scheduling:** cron (Unix), Task Scheduler (Windows), launchd (macOS)
- **State Management:** Persist current schedule, position, history
- **Platform Services:** Native background service registration

### Pros
✅ Builds directly on existing `set` functionality  
✅ High user value — daily convenience improvement  
✅ Natural extension of current workflow  
✅ Differentiates from static wallpaper apps  

### Cons
⚠️ Complex cross-platform service implementation  
⚠️ Daemon lifecycle management (start/stop/crash recovery)  
⚠️ Requires elevated permissions on some platforms  
⚠️ More complex testing (time-based, background processes)

### Estimated Effort: Medium-High (15-20 hours)
- S01: Core scheduling engine (5h)
- S02: Cross-platform daemon/service (6h)
- S03: CLI commands & state management (4h)
- S04: Testing & edge cases (4h)

---

## Concept 2: Multi-Monitor Support

**Working Title:** "Multi-Display Mastery"
**Version:** v1.3 or v1.4

### Vision
Power users often have multiple monitors — extend wallpaper management to support per-monitor wallpapers, spanning images across displays, and detecting monitor configuration changes.

### Core Features

| Feature | Description | Value |
|---------|-------------|-------|
| Per-Monitor Wallpapers | Different wallpaper on each display | Personalization |
| Span Mode | One image across all monitors | Immersive experience |
| Monitor Detection | Auto-detect display count, resolution, positions | Seamless setup |
| Display Profiles | Different configs for home/office setups | Context switching |
| Aspect Ratio Matching | Auto-select best-fit images per monitor | Quality optimization |

### CLI Commands
```bash
wallpaper-cli set --monitor 1 <path>           # Set specific monitor
wallpaper-cli set --span <path>                # Span all monitors
wallpaper-cli monitors list                     # Show detected displays
wallpaper-cli set --profile dual-monitor        # Apply profile
wallpaper-cli set --auto-fit                    # Best-fit per monitor
```

### Technical Approach
- **Monitor Detection:**
  - macOS: NSScreen, CoreGraphics
  - Linux: xrandr, wayland-display-protocol
  - Windows: Win32 API, EnumDisplayMonitors
- **Platform-Specific Setting:**
  - macOS: NSWorkspace per-desktop
  - Linux: feh --xinerama-index, nitrogen per-monitor
  - Windows: SystemParametersInfo with monitor index

### Pros
✅ High-value for power users (developers, creatives)  
✅ Clear market differentiation  
✅ Natural extension of existing platform abstraction  
✅ Mostly builds on existing `set` architecture  

### Cons
⚠️ Highly platform-specific APIs (fragmented implementation)  
⚠️ Linux Wayland support limited/fragmented  
⚠️ Complex testing (requires multiple monitors)  
⚠️ Smaller user base (not everyone has multi-monitor)  

### Estimated Effort: Medium (12-16 hours)
- S01: Monitor detection abstraction (4h)
- S02: Per-monitor wallpaper setting (5h)
- S03: Span/fit algorithms (3h)
- S04: Profiles & CLI integration (3h)

---

## Concept 3: AI-Powered Organization

**Working Title:** "Intelligent Collection"
**Version:** v1.4

### Vision
Collections grow to thousands of wallpapers — leverage AI for automatic tagging, similarity detection, and smart organization without manual effort.

### Core Features

| Feature | Description | Value |
|---------|-------------|-------|
| Auto-Tagging | AI generates tags from image content (anime character, scene type, color palette) | Searchable collection |
| Duplicate Detection | Beyond pHash — semantic similarity detection | Better deduplication |
| Smart Folders | Dynamic collections based on rules ("recent anime landscapes") | Auto-organization |
| Content Filtering | Auto-detect and flag sensitive content | Safety |
| Face/Character Recognition | Identify specific anime characters | Fan utility |

### CLI Commands
```bash
wallpaper-cli ai analyze --all                   # Tag entire collection
wallpaper-cli ai find-similar <path>             # Find visually similar
wallpaper-cli search --ai "landscape blue hair"  # Natural language search
wallpaper-cli smart-folder create "favorites" --rules "rating>4,age<30d"
wallpaper-cli ai dedupe --threshold 0.95         # Semantic deduplication
```

### Technical Approach
- **Local AI:** onnxruntime-go with quantized models (privacy-first)
- **Models:** CLIP for semantic understanding, ResNet for classification
- **Embedding Store:** Vector database (sqlite-vss, chroma)
- **Processing:** Async background indexing with progress tracking

### Pros
✅ Massive differentiation — no CLI tool has this  
✅ Scales with collection size (solves real pain point)  
✅ Privacy-respecting local processing  
✅ Opens future features (recommendations, smart rotation)  

### Cons
⚠️ Binary size increase (+50-100MB for models)  
⚠️ First-run indexing time (minutes for large collections)  
⚠️ Hardware requirements (CPU/GPU for inference)  
⚠️ Complex dependency management (ONNX, models)  
⚠️ Anime-specific models may need fine-tuning  

### Estimated Effort: High (25-35 hours)
- S01: AI pipeline & model integration (8h)
- S02: Embedding storage & vector search (6h)
- S03: Background indexing system (5h)
- S04: CLI commands & smart folders (4h)
- S05: Optimization & edge cases (6h)

---

## Concept 4: Advanced Collection Management

**Working Title:** "Collection Power Tools"
**Version:** v1.3

### Vision
Enhance the core collection experience with features that help users curate, organize, and enjoy their wallpapers at scale — favorites, playlists, ratings, and bulk operations.

### Core Features

| Feature | Description | Value |
|---------|-------------|-------|
| Favorites System | Mark and quickly access favorite wallpapers | Personal curation |
| Playlists | Create sequences of wallpapers to rotate through | Themed experiences |
| Ratings & Metadata | 1-5 star ratings, custom notes | Quality ranking |
| Bulk Operations | Tag, move, delete multiple wallpapers at once | Efficiency |
| Import/Export | Full collection backup with metadata | Portability |
| Search History | Remember and replay common searches | Convenience |

### CLI Commands
```bash
wallpaper-cli favorite add <path>                # Mark favorite
wallpaper-cli playlist create "cozy-winter"       # New playlist
wallpaper-cli playlist add "cozy-winter" <path> # Add to playlist
wallpaper-cli set --playlist "cozy-winter" --random
wallpaper-cli rate <path> 5                      # 5-star rating
wallpaper-cli bulk tag --source wallhaven --rating 4+
wallpaper-cli backup --output collection.zip     # Full backup
```

### Technical Approach
- **Metadata Schema:** Extend database with favorites, ratings, playlists, notes
- **Storage:** SQLite for metadata, filesystem for organization
- **TUI Enhancement:** Inline rating, bulk selection mode, playlist management
- **Import/Export:** ZIP with JSON metadata manifest

### Pros
✅ Moderate complexity — mostly data layer work  
✅ High immediate user value  
✅ Builds on existing collection foundation  
✅ Enables future AI features (train on favorites)  
✅ Keeps binary size small (no new heavy deps)  

### Cons
⚠️ Database migrations for existing users  
⚠️ TUI complexity increase (bulk selection, new views)  
⚠️ Export/import format design decisions  
⚠️ Playlist scheduling conflicts with rotation feature  

### Estimated Effort: Medium (12-16 hours)
- S01: Database schema extensions (3h)
- S02: Favorites & ratings system (3h)
- S03: Playlist management (3h)
- S04: Bulk operations & TUI enhancements (3h)
- S05: Import/export & backup (3h)

---

## Concept 5: Sync & Cloud Integration

**Working Title:** "Cloud Collection"
**Version:** v1.4 or v1.5

### Vision
Users have multiple devices — enable seamless wallpaper synchronization across machines with cloud storage integration and conflict resolution.

### Core Features

| Feature | Description | Value |
|---------|-------------|-------|
| Cloud Sync | iCloud (macOS), OneDrive, Dropbox, GDrive | Multi-device access |
| Selective Sync | Choose which collections sync | Storage management |
| Conflict Resolution | Handle simultaneous edits across devices | Data integrity |
| Cross-Device Handoff | Start browsing on one, set on another | Seamless workflow |
| Shared Collections | Public/shareable wallpaper packs | Community |

### CLI Commands
```bash
wallpaper-cli sync enable --provider icloud      # Enable iCloud
wallpaper-cli sync status                        # Show sync state
wallpaper-cli sync push                          # Force upload
wallpaper-cli sync pull                          # Force download
wallpaper-cli share create "winter-collection"     # Create shareable pack
wallpaper-cli share install <url>                # Install shared pack
```

### Technical Approach
- **Providers:** Native SDKs (where available), rclone-style approach for others
- **Sync Engine:** Differential sync with conflict resolution
- **Metadata Sync:** SQLite replication or JSON-based state
- **Storage Efficiency:** Delta sync for metadata, dedup for images

### Pros
✅ Addresses real multi-device pain point  
✅ Platform-native integration (iCloud on macOS, etc.)  
✅ Enables sharing/community features  

### Cons
⚠️ Complex distributed systems challenges (conflicts, offline)  
⚠️ Rate limits and API costs for cloud providers  
⚠️ Security/privacy concerns (user images in cloud)  
⚠️ Heavy testing burden (multiple devices, network conditions)  
⚠️ iCloud requires Apple Developer account  

### Estimated Effort: High (20-30 hours)
- S01: Sync engine architecture (6h)
- S02: Provider implementations (8h)
- S03: Conflict resolution & offline mode (5h)
- S04: Sharing infrastructure (4h)
- S05: Testing & edge cases (5h)

---

## Summary Comparison

| Concept | User Impact | Technical Complexity | Effort | Binary Impact | Risk |
|---------|-------------|---------------------|--------|---------------|------|
| 1. Automation & Scheduling | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 15-20h | Low | Medium |
| 2. Multi-Monitor | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 12-16h | Low | Medium |
| 3. AI-Powered Organization | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 25-35h | High (+100MB) | High |
| 4. Collection Management | ⭐⭐⭐⭐ | ⭐⭐⭐ | 12-16h | Low | Low |
| 5. Sync & Cloud | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 20-30h | Medium | High |

---

## Recommendation

**Primary Recommendation: Concept 1 (Automation & Scheduling)**

**Why:**
1. **Natural Evolution:** Builds directly on M002's `set` command — daemon just automates what users already do manually
2. **Daily Value:** Every user benefits from hands-free wallpaper rotation
3. **Technical Fit:** Cross-platform services are challenging but well-understood patterns
4. **Foundation for Future:** Scheduling enables time-based themes, which pairs well with later AI features
5. **Manageable Scope:** Can deliver core rotation in 15-20 hours, then iterate

**Alternative Combinations:**
- **Conservative:** Concept 4 (Collection Management) — safer scope, immediate value
- **Ambitious:** Concept 1 + 4 together — automation + better organization
- **Power User:** Concept 2 (Multi-Monitor) — if targeting developer/creative demographic

---

## Next Steps

1. **Select Concept:** Choose which milestone direction to pursue
2. **Scope Definition:** Define exact feature boundaries for selected concept
3. **Research Phase:** Deep-dive into technical implementation (especially for daemon/services)
4. **Slice Planning:** Break into 3-4 slices with estimated tasks
5. **Update Roadmap:** Document M003 in ROADMAP.md

---

*Generated for M003 milestone planning — M002 v1.2.0 complete*
