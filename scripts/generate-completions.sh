#!/bin/bash

# WallFetch Shell Completions Generator
# This script generates shell completions for packaging

set -e

BINARY="${1:-./wallfetch}"
OUTPUT_DIR="${2:-./completions}"

if [ ! -f "$BINARY" ]; then
    echo "Error: Binary not found at $BINARY"
    echo "Usage: $0 [binary_path] [output_dir]"
    exit 1
fi

echo "Generating shell completions..."

# Create output directories
mkdir -p "$OUTPUT_DIR"/{bash,zsh,fish}

# Generate completions
echo "  → Bash completion"
"$BINARY" completion bash > "$OUTPUT_DIR/bash/wallfetch"

echo "  → Zsh completion"
"$BINARY" completion zsh > "$OUTPUT_DIR/zsh/_wallfetch"

echo "  → Fish completion"
"$BINARY" completion fish > "$OUTPUT_DIR/fish/wallfetch.fish"

echo "Shell completions generated successfully in $OUTPUT_DIR/"
echo
echo "To install:"
echo "  Bash: sudo cp $OUTPUT_DIR/bash/wallfetch /etc/bash_completion.d/"
echo "  Zsh:  cp $OUTPUT_DIR/zsh/_wallfetch \"\${fpath[1]}/_wallfetch\""
echo "  Fish: cp $OUTPUT_DIR/fish/wallfetch.fish ~/.config/fish/completions/" 