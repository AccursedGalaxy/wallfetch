package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/AccursedGalaxy/wallfetch/internal/database"
	"github.com/AccursedGalaxy/wallfetch/internal/downloader"
	"github.com/AccursedGalaxy/wallfetch/internal/wallhaven"
	"github.com/spf13/cobra"
)

// newFetchCmd creates the fetch command
func (a *App) newFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch [source]",
		Short: "Fetch wallpapers from a source",
		Long:  "Fetch wallpapers from various sources like Wallhaven",
		Args:  cobra.MaximumNArgs(1),
		RunE:  a.runFetch,
	}

	// Add flags
	cmd.Flags().StringP("categories", "c", "", "Categories to fetch (e.g., anime,nature)")
	cmd.Flags().StringP("resolution", "r", "", "Minimum resolution (e.g., 1920x1080)")
	cmd.Flags().StringP("sort", "s", "", "Sort method (date_added, relevance, random, views, favorites, toplist)")
	cmd.Flags().IntP("limit", "l", 0, "Number of wallpapers to fetch")
	cmd.Flags().IntP("page", "p", 1, "Page number to fetch")
	cmd.Flags().StringP("output", "o", "", "Output directory")
	cmd.Flags().String("query", "", "Search query")
	cmd.Flags().String("purity", "", "Content purity (sfw, sketchy, nsfw)")

	return cmd
}

// newListCmd creates the list command
func (a *App) newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List downloaded wallpapers",
		Long:  "List all downloaded wallpapers with metadata",
		RunE:  a.runList,
	}

	cmd.Flags().StringP("source", "s", "", "Filter by source")
	cmd.Flags().IntP("limit", "l", 50, "Limit number of results")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed information")

	return cmd
}

// newBrowseCmd creates the browse command
func (a *App) newBrowseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "browse [source]",
		Short: "Browse wallpapers",
		Long:  "Browse downloaded wallpapers with preview and external viewer options",
		Args:  cobra.MaximumNArgs(1),
		RunE:  a.runBrowse,
	}

	cmd.Flags().IntP("limit", "l", 10, "Number of wallpapers to browse")
	cmd.Flags().BoolP("random", "r", false, "Browse random wallpapers")
	cmd.Flags().BoolP("preview", "p", false, "Show image preview in terminal")
	cmd.Flags().String("viewer", "", "External image viewer command (e.g., 'feh', 'eog', 'open')")
	cmd.Flags().BoolP("interactive", "i", false, "Interactive browsing mode")

	return cmd
}

// newPruneCmd creates the prune command
func (a *App) newPruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune old wallpapers",
		Long:  "Remove old wallpapers keeping only the most recent ones",
		RunE:  a.runPrune,
	}

	cmd.Flags().IntP("keep", "k", 100, "Number of wallpapers to keep")
	cmd.Flags().BoolP("dry-run", "d", false, "Show what would be deleted without actually deleting")

	return cmd
}

// newDedupeCmd creates the dedupe command
func (a *App) newDedupeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dedupe",
		Short: "Remove duplicate wallpapers",
		Long:  "Remove duplicate wallpapers based on checksums",
		RunE:  a.runDedupe,
	}

	cmd.Flags().BoolP("dry-run", "d", false, "Show what would be deleted without actually deleting")

	return cmd
}

// runFetch handles the fetch command
func (a *App) runFetch(cmd *cobra.Command, args []string) error {
	source := a.config.DefaultSource
	if len(args) > 0 {
		source = args[0]
	}

	switch source {
	case "wallhaven":
		return a.runWallhavenFetch(cmd)
	default:
		return fmt.Errorf("unsupported source: %s", source)
	}
}

