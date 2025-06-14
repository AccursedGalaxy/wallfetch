package downloader

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AccursedGalaxy/wallfetch/internal/database"
	"github.com/AccursedGalaxy/wallfetch/internal/wallhaven"
)

// Downloader handles concurrent wallpaper downloading
type Downloader struct {
	downloadDir   string
	maxConcurrent int
	db            *database.DB
	httpClient    *http.Client
}

// NewDownloader creates a new downloader instance
func NewDownloader(downloadDir string, maxConcurrent int, db *database.DB) *Downloader {
	return &Downloader{
		downloadDir:   downloadDir,
		maxConcurrent: maxConcurrent,
		db:            db,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// DownloadResult represents the result of a download operation
type DownloadResult struct {
	Wallpaper wallhaven.Wallpaper
	LocalPath string
	Checksum  string
	Error     error
	Skipped   bool
	Reason    string
}

// DownloadWallpapers downloads multiple wallpapers concurrently
func (d *Downloader) DownloadWallpapers(wallpapers []wallhaven.Wallpaper, filter *WallpaperFilter) ([]DownloadResult, error) {
	// Ensure download directory exists
	if err := os.MkdirAll(d.downloadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download directory: %w", err)
	}

	// Create channels for work and results
	workChan := make(chan wallhaven.Wallpaper, len(wallpapers))
	resultChan := make(chan DownloadResult, len(wallpapers))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < d.maxConcurrent; i++ {
		wg.Add(1)
		go d.worker(&wg, workChan, resultChan, filter)
	}

	// Send work to workers
	for _, wallpaper := range wallpapers {
		workChan <- wallpaper
	}
	close(workChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []DownloadResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}

// worker is a goroutine that processes wallpaper downloads
func (d *Downloader) worker(wg *sync.WaitGroup, workChan <-chan wallhaven.Wallpaper, resultChan chan<- DownloadResult, filter *WallpaperFilter) {
	defer wg.Done()

	for wallpaper := range workChan {
		result := d.downloadWallpaper(wallpaper, filter)
		resultChan <- result
	}
}

// downloadWallpaper downloads a single wallpaper
func (d *Downloader) downloadWallpaper(wallpaper wallhaven.Wallpaper, filter *WallpaperFilter) DownloadResult {
	result := DownloadResult{
		Wallpaper: wallpaper,
	}

	// Apply filtering if provided
	if filter != nil {
		filterResult := filter.ValidateWallpaper(wallpaper)
		if !filterResult.Passed {
			result.Skipped = true
			result.Reason = fmt.Sprintf("Filtered: %s", filterResult.Reason)
			return result
		}
	}

	// Check if already exists by source ID
	exists, err := d.db.ExistsBySourceID("wallhaven", wallpaper.ID)
	if err != nil {
		result.Error = fmt.Errorf("database check failed: %w", err)
		return result
	}
	if exists {
		result.Skipped = true
		result.Reason = "Already exists (source ID)"
		return result
	}

	// Generate local filename
	filename := d.generateFilename(wallpaper)
	localPath := filepath.Join(d.downloadDir, filename)
	result.LocalPath = localPath

	// Download the file
	resp, err := d.httpClient.Get(wallpaper.Path)
	if err != nil {
		result.Error = fmt.Errorf("failed to download: %w", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Errorf("download failed with status %d", resp.StatusCode)
		return result
	}

	// Create temporary file
	tempFile, err := os.CreateTemp(d.downloadDir, "wallfetch_*.tmp")
	if err != nil {
		result.Error = fmt.Errorf("failed to create temp file: %w", err)
		return result
	}
	defer os.Remove(tempFile.Name())

	// Download and compute checksum simultaneously
	hasher := sha256.New()
	writer := io.MultiWriter(tempFile, hasher)

	_, err = io.Copy(writer, resp.Body)
	tempFile.Close()
	if err != nil {
		result.Error = fmt.Errorf("failed to write file: %w", err)
		return result
	}

	// Compute checksum
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	result.Checksum = checksum

	// Check if file with same checksum already exists
	exists, err = d.db.ExistsByChecksum(checksum)
	if err != nil {
		result.Error = fmt.Errorf("checksum database check failed: %w", err)
		return result
	}
	if exists {
		result.Skipped = true
		result.Reason = "Duplicate (same checksum)"
		return result
	}

	// Move temp file to final location
	if err := os.Rename(tempFile.Name(), localPath); err != nil {
		result.Error = fmt.Errorf("failed to move file: %w", err)
		return result
	}

	// Get file info
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to get file info: %w", err)
		return result
	}

	// Save to database
	tags := d.extractTags(wallpaper)
	dbImage := &database.Image{
		Source:     "wallhaven",
		SourceID:   wallpaper.ID,
		URL:        wallpaper.Path,
		LocalPath:  localPath,
		Checksum:   checksum,
		Tags:       tags,
		Resolution: wallpaper.Resolution,
		FileSize:   fileInfo.Size(),
	}

	if err := d.db.InsertImage(dbImage); err != nil {
		// If database insertion fails, clean up the file
		os.Remove(localPath)
		result.Error = fmt.Errorf("failed to save to database: %w", err)
		return result
	}

	return result
}

// generateFilename creates a filename for the wallpaper
func (d *Downloader) generateFilename(wallpaper wallhaven.Wallpaper) string {
	// Extract file extension from the URL
	ext := filepath.Ext(wallpaper.Path)
	if ext == "" {
		// Default to .jpg if no extension found
		ext = ".jpg"
	}

	// Create filename: wallhaven-{id}.{ext}
	return fmt.Sprintf("wallhaven-%s%s", wallpaper.ID, ext)
}

// extractTags extracts tags from wallpaper metadata
func (d *Downloader) extractTags(wallpaper wallhaven.Wallpaper) string {
	var tags []string
	for _, tag := range wallpaper.Tags {
		tags = append(tags, tag.Name)
	}
	return strings.Join(tags, ",")
}
