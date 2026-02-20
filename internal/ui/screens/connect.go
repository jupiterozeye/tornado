// Package screens contains the main application screens for Tornado.
//
// This file implements the Connection screen - the first screen users see.
// It provides a form for entering database connection details.
//
// Key Learning - Screen as a Model:
//
//	Each screen is a complete Bubble Tea model with Init/Update/View.
//	The App (in app.go) delegates to the active screen.
//
// Key Learning - Form Components:
//
//	This screen uses Bubbles textinput for form fields.
//	It demonstrates how to compose multiple components.
//
// TODO: Implement the connection form screen:
//   - [ ] Define ConnectModel struct with form fields
//   - [ ] Implement NewConnectModel constructor
//   - [ ] Implement Init with focus on first field
//   - [ ] Implement Update for form navigation and submission
//   - [ ] Implement View to render the form
//   - [ ] Add database type selector (SQLite vs PostgreSQL)
//   - [ ] Add connection validation
//   - [ ] Add "recent connections" list
//
// Bubbles Components to Use:
//   - textinput.Model for input fields
//   - list.Model for recent connections
//   - spinner.Model for connection testing animation
//
// References:
//   - https://github.com/charmbracelet/bubbles#text-input
//   - https://github.com/charmbracelet/bubbletea/tree/main/examples/textinputs
package screens

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// ConnectModel is the model for the connection screen.
// It manages the connection form state and user input.
type ConnectModel struct {
	// Form fields
	dbTypeInput   textinput.Model // "sqlite" or "postgres"
	pathInput     textinput.Model // SQLite file path
	hostInput     textinput.Model // Postgres host
	portInput     textinput.Model // Postgres port
	userInput     textinput.Model // Postgres username
	passwordInput textinput.Model // Postgres password
	databaseInput textinput.Model // Postgres database name

	// UI state
	focusIndex   int    // Which field is focused
	isConnecting bool   // Currently attempting connection
	error        string // Error message to display

	// Dimensions
	width  int
	height int

	// Styling
	styles *styles.Styles

	// History
	recentConnections []models.ConnectionHistoryItem
	showRecent        bool
}

// NewConnectModel creates a new connection screen model.
//
// TODO: Initialize all form fields with appropriate defaults
// TODO: Load recent connections from history
func NewConnectModel(s *styles.Styles) *ConnectModel {
	// Initialize database type input
	dbType := textinput.New()
	dbType.Placeholder = "sqlite"
	dbType.Focus()
	dbType.CharLimit = 20

	// Intialize SQLite path input
	path := textinput.New()
	path.Placeholder = "path/to/database.db"
	path.CharLimit = 256

	// Initialize Postgres fields
	host := textinput.New()
	host.Placeholder = "localhost"
	host.CharLimit = 100

	port := textinput.New()
	port.Placeholder = "5432"
	port.CharLimit = 10

	user := textinput.New()
	user.Placeholder = "username"
	user.CharLimit = 50

	password := textinput.New()
	password.Placeholder = "password"
	password.EchoMode = textinput.EchoPassword
	password.CharLimit = 100

	database := textinput.New()
	database.Placeholder = "db_name"
	database.CharLimit = 50

	return &ConnectModel{
		dbTypeInput:   dbType,
		pathInput:     path,
		hostInput:     host,
		portInput:     port,
		userInput:     user,
		passwordInput: password,
		databaseInput: database,
		focusIndex:    0,
		isConnecting:  false,
		error:         "",
		showRecent:    false,
		styles:        s,
	}
}

// Init returns the initial command for the connection screen.
// Usually returns nil since we just want to show the form.
//
// TODO: Consider what should happen on init:
//   - Focus first input field? (handled in constructor)
//   - Load recent connections from file?
//   - Check for default database file?
func (m *ConnectModel) Init() tea.Cmd {
	// TODO: Return initialization commands
	// Example: Load recent connections
	// return loadRecentConnections()
	return nil
}

