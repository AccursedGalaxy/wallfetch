<div align="center">

<pre>
██╗    ██╗ █████╗ ██╗     ██╗     ███████╗███████╗████████╗ ██████╗██╗  ██╗
██║    ██║██╔══██╗██║     ██║     ██╔════╝██╔════╝╚══██╔══╝██╔════╝██║  ██║
██║ █╗ ██║███████║██║     ██║     █████╗  █████╗     ██║   ██║     ███████║
██║███╗██║██╔══██║██║     ██║     ██╔══╝  ██╔══╝     ██║   ██║     ██╔══██║
╚███╔███╔╝██║  ██║███████╗███████╗██║     ███████╗   ██║   ╚██████╗██║  ██║
 ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝     ╚══════╝   ╚═╝    ╚═════╝╚═╝  ╚═╝
</pre>

</div>
<div align="center">
**Professional Wallpaper Management for Linux**

[![Go Version](https://img.shields.io/github/go-mod/go-version/AccursedGalaxy/wallfetch)](https://golang.org/)
[![License](https://img.shields.io/github/license/AccursedGalaxy/wallfetch)](LICENSE)
[![Release](https://img.shields.io/github/v/release/AccursedGalaxy/wallfetch)](https://github.com/AccursedGalaxy/wallfetch/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/AccursedGalaxy/wallfetch/release.yml)](https://github.com/AccursedGalaxy/wallfetch/actions)

</div>

## 🚀 Features

- **Multiple Sources**: Fetch wallpapers from Wallhaven and other sources
- **Smart Filtering**: Filter by resolution, aspect ratio, categories, and more
- **Duplicate Detection**: Intelligent duplicate detection using SHA256 checksums
- **Database Management**: Local SQLite database for metadata and file tracking
- **Concurrent Downloads**: Configurable worker pools for fast downloads
- **Cross-Platform**: Support for Linux, macOS, and Windows
- **Professional CLI**: Modern command-line interface with shell completions

## 📦 Installation

### Quick Install (Recommended)

**One-liner installation for Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/AccursedGalaxy/wallfetch/main/scripts/install.sh | bash
```

### Package Managers

#### Go Install
If you have Go installed:
```bash
go install github.com/AccursedGalaxy/wallfetch/cmd/wallfetch@latest
```

#### Arch Linux (AUR)
```bash
# Using yay
yay -S wallfetch

# Using paru
paru -S wallfetch

# Manual
git clone https://aur.archlinux.org/wallfetch.git
cd wallfetch
makepkg -si
```

#### Ubuntu/Debian
```bash
# Download latest .deb package
wget https://github.com/AccursedGalaxy/wallfetch/releases/latest/download/wallfetch_amd64.deb
sudo dpkg -i wallfetch_amd64.deb
sudo apt-get install -f  # Fix dependencies if needed
```

### Manual Installation

#### Pre-built Binaries
1. Download the latest release for your platform from [GitHub Releases](https://github.com/AccursedGalaxy/wallfetch/releases)
2. Extract and move to your PATH:
```bash
# Linux x64
wget https://github.com/AccursedGalaxy/wallfetch/releases/latest/download/wallfetch-linux-amd64
chmod +x wallfetch-linux-amd64
sudo mv wallfetch-linux-amd64 /usr/local/bin/wallfetch
```

#### Build from Source
```bash
git clone https://github.com/AccursedGalaxy/wallfetch.git
cd wallfetch
make build
sudo make install
```

## 🛠️ Quick Start

### 1. Initialize Configuration
```bash
wallfetch config init
```

### 2. Set Your API Key
Get your free API key from [Wallhaven](https://wallhaven.cc/settings/account) and add it to `~/.config/wallfetch/config.yaml`:
```yaml
wallhaven:
  api_key: "your_api_key_here"
```

### 3. Fetch Wallpapers
```bash
# Fetch 10 wallpapers
wallfetch fetch wallhaven --limit 10

# Fetch with specific filters
wallfetch fetch wallhaven --limit 5 --resolution 1920x1080 --categories general

# Fetch ultrawide wallpapers
wallfetch fetch wallhaven --limit 5 --aspect-ratio 21x9 --only-landscape
```

### 4. Manage Your Collection
```bash
# List downloaded wallpapers
wallfetch list

# Browse wallpapers in your collection
wallfetch browse

# Browse with terminal preview (requires chafa or viu)
wallfetch browse --preview --interactive

# Browse random wallpapers with external viewer
wallfetch browse --random --viewer feh

# Clean up database (remove entries for deleted files)
wallfetch cleanup

# Remove duplicates (with confirmation)
wallfetch dedupe

# Prune old wallpapers, keep only 50 most recent
wallfetch prune --keep 50
```

## ⚙️ Configuration

### Default Configuration Location
- **Linux**: `~/.config/wallfetch/config.yaml`
- **macOS**: `~/.config/wallfetch/config.yaml`
- **Windows**: `%APPDATA%\wallfetch\config.yaml`

### Sample Configuration
```yaml
default_source: "wallhaven"
download_dir: "~/Pictures/Wallpapers"
max_concurrent: 5

wallhaven:
  api_key: "your_api_key_here"

defaults:
  limit: 10
  resolution: "1920x1080"
  sort: "toplist"
  only_landscape: true

filters:
  min_width: 1920
  min_height: 1080
  aspect_ratios: ["16x9", "21x9"]

database:
  path: "~/.local/share/wallfetch/wallfetch.db"
```

### Environment Variables
You can also set configuration via environment variables:
```bash
export WALLHAVEN_API_KEY="your_api_key_here"
export WALLFETCH_DOWNLOAD_DIR="~/Pictures/Wallpapers"
```

## 🎯 Usage Examples

### Basic Usage
```bash
# Fetch 10 random wallpapers
wallfetch fetch wallhaven

# Fetch with specific category
wallfetch fetch wallhaven --categories anime --limit 5

# Fetch top wallpapers
wallfetch fetch wallhaven --sort toplist --limit 10
```

### Advanced Filtering
```bash
# Ultrawide only
wallfetch fetch wallhaven --aspect-ratio 21x9 --only-landscape --limit 5

# High resolution wallpapers
wallfetch fetch wallhaven --min-resolution 2560x1440 --limit 10

# Multiple categories
wallfetch fetch wallhaven --categories general,anime --limit 15
```

### Database Management
```bash
# Show configuration
wallfetch config show

# List all wallpapers with file status
wallfetch list

# Clean up orphaned database entries
wallfetch cleanup --dry-run  # Preview changes
wallfetch cleanup             # Apply cleanup

# Remove duplicates with confirmation
wallfetch dedupe --dry-run    # Preview what would be deleted
wallfetch dedupe              # Actually remove duplicates

# Prune old wallpapers intelligently
wallfetch prune --keep 100 --dry-run  # Preview pruning
wallfetch prune --keep 100            # Keep only 100 most recent

# Delete specific wallpaper
wallfetch delete 12345       # By database ID
wallfetch delete --source-id abc123  # By source ID
```

## 🐚 Shell Completions

Enable shell completions for a better CLI experience:

### Bash
```bash
# Linux
wallfetch completion bash | sudo tee /etc/bash_completion.d/wallfetch

# macOS
wallfetch completion bash > $(brew --prefix)/etc/bash_completion.d/wallfetch
```

### Zsh
```bash
wallfetch completion zsh > "${fpath[1]}/_wallfetch"
```

### Fish
```bash
wallfetch completion fish > ~/.config/fish/completions/wallfetch.fish
```

## 🔧 Development

### Prerequisites
- Go 1.21 or higher
- Make (optional, for using Makefile)

### Building
```bash
# Clone the repository
git clone https://github.com/AccursedGalaxy/wallfetch.git
cd wallfetch

# Build
make build

# Install locally
make install

# Run tests
make test

# Clean build artifacts
make clean
```

### Contributing
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Wallhaven](https://wallhaven.cc/) for providing the wallpaper API
- [Cobra](https://github.com/spf13/cobra) for the excellent CLI framework
- [SQLite](https://www.sqlite.org/) for the embedded database

## 📞 Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/AccursedGalaxy/wallfetch/issues)
- 💡 **Feature Requests**: [GitHub Issues](https://github.com/AccursedGalaxy/wallfetch/issues)
- 📚 **Documentation**: [Wiki](https://github.com/AccursedGalaxy/wallfetch/wiki)

---

<div align="center">
Made with ❤️ for the Linux community
</div>
