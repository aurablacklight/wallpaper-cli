# ROADMAP: Wallpaper CLI Tool

**Active Milestone:** M001 - Wallpaper CLI Tool

**Project Vision:** A resource-efficient, single-binary CLI tool for downloading high-quality anime wallpapers with smart filtering, deduplication, and organization.

---

## Milestones

| ID | Title | Status | Progress | Est. Completion |
|----|-------|--------|----------|-----------------|
| M001 | Wallpaper CLI Tool | 🔵 Active | 0/8 slices | TBD |

---

## M001: Wallpaper CLI Tool

### Slices

| ID | Title | Goal | Risk | Status | Depends On |
|----|-------|------|------|--------|------------|
| S01 | Project Foundation & CLI Scaffold | Project scaffolding with Go modules, CLI framework, and build pipeline | low | [ ] todo | - |
| S02 | CLI Interface & Config System | Complete CLI command structure with all commands and flags from spec | low | [ ] todo | S01 |
| S03 | Wallhaven Source Adapter | Wallhaven API adapter with search, filtering, and metadata extraction | medium | [ ] todo | S02 |
| S04 | Download Manager | Concurrent download manager with progress tracking and resume support | medium | [ ] todo | S03 |
| S05 | Deduplication System | Deduplication system with perceptual hashing and SQLite storage | high | [ ] todo | S04 |
| S06 | Organization & Storage | File organization with directory structure and metadata storage | low | [ ] todo | S05 |
| S07 | Cross-Platform Builds & Optimization | Cross-platform build pipeline and resource efficiency verification | medium | [ ] todo | S06 |
| S08 | Reddit Source Adapter (Stretch) | Reddit source adapter as secondary wallpaper source | medium | [ ] todo | S03, S06 |

### After M001

Users will have a working CLI tool that can fetch anime wallpapers from Wallhaven with smart filtering, deduplication, and organization. The tool will be a single binary under 20MB that works on macOS, Linux, and Windows.

### Future Milestones

- M002: Reddit/Zerochan additional sources
- M003: AI tagging via local LLM
- M004: Multi-monitor and auto-wallpaper features

---

## Navigation

- [M001 ROADMAP](./milestones/M001/M001-ROADMAP.md)
- [Requirements](./REQUIREMENTS.md) (generated from SPEC.md)

---

*Generated from wallpaper-cli-spec.md*
