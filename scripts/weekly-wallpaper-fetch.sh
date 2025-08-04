#!/usr/bin/env bash

# Weekly Wallpaper Fetch Script
# This script fetches new top wallpapers from selected categories weekly
# Requirements: 3440x1440 or larger, not already in wallpaper folder

set -uo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="$HOME/.cache/weekly-wallpaper-fetch.log"
CONFIG_FILE="$HOME/.config/weekly-wallpaper-fetch.conf"

# Default settings
DEFAULT_CATEGORIES="anime,city"
DEFAULT_LIMIT=5
DEFAULT_RESOLUTION="3440x1440"
DEFAULT_SORT="toplist"
DEFAULT_PURITY="sfw"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
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
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# Load configuration
load_config() {
    if [[ -f "$CONFIG_FILE" ]]; then
        log "INFO" "Loading configuration from $CONFIG_FILE"
        source "$CONFIG_FILE"
    else
        log "INFO" "Using default configuration"
        CATEGORIES="$DEFAULT_CATEGORIES"
        LIMIT="$DEFAULT_LIMIT"
        RESOLUTION="$DEFAULT_RESOLUTION"
        SORT="$DEFAULT_SORT"
        PURITY="$DEFAULT_PURITY"
    fi
}

# Create default configuration file
create_default_config() {
    log "INFO" "Creating default configuration file at $CONFIG_FILE"
    cat > "$CONFIG_FILE" << EOF
# Weekly Wallpaper Fetch Configuration
# Categories to fetch (comma-separated)
CATEGORIES="$DEFAULT_CATEGORIES"

# Number of wallpapers to fetch per category
LIMIT=$DEFAULT_LIMIT

# Minimum resolution required
RESOLUTION="$DEFAULT_RESOLUTION"

# Sort method (toplist, date_added, relevance, random, views, favorites)
SORT="$DEFAULT_SORT"

# Content purity (sfw, sketchy, nsfw)
PURITY="$DEFAULT_PURITY"
EOF
    log "INFO" "Configuration file created. Edit $CONFIG_FILE to customize settings."
}

# Check prerequisites
check_prerequisites() {
    log "INFO" "Checking prerequisites..."
    
    # Check if wallfetch is available
    if ! command -v wallfetch >/dev/null 2>&1; then
        log "ERROR" "wallfetch is not installed or not in PATH"
        exit 1
    fi
    
    # Check if wallfetch is configured
    if ! wallfetch config show >/dev/null 2>&1; then
        log "ERROR" "wallfetch is not configured. Run 'wallfetch config init' first"
        exit 1
    fi
    
    # Create log directory
    mkdir -p "$(dirname "$LOG_FILE")"
    
    log "INFO" "Prerequisites check passed"
}

# Get current wallpaper count
get_current_count() {
    local count
    count=$(wallfetch list --limit 1000 2>/dev/null | wc -l)
    echo $((count - 1)) # Subtract header line
}

# Fetch wallpapers for a specific category
fetch_category() {
    local category="$1"
    local count_before
    local count_after
    
    log "INFO" "Fetching wallpapers for category: $category"
    
    # Get count before fetching
    count_before=$(get_current_count)
    
    # Fetch wallpapers
    wallfetch fetch wallhaven \
        --categories "$category" \
        --limit "$LIMIT" \
        --resolution "$RESOLUTION" \
        --sort "$SORT" \
        --purity "$PURITY"
    
    local fetch_exit_code=$?
    count_after=$(get_current_count)
    local new_count=$((count_after - count_before))
    
    if [[ $new_count -gt 0 ]]; then
        log "INFO" "Successfully fetched $new_count new wallpapers for category: $category"
        return 0
    elif [[ $fetch_exit_code -eq 0 ]]; then
        log "WARN" "No new wallpapers found for category: $category"
        return 0
    else
        log "ERROR" "Failed to fetch wallpapers for category: $category (exit code: $fetch_exit_code)"
        return 1
    fi
}

