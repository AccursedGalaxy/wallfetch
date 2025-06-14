package downloader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AccursedGalaxy/wallfetch/internal/config"
	"github.com/AccursedGalaxy/wallfetch/internal/wallhaven"
)

// WallpaperFilter validates wallpapers against configuration requirements
type WallpaperFilter struct {
	config *config.DefaultOptions
}

// NewWallpaperFilter creates a new filter with the given configuration
func NewWallpaperFilter(cfg *config.DefaultOptions) *WallpaperFilter {
	return &WallpaperFilter{config: cfg}
}

// FilterResult represents the result of filtering a wallpaper
type FilterResult struct {
	Passed bool
	Reason string
}

// ValidateWallpaper checks if a wallpaper meets the configuration requirements
func (f *WallpaperFilter) ValidateWallpaper(wallpaper wallhaven.Wallpaper) FilterResult {
	// Parse resolution
	width, height, err := parseResolution(wallpaper.Resolution)
	if err != nil {
		return FilterResult{false, fmt.Sprintf("invalid resolution format: %s", wallpaper.Resolution)}
	}

	// Check minimum dimensions
	if f.config.MinWidth > 0 && width < f.config.MinWidth {
		return FilterResult{false, fmt.Sprintf("width %d < minimum %d", width, f.config.MinWidth)}
	}
	if f.config.MinHeight > 0 && height < f.config.MinHeight {
		return FilterResult{false, fmt.Sprintf("height %d < minimum %d", height, f.config.MinHeight)}
	}

	// Check maximum dimensions
	if f.config.MaxWidth > 0 && width > f.config.MaxWidth {
		return FilterResult{false, fmt.Sprintf("width %d > maximum %d", width, f.config.MaxWidth)}
	}
	if f.config.MaxHeight > 0 && height > f.config.MaxHeight {
		return FilterResult{false, fmt.Sprintf("height %d > maximum %d", height, f.config.MaxHeight)}
	}

	// Check landscape/portrait requirement
	if f.config.OnlyLandscape && height > width {
		return FilterResult{false, fmt.Sprintf("portrait image (%dx%d) - only landscape allowed", width, height)}
	}

	// Check aspect ratios if specified
	if len(f.config.AspectRatios) > 0 {
		aspectRatioValid := false
		currentRatio := calculateAspectRatio(width, height)

		for _, requiredRatio := range f.config.AspectRatios {
			if isAspectRatioMatch(currentRatio, requiredRatio) {
				aspectRatioValid = true
				break
			}
		}

		if !aspectRatioValid {
			return FilterResult{false, fmt.Sprintf("aspect ratio %.2f doesn't match required ratios: %v", currentRatio, f.config.AspectRatios)}
		}
	}

	return FilterResult{true, ""}
}

// parseResolution parses a resolution string like "1920x1080" into width and height
func parseResolution(resolution string) (int, int, error) {
	parts := strings.Split(resolution, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid resolution format: %s", resolution)
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width: %s", parts[0])
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height: %s", parts[1])
	}

	return width, height, nil
}

// calculateAspectRatio calculates the aspect ratio as a float
func calculateAspectRatio(width, height int) float64 {
	return float64(width) / float64(height)
}

// isAspectRatioMatch checks if the current ratio matches the required ratio within tolerance
func isAspectRatioMatch(current float64, required string) bool {
	// Parse required ratio like "16x9", "21x9", etc.
	parts := strings.Split(required, "x")
	if len(parts) != 2 {
		return false
	}

	reqWidth, err1 := strconv.ParseFloat(parts[0], 64)
	reqHeight, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return false
	}

	requiredRatio := reqWidth / reqHeight
	tolerance := 0.1 // Allow 10% tolerance

	return abs(current-requiredRatio) <= tolerance
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
