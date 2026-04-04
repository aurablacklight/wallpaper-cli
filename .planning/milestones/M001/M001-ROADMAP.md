# M001: Wallpaper CLI Tool

**Vision:** A resource-efficient, single-binary CLI tool for downloading high-quality anime wallpapers with smart filtering, deduplication, and organization. Users can fetch wallpapers from multiple sources with confidence that duplicates won't pile up, and the tool will respect their bandwidth and system resources.

---

## Definition of Done

- [ ] Single binary < 20MB verified via build artifacts
- [ ] Memory usage < 10MB at idle verified with runtime profiling
- [ ] Cross-platform builds passing (macOS, Linux, Windows)
- [ ] Wallhaven source adapter fetching and filtering working
- [ ] Deduplication working across sessions with pHash
- [ ] Concurrent downloads (5 parallel) without blocking
- [ ] CLI interface complete with all planned commands
- [ ] All success criteria from SPEC.md met

---

## Success Criteria

1. Single binary under 20MB verified for all platforms
2. Memory usage under 10MB at idle (verified with profiling)
3. Wallhaven source fetching, filtering, and downloading working end-to-end
4. Deduplication prevents re-downloading same images across sessions
5. CLI supports all planned commands with intuitive interface
6. Cross-platform builds for macOS, Linux, Windows

---

## Key Risks

| Risk | Why It Matters |
|------|--------------|
| Perceptual hashing (pHash) implementation complexity | Deduplication is a core requirement; naive hashing won't catch resized/cropped duplicates. May need external library. |
| Reddit API rate limits and authentication | Reddit requires OAuth for API access; rate limiting could slow downloads significantly. |
| WebP conversion complexity vs format preference | Spec prefers WebP for smaller files but conversion adds processing overhead and dependencies. |
| Cross-platform file path handling | Windows paths, macOS ~/ expansion, and Linux permissions all need correct handling. |

---

## Proof Strategy

| Risk/Unknown | What Will Be Proven | Retire In |
|--------------|---------------------|-----------|
| pHash library availability and performance | Go pHash library exists and can process images fast enough (<100ms per image) | S02 |
| Wallhaven API integration complexity | API search, filtering, and metadata extraction works as expected | S03 |
| Concurrent download implementation | Goroutine-based downloads work without memory leaks or race conditions | S06 |
| Cross-platform build pipeline | GoReleaser or similar can produce working binaries for all 3 platforms | S07 |

---

## Boundary Map

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
│   External Boundaries:                                                      │
│   • Wallhaven API v1 (HTTP)                                                 │
│   • Reddit API (HTTP/PRAW)                                                  │
│   • Zerochan (scraping)                                                     │
│   • Filesystem (~/Pictures/wallpapers/)                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Slices

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

---

## After This Milestone

Users will have a working CLI tool that can fetch anime wallpapers from Wallhaven with smart filtering, deduplication, and organization. The tool will be a single binary under 20MB that works on macOS, Linux, and Windows. Future milestones could add Reddit/Zerochan sources, AI tagging, or multi-monitor support.

---

## Verification Contracts

**Code Changes:** Unit tests for all packages >80% coverage. Integration test for full fetch pipeline. Manual verification of binary size and memory usage.

**Integration:** End-to-end test: fetch 10 images with 4k filter, verify all saved correctly, run same fetch again and verify 0 new downloads (dedup working).

**Operational:** Build binaries for all platforms, verify size < 20MB. Run memory profiling during idle state, verify < 10MB.

**UAT:** Manual CLI testing: help readable, flags intuitive, progress visible, output organized as expected. Error messages helpful on bad input.

---

## Requirement Coverage

This milestone covers all Core Requirements from SPEC.md: Multi-Source Fetching (Wallhaven primary, Reddit stretch), Smart Pre-Download Filtering, Deduplication with pHash, Organization by source/tags, and Resource Efficiency targets (single binary, <10MB memory, <50ms startup).
