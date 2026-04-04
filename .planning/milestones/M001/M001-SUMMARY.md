# M001: Wallpaper CLI Tool - Summary

## Milestone Overview

| Field | Value |
|-------|-------|
| **ID** | M001 |
| **Title** | Wallpaper CLI Tool |
| **Status** | 🔵 Active |
| **Slices** | 8 |
| **Total Estimated Tasks** | 40+ |

## Vision

A resource-efficient, single-binary CLI tool for downloading high-quality anime wallpapers with smart filtering, deduplication, and organization. Users can fetch wallpapers from multiple sources with confidence that duplicates won't pile up, and the tool will respect their bandwidth and system resources.

## Key Metrics (Target)

| Metric | Target |
|--------|--------|
| Binary Size | < 20MB |
| Memory (Idle) | < 10MB |
| Startup Time | < 50ms |
| Concurrent Downloads | 5 (default) |

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         WALLPAPER CLI TOOL                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐     │
│   │   CLI Layer      │    │  Source Adapters │    │  Download Mgr    │     │
│   │   (cmd/)         │◄──►│  (sources/)      │◄───►│  (download/)     │     │
│   └──────────────────┘    └──────────────────┘    └──────────────────┘     │
│           │                       │                       │                │
│           ▼                       ▼                       ▼                │
│   ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐     │
│   │  Filter Engine   │    │  Deduplication   │    │  Organization    │     │
│   │  (filter/)       │    │  (dedup/)        │    │  (organize/)     │     │
│   └──────────────────┘    └──────────────────┘    └──────────────────┘     │
│                                             │                              │
│                                             ▼                              │
│                                    ┌──────────────────┐                    │
│                                    │   SQLite Store   │                    │
│                                    │   (data/)        │                    │
│                                    └──────────────────┘                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Slice Overview

| Slice | Goal | Risk | Est. | Key Deliverable |
|-------|------|------|------|-----------------|
| S01 | Project scaffold | Low | 1.5h | Working CLI binary |
| S02 | CLI & config | Low | 2h | All commands/flags |
| S03 | Wallhaven adapter | Medium | 3h | API integration |
| S04 | Download manager | Medium | 2.5h | Concurrent downloads |
| S05 | Deduplication | High | 3h | pHash + SQLite |
| S06 | Organization | Low | 2h | File organization |
| S07 | Cross-platform | Medium | 2.5h | Release binaries |
| S08 | Reddit (stretch) | Medium | 2.5h | Multi-source |

## Risk Register

| Risk | Severity | Mitigation |
|------|----------|------------|
| pHash library performance | High | Research first (S02), fallback to simple hash |
| Reddit API limits | Medium | Use JSON API, rate limiting |
| CGO/cross-compilation | Medium | Use modernc.org/sqlite (CGO-free) |
| Platform path handling | Low | Use filepath.Join, path/filepath |

## Success Criteria

- [ ] Single binary < 20MB (all platforms)
- [ ] Memory < 10MB at idle
- [ ] Wallhaven fetch working end-to-end
- [ ] Deduplication across sessions
- [ ] CLI complete with all commands
- [ ] Cross-platform builds (macOS, Linux, Windows)

## Definition of Done

- All success criteria met
- Unit tests >80% coverage
- Integration tests passing
- Documentation complete
- GitHub releases automated

---

*Milestone plan created based on wallpaper-cli-spec.md*
