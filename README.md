# Wallpaper CLI Tool

A resource-efficient, cross-platform CLI tool for downloading high-quality anime wallpapers from multiple sources with smart filtering, deduplication, and metadata preservation.

![Version](https://img.shields.io/badge/version-v1.2-blue)
![Go](https://img.shields.io/badge/go-1.21+-00ADD8)
![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)
![License](https://img.shields.io/badge/license-MIT-green)

**GitHub:** https://github.com/aurablacklight/wallpaper-cli

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
- **Desktop Integration**: Set wallpapers directly from CLI (`set` command)

---

## 📦 Installation

### Download Pre-built Binary

Download from [GitHub Releases](https://github.com/aurablacklight/wallpaper-cli/releases):

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
git clone https://github.com/aurablacklight/wallpaper-cli.git
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

### Set Command (Desktop Wallpaper)

Set your desktop wallpaper directly from the command line:

```bash
./wallpaper-cli set [path] [flags]
```

#### Set Options

| Flag | Description | Example |
|------|-------------|---------|
| (no flag) | Set specific image file | `set ~/Pictures/wp/image.jpg` |
| `--random` | Set random wallpaper from collection | `set --random` |
| `--latest` | Set most recently downloaded | `set --latest` |
| `--current` | Show currently set wallpaper | `set --current` |

#### Set Examples

```bash
# Set a specific wallpaper
./wallpaper-cli set ~/Pictures/wallpapers/wallhaven/01_abc123.jpg

# Set random wallpaper
./wallpaper-cli set --random

# Set most recently downloaded
./wallpaper-cli set --latest

# Check current wallpaper
./wallpaper-cli set --current
```

#### Platform Support

**macOS**: Uses AppleScript (osascript) - works on Intel and Apple Silicon  
**Linux**: Auto-detects desktop environment (GNOME, KDE, XFCE) with feh/nitrogen fallback  
**Windows**: Uses PowerShell Registry update + rundll32 refresh

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

## 🏗️ Project Structure

```
wallpaper-cli/
├── cmd/                      # CLI commands (cobra)
│   ├── root.go              # Root command & config
│   ├── fetch.go             # Main fetch command
│   ├── config.go            # Config management
│   └── ...
├── internal/
│   ├── sources/             # Source adapters
│   │   ├── wallhaven/      # Wallhaven.cc API
│   │   └── reddit/         # Reddit API
│   ├── download/           # Download manager & progress bar
│   ├── dedup/              # pHash deduplication
│   ├── data/               # SQLite database
│   ├── utils/              # Path & metadata utilities
│   └── validate/           # Input validation
├── .github/                # GitHub templates & workflows
│   ├── ISSUE_TEMPLATE/    # Bug reports & features
│   └── pull_request_template.md
├── CODEOWNERS             # Code ownership rules
├── CONTRIBUTING.md        # Contribution guidelines
└── README.md              # This file
```

---

## 🔒 Repository Protection

This repository uses GitHub's branch protection rules:

- **Pull Request Required:** All changes must go through PR
- **Code Owner Review:** @aurablacklight must approve all changes
- **Linear History:** No merge commits (rebase or squash only)
- **Conversation Resolution:** All review comments must be resolved
- **Admin Enforcement:** Rules apply to everyone including admins

See [CODEOWNERS](CODEOWNERS) and [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

### Quick Steps

1. **Fork** the repository
2. **Create a branch** from `master`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** following our [code standards](CONTRIBUTING.md)
4. **Commit** with clear messages:
   ```bash
   git commit -m "feat: add new sorting option --most-downloaded"
   ```
5. **Push** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
6. **Open a Pull Request** against `master`
   - PR template will auto-populate
   - @aurablacklight will review
   - Address any review comments

### Resources

- [CODEOWNERS](CODEOWNERS) - Who must review changes
- [CONTRIBUTING.md](CONTRIBUTING.md) - Full contribution guidelines
- [Issue Templates](.github/ISSUE_TEMPLATE/) - Bug reports & feature requests

---

## 📝 Changelog

### v1.2 (Current)
- ✅ **Desktop Integration**: Set wallpapers from CLI (`set` command)
  - Set specific images: `set <path>`
  - Random selection: `set --random`
  - Latest download: `set --latest`
  - Current wallpaper: `set --current`
- ✅ Cross-platform wallpaper setting (macOS, Linux, Windows)
- ✅ Config persistence with wallpaper history (last 10)

### v1.1
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

## 📬 Contact & Support

- **Issues:** [GitHub Issues](https://github.com/aurablacklight/wallpaper-cli/issues)
- **Discussions:** Use GitHub Discussions for questions
- **Security:** Report vulnerabilities privately to the maintainer

**Enjoy your wallpapers!** 🎨✨

---

## ⭐ Star History

If you find this tool useful, please consider starring the repo on [GitHub](https://github.com/aurablacklight/wallpaper-cli)!
