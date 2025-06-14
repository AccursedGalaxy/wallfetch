package cli

import (
	"fmt"

	"github.com/AccursedGalaxy/wallfetch/internal/config"
	"github.com/spf13/cobra"
)

// App represents the CLI application
type App struct {
	config  *config.Config
	rootCmd *cobra.Command
}

// NewApp creates a new CLI application
func NewApp(cfg *config.Config) *App {
	app := &App{
		config: cfg,
	}

	app.rootCmd = &cobra.Command{
		Use:   "wallfetch",
		Short: "A powerful CLI tool to fetch and manage wallpapers",
		Long: `WallFetch is a powerful CLI tool to fetch and manage wallpapers from various sources 
with intelligent duplicate detection and local library management.`,
		Version: "1.0.0",
	}

	// Add subcommands
	app.rootCmd.AddCommand(app.newFetchCmd())
	app.rootCmd.AddCommand(app.newListCmd())
	app.rootCmd.AddCommand(app.newBrowseCmd())
	app.rootCmd.AddCommand(app.newPruneCmd())
	app.rootCmd.AddCommand(app.newDedupeCmd())
	app.rootCmd.AddCommand(app.newDeleteCmd())
	app.rootCmd.AddCommand(app.newCleanupCmd())
	app.rootCmd.AddCommand(app.newConfigCmd())

	return app
}

// Run executes the CLI application
func (a *App) Run(args []string) error {
	a.rootCmd.SetArgs(args[1:]) // Skip program name
	return a.rootCmd.Execute()
}

// newConfigCmd creates the config command
func (a *App) newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Manage WallFetch configuration settings",
	}

	// config show
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Configuration:\n")
			fmt.Printf("  Default Source: %s\n", a.config.DefaultSource)
			fmt.Printf("  Download Directory: %s\n", a.config.DownloadDir)
			fmt.Printf("  Max Concurrent: %d\n", a.config.MaxConcurrent)
			fmt.Printf("  Database Path: %s\n", a.config.Database.Path)

			if apiKey := a.config.GetWallhavenAPIKey(); apiKey != "" {
				fmt.Printf("  Wallhaven API Key: %s...%s\n", apiKey[:8], apiKey[len(apiKey)-4:])
			} else {
				fmt.Printf("  Wallhaven API Key: Not set\n")
			}

			return nil
		},
	}

	// config init
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("Configuration file created successfully\n")
			return nil
		},
	}

	cmd.AddCommand(showCmd)
	cmd.AddCommand(initCmd)

	return cmd
}