// runWallhavenFetch handles fetching from Wallhaven
func (a *App) runWallhavenFetch(cmd *cobra.Command) error {
	client := wallhaven.NewClient(a.config.GetWallhavenAPIKey())

	// Get flags
	categories, _ := cmd.Flags().GetString("categories")
	resolution, _ := cmd.Flags().GetString("resolution")
	sort, _ := cmd.Flags().GetString("sort")
	limit, _ := cmd.Flags().GetInt("limit")
	page, _ := cmd.Flags().GetInt("page")
	query, _ := cmd.Flags().GetString("query")
	purity, _ := cmd.Flags().GetString("purity")
	outputDir, _ := cmd.Flags().GetString("output")

	// Use defaults if not specified
	defaults := a.config.Defaults["wallhaven"]
	if categories == "" {
		categories = defaults.Categories
	}
	if resolution == "" {
		resolution = defaults.Resolution
	}
	if sort == "" {
		sort = defaults.Sort
	}
	if limit == 0 {
		limit = defaults.Limit
	}
	if outputDir == "" {
		outputDir = a.config.DownloadDir
	}

	params := wallhaven.SearchParams{
		Query:      query,
		Categories: categories,
		Purity:     purity,
		Sorting:    sort,
		AtLeast:    resolution,
		Page:       page,
	}

	fmt.Printf("Fetching wallpapers from Wallhaven...\n")
	fmt.Printf("  Categories: %s\n", categories)
	fmt.Printf("  Resolution: %s\n", resolution)
	fmt.Printf("  Sort: %s\n", sort)
	fmt.Printf("  Limit: %d\n", limit)
	fmt.Printf("  Page: %d\n", page)
	fmt.Printf("  Output Directory: %s\n", outputDir)

	results, err := client.Search(params)
	if err != nil {
		return fmt.Errorf("failed to search wallpapers: %w", err)
	}

	fmt.Printf("Found %d wallpapers on page %d (total available: %d)\n", len(results.Data), page, results.Meta.Total)

	// Open database
	db, err := database.Open(a.config.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create downloader and filter
	dl := downloader.NewDownloader(outputDir, a.config.MaxConcurrent, db)
	filter := downloader.NewWallpaperFilter(&defaults)

	fmt.Printf("\nStarting download process to get %d wallpapers...\n", limit)

	// Keep track of overall progress
	totalDownloaded := 0
	totalSkipped := 0
	totalFailed := 0
	currentPage := page
	maxPages := results.Meta.LastPage

	for totalDownloaded < limit && currentPage <= maxPages {
		var wallpapers []wallhaven.Wallpaper

		if currentPage == page {
			// Use the results we already have from the first search
			wallpapers = results.Data
		} else {
			// Fetch the next page
			params.Page = currentPage
			fmt.Printf("\nFetching page %d...\n", currentPage)
			pageResults, err := client.Search(params)
			if err != nil {
				return fmt.Errorf("failed to search page %d: %w", currentPage, err)
			}
			wallpapers = pageResults.Data
			fmt.Printf("Found %d wallpapers on page %d\n", len(wallpapers), currentPage)
		}

		if len(wallpapers) == 0 {
			fmt.Printf("No more wallpapers available on page %d\n", currentPage)
			break
		}

		// Download wallpapers from this page
		downloadResults, err := dl.DownloadWallpapers(wallpapers, filter)
		if err != nil {
			return fmt.Errorf("failed to download wallpapers from page %d: %w", currentPage, err)
		}

		// Process results and report progress
		pageDownloaded := 0
		pageSkipped := 0
		pageFailed := 0

		for _, result := range downloadResults {
			if result.Error != nil {
				fmt.Printf("  ❌ %s - Error: %v\n", result.Wallpaper.ID, result.Error)
				pageFailed++
			} else if result.Skipped {
				fmt.Printf("  ⏭️  %s - Skipped: %s\n", result.Wallpaper.ID, result.Reason)
				pageSkipped++
			} else {
				fmt.Printf("  ✅ %s - Downloaded to %s\n", result.Wallpaper.ID, result.LocalPath)
				pageDownloaded++
				totalDownloaded++

				// Stop if we've reached our target
				if totalDownloaded >= limit {
					break
				}
			}
		}

		totalSkipped += pageSkipped
		totalFailed += pageFailed

		fmt.Printf("\nPage %d summary: Downloaded: %d, Skipped: %d, Failed: %d\n",
			currentPage, pageDownloaded, pageSkipped, pageFailed)
		fmt.Printf("Overall progress: %d/%d wallpapers downloaded\n", totalDownloaded, limit)

		// If we've reached our target, stop
		if totalDownloaded >= limit {
			fmt.Printf("\n🎉 Target reached! Downloaded %d wallpapers.\n", totalDownloaded)
			break
		}

		// Move to next page
		currentPage++
	}

	// Final summary
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("FINAL SUMMARY:\n")
	fmt.Printf("  Target: %d\n", limit)
	fmt.Printf("  Downloaded: %d\n", totalDownloaded)
	fmt.Printf("  Skipped: %d\n", totalSkipped)
	fmt.Printf("  Failed: %d\n", totalFailed)
	fmt.Printf("  Pages processed: %d\n", currentPage-page+1)

	if totalDownloaded < limit {
		fmt.Printf("\n⚠️  Could only download %d out of %d requested wallpapers.\n", totalDownloaded, limit)
		fmt.Printf("   This may be due to filters or limited availability.\n")
	}

	return nil
}

// runList handles the list command
func (a *App) runList(cmd *cobra.Command, args []string) error {
	// Get flags
	source, _ := cmd.Flags().GetString("source")
	limit, _ := cmd.Flags().GetInt("limit")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Open database
	db, err := database.Open(a.config.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get images from database
	images, err := db.ListImages(source, limit)
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	if len(images) == 0 {
		fmt.Println("No wallpapers found in database.")
		return nil
	}

	// Get total count
	total, err := db.CountImages()
	if err != nil {
		return fmt.Errorf("failed to count images: %w", err)
	}

	fmt.Printf("Showing %d of %d wallpapers:\n\n", len(images), total)

	// Display images
	for _, img := range images {
		// Check if file exists
		fileExists := true
		if _, err := os.Stat(img.LocalPath); os.IsNotExist(err) {
			fileExists = false
		}

		if verbose {
			fmt.Printf("ID: %d\n", img.ID)
			fmt.Printf("Source: %s (%s)\n", img.Source, img.SourceID)
			fmt.Printf("Resolution: %s\n", img.Resolution)
			fmt.Printf("File Size: %.2f MB\n", float64(img.FileSize)/(1024*1024))
			fmt.Printf("Local Path: %s", img.LocalPath)
			if !fileExists {
				fmt.Printf(" ❌ (FILE MISSING)")
			}
			fmt.Printf("\n")
			fmt.Printf("Tags: %s\n", img.Tags)
			fmt.Printf("Downloaded: %s\n", img.DownloadedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Checksum: %s\n", img.Checksum[:16]+"...")
			fmt.Println(strings.Repeat("-", 50))
		} else {
			status := "✅"
			if !fileExists {
				status = "❌"
			}
			fmt.Printf("%s %-8s | %-12s | %-15s | %s\n",
				status,
				img.SourceID,
				img.Resolution,
				img.DownloadedAt.Format("2006-01-02 15:04"),
				img.LocalPath)
		}
	}

	return nil
}

// runBrowse handles the browse command
func (a *App) runBrowse(cmd *cobra.Command, args []string) error {
	// Get flags
	limit, _ := cmd.Flags().GetInt("limit")
	random, _ := cmd.Flags().GetBool("random")
	preview, _ := cmd.Flags().GetBool("preview")
	viewer, _ := cmd.Flags().GetString("viewer")
	interactive, _ := cmd.Flags().GetBool("interactive")

	source := ""
	if len(args) > 0 {
		source = args[0]
	}

	// Open database
	db, err := database.Open(a.config.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get images from database
	images, err := db.ListImages(source, limit)
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	if len(images) == 0 {
		fmt.Println("No wallpapers found in your collection.")
		fmt.Println("Run 'wallfetch fetch wallhaven' to download some wallpapers first!")
		return nil
	}

	// Filter out missing files
	var validImages []database.Image
	for _, img := range images {
		if _, err := os.Stat(img.LocalPath); err == nil {
			validImages = append(validImages, img)
		}
	}

	if len(validImages) == 0 {
		fmt.Println("No valid wallpaper files found.")
		fmt.Println("Your wallpaper files may have been moved or deleted.")
		fmt.Println("Run 'wallfetch cleanup' to clean up the database.")
		return nil
	}

	images = validImages

	// Shuffle if random mode
	if random {
		// Simple Fisher-Yates shuffle
		for i := len(images) - 1; i > 0; i-- {
			j := i % (i + 1) // Simple pseudo-random
			images[i], images[j] = images[j], images[i]
		}
	}

	// Initialize preview manager
	previewManager := NewPreviewManager()

	// Check if preview is requested but no tools available
	if preview && !previewManager.CanPreview() {
		fmt.Printf("⚠️  Preview mode requested but no preview tools available.\n")
		fmt.Printf("Available tool: %s\n", previewManager.GetAvailableToolName())
		fmt.Print(previewManager.InstallInstructions())
		fmt.Printf("\nContinuing without preview...\n\n")
		preview = false
	}

	// Auto-detect viewer if not specified
	if viewer == "" {
		viewer = a.detectImageViewer()
	}

	fmt.Printf("🖼️  Browsing %d wallpapers", len(images))
	if source != "" {
		fmt.Printf(" from %s", source)
	}
	if random {
		fmt.Printf(" (random order)")
	}
	fmt.Println()

	if preview && previewManager.CanPreview() {
		fmt.Printf("🔍 Using %s for terminal preview\n", previewManager.GetAvailableToolName())
	}
	if viewer != "" {
		fmt.Printf("👁️  External viewer: %s\n", viewer)
	}
	fmt.Println()

	// Interactive browsing mode
	if interactive {
		return a.runInteractiveBrowse(images, previewManager, viewer, preview)
	}

	// Non-interactive mode - show all
	for i, img := range images {
		fmt.Printf("═══ Wallpaper %d/%d ═══\n", i+1, len(images))
		fmt.Printf("ID: %d | Source: %s (%s)\n", img.ID, img.Source, img.SourceID)
		fmt.Printf("Resolution: %s | Size: %.2f MB\n", img.Resolution, float64(img.FileSize)/(1024*1024))
		fmt.Printf("Downloaded: %s\n", img.DownloadedAt.Format("2006-01-02 15:04:05"))
		if img.Tags != "" {
			fmt.Printf("Tags: %s\n", img.Tags)
		}
		fmt.Printf("File: %s\n", img.LocalPath)

		if preview && previewManager.CanPreview() {
			fmt.Println("\n🖼️  Preview:")
			if err := previewManager.PreviewImage(img.LocalPath); err != nil {
				fmt.Printf("Failed to preview: %v\n", err)
			}
		}

		if viewer != "" {
			fmt.Printf("\n🚀 Open with %s? [y/N/q]: ", viewer)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err == nil {
				response = strings.ToLower(strings.TrimSpace(response))
				if response == "q" || response == "quit" {
					break
				}
				if response == "y" || response == "yes" {
					if err := a.openWithViewer(img.LocalPath, viewer); err != nil {
						fmt.Printf("Failed to open viewer: %v\n", err)
					}
				}
			}
		}

		if i < len(images)-1 {
			fmt.Printf("\n⏭️  Press Enter for next, 'q' to quit: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(input)) == "q" {
				break
			}
		}
		fmt.Println()
	}

	fmt.Printf("✅ Finished browsing %d wallpapers\n", len(images))
	return nil
}

// runInteractiveBrowse runs the interactive browsing mode
func (a *App) runInteractiveBrowse(images []database.Image, previewManager *PreviewManager, viewer string, preview bool) error {
	fmt.Println("🎮 Interactive Browsing Mode")
	fmt.Println("Commands: [n]ext, [p]rev, [o]pen, [i]nfo, [q]uit, [h]elp")
	fmt.Println()

	currentIndex := 0

	for {
		img := images[currentIndex]

		// Clear screen (simple version)
		fmt.Print("\033[H\033[2J")

		fmt.Printf("═══ Wallpaper %d/%d ═══\n", currentIndex+1, len(images))

		if preview && previewManager.CanPreview() {
			if err := previewManager.PreviewImage(img.LocalPath); err != nil {
				fmt.Printf("Preview failed: %v\n", err)
			}
		}

		fmt.Printf("\nID: %d | %s (%s) | %s\n", img.ID, img.Source, img.SourceID, img.Resolution)
		fmt.Printf("File: %s\n", filepath.Base(img.LocalPath))

		fmt.Printf("\n[n]ext [p]rev [o]pen [i]nfo [q]uit [h]elp > ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		command := strings.ToLower(strings.TrimSpace(input))

		switch command {
		case "n", "next", "":
			currentIndex = (currentIndex + 1) % len(images)
		case "p", "prev":
			currentIndex = (currentIndex - 1 + len(images)) % len(images)
		case "o", "open":
			if viewer != "" {
				if err := a.openWithViewer(img.LocalPath, viewer); err != nil {
					fmt.Printf("Failed to open: %v\nPress Enter to continue...", err)
					reader.ReadString('\n')
				}
			} else {
				fmt.Printf("No viewer configured. Press Enter to continue...")
				reader.ReadString('\n')
			}
		case "i", "info":
			fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
			fmt.Printf("Detailed Information:\n")
			fmt.Printf("ID: %d\n", img.ID)
			fmt.Printf("Source: %s (%s)\n", img.Source, img.SourceID)
			fmt.Printf("URL: %s\n", img.URL)
			fmt.Printf("Resolution: %s\n", img.Resolution)
			fmt.Printf("File Size: %.2f MB\n", float64(img.FileSize)/(1024*1024))
			fmt.Printf("Downloaded: %s\n", img.DownloadedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Checksum: %s\n", img.Checksum)
			if img.Tags != "" {
				fmt.Printf("Tags: %s\n", img.Tags)
			}
			fmt.Printf("Full Path: %s\n", img.LocalPath)
			previewManager.DisplayImageInfo(img.LocalPath)
			fmt.Printf("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "h", "help":
			fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
			fmt.Printf("Interactive Browsing Commands:\n")
			fmt.Printf("  n, next, Enter  - Next wallpaper\n")
			fmt.Printf("  p, prev         - Previous wallpaper\n")
			fmt.Printf("  o, open         - Open with external viewer\n")
			fmt.Printf("  i, info         - Show detailed information\n")
			fmt.Printf("  q, quit         - Quit browsing\n")
			fmt.Printf("  h, help         - Show this help\n")
			fmt.Printf("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "q", "quit":
			return nil
		default:
			fmt.Printf("Unknown command '%s'. Type 'h' for help.\nPress Enter to continue...", command)
			reader.ReadString('\n')
		}
	}
}

// detectImageViewer attempts to detect available image viewers
func (a *App) detectImageViewer() string {
	viewers := []string{
		"feh",       // Popular Linux image viewer
		"eog",       // GNOME Eye of GNOME
		"gwenview",  // KDE image viewer
		"ristretto", // XFCE image viewer
		"sxiv",      // Simple X Image Viewer
		"nsxiv",     // Neo Simple X Image Viewer
		"qiv",       // Quick Image Viewer
		"xviewer",   // X-Apps image viewer
		"open",      // macOS default
		"xdg-open",  // Linux generic opener
	}

	for _, viewer := range viewers {
		if _, err := exec.LookPath(viewer); err == nil {
			return viewer
		}
	}

	return ""
}

// openWithViewer opens an image with the specified viewer
func (a *App) openWithViewer(imagePath, viewer string) error {
	cmd := exec.Command(viewer, imagePath)

	// For some viewers, we want to run them in background
	backgroundViewers := []string{"feh", "eog", "gwenview", "ristretto", "sxiv", "nsxiv", "qiv", "xviewer"}
	runInBackground := false
	for _, bgViewer := range backgroundViewers {
		if strings.Contains(viewer, bgViewer) {
			runInBackground = true
			break
		}
	}

	if runInBackground {
		return cmd.Start()
	}

	return cmd.Run()
}

// runPrune handles the prune command
func (a *App) runPrune(cmd *cobra.Command, args []string) error {
	keep, _ := cmd.Flags().GetInt("keep")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Open database
	db, err := database.Open(a.config.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get current count
	totalCount, err := db.CountImages()
	if err != nil {
		return fmt.Errorf("failed to count images: %w", err)
	}

	if totalCount <= keep {
		fmt.Printf("✅ Current collection has %d wallpapers (keep target: %d)\n", totalCount, keep)
		fmt.Println("No pruning needed!")
		return nil
	}

	toDelete := totalCount - keep
	fmt.Printf("Collection Management:\n")
	fmt.Printf("  Current wallpapers: %d\n", totalCount)
	fmt.Printf("  Target to keep: %d\n", keep)
	fmt.Printf("  Will delete: %d oldest wallpapers\n", toDelete)

	if dryRun {
		// Get the paths that would be deleted
		pathsToDelete, err := db.DeleteOldImages(keep)
		if err != nil {
			return fmt.Errorf("failed to get old images list: %w", err)
		}

		fmt.Printf("\n🔍 DRY RUN - Would delete %d old wallpapers:\n", len(pathsToDelete))
		for i, path := range pathsToDelete {
			if i < 10 { // Show first 10
				fmt.Printf("  - %s\n", filepath.Base(path))
			} else if i == 10 {
				fmt.Printf("  ... and %d more\n", len(pathsToDelete)-10)
				break
			}
		}
		fmt.Printf("\nRun without --dry-run to actually delete these files\n")
		return nil
	}

	// Ask for confirmation
	fmt.Printf("\n⚠️  This will permanently delete %d old wallpapers from both database and disk.\n", toDelete)
	fmt.Printf("The %d most recently downloaded wallpapers will be kept.\n", keep)
	fmt.Print("Do you want to continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		fmt.Println("❌ Operation cancelled")
		return nil
	}

	// Perform the pruning
	fmt.Printf("\n🗑️  Pruning old wallpapers...\n")

	pathsToDelete, err := db.DeleteOldImages(keep)
	if err != nil {
		return fmt.Errorf("failed to delete old images: %w", err)
	}

	// Delete the actual files
	deleted := 0
	failed := 0
	var totalSize int64

	for _, path := range pathsToDelete {
		// Get file size before deleting
		if fileInfo, err := os.Stat(path); err == nil {
			totalSize += fileInfo.Size()
		}

		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("  ⚠️  Failed to delete file %s: %v\n", filepath.Base(path), err)
				failed++
			} else {
				fmt.Printf("  ✅ %s (file was already missing)\n", filepath.Base(path))
				deleted++
			}
		} else {
			fmt.Printf("  ✅ %s\n", filepath.Base(path))
			deleted++
		}
	}

	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("PRUNE SUMMARY:\n")
	fmt.Printf("  Files deleted: %d\n", deleted)
	if failed > 0 {
		fmt.Printf("  Failed deletions: %d\n", failed)
	}
	fmt.Printf("  Space freed: %.2f MB\n", float64(totalSize)/(1024*1024))
	fmt.Printf("  Remaining wallpapers: %d\n", keep)

	return nil
}

// runDedupe handles the dedupe command
func (a *App) runDedupe(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Open database
	db, err := database.Open(a.config.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Find duplicates
	duplicateGroups, err := db.FindDuplicates()
	if err != nil {
		return fmt.Errorf("failed to find duplicates: %w", err)
	}

	if len(duplicateGroups) == 0 {
		fmt.Println("✅ No duplicates found!")
		return nil
	}

	fmt.Printf("Found %d groups of duplicate wallpapers:\n\n", len(duplicateGroups))

	totalDuplicates := 0
	var toDelete []database.Image

	for i, group := range duplicateGroups {
		fmt.Printf("Duplicate Group %d (%d images):\n", i+1, len(group))
		fmt.Printf("Checksum: %s\n", group[0].Checksum[:16]+"...")

		// Sort by download date (keep the first one in the group)
		for j, img := range group {
			// Check if file still exists
			fileExists := true
			if _, err := os.Stat(img.LocalPath); os.IsNotExist(err) {
				fileExists = false
			}

			status := "✅"
			if !fileExists {
				status = "❌"
			}

			keepMarker := ""
			if j == 0 {
				keepMarker = " (KEEP)"
			} else {
				keepMarker = " (DELETE)"
				toDelete = append(toDelete, img)
				totalDuplicates++
			}

			fmt.Printf("  %s ID: %-4d | %s | %s | %s%s\n",
				status,
				img.ID,
				img.SourceID,
				img.Resolution,
				filepath.Base(img.LocalPath),
				keepMarker)
		}
		fmt.Println()
	}

	if totalDuplicates == 0 {
		fmt.Println("✅ No duplicates to remove!")
		return nil
	}

	fmt.Printf("Summary: Found %d duplicate files to remove\n", totalDuplicates)

	if dryRun {
		fmt.Printf("\n🔍 DRY RUN - Would delete %d duplicate files:\n", totalDuplicates)
		for _, img := range toDelete {
			fmt.Printf("  - ID %d: %s\n", img.ID, filepath.Base(img.LocalPath))
		}
		fmt.Printf("\nRun without --dry-run to actually remove duplicates\n")
		return nil
	}

	// Ask for confirmation
	fmt.Printf("\n⚠️  This will permanently delete %d duplicate files from both database and disk.\n", totalDuplicates)
	fmt.Print("Do you want to continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		fmt.Println("❌ Operation cancelled")
		return nil
	}

	// Delete duplicates
	deleted := 0
	failed := 0

	fmt.Printf("\n🗑️  Removing duplicates...\n")
	for _, img := range toDelete {
		// Delete from database
		_, err := db.DeleteImage(img.ID)
		if err != nil {
			fmt.Printf("  ❌ Failed to delete ID %d from database: %v\n", img.ID, err)
			failed++
			continue
		}

		// Delete file from disk
		if err := os.Remove(img.LocalPath); err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("  ⚠️  Deleted from database but failed to delete file %s: %v\n", img.LocalPath, err)
			} else {
				fmt.Printf("  ✅ ID %d: %s (file was already missing)\n", img.ID, filepath.Base(img.LocalPath))
			}
		} else {
			fmt.Printf("  ✅ ID %d: %s\n", img.ID, filepath.Base(img.LocalPath))
		}
		deleted++
	}

	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("DEDUPE SUMMARY:\n")
	fmt.Printf("  Successfully deleted: %d\n", deleted)
	if failed > 0 {
		fmt.Printf("  Failed: %d\n", failed)
	}
	fmt.Printf("  Space potentially freed: calculating...\n")

	// Calculate space saved (rough estimate)
	var totalSize int64
	for _, img := range toDelete {
		if deleted > failed { // only count if we successfully deleted most
			totalSize += img.FileSize
		}
	}

	if totalSize > 0 {
		fmt.Printf("  Approximate space freed: %.2f MB\n", float64(totalSize)/(1024*1024))
	}

	return nil
}

// newDeleteCmd creates the delete command
func (a *App) newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [wallpaper-id]",
		Short: "Delete a wallpaper",
		Long:  "Delete a wallpaper by ID or source ID",
		RunE:  a.runDelete,
	}

	cmd.Flags().Bool("file", false, "Also delete the file from disk")
	cmd.Flags().StringP("source-id", "s", "", "Delete by source ID (e.g., wallhaven ID)")

	return cmd
}

// newCleanupCmd creates the cleanup command
func (a *App) newCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up missing files",
		Long:  "Remove database entries for wallpapers that no longer exist on disk",
		RunE:  a.runCleanup,
	}

	cmd.Flags().BoolP("dry-run", "d", false, "Show what would be cleaned without actually cleaning")

	return cmd
}

// runDelete handles the delete command
func (a *App) runDelete(cmd *cobra.Command, args []string) error {
	deleteFile, _ := cmd.Flags().GetBool("file")
	sourceID, _ := cmd.Flags().GetString("source-id")

	if len(args) == 0 && sourceID == "" {
		return fmt.Errorf("must provide either wallpaper ID or --source-id")
	}

	// Open database
	db, err := database.Open(a.config.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	var localPath string

	if sourceID != "" {
		// Delete by source ID
		localPath, err = db.DeleteImageBySourceID("wallhaven", sourceID)
		if err != nil {
			return fmt.Errorf("failed to delete wallpaper by source ID: %w", err)
		}
		fmt.Printf("Deleted wallpaper %s from database\n", sourceID)
	} else {
		// Delete by ID
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid wallpaper ID: %s", args[0])
		}

		localPath, err = db.DeleteImage(id)
		if err != nil {
			return fmt.Errorf("failed to delete wallpaper: %w", err)
		}
		fmt.Printf("Deleted wallpaper ID %d from database\n", id)
	}

	// Delete file if requested
	if deleteFile {
		if err := os.Remove(localPath); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to delete file %s: %w", localPath, err)
			}
			fmt.Printf("File %s was already missing\n", localPath)
		} else {
			fmt.Printf("Deleted file: %s\n", localPath)
		}
	} else {
		fmt.Printf("File preserved: %s\n", localPath)
		fmt.Printf("Use --file flag to also delete the file from disk\n")
	}

	return nil
}

