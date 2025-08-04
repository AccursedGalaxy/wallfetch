#!/usr/bin/env bash

# Weekly Wallpaper Fetch Automation Installer
# This script installs the weekly wallpaper fetch automation

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SERVICE_NAME="weekly-wallpaper-fetch"
SERVICE_FILE="$SCRIPT_DIR/${SERVICE_NAME}.service"
TIMER_FILE="$SCRIPT_DIR/${SERVICE_NAME}.timer"
SYSTEMD_USER_DIR="$HOME/.config/systemd/user"

# Logging function
log() {
    local level="$1"
    shift
    local message="$*"
    
    case "$level" in
        "INFO")
            echo -e "${GREEN}[INFO]${NC} $message"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
        "DEBUG")
            echo -e "${BLUE}[DEBUG]${NC} $message"
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    log "INFO" "Checking prerequisites..."
    
    # Check if wallfetch is available
    if ! command -v wallfetch >/dev/null 2>&1; then
        log "ERROR" "wallfetch is not installed or not in PATH"
        log "ERROR" "Please install wallfetch first: https://github.com/AccursedGalaxy/wallfetch"
        exit 1
    fi
    
    # Check if wallfetch is configured
    if ! wallfetch config show >/dev/null 2>&1; then
        log "ERROR" "wallfetch is not configured. Run 'wallfetch config init' first"
        exit 1
    fi
    
    # Check if systemd user services are available
    if ! systemctl --user --version >/dev/null 2>&1; then
        log "ERROR" "systemd user services are not available"
        exit 1
    fi
    
    log "INFO" "Prerequisites check passed"
}

# Create systemd user directory
create_systemd_dir() {
    if [[ ! -d "$SYSTEMD_USER_DIR" ]]; then
        log "INFO" "Creating systemd user directory: $SYSTEMD_USER_DIR"
        mkdir -p "$SYSTEMD_USER_DIR"
    fi
}

# Install service and timer files
install_systemd_files() {
    log "INFO" "Installing systemd service and timer files..."
    
    # Copy and customize service file
    if [[ -f "$SERVICE_FILE" ]]; then
        # Replace the hardcoded path with the actual script directory
        sed "s|%h/Projects/wallfetch/scripts/weekly-wallpaper-fetch.sh|$SCRIPT_DIR/weekly-wallpaper-fetch.sh|g" \
            "$SERVICE_FILE" > "$SYSTEMD_USER_DIR/${SERVICE_NAME}.service"
        
        # Also update the working directory
        sed -i "s|WorkingDirectory=%h/Projects/wallfetch|WorkingDirectory=$PROJECT_DIR|g" \
            "$SYSTEMD_USER_DIR/${SERVICE_NAME}.service"
        
        log "INFO" "Installed service file: $SYSTEMD_USER_DIR/${SERVICE_NAME}.service"
    else
        log "ERROR" "Service file not found: $SERVICE_FILE"
        exit 1
    fi
    
    # Copy timer file (no changes needed)
    if [[ -f "$TIMER_FILE" ]]; then
        cp "$TIMER_FILE" "$SYSTEMD_USER_DIR/"
        log "INFO" "Installed timer file: $SYSTEMD_USER_DIR/${SERVICE_NAME}.timer"
    else
        log "ERROR" "Timer file not found: $TIMER_FILE"
        exit 1
    fi
}

# Enable and start the timer
enable_timer() {
    log "INFO" "Enabling and starting weekly wallpaper fetch timer..."
    
    # Reload systemd user daemon
    systemctl --user daemon-reload
    
    # Enable the timer
    systemctl --user enable "${SERVICE_NAME}.timer"
    
    # Start the timer
    systemctl --user start "${SERVICE_NAME}.timer"
    
    log "INFO" "Timer enabled and started successfully"
}

