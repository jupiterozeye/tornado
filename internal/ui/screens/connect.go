// Package screens contains the main application screens for Tornado.
//
// This file implements the Connection screen - the first screen users see.
// It provides a clean, minimal interface that opens a modal for connection details.
package screens

import (
	"strconv"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jupiterozeye/tornado/internal/assets"
	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// ConnectionState represents the current state of the connection screen
type ConnectionState int

const (
	StateWelcome ConnectionState = iota
	StateForm
	StateConnecting
)

// ConnectModel is the model for the connection screen.
type ConnectModel struct {
	// State
	state ConnectionState

	// Form fields
	dbTypeInput   textinput.Model
	pathInput     textinput.Model
	hostInput     textinput.Model
	portInput     textinput.Model
	userInput     textinput.Model
	passwordInput textinput.Model
	databaseInput textinput.Model

	// UI state
	focusIndex int
	errorMsg   string

	// Dimensions
	width  int
	height int

	// Styling
	styles *styles.Styles

	// Loading spinner
	spinner spinner.Model
}

// NewConnectModel creates a new connection screen model.
func NewConnectModel() *ConnectModel {
	s := styles.Default()

	// Initialize spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	// Initialize form fields
	dbType := textinput.New()
	dbType.Placeholder = "sqlite"
	dbType.CharLimit = 20

	path := textinput.New()
	path.Placeholder = "/path/to/database.db"
	path.CharLimit = 256

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
	database.Placeholder = "database"
	database.CharLimit = 50

	return &ConnectModel{
		state:         StateWelcome,
		styles:        s,
		spinner:       sp,
		dbTypeInput:   dbType,
		pathInput:     path,
		hostInput:     host,
		portInput:     port,
		userInput:     user,
		passwordInput: password,
		databaseInput: database,
		focusIndex:    0,
	}
}

// Init returns the initial command for the connection screen.
func (m *ConnectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the connection screen.
func (m *ConnectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.state {
		case StateWelcome:
			switch msg.String() {
			case " ":
				m.state = StateForm
				m.focusIndex = 0
				m.dbTypeInput.Focus()
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}

		case StateForm:
			return m.handleFormKeys(msg)

		case StateConnecting:
			// Only allow quit during connection
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		if m.state == StateConnecting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case ConnectSuccessMsg:
		// Connection succeeded - app.go will handle transition
		return m, nil

	case ConnectErrorMsg:
		m.state = StateForm
		m.errorMsg = msg.Err
		return m, nil
	}

	return m, nil
}

func (m *ConnectModel) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateWelcome
		m.blurAllFields()
		m.errorMsg = ""
		return m, nil

	case "tab":
		m.nextField()
		return m, nil

	case "shift+tab":
		m.prevField()
		return m, nil

	case "up":
		m.prevField()
		return m, nil

	case "down":
		m.nextField()
		return m, nil

	case "enter":
		if m.focusIndex == m.getMaxFieldIndex() {
			return m, m.startConnection()
		}
		m.nextField()
		return m, nil

	default:
		// Update the focused input
		var cmd tea.Cmd
		switch m.focusIndex {
		case 0:
			oldType := m.dbTypeInput.Value()
			m.dbTypeInput, cmd = m.dbTypeInput.Update(msg)
			// Adjust focus if db type changed
			if oldType != m.dbTypeInput.Value() {
				m.adjustFocusForDbType()
			}
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
}

// View renders the connection screen.
func (m *ConnectModel) View() string {
	switch m.state {
	case StateWelcome:
		return m.viewWelcome()
	case StateForm:
		return m.viewForm()
	case StateConnecting:
		return m.viewConnecting()
	default:
		return m.viewWelcome()
	}
}

func (m *ConnectModel) viewWelcome() string {
	// Center everything vertically and horizontally
	centerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	// Logo with primary color
	logoStyle := lipgloss.NewStyle().
		Foreground(styles.Primary)
	logo := logoStyle.Render(assets.Logo)

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		MarginTop(2)
	help := helpStyle.Render("Space: Connect | Ctrl+C: Quit")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		logo,
		help,
	)

	return centerStyle.Render(content)
}

func (m *ConnectModel) viewForm() string {
	// Center the modal
	centerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	// Modal box style (matching explorer/query/results boxes)
	isSQLite := m.isSQLite()
	height := 12
	if !isSQLite {
		height = 20
	}

	modalStyle := lipgloss.NewStyle().
		Width(60).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderFocus).
		Padding(1, 2)

	// Title
	title := m.styles.Subheader.Render("Database Connection")

	// Build form content
	var fields []string

	// DB Type selector
	dbTypeLabel := m.styles.Muted.Render("Database Type:")
	dbTypeField := m.renderField(0)
	fields = append(fields, dbTypeLabel, dbTypeField, "")

	if isSQLite {
		// SQLite fields only
		pathLabel := m.styles.Muted.Render("File Path:")
		pathField := m.renderField(1)
		fields = append(fields, pathLabel, pathField)
	} else {
		// PostgreSQL fields
		hostLabel := m.styles.Muted.Render("Host:")
		hostField := m.renderField(2)
		fields = append(fields, hostLabel, hostField)

		portLabel := m.styles.Muted.Render("Port:")
		portField := m.renderField(3)
		fields = append(fields, "", portLabel, portField)

		userLabel := m.styles.Muted.Render("Username:")
		userField := m.renderField(4)
		fields = append(fields, "", userLabel, userField)

		passLabel := m.styles.Muted.Render("Password:")
		passField := m.renderField(5)
		fields = append(fields, "", passLabel, passField)

		dbLabel := m.styles.Muted.Render("Database:")
		dbField := m.renderField(6)
		fields = append(fields, "", dbLabel, dbField)
	}

	// Error message
	if m.errorMsg != "" {
		fields = append(fields, "", m.styles.Error.Render(m.errorMsg))
	}

	// Help text
	fields = append(fields, "", m.styles.Muted.Render("Tab: Navigate | Enter: Connect | Esc: Cancel"))

	formContent := lipgloss.JoinVertical(lipgloss.Left, fields...)
	modalContent := lipgloss.JoinVertical(lipgloss.Left, title, "", formContent)

	return centerStyle.Render(modalStyle.Render(modalContent))
}

