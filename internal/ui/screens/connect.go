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
//   - [x] Define ConnectModel struct with form fields
//   - [x] Implement NewConnectModel constructor
//   - [x] Implement Init with focus on first field
//   - [x] Implement Update for form navigation and submission
//   - [x] Implement View to render the form
//   - [x] Add database type selector (SQLite vs PostgreSQL)
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
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jupiterozeye/tornado/internal/assets"
	"github.com/jupiterozeye/tornado/internal/db"
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
	errorMsg     string // Error message to display

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
// Initializes all form fields with appropriate defaults
// TODO: Load recent connections from history
func NewConnectModel() *ConnectModel {
	s := styles.Default()

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
		errorMsg:      "",
		showRecent:    false,
		styles:        s,
	}
}

// Init returns the initial command for the connection screen.
// Returns textinput.Blink to make the cursor blink.
//
// TODO: Consider what should happen on init:
//   - Focus first input field? (handled in constructor)
//   - Load recent connections from file?
//   - Check for default database file?
func (m *ConnectModel) Init() tea.Cmd {
	return textinput.Blink
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
func (m *ConnectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// First, update the focused input for cursor blinking and state
		var cmd tea.Cmd

		switch msg.String() {
		case "tab":
			m.nextField()
		case "shift+tab":
			m.prevField()
		case "up":
			m.prevField()
		case "down":
			m.nextField()
		case "enter":
			return m.handleEnter()
		default:
			// Pass character input to focused field
			// Capture old db type to detect changes
			oldDbType := m.dbTypeInput.Value()
			newModel, cmd := m.updateFocusedInput(msg)
			// Check if db type changed
			if m.focusIndex == 0 && oldDbType != m.dbTypeInput.Value() {
				// DB type changed, adjust focus if needed
				maxIndex := m.getMaxFieldIndex()
				if m.focusIndex > maxIndex {
					m.focusIndex = maxIndex
					m.blurAllFields()
					m.focusCurrentField()
				}
			}
			return newModel, cmd
		}

		// For navigation keys, still update focused input for cursor blinking
		_, cmd = m.updateFocusedInput(msg)
		return m, cmd

	case ConnectSuccessMsg:
		// Return the success message - app.go will handle the transition
		return m, nil

	case ConnectErrorMsg:
		m.isConnecting = false
		m.errorMsg = msg.Error
		return m, nil
	}

	return m, nil
}

// View renders the connection screen.
//
// Layout structure:
//   - Title/header (centered logo)
//   - Database type selector
//   - Form fields (different based on db type)
//   - Error message (if any)
//   - Help/keybindings
func (m *ConnectModel) View() string {
	// Centered logo header
	logoStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Align(lipgloss.Center)
	logo := logoStyle.Render(assets.Logo)

	// Show connecting message if connecting
	if m.isConnecting {
		connectingMsg := m.styles.Muted.Render("Connecting...")
		return lipgloss.JoinVertical(
			lipgloss.Left,
			logo,
			"",
			connectingMsg,
		)
	}

	// Determine which DB type is selected
	dbType := m.dbTypeInput.Value()
	if dbType == "" {
		dbType = "sqlite" // default
	}
	isSQLite := dbType == "sqlite" || dbType == "SQLite"

	// Form section header
	formHeader := m.styles.Subheader.Render("Connection Details")

	// Build dynamic content based on DB type
	var content []string
	content = append(content, logo, "", formHeader, "")

	// Database type field (always shown)
	dbTypeLabel := "Database Type (sqlite/postgres):"
	dbTypeField := m.renderField(dbTypeLabel, m.dbTypeInput.View(), 0)
	content = append(content, dbTypeField, "")

	if isSQLite {
		// SQLite section only
		sqliteHeader := m.styles.Header.Render("SQLite Settings")
		pathLabel := "File Path:"
		pathField := m.renderField(pathLabel, m.pathInput.View(), 1)
		content = append(content, sqliteHeader, pathField)
	} else {
		// PostgreSQL section only
		postgresHeader := m.styles.Header.Render("PostgreSQL Settings")
		hostLabel := "Host:"
		hostField := m.renderField(hostLabel, m.hostInput.View(), 2)
		portLabel := "Port:"
		portField := m.renderField(portLabel, m.portInput.View(), 3)
		userLabel := "Username:"
		userField := m.renderField(userLabel, m.userInput.View(), 4)
		passwordLabel := "Password:"
		passwordField := m.renderField(passwordLabel, m.passwordInput.View(), 5)
		dbLabel := "Database Name:"
		dbField := m.renderField(dbLabel, m.databaseInput.View(), 6)
		content = append(content,
			postgresHeader,
			hostField,
			portField,
			userField,
			passwordField,
			dbField,
		)
	}

	// Error message
	if m.errorMsg != "" {
		content = append(content, "", m.styles.Error.Render("Error: "+m.errorMsg))
	}

	// Help text
	help := m.styles.Muted.Render("Tab/Shift+Tab: Navigate | Up/Down: Move | Enter: Connect | Ctrl+C: Quit")
	content = append(content, "", help)

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// renderField renders a form field with label, applying focus styling if focused
func (m *ConnectModel) renderField(label, value string, index int) string {
	labelStyle := m.styles.Muted
	if m.focusIndex == index {
		labelStyle = m.styles.Bold
		value = m.styles.InputFocus.Render(value)
	} else {
		value = m.styles.Input.Render(value)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(label+" "),
		value,
	)
}

// getMaxFieldIndex returns the maximum field index based on current DB type
func (m *ConnectModel) getMaxFieldIndex() int {
	dbType := m.dbTypeInput.Value()
	if dbType == "" {
		dbType = "sqlite"
	}
	if dbType == "sqlite" || dbType == "SQLite" {
		return 1 // dbType (0) and path (1)
	}
	return 6 // All PostgreSQL fields
}

// isFieldVisible returns true if the field should be shown for current DB type
func (m *ConnectModel) isFieldVisible(index int) bool {
	maxIndex := m.getMaxFieldIndex()
	return index <= maxIndex
}

// nextField moves focus to the next visible field
func (m *ConnectModel) nextField() {
	m.blurCurrentField()
	maxIndex := m.getMaxFieldIndex()
	m.focusIndex++
	if m.focusIndex > maxIndex {
		m.focusIndex = 0
	}
	m.focusCurrentField()
}

// prevField moves focus to the previous visible field
func (m *ConnectModel) prevField() {
	m.blurCurrentField()
	maxIndex := m.getMaxFieldIndex()
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = maxIndex
	}
	m.focusCurrentField()
}

