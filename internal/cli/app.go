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
		Version: "1.1.0",
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
	app.rootCmd.AddCommand(app.newCompletionCmd())

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

// newCompletionCmd creates the completion command
func (a *App) newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for wallfetch.

To load completions:

Bash:
  # Linux:
  $ wallfetch completion bash > /etc/bash_completion.d/wallfetch
  # macOS:
  $ wallfetch completion bash > $(brew --prefix)/etc/bash_completion.d/wallfetch

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ wallfetch completion zsh > "${fpath[1]}/_wallfetch"

Fish:
  $ wallfetch completion fish | source

  # To load completions for each session, execute once:
  $ wallfetch completion fish > ~/.config/fish/completions/wallfetch.fish

PowerShell:
  PS> wallfetch completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> wallfetch completion powershell > wallfetch.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return a.rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return a.rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return a.rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return a.rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				return cmd.Help()
			}
		},
	}
}
