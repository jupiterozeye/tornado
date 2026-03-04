// Package main is the entry point for the Tornado TUI application.
// References:
//   - https://charm.land/bubbletea/v2#quick-start
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/jupiterozeye/tornado/internal/app"
	"github.com/jupiterozeye/tornado/internal/config"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

func main() {
	// Check for DEBUG environment variable and set up file logging
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	// Load configuration from file
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		// Continue with defaults
	}

	// Apply saved theme
	if cfg != nil && cfg.GetTheme() != "" {
		styles.SetTheme(cfg.GetTheme())
	}

	// Create the root application model
	application := app.New()

	// Create and run the Bubble Tea program
	p := tea.NewProgram(application)

	// Run the program - this blocks until the user quits
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Tornado: %v\n", err)
		os.Exit(1)
	}
}