// blurCurrentField removes focus from the current field
func (m *ConnectModel) blurCurrentField() {
	switch m.focusIndex {
	case 0:
		m.dbTypeInput.Blur()
	case 1:
		m.pathInput.Blur()
	case 2:
		m.hostInput.Blur()
	case 3:
		m.portInput.Blur()
	case 4:
		m.userInput.Blur()
	case 5:
		m.passwordInput.Blur()
	case 6:
		m.databaseInput.Blur()
	}
}

// blurAllFields removes focus from all fields
func (m *ConnectModel) blurAllFields() {
	m.dbTypeInput.Blur()
	m.pathInput.Blur()
	m.hostInput.Blur()
	m.portInput.Blur()
	m.userInput.Blur()
	m.passwordInput.Blur()
	m.databaseInput.Blur()
}

// focusCurrentField sets focus on the current field
func (m *ConnectModel) focusCurrentField() {
	switch m.focusIndex {
	case 0:
		m.dbTypeInput.Focus()
	case 1:
		m.pathInput.Focus()
	case 2:
		m.hostInput.Focus()
	case 3:
		m.portInput.Focus()
	case 4:
		m.userInput.Focus()
	case 5:
		m.passwordInput.Focus()
	case 6:
		m.databaseInput.Focus()
	}
}

// handleEnter handles the enter key press
func (m *ConnectModel) handleEnter() (tea.Model, tea.Cmd) {
	// If on last visible field, start connection
	if m.focusIndex == m.getMaxFieldIndex() {
		return m, m.startConnection()
	}
	// Otherwise, move to next field
	m.nextField()
	return m, nil
}

// updateFocusedInput passes keyboard input to the focused input field
func (m *ConnectModel) updateFocusedInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focusIndex {
	case 0:
		m.dbTypeInput, cmd = m.dbTypeInput.Update(msg)
	case 1:
		m.pathInput, cmd = m.pathInput.Update(msg)
	case 2:
		m.hostInput, cmd = m.hostInput.Update(msg)
	case 3:
		m.portInput, cmd = m.portInput.Update(msg)
	case 4:
		m.userInput, cmd = m.userInput.Update(msg)
	case 5:
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	case 6:
		m.databaseInput, cmd = m.databaseInput.Update(msg)
	}

	return m, cmd
}

// getConfig returns the connection config from form inputs.
func (m *ConnectModel) getConfig() models.ConnectionConfig {
	port, _ := strconv.Atoi(m.portInput.Value())
	if port == 0 {
		port = 5432
	}

	return models.ConnectionConfig{
		Type:     m.dbTypeInput.Value(),
		Path:     m.pathInput.Value(),
		Host:     m.hostInput.Value(),
		Port:     port,
		User:     m.userInput.Value(),
		Password: m.passwordInput.Value(),
		Database: m.databaseInput.Value(),
	}
}

// startConnection initiates a database connection.
// Returns a command that performs the connection asynchronously.
//
// Key Learning - Async Operations:
//   - Return a tea.Cmd that performs the operation
//   - The Cmd runs in a goroutine
//   - When done, it returns a message with the result
func (m *ConnectModel) startConnection() tea.Cmd {
	m.isConnecting = true
	m.errorMsg = ""

	config := m.getConfig()

	return func() tea.Msg {
		database, err := db.Open(config)
		if err != nil {
			return ConnectErrorMsg{Error: err.Error()}
		}
		return ConnectSuccessMsg{DB: database}
	}
}

// ConnectAttemptMsg signals that a connection attempt is starting.
type ConnectAttemptMsg struct{}

// ConnectSuccessMsg signals successful database connection.
type ConnectSuccessMsg struct {
	DB db.Database
}

// ConnectErrorMsg signals a failed connection attempt.
type ConnectErrorMsg struct {
	Error string
}