// runCleanup handles the cleanup command
func (a *App) runCleanup(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Open database
	db, err := database.Open(a.config.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if dryRun {
		// Get all images and check which ones are missing
		images, err := db.ListImages("", 0)
		if err != nil {
			return fmt.Errorf("failed to list images: %w", err)
		}

		var missingFiles []string
		for _, img := range images {
			if _, err := os.Stat(img.LocalPath); os.IsNotExist(err) {
				missingFiles = append(missingFiles, img.LocalPath)
			}
		}

		if len(missingFiles) == 0 {
			fmt.Println("No missing files found")
		} else {
			fmt.Printf("Would clean up %d missing files:\n", len(missingFiles))
			for _, path := range missingFiles {
				fmt.Printf("  - %s\n", path)
			}
			fmt.Printf("\nRun without --dry-run to actually clean up\n")
		}
	} else {
		// Actually clean up
		deletedPaths, err := db.CleanupMissingFiles()
		if err != nil {
			return fmt.Errorf("failed to cleanup missing files: %w", err)
		}

		if len(deletedPaths) == 0 {
			fmt.Println("No missing files found")
		} else {
			fmt.Printf("Cleaned up %d missing files:\n", len(deletedPaths))
			for _, path := range deletedPaths {
				fmt.Printf("  - %s\n", path)
			}
		}
	}

	return nil
}

// newUpdateCmd creates the update command
func (a *App) newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update wallfetch to the latest version",
		Long:  "Check for and install the latest version of wallfetch",
		RunE:  a.runUpdate,
	}

	cmd.Flags().BoolP("check", "c", false, "Only check for updates without installing")
	cmd.Flags().BoolP("force", "f", false, "Force update even if already on latest version")

	return cmd
}

