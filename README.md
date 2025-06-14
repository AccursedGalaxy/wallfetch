<div align="center">

<pre>
██╗    ██╗ █████╗ ██╗     ██╗     ███████╗███████╗████████╗ ██████╗██╗  ██╗
██║    ██║██╔══██╗██║     ██║     ██╔════╝██╔════╝╚══██╔══╝██╔════╝██║  ██║
██║ █╗ ██║███████║██║     ██║     █████╗  █████╗     ██║   ██║     ███████║
██║███╗██║██╔══██║██║     ██║     ██╔══╝  ██╔══╝     ██║   ██║     ██╔══██║
╚███╔███╔╝██║  ██║███████╗███████╗██║     ███████╗   ██║   ╚██████╗██║  ██║
 ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝     ╚══════╝   ╚═╝    ╚═════╝╚═╝  ╚═╝
</pre>

### 🎭 *Professional Wallpaper Management for Linux*

> *Transform your desktop with curated wallpapers from across the web*

---

<p align="center">
  <a href="https://golang.org/">
    <img src="https://img.shields.io/github/go-mod/go-version/AccursedGalaxy/wallfetch?style=for-the-badge&logo=go&logoColor=white&color=00ADD8" alt="Go Version">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/AccursedGalaxy/wallfetch?style=for-the-badge&color=green" alt="License">
  </a>
  <a href="https://github.com/AccursedGalaxy/wallfetch/releases">
    <img src="https://img.shields.io/github/v/release/AccursedGalaxy/wallfetch?style=for-the-badge&logo=github&color=blue" alt="Release">
  </a>
  <a href="https://github.com/AccursedGalaxy/wallfetch/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/AccursedGalaxy/wallfetch/release.yml?style=for-the-badge&logo=github-actions&logoColor=white" alt="Build Status">
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=for-the-badge" alt="Platform">
  <img src="https://img.shields.io/github/downloads/AccursedGalaxy/wallfetch/total?style=for-the-badge&color=success" alt="Downloads">
  <img src="https://img.shields.io/github/stars/AccursedGalaxy/wallfetch?style=for-the-badge&color=yellow" alt="Stars">
</p>

</div>

---

## ✨ **What is WallFetch?**

**WallFetch** is a powerful, modern CLI tool designed for Linux enthusiasts who want to **effortlessly discover, download, and manage stunning wallpapers** from multiple sources. Built with performance and simplicity in mind, it transforms your wallpaper collection workflow into a seamless experience.

<div align="center">

### 🎯 **Perfect for...**
**Desktop Customizers** • **Linux Power Users** • **Aesthetic Enthusiasts** • **Developers**

</div>

---

## 🚀 **Core Features**

<table>
<tr>
<td width="50%">

### 🌐 **Multi-Source Support**
> Fetch from **Wallhaven** and other premium sources
> - 🔍 Advanced search capabilities
> - 🏷️ Category-based filtering
> - ⭐ Quality-rated content

### 🧠 **Smart Intelligence**
> Intelligent duplicate detection & management
> - 🔐 SHA256 checksum verification
> - 📊 Metadata tracking
> - 🗃️ SQLite database backend

### ⚡ **Performance Optimized**
> Lightning-fast concurrent downloads
> - 🔧 Configurable worker pools
> - 📈 Batch processing
> - 💾 Efficient storage management

</td>
<td width="50%">

### 🎨 **Advanced Filtering**
> Precision control over your collection
> - 📐 Resolution & aspect ratio filters
> - 🎭 Category & tag-based sorting
> - 🖼️ Orientation preferences

### 🖥️ **Cross-Platform Ready**
> Seamless experience across platforms
> - 🐧 **Linux** (Primary focus)
> - 🍎 **macOS** compatibility
> - 🪟 **Windows** support

### 🛠️ **Developer Experience**
> Modern CLI with professional features
> - 🐚 Shell completions (Bash/Zsh/Fish)
> - 📋 Rich configuration options
> - 🔄 Easy automation & scripting

</td>
</tr>
</table>

---

<div align="center">

### 🎪 **Feature Highlights**

| Feature | Description | Status |
|---------|-------------|---------|
| **🔄 Concurrent Downloads** | Multi-threaded downloading for maximum speed | ✅ **Active** |
| **🔍 Smart Deduplication** | Never download the same wallpaper twice | ✅ **Active** |
| **📱 Interactive Browser** | Preview wallpapers directly in terminal | ✅ **Active** |
| **🗂️ Collection Management** | Organize, prune, and maintain your library | ✅ **Active** |
| **⚙️ Flexible Configuration** | YAML config with environment variable support | ✅ **Active** |
| **🎯 Precision Filtering** | Fine-grained control over wallpaper selection | ✅ **Active** |

</div>

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
