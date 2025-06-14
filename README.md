# WallFetch

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/AccursedGalaxy/wallfetch)](https://goreportcard.com/report/github.com/AccursedGalaxy/wallfetch)

A powerful CLI tool to fetch and manage wallpapers from various sources with intelligent duplicate detection and local library management.

## ✨ Features

- **Multi-Source Support**: Fetch wallpapers from popular sources like Wallhaven
- **Smart Duplicate Detection**: Prevents downloading duplicates using SHA256 checksums and source IDs
- **Local Library Management**: Organize, browse, and prune your wallpaper collection
- **Flexible Filtering**: Sort by trending, categories, tags, and more
- **Automated Scheduling**: Set up cron jobs for periodic wallpaper fetching
- **Lightweight Database**: SQLite-based metadata storage for fast operations
- **Cross-Platform**: Works on Linux, macOS, and Windows

## 🚀 Installation

### From Source

1. **Prerequisites**: Ensure you have Go 1.21+ installed
2. **Clone and build**:
   ```bash
   git clone https://github.com/AccursedGalaxy/wallfetch.git
   cd wallfetch
   make build
   ```
3. **Install globally** (optional):
   ```bash
   sudo make install
   ```

### Quick Setup

1. **Initialize configuration**:
   ```bash
   wallfetch config init
   ```
2. **Edit your config** at `~/.config/wallfetch/config.yaml`:
   ```yaml
   api_keys:
     wallhaven: "your_api_key_here"  # Get from https://wallhaven.cc/settings/account
   ```
3. **Test the setup**:
   ```bash
   wallfetch config show
   ```

## 📖 Usage

### Basic Commands

#### Fetch Wallpapers

```bash
# Fetch trending anime wallpapers from Wallhaven
wallfetch fetch wallhaven --sort toplist --categories anime --limit 5

# Fetch with specific resolution  
wallfetch fetch wallhaven --resolution 1920x1080 --limit 10

# Fetch from multiple categories
wallfetch fetch wallhaven --categories "anime,nature" --sort toplist
```

#### Manage Your Collection

```bash
# List all downloaded wallpapers
wallfetch list

# List with detailed information
wallfetch list --verbose

# List with filtering
wallfetch list --source wallhaven --limit 20

# Browse wallpapers (planned feature)
wallfetch browse

# Browse specific source (planned feature)
wallfetch browse wallhaven --limit 10
```

#### Collection Maintenance

```bash
# Keep only the 100 most recent wallpapers
wallfetch prune --keep 100

# Remove duplicates (dry run)
wallfetch dedupe --dry-run

# Remove duplicates
wallfetch dedupe
```

### Advanced Usage

#### Batch Operations

```bash
# Fetch from multiple sources
wallfetch fetch wallhaven --categories anime --limit 5
wallfetch fetch unsplash --query "nature" --limit 5

# Set up automated fetching (add to crontab)
# Fetch 10 new wallpapers daily at 9 AM
0 9 * * * /usr/local/bin/wallfetch fetch wallhaven --categories anime --limit 10
```

#### Custom Download Directory

```bash
# Set custom download directory
wallfetch fetch wallhaven --output ~/Pictures/Wallpapers --categories nature

# Use config file
wallfetch --config ~/.config/wallfetch/config.yaml fetch wallhaven
```

## 🗂️ Database Schema

WallFetch stores metadata in a SQLite database (`~/.local/share/wallfetch/wallpapers.db`):

```sql
CREATE TABLE images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    url TEXT NOT NULL,
    local_path TEXT NOT NULL,
    checksum TEXT NOT NULL,
    tags TEXT,
    resolution TEXT,
    file_size INTEGER,
    downloaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 🔄 Duplicate Detection

WallFetch uses a two-layer approach to prevent duplicates:

1. **Source ID Check**: Prevents re-downloading the same image from the same source
2. **SHA256 Checksum**: Detects identical images from different sources or URLs

Process:

```
Download → Compute SHA256 → Check Database → Save if Unique
```

## 📊 Current Status

### ✅ Working Features
- **Wallhaven API Integration**: Full search and download support
- **Database Management**: SQLite-based metadata storage
- **Duplicate Detection**: Both source ID and checksum-based
- **Concurrent Downloads**: Configurable worker pool
- **Configuration Management**: YAML-based config system
- **CLI Interface**: Full command-line interface with help

### 🚧 Planned Features
- **Browse Command**: Open wallpapers in image viewer
- **Prune Command**: Remove old wallpapers
- **Dedupe Command**: Remove duplicate files
- **Additional Sources**: Unsplash, Reddit support

## 🛠️ Supported Sources

| Source    | Status     | Features                               |
|-----------|------------|----------------------------------------|
| Wallhaven | ✅ Active  | Categories, tags, sorting, resolutions |
| Unsplash  | 🚧 Planned | Collections, search, user galleries    |
| Reddit    | 🚧 Planned | Subreddit scraping, top posts          |

## ⚙️ Configuration File

Create `~/.config/wallfetch/config.yaml`:

```yaml
# Default settings
default_source: wallhaven
download_dir: ~/Pictures/Wallpapers
max_concurrent: 5

# API Keys
api_keys:
  wallhaven: "your_api_key_here"

# Default fetch options
defaults:
  wallhaven:
    categories: "anime,nature"
    resolution: "1920x1080"
    sort: "toplist"
    limit: 10

# Database settings
database:
  path: "~/.local/share/wallfetch/wallpapers.db"
  auto_vacuum: true
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Wallhaven](https://wallhaven.cc/) for their excellent API
- The Go community for amazing libraries and tools

## 📞 Support

- 🐛 Found a bug? [Open an issue](https://github.com/AccursedGalaxy/wallfetch/issues)
- 💡 Have a feature request? [Start a discussion](https://github.com/AccursedGalaxy/wallfetch/discussions)
- 📧 Need help? Check out the [documentation](https://github.com/AccursedGalaxy/wallfetch/wiki)

---

⭐ If you find WallFetch useful, please consider giving it a star on GitHub!
