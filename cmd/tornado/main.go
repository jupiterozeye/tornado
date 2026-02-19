// Package main is the entry point for the Tornado TUI application.
//
// This file is responsible for:
//   - Initializing the application with its starting state
//   - Setting up logging (when DEBUG is set)
//   - Creating and running the Bubble Tea program
//
// TODO: Implement the following:
//   - [ ] Parse command-line flags (e.g., --connect, --db-type)
//   - [ ] Load configuration from file (~/.tornado.yaml or similar)
//   - [ ] Handle graceful shutdown
//   - [ ] Set up signal handling for SIGINT/SIGTERM
//
// Key Bubble Tea Concepts Used Here:
//   - tea.NewProgram() creates the TUI program
//   - tea.WithAltScreen() uses the alternate screen buffer (cleaner exit)
//   - tea.WithMouseCellMotion() enables mouse support for clicking
//
// References:
//   - https://github.com/charmbracelet/bubbletea#quick-start
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yourusername/tornado/internal/app"
)

func main() {
	// TODO: Check for DEBUG environment variable and set up file logging
	// Example from Bubble Tea docs:
	// if len(os.Getenv("DEBUG")) > 0 {
	//     f, err := tea.LogToFile("debug.log", "debug")
	//     if err != nil {
	//         fmt.Println("fatal:", err)
	//         os.Exit(1)
	//     }
	//     defer f.Close()
	// }

	// TODO: Parse command line arguments
	// Could support flags like:
	//   --db-type [sqlite|postgres]
	//   --connect "connection-string"
	//   --config path/to/config.yaml

	// Create the root application model
	// This is the main Model in the Elm Architecture
	// It will contain references to all screens and manage navigation
	application := app.New()

	// Create the Bubble Tea program with options:
	// - WithAltScreen: Uses alternate screen buffer (your terminal returns to
	//   its previous state when the app exits)
	// - WithMouseCellMotion: Enables mouse click and drag support
	p := tea.NewProgram(
		application,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program - this blocks until the user quits
	// The returned model is the final state (usually not needed)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Tornado: %v\n", err)
		os.Exit(1)
	}
}