// runUpdate handles the update command
func (a *App) runUpdate(cmd *cobra.Command, args []string) error {
	checkOnly, _ := cmd.Flags().GetBool("check")
	force, _ := cmd.Flags().GetBool("force")

	// Get current version
	currentVersion := a.rootCmd.Version

	fmt.Printf("Current version: %s\n", currentVersion)

	// Get latest release info from GitHub
	fmt.Println("Checking for updates...")

	latestVersion, downloadURL, err := a.getLatestReleaseInfo()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	fmt.Printf("Latest version: %s\n", latestVersion)

	// Normalize versions for comparison (remove "v" prefix and "-dirty" suffix)
	currentVersionClean := strings.TrimPrefix(currentVersion, "v")
	currentVersionClean = strings.Split(currentVersionClean, "-")[0]
	latestVersionClean := strings.TrimPrefix(latestVersion, "v")

	// Compare versions
	if currentVersionClean == latestVersionClean && !force {
		fmt.Println("✅ You are already on the latest version!")
		return nil
	}

	if checkOnly {
		if currentVersionClean != latestVersionClean {
			fmt.Printf("🆕 Update available: %s → %s\n", currentVersion, latestVersion)
			fmt.Println("Run 'wallfetch update' to install the latest version")
		}
		return nil
	}

	// Find current installation path
	execPath, err := a.findWallfetchPath()
	if err != nil {
		return fmt.Errorf("failed to find wallfetch installation: %w", err)
	}

	fmt.Printf("Found wallfetch at: %s\n", execPath)

	// Ask for confirmation
	fmt.Printf("\n⚠️  This will update wallfetch from %s to %s\n", currentVersion, latestVersion)
	fmt.Printf("Installation path: %s\n", execPath)
	fmt.Print("Do you want to continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		fmt.Println("❌ Update cancelled")
		return nil
	}

	// Download new binary
	fmt.Printf("\n📥 Downloading wallfetch %s...\n", latestVersion)

	tempFile, err := a.downloadBinary(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(tempFile)

	// Make it executable
	if err := os.Chmod(tempFile, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// Backup current binary
	backupPath := execPath + ".backup"
	fmt.Printf("📦 Creating backup at %s...\n", backupPath)

	// Check if we need sudo
	needSudo := false
	if err := os.Rename(execPath, backupPath); err != nil {
		if os.IsPermission(err) {
			needSudo = true
		} else {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	} else {
		// Restore the original for now
		os.Rename(backupPath, execPath)
	}

	// Replace binary
	fmt.Println("🔄 Installing new version...")

	if needSudo {
		fmt.Println("🔐 Administrator privileges required...")

		// Use sudo to backup and replace
		backupCmd := exec.Command("sudo", "mv", execPath, backupPath)
		if err := backupCmd.Run(); err != nil {
			return fmt.Errorf("failed to create backup with sudo: %w", err)
		}

		installCmd := exec.Command("sudo", "cp", tempFile, execPath)
		if err := installCmd.Run(); err != nil {
			// Try to restore backup
			restoreCmd := exec.Command("sudo", "mv", backupPath, execPath)
			restoreCmd.Run()
			return fmt.Errorf("failed to install update: %w", err)
		}

		// Remove backup
		removeCmd := exec.Command("sudo", "rm", "-f", backupPath)
		removeCmd.Run()
	} else {
		// Direct file operations
		if err := os.Rename(execPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		if err := copyFile(tempFile, execPath); err != nil {
			// Try to restore backup
			os.Rename(backupPath, execPath)
			return fmt.Errorf("failed to install update: %w", err)
		}

		// Set permissions
		if err := os.Chmod(execPath, 0755); err != nil {
			fmt.Printf("⚠️  Warning: failed to set permissions: %v\n", err)
		}

		// Remove backup
		os.Remove(backupPath)
	}

	fmt.Printf("\n✅ Successfully updated to %s!\n", latestVersion)
	fmt.Println("🎉 Please run 'wallfetch --version' to verify the update")

	return nil
}

// getLatestReleaseInfo fetches the latest release information from GitHub
func (a *App) getLatestReleaseInfo() (version string, downloadURL string, err error) {
	repo := "AccursedGalaxy/wallfetch"
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	version = release.TagName

	// Find the right asset for current platform
	osName := runtime.GOOS
	archName := runtime.GOARCH
	if archName == "amd64" {
		archName = "amd64"
	}

	assetName := fmt.Sprintf("wallfetch-%s-%s", osName, archName)

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return "", "", fmt.Errorf("no binary found for %s-%s", osName, archName)
	}

	return version, downloadURL, nil
}

// findWallfetchPath finds the current wallfetch binary path
func (a *App) findWallfetchPath() (string, error) {
	// First, try to get the path of the current executable
	execPath, err := os.Executable()
	if err == nil {
		// Resolve any symlinks
		realPath, err := filepath.EvalSymlinks(execPath)
		if err == nil {
			return realPath, nil
		}
		return execPath, nil
	}

	// Fallback: use which command
	path, err := exec.LookPath("wallfetch")
	if err != nil {
		return "", fmt.Errorf("wallfetch not found in PATH")
	}

	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path, nil
	}

	return realPath, nil
}

// downloadBinary downloads the binary from the given URL
func (a *App) downloadBinary(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file
	tempFile, err := os.CreateTemp("", "wallfetch-update-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Download with progress
	written, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	fmt.Printf("✅ Downloaded %.2f MB\n", float64(written)/(1024*1024))

	return tempFile.Name(), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
