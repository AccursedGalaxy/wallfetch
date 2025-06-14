package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// PreviewTool represents available preview tools
type PreviewTool int

const (
	PreviewToolKitty PreviewTool = iota // Kitty graphics protocol (highest quality)
	PreviewToolSixel                    // Sixel graphics (high quality)
	PreviewToolChafa                    // Chafa (good quality, widely supported)
	PreviewToolViu                      // Viu (basic quality)
	PreviewToolNone                     // No preview available
)

// PreviewManager handles image previews in terminal
type PreviewManager struct {
	availableTool PreviewTool
	terminalSize  struct {
		width  int
		height int
	}
}

// NewPreviewManager creates a new preview manager
func NewPreviewManager() *PreviewManager {
	pm := &PreviewManager{}
	pm.detectAvailableTools()
	pm.getTerminalSize()
	return pm
}

// detectAvailableTools checks which preview tools are available
func (pm *PreviewManager) detectAvailableTools() {
	// Check for Kitty graphics protocol (highest quality)
	if pm.supportsKittyGraphics() {
		pm.availableTool = PreviewToolKitty
		return
	}

	// Check for Sixel graphics (high quality)
	if pm.supportsSixel() {
		pm.availableTool = PreviewToolSixel
		return
	}

	// Check for chafa (good quality, widely supported)
	if pm.isCommandAvailable("chafa") {
		pm.availableTool = PreviewToolChafa
		return
	}

	// Check for viu as fallback
	if pm.isCommandAvailable("viu") {
		pm.availableTool = PreviewToolViu
		return
	}

	pm.availableTool = PreviewToolNone
}

// isCommandAvailable checks if a command is available in PATH
func (pm *PreviewManager) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// getTerminalSize gets the current terminal dimensions
func (pm *PreviewManager) getTerminalSize() {
	// Try to get terminal size from environment or use defaults
	pm.terminalSize.width = 80
	pm.terminalSize.height = 24

	// Try to get actual terminal size
	if cmd := exec.Command("stty", "size"); cmd.Err == nil {
		if output, err := cmd.Output(); err == nil {
			var h, w int
			if n, err := fmt.Sscanf(string(output), "%d %d", &h, &w); n == 2 && err == nil {
				pm.terminalSize.height = h
				pm.terminalSize.width = w
			}
		}
	}
}

// supportsFullColor checks if terminal supports 24-bit color
func (pm *PreviewManager) supportsFullColor() bool {
	// Check for truecolor support via environment variables
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return true
	}

	// Check TERM variable for modern terminals
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	// Modern terminals that support truecolor
	truecolorTerms := []string{
		"xterm-256color", "screen-256color", "tmux-256color",
		"alacritty", "kitty", "iterm2", "gnome-terminal",
		"konsole", "wezterm", "rio", "ghostty", "foot",
		"st-256color", "xterm-kitty", "contour",
	}

	for _, trueTerm := range truecolorTerms {
		if strings.Contains(term, trueTerm) {
			return true
		}
	}

	// Check TERM_PROGRAM for additional terminal apps
	modernTermPrograms := []string{
		"iTerm", "Hyper", "Tabby", "Terminus", "Warp",
		"WezTerm", "Alacritty", "kitty", "rio", "ghostty",
	}

	for _, program := range modernTermPrograms {
		if strings.Contains(termProgram, program) {
			return true
		}
	}

	return false
}

// supportsAdvancedSymbols checks if terminal supports advanced Unicode symbols
func (pm *PreviewManager) supportsAdvancedSymbols() bool {
	// Check if we're in a modern terminal that supports sextants and advanced Unicode
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	// Terminals with excellent Unicode support
	advancedTerms := []string{
		"xterm-256color", "screen-256color", "tmux-256color",
		"alacritty", "kitty", "iterm2", "gnome-terminal",
		"konsole", "terminal", "wezterm", "rio", "ghostty",
		"foot", "st-256color", "xterm-kitty", "contour",
	}

	for _, advTerm := range advancedTerms {
		if strings.Contains(term, advTerm) {
			return true
		}
	}

	// Check TERM_PROGRAM for modern terminal apps
	modernTermPrograms := []string{
		"iTerm", "Hyper", "Tabby", "Terminus", "Warp",
		"WezTerm", "Alacritty", "kitty", "rio", "ghostty",
	}

	for _, program := range modernTermPrograms {
		if strings.Contains(termProgram, program) {
			return true
		}
	}

	// Check if LANG/LC_ALL indicates UTF-8 support
	lang := os.Getenv("LANG")
	if lang != "" && (strings.Contains(lang, "UTF-8") || strings.Contains(lang, "utf8")) {
		// Also check that we're not in a very basic terminal
		if term != "" && !strings.Contains(term, "linux") && !strings.Contains(term, "vt") {
			return true
		}
	}

	return false
}