# Main fetch function
fetch_all_categories() {
    log "INFO" "Starting weekly wallpaper fetch..."
    log "INFO" "Categories: $CATEGORIES"
    log "INFO" "Limit per category: $LIMIT"
    log "INFO" "Minimum resolution: $RESOLUTION"
    log "INFO" "Sort method: $SORT"
    
    local total_before
    local total_after
    local success_count=0
    local fail_count=0
    
    # Get total count before
    total_before=$(get_current_count)
    
    # Split categories and fetch each one
    IFS=',' read -ra CATEGORY_ARRAY <<< "$CATEGORIES"
    for category in "${CATEGORY_ARRAY[@]}"; do
        category=$(echo "$category" | xargs) # Trim whitespace
        
        if [[ -n "$category" ]]; then
            if fetch_category "$category"; then
                ((success_count++))
            else
                ((fail_count++))
            fi
            
            # Small delay between categories to be respectful to the API
            sleep 2
        fi
    done
    
    # Get total count after
    total_after=$(get_current_count)
    local total_new=$((total_after - total_before))
    
    log "INFO" "Weekly fetch completed!"
    log "INFO" "Categories processed successfully: $success_count"
    log "INFO" "Categories failed: $fail_count"
    log "INFO" "Total new wallpapers added: $total_new"
    
    # Send notification if available
    if command -v notify-send >/dev/null 2>&1; then
        notify-send "Weekly Wallpaper Fetch" \
            "Added $total_new new wallpapers\nSuccess: $success_count categories\nFailed: $fail_count categories" \
            -a "Weekly Wallpaper Fetch" \
            -i "preferences-desktop-wallpaper" \
            -t 5000
    fi
}

# Cleanup old wallpapers (optional)
cleanup_old_wallpapers() {
    log "INFO" "Cleaning up old wallpapers..."
    
    # Keep only the 500 most recent wallpapers
    if wallfetch prune --keep 500 --dry-run 2>/dev/null | grep -q "would be deleted"; then
        log "INFO" "Removing old wallpapers..."
        wallfetch prune --keep 500
    else
        log "INFO" "No cleanup needed"
    fi
}

# Show help
show_help() {
    cat << EOF
Weekly Wallpaper Fetch Script

Usage: $0 [OPTIONS]

Options:
    -c, --config          Create default configuration file
    -f, --fetch           Fetch wallpapers (default action)
    -l, --logs            Show recent logs
    -s, --status          Show current wallpaper count
    -h, --help            Show this help message

Configuration:
    Edit $CONFIG_FILE to customize:
    - CATEGORIES: Comma-separated list of categories
    - LIMIT: Number of wallpapers per category
    - RESOLUTION: Minimum resolution required
    - SORT: Sort method (toplist, date_added, etc.)
    - PURITY: Content purity (sfw, sketchy, nsfw)

Examples:
    $0 --config          # Create default config
    $0 --fetch           # Fetch wallpapers
    $0 --status          # Show current count
    $0 --logs            # Show recent logs

EOF
}

# Show logs
show_logs() {
    if [[ -f "$LOG_FILE" ]]; then
        echo "=== Recent Weekly Wallpaper Fetch Logs ==="
        tail -50 "$LOG_FILE"
    else
        echo "No log file found at $LOG_FILE"
    fi
}

# Show status
show_status() {
    local count
    count=$(get_current_count)
    echo "Current wallpaper count: $count"
    
    if [[ -f "$CONFIG_FILE" ]]; then
        echo "Configuration file: $CONFIG_FILE"
        echo "Categories: $CATEGORIES"
        echo "Limit per category: $LIMIT"
        echo "Minimum resolution: $RESOLUTION"
    else
        echo "Using default configuration"
    fi
}

# Main script logic
main() {
    local action="fetch"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--config)
                action="config"
                shift
                ;;
            -f|--fetch)
                action="fetch"
                shift
                ;;
            -l|--logs)
                action="logs"
                shift
                ;;
            -s|--status)
                action="status"
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
        "config")
            create_default_config
            ;;
        "fetch")
            check_prerequisites
            load_config
            fetch_all_categories
            cleanup_old_wallpapers
            # Always exit successfully for fetch operations
            exit 0
            ;;
        "logs")
            show_logs
            ;;
        "status")
            load_config
            show_status
            ;;
    esac
}

# Run main function with all arguments
main "$@" 