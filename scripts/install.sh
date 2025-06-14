#!/bin/bash

# WallFetch Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/AccursedGalaxy/wallfetch/main/scripts/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="AccursedGalaxy/wallfetch"
BINARY_NAME="wallfetch"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/wallfetch"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    armv7*) ARCH="armv7" ;;
    *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

# Functions
print_banner() {
    echo -e "${BLUE}"
    echo "██╗    ██╗ █████╗ ██╗     ██╗     ███████╗███████╗████████╗ ██████╗██╗  ██╗"
    echo "██║    ██║██╔══██╗██║     ██║     ██╔════╝██╔════╝╚══██╔══╝██╔════╝██║  ██║"
    echo "██║ █╗ ██║███████║██║     ██║     █████╗  █████╗     ██║   ██║     ███████║"
    echo "██║███╗██║██╔══██║██║     ██║     ██╔══╝  ██╔══╝     ██║   ██║     ██╔══██║"
    echo "╚███╔███╔╝██║  ██║███████╗███████╗██║     ███████╗   ██║   ╚██████╗██║  ██║"
    echo " ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝     ╚══════╝   ╚═╝    ╚═════╝╚═╝  ╚═╝"
    echo -e "${NC}"
    echo -e "${GREEN}Professional Wallpaper Management for Linux${NC}"
    echo
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed."
        exit 1
    fi
    
    if ! command -v tar >/dev/null 2>&1; then
        log_error "tar is required but not installed."
        exit 1
    fi
    
    log_success "All dependencies satisfied"
}

get_latest_release() {
    log_info "Getting latest release information..."
    
    local api_url="https://api.github.com/repos/$REPO/releases/latest"
    local latest_release
    
    latest_release=$(curl -fsSL "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$latest_release" ]; then
        log_error "Failed to get latest release information"
        exit 1
    fi
    
    echo "$latest_release"
}

download_binary() {
    local version="$1"
    local download_url="https://github.com/$REPO/releases/download/$version/${BINARY_NAME}-${OS}-${ARCH}"
    local temp_file="/tmp/${BINARY_NAME}"
    
    log_info "Downloading WallFetch $version for $OS-$ARCH..."
    
    if curl -fsSL "$download_url" -o "$temp_file"; then
        chmod +x "$temp_file"
        log_success "Downloaded successfully"
        echo "$temp_file"
    else
        log_error "Failed to download binary from $download_url"
        exit 1
    fi
}

install_binary() {
    local temp_file="$1"
    
    log_info "Installing WallFetch to $INSTALL_DIR..."
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        cp "$temp_file" "$INSTALL_DIR/$BINARY_NAME"
    else
        if command -v sudo >/dev/null 2>&1; then
            sudo cp "$temp_file" "$INSTALL_DIR/$BINARY_NAME"
        else
            log_error "Need write access to $INSTALL_DIR but sudo is not available"
            log_info "Please run: cp $temp_file $INSTALL_DIR/$BINARY_NAME as root"
            exit 1
        fi
    fi
    
    # Cleanup
    rm -f "$temp_file"
    
    log_success "WallFetch installed successfully"
}

setup_config() {
    log_info "Setting up configuration..."
    
    # Create config directory
    mkdir -p "$CONFIG_DIR"
    
    # Initialize config if it doesn't exist
    if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
        log_info "Creating default configuration..."
        "$INSTALL_DIR/$BINARY_NAME" config init
        log_success "Configuration initialized at $CONFIG_DIR/config.yaml"
        log_warning "Don't forget to set your Wallhaven API key!"
    else
        log_info "Configuration already exists at $CONFIG_DIR/config.yaml"
    fi
}

verify_installation() {
    log_info "Verifying installation..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version
        version=$("$BINARY_NAME" --version 2>/dev/null | head -1 || echo "unknown")
        log_success "WallFetch is installed and working: $version"
        return 0
    else
        log_error "Installation verification failed"
        log_info "You may need to restart your shell or add $INSTALL_DIR to your PATH"
        return 1
    fi
}

show_next_steps() {
    echo
    echo -e "${GREEN}🎉 Installation Complete!${NC}"
    echo
    echo -e "${YELLOW}Next Steps:${NC}"
    echo "1. Get your Wallhaven API key: https://wallhaven.cc/settings/account"
    echo "2. Edit config: $CONFIG_DIR/config.yaml"
    echo "3. Set your API key in the config file"
    echo "4. Test: wallfetch config show"
    echo "5. Fetch wallpapers: wallfetch fetch wallhaven --limit 5"
    echo
    echo -e "${BLUE}Documentation:${NC} https://github.com/$REPO"
    echo -e "${BLUE}Issues:${NC} https://github.com/$REPO/issues"
    echo
}

main() {
    print_banner
    
    # Check if running as root
    if [ "$EUID" -eq 0 ]; then
        log_warning "Running as root. Consider using a regular user account."
    fi
    
    check_dependencies
    
    local version
    version=$(get_latest_release)
    log_info "Latest version: $version"
    
    local temp_file
    temp_file=$(download_binary "$version")
    
    install_binary "$temp_file"
    setup_config
    
    if verify_installation; then
        show_next_steps
    fi
}

# Handle Ctrl+C
trap 'echo -e "\n${RED}Installation cancelled${NC}"; exit 1' INT

# Run main function
main "$@" 