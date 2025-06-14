package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// PreviewTool represents available preview tools
type PreviewTool int

const (
	PreviewToolChafa PreviewTool = iota
	PreviewToolViu
	PreviewToolNone
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
	// Check for chafa first (generally better quality)
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

// CanPreview returns true if image preview is available
func (pm *PreviewManager) CanPreview() bool {
	return pm.availableTool != PreviewToolNone
}

// GetAvailableToolName returns the name of the available preview tool
func (pm *PreviewManager) GetAvailableToolName() string {
	switch pm.availableTool {
	case PreviewToolChafa:
		return "chafa"
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
	case PreviewToolChafa:
		return pm.previewWithChafa(imagePath)
	case PreviewToolViu:
		return pm.previewWithViu(imagePath)
	default:
		return fmt.Errorf("no preview tool available")
	}
}

// previewWithChafa uses chafa to display image preview
func (pm *PreviewManager) previewWithChafa(imagePath string) error {
	// Calculate preview size (leave some space for text)
	previewWidth := pm.terminalSize.width - 4
	previewHeight := pm.terminalSize.height - 10

	if previewWidth < 20 {
		previewWidth = 20
	}
	if previewHeight < 10 {
		previewHeight = 10
	}

	cmd := exec.Command("chafa",
		"--size", fmt.Sprintf("%dx%d", previewWidth, previewHeight),
		"--symbols", "block",
		"--colors", "256",
		imagePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// previewWithViu uses viu to display image preview
func (pm *PreviewManager) previewWithViu(imagePath string) error {
	// Calculate preview size
	previewWidth := pm.terminalSize.width - 4
	previewHeight := pm.terminalSize.height - 10

	if previewWidth < 20 {
		previewWidth = 20
	}
	if previewHeight < 10 {
		previewHeight = 10
	}

	cmd := exec.Command("viu",
		"-w", strconv.Itoa(previewWidth),
		"-h", strconv.Itoa(previewHeight),
		imagePath)

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