// supportsKittyGraphics checks if terminal supports Kitty graphics protocol
func (pm *PreviewManager) supportsKittyGraphics() bool {
	term := os.Getenv("TERM")
	return term == "xterm-kitty" || strings.Contains(os.Getenv("TERM_PROGRAM"), "kitty")
}

// supportsSixel checks if terminal supports Sixel graphics
func (pm *PreviewManager) supportsSixel() bool {
	// Check for terminals known to support Sixel
	term := os.Getenv("TERM")
	sixelTerms := []string{
		"mlterm", "xterm-sixel", "alacritty", "contour",
	}

	for _, sixelTerm := range sixelTerms {
		if strings.Contains(term, sixelTerm) {
			return true
		}
	}

	// Check if chafa can output sixel (it can detect sixel support)
	if pm.isCommandAvailable("chafa") {
		cmd := exec.Command("chafa", "--help")
		if output, err := cmd.Output(); err == nil {
			return strings.Contains(string(output), "sixel")
		}
	}

	return false
}

// CanPreview returns true if image preview is available
func (pm *PreviewManager) CanPreview() bool {
	return pm.availableTool != PreviewToolNone
}

// GetAvailableToolName returns the name of the available preview tool
func (pm *PreviewManager) GetAvailableToolName() string {
	switch pm.availableTool {
	case PreviewToolKitty:
		return "kitty graphics (highest quality)"
	case PreviewToolSixel:
		return "sixel graphics (high quality)"
	case PreviewToolChafa:
		return "chafa (enhanced quality)"
	case PreviewToolViu:
		return "viu"
	default:
		return "none"
	}
}

// PreviewImage displays an image preview in the terminal
func (pm *PreviewManager) PreviewImage(imagePath string) error {
	if pm.availableTool == PreviewToolNone {
		return fmt.Errorf("no image preview tool available (install 'chafa' or 'viu')")
	}

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Errorf("image file not found: %s", imagePath)
	}

	switch pm.availableTool {
	case PreviewToolKitty:
		return pm.previewWithKitty(imagePath)
	case PreviewToolSixel:
		return pm.previewWithSixel(imagePath)
	case PreviewToolChafa:
		return pm.previewWithChafa(imagePath)
	case PreviewToolViu:
		return pm.previewWithViu(imagePath)
	default:
		return fmt.Errorf("no preview tool available")
	}
}

