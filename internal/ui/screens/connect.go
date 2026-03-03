// Package screens contains the main application screens for Tornado.
//
// This file implements the Connection screen - the first screen users see.
// It provides a clean, minimal interface with connection history and file browser.
package screens

import (
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jupiterozeye/tornado/internal/assets"
	"github.com/jupiterozeye/tornado/internal/config"
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
	StateFilePicker
)

// DBTypeItem represents a database type option for the list
type DBTypeItem struct {
	name        string
	description string
}

func (i DBTypeItem) FilterValue() string { return i.name }
func (i DBTypeItem) Title() string       { return i.name }
func (i DBTypeItem) Description() string { return i.description }

// ConnectionItem represents a saved connection in the history list
type ConnectionItem struct {
	entry config.ConnectionEntry
}

func (i ConnectionItem) FilterValue() string { return i.entry.Name }
func (i ConnectionItem) Title() string       { return i.entry.Name }
func (i ConnectionItem) Description() string {
	if i.entry.Type == "sqlite" {
		return "SQLite: " + i.entry.Path
	}
	return "PostgreSQL: " + i.entry.Host
}

// ConnectModel is the model for the connection screen.
type ConnectModel struct {
	// State
	state ConnectionState

	// Form fields
	dbTypeList    list.Model
	pathInput     textinput.Model
	hostInput     textinput.Model
	portInput     textinput.Model
	userInput     textinput.Model
	passwordInput textinput.Model
	databaseInput textinput.Model

	// Connection history
	showHistory    bool
	connectionList list.Model
	connections    []config.ConnectionEntry

	// File picker
	filepicker     filepicker.Model
	showFilePicker bool

	// UI state
	focusIndex int // 0=dbType, 1=path, 2=host, 3=port, 4=user, 5=password, 6=database
	showDbList bool
	errorMsg   string
	selectedDb string

	// Dimensions
	width  int
	height int

	// Styling
	styles *styles.Styles

	// Loading spinner
	spinner      spinner.Model
	spinnerFrame int
	animT        float64
}

type connectAnimTickMsg time.Time

func connectAnimTick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return connectAnimTickMsg(t)
	})
}

// NewConnectModel creates a new connection screen model.
func NewConnectModel() *ConnectModel {
	s := styles.Default()

	// Initialize spinner with custom parenthsis spinner
	sp := spinner.New()
	sp.Spinner = spinner.Spinner{
		Frames: []string{"⎛", "⎜", "⎝", "⎞", "⎟", "⎠"},
		FPS:    time.Second / 8,
	}
	sp.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	// Initialize DB type list
	dbItems := []list.Item{
		DBTypeItem{name: "SQLite", description: "Local file-based database"},
		DBTypeItem{name: "PostgreSQL", description: "Network database server"},
	}

	dbList := list.New(dbItems, list.NewDefaultDelegate(), 40, 6)
	dbList.Title = "Database Type"
	dbList.SetShowStatusBar(false)
	dbList.SetShowHelp(false)
	dbList.SetShowPagination(false)
	dbList.SetFilteringEnabled(false)
	dbList.SetShowTitle(false)

	// Initialize form fields
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

	// Initialize file picker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".db", ".sqlite", ".sqlite3", ".db3"}
	homeDir, _ := os.UserHomeDir()
	fp.CurrentDirectory = homeDir

	// Load connection history
	connections := []config.ConnectionEntry{}
	if cfg := config.Get(); cfg != nil {
		connections = cfg.GetConnections()
	}

	// Create connection history list
	var connItems []list.Item
	for _, conn := range connections {
		connItems = append(connItems, ConnectionItem{entry: conn})
	}

	connList := list.New(connItems, list.NewDefaultDelegate(), 50, 6)
	connList.Title = "Recent Connections"
	connList.SetShowStatusBar(false)
	connList.SetShowHelp(false)
	connList.SetShowPagination(false)
	connList.SetFilteringEnabled(false)
	connList.SetShowTitle(true)

	m := &ConnectModel{
		state:          StateWelcome,
		styles:         s,
		spinner:        sp,
		dbTypeList:     dbList,
		pathInput:      path,
		hostInput:      host,
		portInput:      port,
		userInput:      user,
		passwordInput:  password,
		databaseInput:  database,
		filepicker:     fp,
		focusIndex:     0,
		showDbList:     true,
		selectedDb:     "SQLite",
		showHistory:    len(connections) > 0,
		connectionList: connList,
		connections:    connections,
	}

	// Pre-fill with most recent connection if available
	if len(connections) > 0 {
		m.fillFormFromHistory(connections[0])
	}

	return m
}

