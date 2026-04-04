# M002: Desktop Integration - Summary

## Milestone Overview

| Field | Value |
|-------|-------|
| **ID** | M002 |
| **Title** | Desktop Integration |
| **Version** | v1.2 |
| **Status** | 🚧 Planned |
| **Slices** | 3 |
| **Estimated Tasks** | 21 |

---

## Vision

Transform the wallpaper CLI from a download-only tool into a complete desktop wallpaper management solution. Users can not only fetch wallpapers but also preview, select, and automatically set them as their desktop background across macOS, Linux, and Windows.

---

## Key Deliverables

### Auto-Set Wallpaper
- `wallpaper-cli set <path>` — Set specific wallpaper
- `wallpaper-cli set --random` — Set random from collection  
- `wallpaper-cli set --latest` — Set most recent download

### TUI Browser
- `wallpaper-cli browse` — Interactive terminal UI
- Thumbnail previews with metadata
- Arrow key navigation
- Real-time fuzzy search (`/` to search)
- Enter to set wallpaper immediately

### Platform Support
- macOS (AppleScript / NSWorkspace)
- Linux (GNOME, KDE, XFCE, fallback)
- Windows (PowerShell / Registry)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      DESKTOP INTEGRATION (v1.2)                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐     │
│   │   CLI Layer      │    │  Platform APIs   │    │   TUI Layer      │     │
│   │   (cmd/)         │◄──►│  (platform/)     │◄───►│  (tui/)          │     │
│   │                  │    │                  │    │                  │     │
│   │  • set command   │    │  • macOS         │    │  • browse        │     │
│   │  • set --random  │    │  • Linux         │    │  • fuzzy search  │     │
│   │  • set --latest  │    │  • Windows       │    │  • preview       │     │
│   └──────────────────┘    └──────────────────┘    └──────────────────┘     │
│           │                       │                       │                │
│           ▼                       ▼                       ▼                │
│   ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐     │
│   │  Config Store    │    │  System APIs     │    │  Image Cache     │     │
│   │  (config.json)   │    │  • osascript     │    │  (thumbs/)       │     │
│   │                  │    │  • gsettings     │    │                  │     │
│   │  Current         │    │  • nitrogen      │    │  • 256x256       │     │
│   │  wallpaper path  │    │  • Registry      │    │    thumbnails    │     │
│   └──────────────────┘    └──────────────────┘    └──────────────────┘     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Slice Overview

| Slice | Goal | Risk | Est. | Key Deliverable |
|-------|------|------|------|-----------------|
| S01 | Cross-platform wallpaper setting | High | 4.5h | `set` command working on macOS/Linux/Windows |
| S02 | TUI with Bubble Tea | Medium | 5h | Interactive `browse` command with thumbnails |
| S03 | Fuzzy search & integration | Low | 3h | Search/filter with Enter-to-set |
| S04 | macOS App Integration | Low | 2.5h | Auto-discover CLI downloads in WallpaperEngine app |

**Total Estimate:** 15 hours (was 12.5, +2.5 for S04)

---

## macOS Integration Highlight

**S04: macOS App Integration** enables a powerful workflow:
1. Use CLI to batch download from Wallhaven/Reddit (with deduplication)
2. Launch macOS WallpaperEngine app
3. CLI downloads appear automatically in browser
4. Enjoy live wallpapers (video, web, scene) with native performance

See: [Integration Spec](../../INTEGRATION-macOS-WallpaperEngine.md)

---

## Technology Stack

| Component | Library | Purpose |
|-----------|---------|---------|
| TUI Framework | charmbracelet/bubbletea | Interactive terminal UI |
| Styling | charmbracelet/lipgloss | CSS-like styling in terminal |
| Components | charmbracelet/bubbles | Pre-built TUI components |
| Fuzzy Search | sahilm/fuzzy | String matching algorithm |
| Platform Detection | runtime.GOOS + env vars | OS and DE detection |

---

## Risk Register

| Risk | Severity | Mitigation |
|------|----------|------------|
| macOS sandboxing/permissions | High | Test early, document permission requirements |
| Linux DE fragmentation | High | Support top 3 DEs, graceful fallback |
| Windows API differences | Medium | Test on Win10/11, use PowerShell as common interface |
| TUI performance with many images | Medium | Lazy loading, pagination, async thumbnail gen |
| Cross-platform testing burden | Medium | CI testing where possible, manual on key platforms |

---

## Success Criteria

- [ ] `wallpaper-cli set <path>` works on macOS
- [ ] `wallpaper-cli set <path>` works on Linux (GNOME/KDE/XFCE)
- [ ] `wallpaper-cli set <path>` works on Windows
- [ ] `wallpaper-cli set --random` works on all platforms
- [ ] `wallpaper-cli set --latest` works on all platforms
- [ ] `wallpaper-cli browse` launches TUI
- [ ] TUI displays thumbnails
- [ ] Fuzzy search filters in <100ms
- [ ] Selection in TUI sets wallpaper
- [ ] Binary size remains <20MB

---

## Definition of Done

- All success criteria met
- Unit tests for platform backends
- Manual testing on macOS (Intel + ARM)
- Manual testing on Linux (Ubuntu GNOME + KDE)
- Manual testing on Windows 10/11
- Documentation updated
- Demo GIF/screenshots added to README

---

## Dependencies

- M001 complete (provides base CLI, fetch, download, storage)
- M001 v1.1 features (Reddit, progress bar, sorting) — all verified working

---

## After This Milestone

Users will have a complete wallpaper management workflow:
1. `fetch` — Download from Wallhaven/Reddit
2. `browse` — Interactive TUI to explore collection
3. `set` — Set wallpapers directly from CLI

Future milestones could add:
- Auto-rotation/scheduling
- Multi-monitor support
- AI auto-tagging
- Additional sources (Zerochan, etc.)

---

## Verification Contracts

**Code Changes:**
- Unit tests for each platform backend (mock system calls)
- TUI component tests
- Fuzzy search accuracy tests

**Integration:**
- End-to-end: fetch → browse → set workflow
- Test with 1000+ wallpaper collection

**Operational:**
- Cross-platform build pipeline
- Performance benchmarking

**UAT:**
- Manual testing checklist per platform
- Screenshot verification

---

*M002 planned based on v1.0/v1.1 completion and user workflow needs*