# Create default configuration
create_config() {
    log "INFO" "Creating default configuration..."
    
    # Run the script to create default config
    if "$SCRIPT_DIR/weekly-wallpaper-fetch.sh" --config; then
        log "INFO" "Default configuration created"
    else
        log "WARN" "Failed to create default configuration"
    fi
}

# Test the script
test_script() {
    log "INFO" "Testing the weekly wallpaper fetch script..."
    
    if "$SCRIPT_DIR/weekly-wallpaper-fetch.sh" --status; then
        log "INFO" "Script test passed"
    else
        log "WARN" "Script test failed, but continuing installation"
    fi
}

# Show status
show_status() {
    log "INFO" "=== Weekly Wallpaper Fetch Automation Status ==="
    
    echo ""
    echo "Service Status:"
    systemctl --user status "${SERVICE_NAME}.service" --no-pager -l || true
    
    echo ""
    echo "Timer Status:"
    systemctl --user status "${SERVICE_NAME}.timer" --no-pager -l || true
    
    echo ""
    echo "Next Run:"
    systemctl --user list-timers "${SERVICE_NAME}.timer" --no-pager || true
    
    echo ""
    echo "Configuration:"
    "$SCRIPT_DIR/weekly-wallpaper-fetch.sh" --status || true
}

# Show help
show_help() {
    cat << EOF
Weekly Wallpaper Fetch Automation Installer

Usage: $0 [OPTIONS]

Options:
    install     Install the automation (default)
    uninstall   Remove the automation
    status      Show current status
    test        Test the script
    help        Show this help message

Examples:
    $0 install    # Install the automation
    $0 status     # Show status
    $0 uninstall  # Remove the automation

EOF
}

# Uninstall function
uninstall() {
    log "INFO" "Uninstalling weekly wallpaper fetch automation..."
    
    # Stop and disable timer
    systemctl --user stop "${SERVICE_NAME}.timer" 2>/dev/null || true
    systemctl --user disable "${SERVICE_NAME}.timer" 2>/dev/null || true
    
    # Stop and disable service
    systemctl --user stop "${SERVICE_NAME}.service" 2>/dev/null || true
    systemctl --user disable "${SERVICE_NAME}.service" 2>/dev/null || true
    
    # Remove service files
    rm -f "$SYSTEMD_USER_DIR/${SERVICE_NAME}.service"
    rm -f "$SYSTEMD_USER_DIR/${SERVICE_NAME}.timer"
    
    # Reload systemd
    systemctl --user daemon-reload
    
    log "INFO" "Automation uninstalled successfully"
}

# Main installation function
install() {
    log "INFO" "Installing weekly wallpaper fetch automation..."
    
    check_prerequisites
    create_systemd_dir
    install_systemd_files
    create_config
    test_script
    enable_timer
    
    log "INFO" "Installation completed successfully!"
    log "INFO" ""
    log "INFO" "The automation will:"
    log "INFO" "  - Run every week with a random delay of up to 1 hour"
    log "INFO" "  - Start 5 minutes after boot"
    log "INFO" "  - Fetch top wallpapers from configured categories"
    log "INFO" "  - Ensure minimum resolution of 3440x1440"
    log "INFO" "  - Avoid duplicates automatically"
    log "INFO" ""
    log "INFO" "Configuration file: ~/.config/weekly-wallpaper-fetch.conf"
    log "INFO" "Log file: ~/.cache/weekly-wallpaper-fetch.log"
    log "INFO" ""
    log "INFO" "Use '$0 status' to check the current status"
    log "INFO" "Use '$0 uninstall' to remove the automation"
}

# Main script logic
main() {
    local action="install"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            install)
                action="install"
                shift
                ;;
            uninstall)
                action="uninstall"
                shift
                ;;
            status)
                action="status"
                shift
                ;;
            test)
                action="test"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    case "$action" in
        "install")
            install
            ;;
        "uninstall")
            uninstall
            ;;
        "status")
            show_status
            ;;
        "test")
            test_script
            ;;
    esac
}

# Run main function with all arguments
main "$@" 