// previewWithKitty uses Kitty graphics protocol for highest quality
func (pm *PreviewManager) previewWithKitty(imagePath string) error {
	// Use chafa with kitty protocol for best quality
	args := []string{
		"--format", "kitty",
		"--size", fmt.Sprintf("%dx%d", pm.terminalSize.width-2, pm.terminalSize.height-6),
	}

	args = append(args, imagePath)

	cmd := exec.Command("chafa", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// previewWithSixel uses Sixel graphics for high quality
func (pm *PreviewManager) previewWithSixel(imagePath string) error {
	// Use chafa with sixel protocol for high quality
	args := []string{
		"--format", "sixels",
		"--size", fmt.Sprintf("%dx%d", pm.terminalSize.width-2, pm.terminalSize.height-6),
	}

	args = append(args, imagePath)

	cmd := exec.Command("chafa", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// previewWithChafa uses chafa to display image preview
func (pm *PreviewManager) previewWithChafa(imagePath string) error {
	// Calculate preview size with better space utilization
	// Reserve less space for better preview quality
	previewWidth := pm.terminalSize.width - 2   // Minimal side margins
	previewHeight := pm.terminalSize.height - 6 // Reserve space for UI elements

	// Ensure minimum viable size but allow larger previews
	if previewWidth < 40 {
		previewWidth = 40
	}
	if previewHeight < 20 {
		previewHeight = 20
	}

	// Build optimized chafa arguments for maximum quality
	args := []string{
		"--size", fmt.Sprintf("%dx%d", previewWidth, previewHeight),
		"--stretch", // Allow stretching for better fit
	}

	// Optimize color settings based on terminal capabilities
	if pm.supportsFullColor() {
		args = append(args, "--colors", "full") // 24-bit color
	} else {
		args = append(args, "--colors", "256") // 256 colors fallback
	}

	// Use the best symbol combination for maximum detail and quality
	if pm.supportsAdvancedSymbols() {
		// Use all symbols for maximum quality and detail
		args = append(args, "--symbols", "all")
	} else {
		// Fallback to a high-quality combination for older terminals
		args = append(args, "--symbols", "block+border+space")
	}

	// Advanced quality settings for crisp images
	args = append(args,
		"--dither", "none", // No dithering for cleaner look
		"--color-space", "rgb", // Use RGB color space for accuracy
		"--optimize", "9", // Maximum optimization level
		"--polite", "off", // Use all available colors
		"--passthrough", "none", // Disable passthrough for consistency
		"--preprocess", "on", // Enable preprocessing for better quality
		"--work", "9", // Maximum work factor for quality
	)

	args = append(args, imagePath)

	cmd := exec.Command("chafa", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// previewWithViu uses viu to display image preview
func (pm *PreviewManager) previewWithViu(imagePath string) error {
	// Calculate preview size with better space utilization
	previewWidth := pm.terminalSize.width - 2   // Minimal side margins
	previewHeight := pm.terminalSize.height - 6 // Reserve space for UI elements

	// Ensure minimum viable size
	if previewWidth < 40 {
		previewWidth = 40
	}
	if previewHeight < 20 {
		previewHeight = 20
	}

	args := []string{
		"-w", strconv.Itoa(previewWidth),
		"-h", strconv.Itoa(previewHeight),
	}

	// Add quality improvements for viu
	if pm.supportsFullColor() {
		args = append(args, "-t") // Use true color if available
	}

	// Additional viu quality options
	args = append(args,
		"-s", // Use blocks for better quality
	)

	args = append(args, imagePath)

	cmd := exec.Command("viu", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DisplayImageInfo shows basic information about an image file
func (pm *PreviewManager) DisplayImageInfo(imagePath string) {
	info, err := os.Stat(imagePath)
	if err != nil {
		fmt.Printf("❌ File not found: %s\n", imagePath)
		return
	}

	fmt.Printf("📁 File: %s\n", filepath.Base(imagePath))
	fmt.Printf("📏 Size: %.2f MB\n", float64(info.Size())/(1024*1024))
	fmt.Printf("📅 Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))

	// Try to get image dimensions if possible
	pm.tryGetImageDimensions(imagePath)
}

// tryGetImageDimensions attempts to get image dimensions using available tools
func (pm *PreviewManager) tryGetImageDimensions(imagePath string) {
	// Try using identify from ImageMagick
	if pm.isCommandAvailable("identify") {
		cmd := exec.Command("identify", "-format", "%wx%h", imagePath)
		if output, err := cmd.Output(); err == nil {
			fmt.Printf("🖼️  Dimensions: %s\n", string(output))
			return
		}
	}

	// Try using file command for basic info
	if pm.isCommandAvailable("file") {
		cmd := exec.Command("file", imagePath)
		if output, err := cmd.Output(); err == nil {
			fmt.Printf("ℹ️  Info: %s", string(output))
		}
	}
}

// InstallInstructions provides installation instructions for preview tools
func (pm *PreviewManager) InstallInstructions() string {
	return `
📦 Install image preview tools:

Ubuntu/Debian:
  sudo apt install chafa
  # or
  sudo apt install viu

Arch Linux:
  sudo pacman -S chafa
  # or
  yay -S viu

macOS:
  brew install chafa
  # or
  brew install viu

After installation, you'll be able to preview images directly in the terminal!
`
}
