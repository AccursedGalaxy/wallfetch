package cli

import (
	"fmt"
	"os"
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
		Long:  "Browse downloaded wallpapers in the default image viewer",
		Args:  cobra.MaximumNArgs(1),
		RunE:  a.runBrowse,
	}

	cmd.Flags().IntP("limit", "l", 10, "Number of wallpapers to browse")
	cmd.Flags().BoolP("random", "r", false, "Browse random wallpapers")

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
	fmt.Println("Browse command not yet implemented")
	// TODO: Implement browsing functionality
	return nil
}

// runPrune handles the prune command
func (a *App) runPrune(cmd *cobra.Command, args []string) error {
	fmt.Println("Prune command not yet implemented")
	// TODO: Implement pruning functionality
	return nil
}

// runDedupe handles the dedupe command
func (a *App) runDedupe(cmd *cobra.Command, args []string) error {
	fmt.Println("Dedupe command not yet implemented")
	// TODO: Implement deduplication functionality
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