func (m *ConnectModel) viewConnecting() string {
	// Center everything
	centerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	// Spinner + text
	spinnerText := m.spinner.View() + " Connecting..."
	content := m.styles.Body.Render(spinnerText)

	return centerStyle.Render(content)
}

func (m *ConnectModel) renderField(index int) string {
	var value string
	var isFocused bool

	switch index {
	case 0:
		value = m.dbTypeInput.View()
		isFocused = m.focusIndex == 0
	case 1:
		value = m.pathInput.View()
		isFocused = m.focusIndex == 1
	case 2:
		value = m.hostInput.View()
		isFocused = m.focusIndex == 2
	case 3:
		value = m.portInput.View()
		isFocused = m.focusIndex == 3
	case 4:
		value = m.userInput.View()
		isFocused = m.focusIndex == 4
	case 5:
		value = m.passwordInput.View()
		isFocused = m.focusIndex == 5
	case 6:
		value = m.databaseInput.View()
		isFocused = m.focusIndex == 6
	}

	if isFocused {
		return m.styles.InputFocus.Render(value)
	}
	return m.styles.Input.Render(value)
}

// Helper methods

func (m *ConnectModel) isSQLite() bool {
	dbType := m.dbTypeInput.Value()
	return dbType == "" || dbType == "sqlite" || dbType == "SQLite"
}

func (m *ConnectModel) getMaxFieldIndex() int {
	if m.isSQLite() {
		return 1 // dbType and path
	}
	return 6 // all PostgreSQL fields
}

func (m *ConnectModel) adjustFocusForDbType() {
	maxIndex := m.getMaxFieldIndex()
	if m.focusIndex > maxIndex {
		m.focusIndex = maxIndex
		m.blurAllFields()
		m.focusCurrentField()
	}
}

func (m *ConnectModel) nextField() {
	m.blurCurrentField()
	maxIndex := m.getMaxFieldIndex()
	m.focusIndex++
	if m.focusIndex > maxIndex {
		m.focusIndex = 0
	}
	m.focusCurrentField()
}

func (m *ConnectModel) prevField() {
	m.blurCurrentField()
	maxIndex := m.getMaxFieldIndex()
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = maxIndex
	}
	m.focusCurrentField()
}

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

func (m *ConnectModel) blurAllFields() {
	m.dbTypeInput.Blur()
	m.pathInput.Blur()
	m.hostInput.Blur()
	m.portInput.Blur()
	m.userInput.Blur()
	m.passwordInput.Blur()
	m.databaseInput.Blur()
}

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

func (m *ConnectModel) startConnection() tea.Cmd {
	m.state = StateConnecting
	m.errorMsg = ""

	// Start spinner
	spinnerCmd := m.spinner.Tick

	// Start connection
	config := m.getConfig()
	connectCmd := func() tea.Msg {
		database, err := db.Open(config)
		if err != nil {
			return ConnectErrorMsg{Err: err.Error()}
		}
		return ConnectSuccessMsg{DB: database}
	}

	return tea.Batch(spinnerCmd, connectCmd)
}

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

// Message types

type ConnectSuccessMsg struct {
	DB db.Database
}

type ConnectErrorMsg struct {
	Err string
}