// fillFormFromHistory fills the form fields from a connection entry
func (m *ConnectModel) fillFormFromHistory(entry config.ConnectionEntry) {
	if entry.Type == "sqlite" {
		m.selectedDb = "SQLite"
		dbIndex := 0
		m.dbTypeList.Select(dbIndex)
		m.pathInput.SetValue(entry.Path)
	} else {
		m.selectedDb = "PostgreSQL"
		dbIndex := 1
		m.dbTypeList.Select(dbIndex)
		m.hostInput.SetValue(entry.Host)
		if entry.Port != 0 {
			m.portInput.SetValue(strconv.Itoa(entry.Port))
		}
		m.userInput.SetValue(entry.User)
		m.databaseInput.SetValue(entry.Database)
	}
}

// Init returns the initial command for the connection screen.
func (m *ConnectModel) Init() tea.Cmd {
	return tea.Batch(
		connectAnimTick(),
		m.filepicker.Init(),
	)
}

// Update handles messages for the connection screen.
func (m *ConnectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch m.state {
		case StateWelcome:
			switch msg.String() {
			case "space":
				m.state = StateForm
				m.focusIndex = 0
				m.showDbList = true
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}

		case StateFilePicker:
			switch msg.String() {
			case "esc", "q":
				m.state = StateForm
				m.showFilePicker = false
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}

			var cmd tea.Cmd
			m.filepicker, cmd = m.filepicker.Update(msg)

			// Check if user selected a file
			if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
				m.pathInput.SetValue(path)
				m.state = StateForm
				m.showFilePicker = false
				return m, nil
			}

			return m, cmd

		case StateForm:
			return m.handleFormKeys(msg)

		case StateConnecting:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			// If there's an error showing, any key returns to form
			if m.errorMsg != "" {
				m.state = StateForm
				m.errorMsg = ""
				return m, nil
			}
		}

	case spinner.TickMsg:
		if m.state == StateConnecting {
			m.spinnerFrame++
			return m, m.spinner.Tick
		}

	case connectAnimTickMsg:
		m.animT += 0.06
		return m, connectAnimTick()

	// Pass through connection messages so they bubble up to App
	case ConnectSuccessMsg:
		// Save successful connection to history
		if cfg := config.Get(); cfg != nil {
			cfg.AddConnection(m.getConfig())
		}
		return m, func() tea.Msg { return msg }

	case ConnectErrorMsg:
		m.errorMsg = msg.Err
		return m, func() tea.Msg { return msg }
	}

	// Pass messages to active lists
	if m.state == StateForm {
		if m.showDbList && m.focusIndex == 0 {
			var cmd tea.Cmd
			newListModel, cmd := m.dbTypeList.Update(msg)
			m.dbTypeList = newListModel

			// Check if an item was selected
			if item, ok := m.dbTypeList.SelectedItem().(DBTypeItem); ok {
				if item.name != m.selectedDb {
					m.selectedDb = item.name
					// Reset other fields when switching DB types
					if m.selectedDb == "SQLite" {
						m.focusIndex = 0
					}
				}
			}

			return m, cmd
		}

		if m.showHistory {
			var cmd tea.Cmd
			newListModel, cmd := m.connectionList.Update(msg)
			m.connectionList = newListModel
			return m, cmd
		}
	}

	return m, nil
}

