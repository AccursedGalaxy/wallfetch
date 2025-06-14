package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AccursedGalaxy/wallfetch/internal/cli"
	"github.com/AccursedGalaxy/wallfetch/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		// Continue with default config
		cfg = config.Default()
	}

	// Create and run CLI
	app := cli.NewApp(cfg)
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
} 