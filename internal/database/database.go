package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// Image represents an image record in the database
type Image struct {
	ID           int       `json:"id"`
	Source       string    `json:"source"`
	SourceID     string    `json:"source_id"`
	URL          string    `json:"url"`
	LocalPath    string    `json:"local_path"`
	Checksum     string    `json:"checksum"`
	Tags         string    `json:"tags"`
	Resolution   string    `json:"resolution"`
	FileSize     int64     `json:"file_size"`
	DownloadedAt time.Time `json:"downloaded_at"`
}

// Open opens the database connection
func Open(dbPath string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize the database schema
	if err := db.init(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// init initializes the database schema
func (db *DB) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source TEXT NOT NULL,
		source_id TEXT NOT NULL,
		url TEXT NOT NULL,
		local_path TEXT NOT NULL,
		checksum TEXT NOT NULL,
		tags TEXT,
		resolution TEXT,
		file_size INTEGER,
		downloaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source, source_id),
		UNIQUE(checksum)
	);

	CREATE INDEX IF NOT EXISTS idx_source ON images(source);
	CREATE INDEX IF NOT EXISTS idx_checksum ON images(checksum);
	CREATE INDEX IF NOT EXISTS idx_downloaded_at ON images(downloaded_at);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// InsertImage inserts a new image record
func (db *DB) InsertImage(img *Image) error {
	query := `
	INSERT INTO images (source, source_id, url, local_path, checksum, tags, resolution, file_size)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, img.Source, img.SourceID, img.URL, img.LocalPath,
		img.Checksum, img.Tags, img.Resolution, img.FileSize)
	return err
}

// ExistsBySourceID checks if an image exists by source and source ID
func (db *DB) ExistsBySourceID(source, sourceID string) (bool, error) {
	query := `SELECT COUNT(*) FROM images WHERE source = ? AND source_id = ?`
	var count int
	err := db.conn.QueryRow(query, source, sourceID).Scan(&count)
	return count > 0, err
}

// ExistsByChecksum checks if an image exists by checksum
func (db *DB) ExistsByChecksum(checksum string) (bool, error) {
	query := `SELECT COUNT(*) FROM images WHERE checksum = ?`
	var count int
	err := db.conn.QueryRow(query, checksum).Scan(&count)
	return count > 0, err
}

// ListImages lists images with optional filtering
func (db *DB) ListImages(source string, limit int) ([]Image, error) {
	query := `SELECT id, source, source_id, url, local_path, checksum, tags, resolution, file_size, downloaded_at FROM images`
	args := []interface{}{}

	if source != "" {
		query += ` WHERE source = ?`
		args = append(args, source)
	}

	query += ` ORDER BY downloaded_at DESC`

	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		err := rows.Scan(&img.ID, &img.Source, &img.SourceID, &img.URL, &img.LocalPath,
			&img.Checksum, &img.Tags, &img.Resolution, &img.FileSize, &img.DownloadedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, rows.Err()
}

// CountImages returns the total number of images
func (db *DB) CountImages() (int, error) {
	query := `SELECT COUNT(*) FROM images`
	var count int
	err := db.conn.QueryRow(query).Scan(&count)
	return count, err
}

// DeleteOldImages deletes old images keeping only the most recent ones
func (db *DB) DeleteOldImages(keepCount int) ([]string, error) {
	// First, get the paths of images that will be deleted
	query := `
	SELECT local_path FROM images 
	ORDER BY downloaded_at DESC 
	LIMIT -1 OFFSET ?
	`
	rows, err := db.conn.Query(query, keepCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}

	// Delete the old records
	deleteQuery := `
	DELETE FROM images 
	WHERE id NOT IN (
		SELECT id FROM images 
		ORDER BY downloaded_at DESC 
		LIMIT ?
	)
	`
	_, err = db.conn.Exec(deleteQuery, keepCount)
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// FindDuplicates finds duplicate images by checksum
func (db *DB) FindDuplicates() ([][]Image, error) {
	query := `
	SELECT id, source, source_id, url, local_path, checksum, tags, resolution, file_size, downloaded_at
	FROM images 
	WHERE checksum IN (
		SELECT checksum FROM images 
		GROUP BY checksum 
		HAVING COUNT(*) > 1
	)
	ORDER BY checksum, downloaded_at
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	duplicateGroups := make(map[string][]Image)
	for rows.Next() {
		var img Image
		err := rows.Scan(&img.ID, &img.Source, &img.SourceID, &img.URL, &img.LocalPath,
			&img.Checksum, &img.Tags, &img.Resolution, &img.FileSize, &img.DownloadedAt)
		if err != nil {
			return nil, err
		}
		duplicateGroups[img.Checksum] = append(duplicateGroups[img.Checksum], img)
	}

	var result [][]Image
	for _, group := range duplicateGroups {
		result = append(result, group)
	}

	return result, rows.Err()
}

// DeleteImage deletes an image by ID and returns the local path
func (db *DB) DeleteImage(id int) (string, error) {
	// First get the local path
	var localPath string
	query := `SELECT local_path FROM images WHERE id = ?`
	err := db.conn.QueryRow(query, id).Scan(&localPath)
	if err != nil {
		return "", err
	}

	// Delete from database
	deleteQuery := `DELETE FROM images WHERE id = ?`
	_, err = db.conn.Exec(deleteQuery, id)
	if err != nil {
		return "", err
	}

	return localPath, nil
}

// DeleteImageBySourceID deletes an image by source and source_id
func (db *DB) DeleteImageBySourceID(source, sourceID string) (string, error) {
	// First get the local path
	var localPath string
	query := `SELECT local_path FROM images WHERE source = ? AND source_id = ?`
	err := db.conn.QueryRow(query, source, sourceID).Scan(&localPath)
	if err != nil {
		return "", err
	}

	// Delete from database
	deleteQuery := `DELETE FROM images WHERE source = ? AND source_id = ?`
	_, err = db.conn.Exec(deleteQuery, source, sourceID)
	if err != nil {
		return "", err
	}

	return localPath, nil
}

// CleanupMissingFiles removes database entries for files that no longer exist on disk
func (db *DB) CleanupMissingFiles() ([]string, error) {
	// Get all images
	images, err := db.ListImages("", 0)
	if err != nil {
		return nil, err
	}

	var deletedPaths []string
	for _, img := range images {
		// Check if file exists
		if _, err := os.Stat(img.LocalPath); os.IsNotExist(err) {
			// File doesn't exist, remove from database
			_, err = db.conn.Exec(`DELETE FROM images WHERE id = ?`, img.ID)
			if err != nil {
				return deletedPaths, fmt.Errorf("failed to delete image %d: %w", img.ID, err)
			}
			deletedPaths = append(deletedPaths, img.LocalPath)
		}
	}

	return deletedPaths, nil
}

// GetImageByID gets an image by ID
func (db *DB) GetImageByID(id int) (*Image, error) {
	query := `SELECT id, source, source_id, url, local_path, checksum, tags, resolution, file_size, downloaded_at FROM images WHERE id = ?`
	var img Image
	err := db.conn.QueryRow(query, id).Scan(&img.ID, &img.Source, &img.SourceID, &img.URL, &img.LocalPath,
		&img.Checksum, &img.Tags, &img.Resolution, &img.FileSize, &img.DownloadedAt)
	if err != nil {
		return nil, err
	}
	return &img, nil
}
