# wallpaper-cli

A resource-efficient CLI tool for downloading anime wallpapers from multiple sources with smart filtering, deduplication, and cross-platform desktop integration.

![Go](https://img.shields.io/badge/go-1.21+-00ADD8)
![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)

**GitHub:** https://github.com/aurablacklight/wallpaper-cli

---

## Features

- **Multiple Sources**: Wallhaven.cc + Reddit r/Animewallpaper
- **Smart Filtering**: Resolution, aspect ratio, tags, time periods
- **Popularity Sorting**: Top, favorites, most viewed, hot, latest
- **Deduplication**: Perceptual hashing (pHash) with SQLite tracking
- **Progress Bar**: Visual download progress with speed and ETA
- **Cross-Platform**: macOS, Linux, Windows wallpaper setting
- **Collections**: Favorites, ratings, playlists
- **Metadata Export**: JSON export for integration with other tools

---

## Installation

### Build from Source

```bash
git clone https://github.com/aurablacklight/wallpaper-cli.git
cd wallpaper-cli
go build -o wallpaper-cli .
```

---

## Quick Start

```bash
# Top 10 anime wallpapers from Wallhaven (4K)
wallpaper-cli fetch --resolution 4k --tags "anime" --limit 10

# Top 10 from Reddit
wallpaper-cli fetch --source reddit --sort=hot --limit 10

# Set a random wallpaper from your collection
wallpaper-cli set --random

# List what you've downloaded
wallpaper-cli list --source wallhaven --since 7d
```

---

## Commands

### fetch

Download wallpapers with filtering and sorting.

```bash
wallpaper-cli fetch [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--source` | wallhaven, reddit, all | wallhaven |
| `--resolution` | 1080p, 1440p, 4k, 8k, WxH | - |
| `--aspect-ratio` | 16:9, 21:9, etc. | - |
| `--tags` | Comma-separated tags | - |
| `--popular` | Sort by top rated | - |
| `--favorites` | Sort by most favorited | - |
| `--views` | Sort by most viewed | - |
| `--latest` | Sort by newest | - |
| `--day/--week/--month/--year/--all-time` | Time period | - |
| `--limit` | Max downloads | 10 |
| `--output` | Output directory | ~/Pictures/wallpapers/ |
| `--dry-run` | Preview without downloading | - |

### set

Set your desktop wallpaper.

```bash
wallpaper-cli set [path]         # Set specific image
wallpaper-cli set --random       # Random from collection
wallpaper-cli set --latest       # Most recently downloaded
wallpaper-cli set --current      # Show current wallpaper
```

**Platform support:** macOS (AppleScript), Linux (GNOME/KDE/XFCE/feh), Windows (PowerShell)

### list

Query downloaded wallpapers from the database.

```bash
wallpaper-cli list                          # All wallpapers
wallpaper-cli list --source wallhaven       # Filter by source
wallpaper-cli list --since 7d              # Recent downloads
wallpaper-cli list --json                   # Machine-readable
wallpaper-cli list --path-only             # For piping
```

### export

Export wallpaper metadata to JSON.

```bash
wallpaper-cli export                                          # To stdout
wallpaper-cli export --output metadata.json                   # To file
wallpaper-cli export --source wallhaven --since 7d           # Filtered
```

### collections

Manage favorites, ratings, and playlists.

```bash
wallpaper-cli favorite <path>                    # Toggle favorite
wallpaper-cli rate <path> <1-5>                  # Rate wallpaper
wallpaper-cli playlist create <name>             # Create playlist
wallpaper-cli playlist add <name> <path>         # Add to playlist
wallpaper-cli playlist list                      # List playlists
```

### config

Manage configuration.

```bash
wallpaper-cli config init                        # Create default config
wallpaper-cli config list                        # Show all settings
wallpaper-cli config get default_resolution      # Get value
wallpaper-cli config set default_resolution 4k   # Set value
```

### stats

Show collection statistics.

```bash
wallpaper-cli stats
```

---

## Configuration

Config lives at `~/.config/wallpaper-cli/config.json`:

```json
{
  "default_source": "wallhaven",
  "default_resolution": "4k",
  "output_directory": "~/Pictures/wallpapers",
  "dedup": true,
  "concurrent_downloads": 5,
  "sources": {
    "wallhaven": { "enabled": true },
    "reddit": { "enabled": true, "subreddits": ["Animewallpaper"] }
  }
}
```

---

## File Organization

```
~/Pictures/wallpapers/
├── wallhaven/
│   └── 01_abc123_3840x2160.jpg
└── reddit/
    └── 01_1rnel2q_1440x2560_Title.jpg
```

**Filename format:** `RANK_ID_RESOLUTION.ext`

Source URLs are preserved in extended file attributes (`xattr -l file.jpg`).

---

## Development

```bash
make build          # Build binary
make build-all      # Cross-platform builds
make test           # Run tests
```

---

## License

MIT