// Update handles messages for the connection screen.
//
// Key events to handle:
//   - Tab/Shift+Tab: Move focus between fields
//   - Up/Down: Navigate fields
//   - Enter: Submit form (if on submit button)
//   - Ctrl+R: Toggle recent connections list
//   - Character input: Type into focused field
//
// Messages to handle:
//   - tea.KeyMsg: Keyboard input
//   - tea.WindowSizeMsg: Terminal resize
//   - ConnectSuccessMsg: Connection succeeded
//   - ConnectErrorMsg: Connection failed
//
// TODO: Implement complete form handling
func (m *ConnectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: Implement message handling
	//
	// switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	//     m.width = msg.Width
	//     m.height = msg.Height
	//
	// case tea.KeyMsg:
	//     switch msg.String() {
	//     case "tab":
	//         m.nextField()
	//     case "shift+tab":
	//         m.prevField()
	//     case "enter":
	//         return m.handleEnter()
	//     case "up":
	//         m.prevField()
	//     case "down":
	//         m.nextField()
	//     default:
	//         // Pass to focused input
	//         return m.updateFocusedInput(msg)
	//     }
	//
	// case ConnectSuccessMsg:
	//     // Transition to browser screen
	//     return m, func() tea.Msg {
	//         return ScreenChangeMsg{Screen: ScreenBrowser}
	//     }
	//
	// case ConnectErrorMsg:
	//     m.isConnecting = false
	//     m.error = msg.Error()
	// }

	return m, nil
}

// View renders the connection screen.
//
// Layout structure:
//   - Title/header
//   - Database type selector
//   - Form fields (different based on db type)
//   - Error message (if any)
//   - Help/keybindings
//
// TODO: Implement form rendering with Lip Gloss styling
func (m *ConnectModel) View() string {
	// TODO: Implement view composition
	//
	// Structure:
	// 1. Title: "Tornado - Database Connection"
	// 2. Database type toggle: [SQLite] [PostgreSQL]
	// 3. Form fields based on type:
	//    SQLite:    File Path: [________________]
	//    PostgreSQL: Host: [________] Port: [____]
	//               User: [________] Password: [________]
	//               Database: [________]
	// 4. Connect button: [Connect]
	// 5. Error message in red if any
	// 6. Help at bottom: Tab: switch field, Enter: connect, Ctrl+R: recent
	//
	// Use lipgloss.JoinVertical and lipgloss.JoinHorizontal for layout

	return "Connection Screen\n\nTODO: Implement connection form"
}

// Focus handlers and form logic
// TODO: Implement these helper methods

func (m *ConnectModel) nextField() {
	// TODO: Move focus to next field
}

func (m *ConnectModel) prevField() {
	// TODO: Move focus to previous field
}

func (m *ConnectModel) handleEnter() (tea.Model, tea.Cmd) {
	// TODO: If on connect button, start connection
	// Otherwise, move to next field
	return m, nil
}

// getConfig returns the connection config from form inputs.
//
// TODO: Implement config extraction
func (m *ConnectModel) getConfig() models.ConnectionConfig {
	return models.ConnectionConfig{}
}

// startConnection initiates a database connection.
// Returns a command that performs the connection asynchronously.
//
// Key Learning - Async Operations:
//   - Return a tea.Cmd that performs the operation
//   - The Cmd runs in a goroutine
//   - When done, it returns a message with the result
//
// TODO: Implement async connection
func (m *ConnectModel) startConnection() tea.Cmd {
	// TODO: Return a command that:
	// 1. Creates the appropriate Database implementation
	// 2. Calls Connect with the form config
	// 3. Returns ConnectSuccessMsg or ConnectErrorMsg
	return nil
}

// ConnectAttemptMsg signals that a connection attempt is starting.
type ConnectAttemptMsg struct{}

// ConnectSuccessMsg signals successful database connection.
type ConnectSuccessMsg struct {
	Config models.ConnectionConfig
}

// ConnectErrorMsg signals a failed connection attempt.
type ConnectErrorMsg struct {
	Error string
}