func (m *ConnectModel) handleFormKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateWelcome
		m.blurAllFields()
		m.errorMsg = ""
		m.showHistory = false
		return m, nil

	case "tab":
		m.nextField()
		return m, nil

	case "shift+tab":
		m.prevField()
		return m, nil

	case "up":
		if m.focusIndex == 0 && m.showDbList {
			// Let the list handle it
			return m, nil
		}
		m.prevField()
		return m, nil

	case "down":
		if m.focusIndex == 0 && m.showDbList {
			// Let the list handle it
			return m, nil
		}
		m.nextField()
		return m, nil

	case "enter":
		// Handle history list selection
		if m.showHistory {
			if item, ok := m.connectionList.SelectedItem().(ConnectionItem); ok {
				m.fillFormFromHistory(item.entry)
				m.showHistory = false
				return m, nil
			}
		}

		if m.focusIndex == 0 && m.showDbList {
			// Select DB type and move to next field
			m.showDbList = false
			m.nextField()
			return m, nil
		}

		// Handle path field - open file picker for SQLite
		if m.focusIndex == 1 && m.isSQLite() {
			m.state = StateFilePicker
			m.showFilePicker = true
			return m, nil
		}

		if m.focusIndex == m.getMaxFieldIndex() {
			return m, m.startConnection()
		}
		m.nextField()
		return m, nil

	case "ctrl+b":
		// Browse button for SQLite path
		if m.isSQLite() {
			m.state = StateFilePicker
			m.showFilePicker = true
			return m, nil
		}

	case "ctrl+h":
		// Toggle connection history
		if len(m.connections) > 0 {
			m.showHistory = !m.showHistory
			return m, nil
		}

	default:
		// Update the focused text input
		return m.updateFocusedInput(msg)
	}

	return m, nil
}

func (m *ConnectModel) updateFocusedInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focusIndex {
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

// View renders the connection screen.
func (m *ConnectModel) View() tea.View {
	var content string
	switch m.state {
	case StateWelcome:
		content = m.viewWelcome()
	case StateForm:
		content = m.viewFormScreen()
	case StateConnecting:
		content = m.viewConnectingScreen()
	case StateFilePicker:
		content = m.viewFilePickerScreen()
	default:
		content = m.viewWelcome()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// viewFormScreen renders the form dialog on a solid background
func (m *ConnectModel) viewFormScreen() string {
	dialog := m.viewForm()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(styles.BgDefault)))
}

// viewConnectingScreen renders the connecting dialog on a solid background
func (m *ConnectModel) viewConnectingScreen() string {
	dialog := m.viewConnecting()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(styles.BgDefault)))
}

// viewFilePickerScreen renders the file picker modal overlay
func (m *ConnectModel) viewFilePickerScreen() string {
	// First render the form as background
	background := m.viewFormScreen()

	// Then render the file picker as a modal on top
	pickerContent := m.viewFilePicker()

	// Composite the picker on top of the background
	return compositeOverlay(background, pickerContent, m.width, m.height)
}

// viewFilePicker renders the file picker dialog
func (m *ConnectModel) viewFilePicker() string {
	bodyWidth := 60

	var content []string

	// Title
	title := m.styles.Header.Render("Select Database File")
	content = append(content, title)
	content = append(content, "")

	// Current directory
	dirStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)
	content = append(content, dirStyle.Render("Current: "+m.filepicker.CurrentDirectory))
	content = append(content, "")

	// File picker view
	pickerView := m.filepicker.View()
	content = append(content, pickerView)

	// Help
	content = append(content, "")
	helpStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)
	content = append(content, helpStyle.Render("enter: select • esc: cancel"))

	body := strings.Join(content, "\n")

	// Build dialog box
	return renderDialogBox("File Browser", []string{body}, "ctrl+c quit", bodyWidth)
}

func (m *ConnectModel) viewWelcome() string {
	logoStyle := lipgloss.NewStyle().Foreground(styles.Primary)
	logo := logoStyle.Render(assets.Logo)
	anim := m.renderTornadoAnimation()

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		MarginTop(2)
	help := helpStyle.Render("Space: Connect | Ctrl+C: Quit")

	fullLogo := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logo)
	fullHelp := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	content := lipgloss.JoinVertical(lipgloss.Left, fullLogo, anim, fullHelp)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

var tornadoAnimLines = []string{
	"                          ██████                            ",
	"          ██████████                          ████████      ",
	"    ██████████                                    ████████  ",
	"  ████████      ████                      ██████      ██████",
	"████████    ████                              ████    ██████",
	"██████    ████    ██████    ████████    ██      ██████████  ",
	"  ██████  ████  ████      ████████████    ██  ██████████    ",
	"    ████████████  ████                  ████████████    ██  ",
	"        ████  ██████████████████████████████████    ██████  ",
	"          ████                                  ████████    ",
	"            ████████    ██████████████████████████████      ",
	"                  ████████                            ██    ",
	"              ████      ██████████████████████████████      ",
	"                ████████        ██████████████              ",
	"                    ████████████              ██████        ",
	"                            ██████████████████████          ",
	"                      ██████                                ",
	"                        ████████████████████                ",
	"                      ██      ████████████                  ",
	"                        ██████                              ",
	"                          ██████████████                    ",
	"                      ██      ██████                        ",
	"                      ██████                                ",
	"                        ████████████                        ",
	"                      ██                                    ",
	"                        ██████                              ",
	"                          ██                                ",
}

