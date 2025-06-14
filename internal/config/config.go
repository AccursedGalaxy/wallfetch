package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// Default settings
	DefaultSource string `yaml:"default_source"`
	DownloadDir   string `yaml:"download_dir"`
	MaxConcurrent int    `yaml:"max_concurrent"`

	// API Keys
	APIKeys map[string]string `yaml:"api_keys"`

	// Default fetch options
	Defaults map[string]DefaultOptions `yaml:"defaults"`

	// Database settings
	Database DatabaseConfig `yaml:"database"`
}

// DefaultOptions represents default options for each source
type DefaultOptions struct {
	Categories    string   `yaml:"categories"`
	Resolution    string   `yaml:"resolution"`
	Sort          string   `yaml:"sort"`
	Limit         int      `yaml:"limit"`
	AspectRatios  []string `yaml:"aspect_ratios"`
	MinWidth      int      `yaml:"min_width"`
	MinHeight     int      `yaml:"min_height"`
	MaxWidth      int      `yaml:"max_width"`
	MaxHeight     int      `yaml:"max_height"`
	OnlyLandscape bool     `yaml:"only_landscape"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path       string `yaml:"path"`
	AutoVacuum bool   `yaml:"auto_vacuum"`
}

// Load loads configuration from the config file and environment
func Load() (*Config, error) {
	cfg := Default()

	// Try to load from config file
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables if they exist
	if apiKey := os.Getenv("WALLHAVEN_API_KEY"); apiKey != "" {
		if cfg.APIKeys == nil {
			cfg.APIKeys = make(map[string]string)
		}
		cfg.APIKeys["wallhaven"] = apiKey
	}

	// Expand paths
	cfg.DownloadDir = expandPath(cfg.DownloadDir)
	cfg.Database.Path = expandPath(cfg.Database.Path)

	return cfg, nil
}

// Default returns a configuration with default values
func Default() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		DefaultSource: "wallhaven",
		DownloadDir:   filepath.Join(homeDir, "Pictures", "Wallpapers"),
		MaxConcurrent: 5,
		APIKeys:       make(map[string]string),
		Defaults: map[string]DefaultOptions{
			"wallhaven": {
				Categories:    "anime,nature",
				Resolution:    "1920x1080",
				Sort:          "toplist",
				Limit:         10,
				AspectRatios:  []string{"16x9", "21x9", "32x9"}, // Support ultrawide
				MinWidth:      1920,
				MinHeight:     1080,
				OnlyLandscape: true, // Prevent portrait images
			},
		},
		Database: DatabaseConfig{
			Path:       filepath.Join(homeDir, ".local", "share", "wallfetch", "wallpapers.db"),
			AutoVacuum: true,
		},
	}
}

// GetWallhavenAPIKey returns the Wallhaven API key from config or environment
func (c *Config) GetWallhavenAPIKey() string {
	if c.APIKeys != nil {
		if key, exists := c.APIKeys["wallhaven"]; exists {
			return key
		}
	}
	return os.Getenv("WALLHAVEN_API_KEY")
}

// Save saves the current configuration to the config file
func (c *Config) Save() error {
	configPath := getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "wallfetch", "config.yaml")
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "wallfetch", "config.yaml")
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if path == "" {
		return path
	}

	if path[0] == '~' {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[1:])
	}

	return path
}
