// Package components - Status bar component for displaying application state.
//
// This file implements StatusBar - a bottom bar showing:
//   - Current database connection info
//   - Current screen/mode
//   - Keybinding hints
//   - Error messages
//
// Layout (envisioned):
//   ┌─────────────────────────────────────────────────────────────────┐
//   │ ... main content ...                                            │
//   ├─────────────────────────────────────────────────────────────────┤
//   │ 🗄️ mydb.db │ Browser │ 123 rows │ [q]uit [?]help [r]efresh    │
//   └─────────────────────────────────────────────────────────────────┘
//
// TODO: Implement the status bar:
//   - [ ] Define StatusBar struct
//   - [ ] Implement NewStatusBar constructor
//   - [ ] Implement View method (no Update needed - passive display)
//   - [ ] Add connection status display
//   - [ ] Add keybinding hints
//   - [ ] Add error message display
//   - [ ] Add loading indicator
//
// Key Learning - Passive Components:
//   Some components only display data and don't handle input.
//   They still implement the tea.Model interface but Update just returns.
//
// References:
//   - https://github.com/charmbracelet/lipgloss#joining
package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// StatusBar displays application status at the bottom of the screen.
// It's a passive component - it displays info but doesn't handle input.
//
// TODO: Complete the implementation
type StatusBar struct {
	// TODO: Add these fields
	//
	// ===== Display Content =====
	// connectionName  string
	// currentScreen   string
	// rowCount        int
	// errorMessage    string
	// isLoading       bool
	//
	// ===== Keybinding Hints =====
	// keybindings     []Keybinding
	//
	// ===== Dimensions =====
	// width           int
	//
	// ===== Styling =====
	// styles          *styles.Styles
}

// Keybinding represents a keybinding hint for display.
type Keybinding struct {
	Key  string
	Desc string
}

// NewStatusBar creates a new status bar component.
//
// TODO: Initialize with default values
func NewStatusBar(width int) *StatusBar {
	return &StatusBar{
		// TODO: Initialize fields
		// width: width,
		// keybindings: []Keybinding{
		//     {"q", "quit"},
		//     {"?", "help"},
		// },
	}
}

// Init returns nil (passive component).
func (s *StatusBar) Init() tea.Cmd {
	return nil
}

// Update returns the model unchanged (passive component).
func (s *StatusBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

// View renders the status bar.
//
// TODO: Implement status bar layout
// Structure:
//   [Connection] | [Screen] | [Info] | [Keybindings]
//   If error: show error in red
//   If loading: show spinner
func (s *StatusBar) View() string {
	// TODO: Build status bar with lipgloss.JoinHorizontal
	//
	// leftSection := s.renderConnection()
	// middleSection := s.renderInfo()
	// rightSection := s.renderKeybindings()
	//
	// return lipgloss.JoinHorizontal(
	//     lipgloss.Top,
	//     leftSection,
	//     middleSection,
	//     rightSection,
	// )

	return "StatusBar - TODO"
}

// SetConnection sets the connection display name.
func (s *StatusBar) SetConnection(name string) {
	// s.connectionName = name
}

// SetScreen sets the current screen name.
func (s *StatusBar) SetScreen(screen string) {
	// s.currentScreen = screen
}

// SetRowCount sets the row count display.
func (s *StatusBar) SetRowCount(count int) {
	// s.rowCount = count
}

// SetError sets an error message to display.
func (s *StatusBar) SetError(err string) {
	// s.errorMessage = err
}

// ClearError clears the error message.
func (s *StatusBar) ClearError() {
	// s.errorMessage = ""
}

// SetLoading sets the loading indicator.
func (s *StatusBar) SetLoading(loading bool) {
	// s.isLoading = loading
}

// SetKeybindings sets the keybinding hints.
func (s *StatusBar) SetKeybindings(bindings []Keybinding) {
	// s.keybindings = bindings
}

// SetWidth updates the status bar width.
func (s *StatusBar) SetWidth(width int) {
	// s.width = width
}

// renderConnection renders the connection info section.
//
// TODO: Implement
func (s *StatusBar) renderConnection() string {
	// Show database icon and name
	// Style based on connection status
	return ""
}

// renderInfo renders the middle info section.
//
// TODO: Implement
func (s *StatusBar) renderInfo() string {
	// Show current screen, row count, etc.
	return ""
}

// renderKeybindings renders the keybinding hints section.
//
// TODO: Implement
func (s *StatusBar) renderKeybindings() string {
	// Show [key] description format
	// Right-align this section
	return ""
}

// renderError renders an error overlay if there's an error.
//
// TODO: Implement
func (s *StatusBar) renderError() string {
	// Show error in red box above status bar
	return ""
}

// Import tea for the Model interface
import "github.com/charmbracelet/bubbletea"