func (m *ConnectModel) renderTornadoAnimation() string {
	if m.width == 0 {
		return ""
	}

	var out []string
	n := len(tornadoAnimLines)
	for i, line := range tornadoAnimLines {
		funnel := math.Pow(float64(i)/float64(maxConnectInt(1, n-1)), 1.2)
		topWeight := 1.0 - funnel
		amp := 5.0*topWeight + 2.0*funnel
		phase := m.animT*3.1 + float64(i)*0.55
		sway := int(math.Sin(phase) * amp)
		pad := (m.width-lipgloss.Width(line))/2 + sway
		if pad < 0 {
			pad = 0
		}
		styled := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(line)
		if i < 2 {
			styled = lipgloss.NewStyle().Foreground(styles.Primary).Render(line)
		}
		lineOut := strings.Repeat(" ", pad) + styled
		lineOut = padToVisualWidth(lineOut, m.width)
		out = append(out, lineOut)
	}
	return strings.Join(out, "\n")
}

func padToVisualWidth(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func maxConnectInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *ConnectModel) viewForm() string {
	isSQLite := m.isSQLite()
	bodyWidth := 52
	fieldWidth := bodyWidth - 2
	var fields []string

	// Show connection history if available and toggled
	if m.showHistory && len(m.connections) > 0 {
		m.connectionList.SetWidth(fieldWidth - 4)
		m.connectionList.SetHeight(6)
		historySection := m.styles.Subheader.Render("Recent Connections (Ctrl+H to toggle)")
		fields = append(fields, historySection)
		fields = append(fields, m.connectionList.View())
		fields = append(fields, "")
	}

	if m.showDbList && m.focusIndex == 0 {
		m.dbTypeList.SetWidth(fieldWidth - 4)
		m.dbTypeList.SetHeight(4)
		fields = append(fields, m.renderFieldContainer("Database Type", m.dbTypeList.View(), true, fieldWidth))
	} else {
		dbValue := m.selectedDb
		if m.focusIndex == 0 {
			dbValue = "▾ " + dbValue
		}
		fields = append(fields, m.renderFieldContainer("Database Type", dbValue, m.focusIndex == 0, fieldWidth))
	}

	if isSQLite {
		// Path field with browse button hint
		pathValue := m.displayInput(&m.pathInput, m.focusIndex == 1)
		pathLabel := "Database File"
		if m.focusIndex == 1 {
			pathLabel += " (Ctrl+B to browse)"
		}
		fields = append(fields, m.renderFieldContainer(pathLabel, pathValue, m.focusIndex == 1, fieldWidth))
	} else {
		fields = append(fields,
			m.renderFieldContainer("Host", m.displayInput(&m.hostInput, m.focusIndex == 2), m.focusIndex == 2, fieldWidth),
			m.renderFieldContainer("Port", m.displayInput(&m.portInput, m.focusIndex == 3), m.focusIndex == 3, fieldWidth),
			m.renderFieldContainer("Username", m.displayInput(&m.userInput, m.focusIndex == 4), m.focusIndex == 4, fieldWidth),
			m.renderFieldContainer("Password", m.displayInput(&m.passwordInput, m.focusIndex == 5), m.focusIndex == 5, fieldWidth),
			m.renderFieldContainer("Database", m.displayInput(&m.databaseInput, m.focusIndex == 6), m.focusIndex == 6, fieldWidth),
		)
	}

	if m.errorMsg != "" {
		fields = append(fields, m.styles.Error.Render(truncateToWidth(m.errorMsg, fieldWidth-2)))
	}

	helpText := "tab Next · shift+tab Prev · esc Cancel"
	if m.focusIndex == 0 && m.showDbList {
		helpText = "up/down Select · enter Confirm · esc Cancel"
	} else if m.focusIndex == m.getMaxFieldIndex() {
		helpText = "enter Connect · tab Previous · esc Cancel"
	}

	if len(m.connections) > 0 && !m.showHistory {
		helpText += " · ctrl+h History"
	}

	if isSQLite && m.focusIndex == 1 {
		helpText = "ctrl+b Browse · " + helpText
	}

	return renderDialogBox("Connect to Database", fields, helpText, bodyWidth+2)
}

