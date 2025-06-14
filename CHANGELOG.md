# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2025-06-14

### Added

- **Complete Missing Core Commands**
  - Implemented `dedupe` command - Remove duplicate wallpapers with user confirmation and dry-run support
  - Implemented `prune` command - Clean up old wallpapers intelligently, keeping only the most recent ones
  - Implemented `browse` command - View wallpapers with image viewer integration and terminal preview
- **Wallpaper Preview & Display**
  - Terminal image preview using `chafa` or `viu` tools
  - Interactive browsing mode with navigation controls
  - External image viewer integration with auto-detection
  - Detailed image information display with file stats
- **Enhanced User Experience**
  - User confirmation prompts for destructive operations
  - Comprehensive dry-run modes for safe testing
  - Real-time file existence checking
  - Space usage calculations and reporting
  - Random wallpaper browsing mode

### Changed

- Enhanced browse command with multiple viewing modes (interactive, preview, external viewer)
- Improved error handling and user feedback throughout all commands
- Updated version to 1.1.0

### Fixed

- All previously missing command implementations now complete and functional

## [1.0.0] - 2024-01-XX

### Added

- Initial release
- Wallhaven API integration
- Intelligent duplicate detection using SHA256 checksums
- Smart filtering by resolution, aspect ratio, and categories
- SQLite database for metadata storage
- Concurrent downloads with configurable worker pools
- YAML configuration system
- Complete CLI interface with fetch, list, browse, cleanup, delete, prune, and dedupe commands
- Cross-platform support for Linux, macOS, and Windows 