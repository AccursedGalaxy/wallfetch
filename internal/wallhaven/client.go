package wallhaven

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	BaseURL = "https://wallhaven.cc/api/v1"
)

// Client represents a Wallhaven API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Wallhaven API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchParams represents search parameters for the Wallhaven API
type SearchParams struct {
	Query      string // Search query
	Categories string // Categories: general, anime, people (100/010/001)
	Purity     string // Purity: sfw, sketchy, nsfw (100/010/001)
	Sorting    string // Sorting: date_added, relevance, random, views, favorites, toplist
	Order      string // Order: desc, asc
	TopRange   string // Top range: 1d, 3d, 1w, 1M, 3M, 6M, 1y
	AtLeast    string // Minimum resolution (e.g., 1920x1080)
	Ratios     string // Aspect ratios (e.g., 16x9,16x10)
	Colors     string // Color search
	Page       int    // Page number
	Seed       string // Random seed
}

// SearchResult represents the search API response
type SearchResult struct {
	Data []Wallpaper `json:"data"`
	Meta Meta        `json:"meta"`
}

// Wallpaper represents a wallpaper from the API
type Wallpaper struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	ShortURL   string    `json:"short_url"`
	Views      int       `json:"views"`
	Favorites  int       `json:"favorites"`
	Source     string    `json:"source"`
	Purity     string    `json:"purity"`
	Category   string    `json:"category"`
	DimensionX int       `json:"dimension_x"`
	DimensionY int       `json:"dimension_y"`
	Resolution string    `json:"resolution"`
	Ratio      string    `json:"ratio"`
	FileSize   int       `json:"file_size"`
	FileType   string    `json:"file_type"`
	CreatedAt  string    `json:"created_at"`
	Colors     []string  `json:"colors"`
	Path       string    `json:"path"`
	Thumbs     Thumbs    `json:"thumbs"`
	Tags       []Tag     `json:"tags,omitempty"`
	Uploader   *Uploader `json:"uploader,omitempty"`
}

// Thumbs represents thumbnail URLs
type Thumbs struct {
	Large    string `json:"large"`
	Original string `json:"original"`
	Small    string `json:"small"`
}

// Tag represents a wallpaper tag
type Tag struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Alias      string `json:"alias"`
	CategoryID int    `json:"category_id"`
	Category   string `json:"category"`
	Purity     string `json:"purity"`
	CreatedAt  string `json:"created_at"`
}

// Uploader represents the wallpaper uploader
type Uploader struct {
	Username string            `json:"username"`
	Group    string            `json:"group"`
	Avatar   map[string]string `json:"avatar"`
}

// Meta represents pagination metadata
type Meta struct {
	CurrentPage int    `json:"current_page"`
	LastPage    int    `json:"last_page"`
	PerPage     string `json:"per_page"` // Wallhaven API returns this as string
	Total       int    `json:"total"`
	Query       string `json:"query,omitempty"`
	Seed        string `json:"seed,omitempty"`
}

// WallpaperDetail represents detailed wallpaper information
type WallpaperDetail struct {
	Data Wallpaper `json:"data"`
}

// Search searches for wallpapers
func (c *Client) Search(params SearchParams) (*SearchResult, error) {
	u, err := url.Parse(fmt.Sprintf("%s/search", BaseURL))
	if err != nil {
		return nil, err
	}

	q := u.Query()

	// Add search parameters
	if params.Query != "" {
		q.Set("q", params.Query)
	}
	if params.Categories != "" {
		q.Set("categories", c.convertCategories(params.Categories))
	}
	if params.Purity != "" {
		q.Set("purity", c.convertPurity(params.Purity))
	}
	if params.Sorting != "" {
		q.Set("sorting", params.Sorting)
	}
	if params.Order != "" {
		q.Set("order", params.Order)
	}
	if params.TopRange != "" {
		q.Set("topRange", params.TopRange)
	}
	if params.AtLeast != "" {
		q.Set("atleast", params.AtLeast)
	}
	if params.Ratios != "" {
		q.Set("ratios", params.Ratios)
	}
	if params.Colors != "" {
		q.Set("colors", params.Colors)
	}
	if params.Page > 0 {
		q.Set("page", strconv.Itoa(params.Page))
	}
	if params.Seed != "" {
		q.Set("seed", params.Seed)
	}

	// Add API key if available
	if c.apiKey != "" {
		q.Set("apikey", c.apiKey)
	}

	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWallpaper gets detailed information about a specific wallpaper
func (c *Client) GetWallpaper(id string) (*WallpaperDetail, error) {
	u := fmt.Sprintf("%s/w/%s", BaseURL, id)
	if c.apiKey != "" {
		u += "?apikey=" + c.apiKey
	}

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var result WallpaperDetail
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// convertCategories converts category names to API format
func (c *Client) convertCategories(categories string) string {
	// Convert comma-separated category names to binary format
	// general=100, anime=010, people=001
	result := "000"
	if contains(categories, "general") {
		result = "1" + result[1:]
	}
	if contains(categories, "anime") {
		result = result[:1] + "1" + result[2:]
	}
	if contains(categories, "people") {
		result = result[:2] + "1"
	}
	return result
}

// convertPurity converts purity names to API format
func (c *Client) convertPurity(purity string) string {
	// Convert comma-separated purity names to binary format
	// sfw=100, sketchy=010, nsfw=001
	result := "000"
	if contains(purity, "sfw") {
		result = "1" + result[1:]
	}
	if contains(purity, "sketchy") {
		result = result[:1] + "1" + result[2:]
	}
	if contains(purity, "nsfw") {
		result = result[:2] + "1"
	}
	return result
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)+1] == substr+"," ||
					s[len(s)-len(substr)-1:] == ","+substr ||
					contains(s[len(substr)+1:], substr)))
}