func (m *ConnectModel) viewConnecting() string {
	body := []string{}
	if m.errorMsg != "" {
		body = append(body,
			m.styles.Error.Render("Connection Failed"),
			truncateToWidth(m.errorMsg, 50),
			"",
			m.styles.Muted.Render("Press any key to return"),
		)
		return renderDialogBox("Connecting", body, "any key Back", 54)
	}

	frames := []string{"⎛", "⎜", "⎝", "⎞", "⎟", "⎠"}
	frame := frames[m.spinnerFrame%len(frames)]
	body = append(body,
		lipgloss.NewStyle().Foreground(styles.Primary).Render(frame)+"  Connecting to database...",
		"",
		m.styles.Muted.Render("Please wait"),
	)
	return renderDialogBox("Connecting", body, "esc Disabled while connecting", 54)
}

func (m *ConnectModel) renderFieldContainer(label, content string, focused bool, width int) string {
	if width < 8 {
		width = 8
	}
	innerWidth := width - 2
	borderColor := styles.Border
	if focused {
		borderColor = styles.BorderFocus
	}
	borderStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Background(styles.BgDark)
	bodyStyle := lipgloss.NewStyle().Background(styles.BgDark)

	top := makeDialogTopBorder(label, innerWidth)
	contentLines := strings.Split(content, "\n")
	if len(contentLines) == 0 {
		contentLines = []string{""}
	}
	bottom := strings.Repeat("─", innerWidth)

	out := []string{borderStyle.Render("┌" + top + "┐")}
	for _, contentLine := range contentLines {
		line := truncateToWidth(contentLine, innerWidth)
		line = padToWidth(line, innerWidth)
		out = append(out, borderStyle.Render("│")+bodyStyle.Render(line)+borderStyle.Render("│"))
	}
	out = append(out, borderStyle.Render("└"+bottom+"┘"))

	return strings.Join(out, "\n")
}

func (m *ConnectModel) displayInput(input *textinput.Model, focused bool) string {
	if focused {
		return input.View()
	}
	if input.Value() != "" {
		return input.View()
	}
	return input.Placeholder
}

// Helper methods

func (m *ConnectModel) isSQLite() bool {
	return strings.EqualFold(m.selectedDb, "SQLite")
}

func (m *ConnectModel) getMaxFieldIndex() int {
	if m.isSQLite() {
		return 1 // dbType and path
	}
	return 6 // all PostgreSQL fields
}

func (m *ConnectModel) nextField() {
	m.blurCurrentField()
	maxIndex := m.getMaxFieldIndex()
	m.focusIndex++
	if m.focusIndex > maxIndex {
		m.focusIndex = 0
		m.showDbList = true
	}
	m.focusCurrentField()
}

func (m *ConnectModel) prevField() {
	m.blurCurrentField()
	maxIndex := m.getMaxFieldIndex()
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = maxIndex
		m.showDbList = false
	}
	if m.focusIndex == 0 {
		m.showDbList = true
	}
	m.focusCurrentField()
}

func (m *ConnectModel) blurCurrentField() {
	switch m.focusIndex {
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
	m.pathInput.Blur()
	m.hostInput.Blur()
	m.portInput.Blur()
	m.userInput.Blur()
	m.passwordInput.Blur()
	m.databaseInput.Blur()
}

func (m *ConnectModel) focusCurrentField() {
	switch m.focusIndex {
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
	m.spinnerFrame = 0
	m.showHistory = false

	// Get config
	config := m.getConfig()

	// Start spinner animation
	spinnerCmd := tea.Tick(time.Second/8, func(t time.Time) tea.Msg {
		return spinner.TickMsg{}
	})

	return tea.Batch(
		spinnerCmd,
		func() tea.Msg {
			// Attempt connection
			database, err := db.Open(config)
			if err != nil {
				return ConnectErrorMsg{Err: err.Error()}
			}
			return ConnectSuccessMsg{DB: database}
		},
	)
}

func (m *ConnectModel) getConfig() models.ConnectionConfig {
	port, _ := strconv.Atoi(m.portInput.Value())
	if port == 0 {
		port = 5432
	}

	dbType := "sqlite"
	if !m.isSQLite() {
		dbType = "postgres"
	}

	return models.ConnectionConfig{
		Type:     dbType,
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
