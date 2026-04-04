# Wallpaper CLI Tool

A resource-efficient, cross-platform CLI tool for downloading high-quality anime wallpapers from multiple sources with smart filtering, deduplication, and metadata preservation.

![Version](https://img.shields.io/badge/version-v1.1-blue)
![Go](https://img.shields.io/badge/go-1.21+-00ADD8)
![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)

---

## ✨ Features

- **Multiple Sources**: Wallhaven.cc (primary) + Reddit r/Animewallpaper
- **Smart Filtering**: Resolution, aspect ratio, tags, time periods
- **Popularity Sorting**: Top, favorites, most viewed, hot, latest
- **Deduplication**: Perceptual hashing (pHash) with SQLite tracking
- **File Metadata**: Original source URLs saved in extended attributes
- **Progress Bar**: Visual download progress with speed and ETA
- **Organization**: By source, date, or tags
- **Cross-Platform**: macOS, Linux, Windows

---

## 📦 Installation

### Download Pre-built Binary

```bash
# macOS (Apple Silicon)
./wallpaper-cli-darwin-arm64 --version

# macOS (Intel)
./wallpaper-cli-darwin-amd64 --version

# Linux
./wallpaper-cli-linux-amd64 --version

# Windows
./wallpaper-cli-windows-amd64.exe --version
```

### Build from Source

```bash
git clone https://github.com/user/wallpaper-cli
cd wallpaper-cli
go build -o wallpaper-cli .
```

---

## 🚀 Quick Start

### Download Top Wallpapers

```bash
# Top 10 anime wallpapers from Wallhaven (4K)
./wallpaper-cli fetch --resolution 4k --tags "anime" --limit 10

# Top 10 from Reddit (hot posts)
./wallpaper-cli fetch --source reddit --sort=hot --limit 10

# Top 20 from both sources
./wallpaper-cli fetch --source all --limit 20
```

### Sorting Options

```bash
# Most favorited of all time
./wallpaper-cli fetch --favorites --all-time --limit 10

# Top this week
./wallpaper-cli fetch --popular --week --limit 10

# Most viewed this month
./wallpaper-cli fetch --views --month --limit 10

# Latest uploads
./wallpaper-cli fetch --latest --limit 10
```

### Dry Run (Preview Before Download)

```bash
./wallpaper-cli fetch --limit 5 --dry-run
```

---

## 📋 Command Reference

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Config file path | `~/.config/wallpaper-cli/config.json` |
| `-v, --version` | Show version | - |
| `-h, --help` | Show help | - |

### Fetch Command

```bash
./wallpaper-cli fetch [flags]
```

#### Source Selection

| Flag | Description | Default |
|------|-------------|---------|
| `--source` | Source: wallhaven, reddit, all | wallhaven |

#### Filtering

| Flag | Description | Example |
|------|-------------|---------|
| `--resolution` | Target resolution | 1080p, 1440p, 4k, 8k, 1920x1080 |
| `--aspect-ratio` | Aspect ratio filter | 16:9, 21:9, 32:9 |
| `--tags` | Comma-separated tags | "anime,landscape,night" |
| `--anime` | Anime category only | - |

#### Sorting

| Flag | Description |
|------|-------------|
| `--popular` or `--sort=top` | Top rated |
| `--favorites` | Most favorited |
| `--views` | Most viewed |
| `--latest` | Newest uploads |
| `--sort=hot` | Hot/trending (Reddit) |
| `--sort=new` | Newest (Reddit) |
| `--sort=random` | Random (default) |

#### Time Periods

| Flag | Description |
|------|-------------|
| `--day` or `--time=day` | Last 24 hours |
| `--week` or `--time=week` | Last 7 days |
| `--month` or `--time=month` | Last 30 days |
| `--year` or `--time=year` | Last year |
| `--all-time` or `--time=all` | All time |

#### Output Options

| Flag | Description | Default |
|------|-------------|---------|
| `--limit` | Maximum downloads | 10 |
| `--output` | Output directory | ~/Pictures/wallpapers/ |
| `--organize-by` | Organization: source, date, tags | source |
| `--format` | Preferred format: webp, jpg, png, original | original |
| `--concurrent` | Concurrent downloads | 5 |
| `--dedup` | Enable deduplication | true |
| `--dry-run` | Preview without downloading | - |

---

## 🎯 Examples

### Download Top Weekly Wallpapers

```bash
./wallpaper-cli fetch --popular --week --limit 20 --resolution 4k
```

### Download from Reddit Only

```bash
./wallpaper-cli fetch --source reddit --sort=top --month --limit 15
```

### Custom Output with Date Organization

```bash
./wallpaper-cli fetch --output ~/Wallpapers --organize-by date --limit 50
```

### Multi-Source with Tags

```bash
./wallpaper-cli fetch --source all --tags "landscape,night" --favorites --limit 30
```

---

## 📁 File Organization

### By Source (Default)

```
~/Pictures/wallpapers/
├── wallhaven/
│   ├── 01_abc123_3840x2160.jpg
│   └── 02_def456_3840x2160.png
└── reddit/
    ├── 01_1rnel2q_1440x2560_Title.jpg
    └── 02_1rsypsj_2560x1440_Title.png
```

### By Date

```
~/Pictures/wallpapers/
├── 2024/
│   └── 04/
│       ├── 01_abc123_3840x2160.jpg
│       └── 02_def456_3840x2160.png
```

### Filename Format

**Wallhaven:** `RANK_ID_RESOLUTION.ext`  
Example: `01_zpqr1w_3840x2160.png`

**Reddit:** `RANK_ID_RESOLUTION_TITLE.ext`  
Example: `01_1rnel2q_1440x2560_Cyrene [Honkai- Star Rail].jpg`

---

## 🔍 Accessing Source URLs (File Metadata)

Original source URLs are saved as **extended file attributes** (metadata).

### macOS/Linux: Command Line

```bash
# Get Reddit URL from file
xattr -p user.reddit_url "filename.jpg"

# Get Wallhaven URL from file
xattr -p user.wallhaven_url "filename.png"

# List all metadata
xattr -l filename.jpg
```

### Open Original Source

```bash
# Copy URL to clipboard
xattr -p user.reddit_url "01_1rnel2q_1440x2560_Cyrene.jpg" | pbcopy

# Open URL in browser
open $(xattr -p user.reddit_url "01_1rnel2q_1440x2560_Cyrene.jpg")
```

---

## ⚙️ Configuration

### Config Commands

```bash
# Initialize default config
./wallpaper-cli config init

# View current config
./wallpaper-cli config list

# Get specific value
./wallpaper-cli config get default_resolution

# Set value
./wallpaper-cli config set default_resolution 4k
./wallpaper-cli config set output_directory ~/Wallpapers
```

### Default Config Location

- **macOS/Linux:** `~/.config/wallpaper-cli/config.json`
- **Windows:** `%APPDATA%\wallpaper-cli\config.json`

### Config Schema

```json
{
  "default_source": "wallhaven",
  "default_resolution": "4k",
  "output_directory": "/Users/name/Pictures/wallpapers",
  "organization": "source",
  "format": "original",
  "dedup": true,
  "dedup_threshold": 10,
  "concurrent_downloads": 5,
  "sources": {
    "wallhaven": { "enabled": true },
    "reddit": {
      "enabled": true,
      "subreddits": ["Animewallpaper"]
    }
  }
}
```

---

## 🛠️ Development

### Build

```bash
make build
```

### Cross-Platform Build

```bash
make build-all
```

### Test

```bash
make test
```

---

## 📊 System Requirements

- **Binary Size:** < 15 MB
- **Memory:** < 10 MB at idle
- **Go Version:** 1.21+
- **Platforms:** macOS (Intel/Apple Silicon), Linux, Windows

---

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📝 Changelog

### v1.1 (Current)
- ✅ Progress bar with visual display
- ✅ Reddit source adapter
- ✅ Popularity sorting (top, favorites, views, hot)
- ✅ Time period filtering (day, week, month, year, all)
- ✅ File metadata (original source URLs in extended attributes)
- ✅ Multi-source fetching (--source all)
- ✅ Ranking numbers in filenames

### v1.0
- ✅ Core download functionality
- ✅ Wallhaven API integration
- ✅ Concurrent downloads
- ✅ Perceptual hash deduplication
- ✅ SQLite database tracking
- ✅ Organization modes (source/date/tags)
- ✅ Cross-platform builds

---

## 📄 License

MIT License - see LICENSE file for details

---

## 🙏 Acknowledgments

- [Wallhaven.cc](https://wallhaven.cc) for the amazing wallpaper API
- Reddit r/Animewallpaper community
- [schollz/progressbar](https://github.com/schollz/progressbar) for the visual progress bar
- [spf13/cobra](https://github.com/spf13/cobra) for CLI framework

---

## 📬 Contact

For issues, feature requests, or contributions, please open a GitHub issue.

**Enjoy your wallpapers!** 🎨✨
