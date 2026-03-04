// Package screens contains the main application screens for Tornado.
//
// This file implements the Connection screen - the first screen users see.
// It provides a clean, minimal interface with connection history.
package screens

import (
	"math"
	"strings"
	"time"

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
)

// ConnectionItem represents a saved connection in the history list
type ConnectionItem struct {
	entry config.ConnectionEntry
}

func (i ConnectionItem) FilterValue() string { return i.entry.Name }
func (i ConnectionItem) Title() string       { return i.entry.Name }
func (i ConnectionItem) Description() string {
	return i.entry.Path
}

// ConnectModel is the model for the connection screen.
type ConnectModel struct {
	// State
	state ConnectionState

	// Form fields
	pathInput textinput.Model

	// Connection history
	showHistory    bool
	connectionList list.Model
	connections    []config.ConnectionEntry

	// UI state
	errorMsg string

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

	// Initialize form fields
	path := textinput.New()
	path.Placeholder = "/path/to/database.db"
	path.CharLimit = 256

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

	connList := list.New(connItems, list.NewDefaultDelegate(), 40, 5)
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
		pathInput:      path,
		showHistory:    len(connections) > 0,
		connectionList: connList,
		connections:    connections,
	}

	// Pre-fill with most recent connection if available
	if len(connections) > 0 {
		m.pathInput.SetValue(connections[0].Path)
	}

	return m
}

// Init returns the initial command for the connection screen.
func (m *ConnectModel) Init() tea.Cmd {
	return connectAnimTick()
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
				m.pathInput.Focus()
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}

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
				m.pathInput.Focus()
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

	// Pass messages to connection list if showing
	if m.state == StateForm && m.showHistory {
		var cmd tea.Cmd
		newListModel, cmd := m.connectionList.Update(msg)
		m.connectionList = newListModel
		return m, cmd
	}

	// Pass messages to path input
	if m.state == StateForm {
		var cmd tea.Cmd
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *ConnectModel) handleFormKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateWelcome
		m.pathInput.Blur()
		m.errorMsg = ""
		m.showHistory = false
		return m, nil

	case "ctrl+h":
		// Toggle connection history
		if len(m.connections) > 0 {
			m.showHistory = !m.showHistory
			return m, nil
		}

	case "enter":
		// Handle history list selection
		if m.showHistory {
			if item, ok := m.connectionList.SelectedItem().(ConnectionItem); ok {
				m.pathInput.SetValue(item.entry.Path)
				m.showHistory = false
				return m, nil
			}
		}

		// Connect
		return m, m.startConnection()

	default:
		// Pass to path input for editing (including paste)
		var cmd tea.Cmd
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	}

	return m, nil
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
	default:
		content = m.viewWelcome()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// viewFormScreen renders the form dialog in the bottom right
func (m *ConnectModel) viewFormScreen() string {
	// Render the welcome background first
	background := m.viewWelcomeBackground()

	// Render the form dialog
	dialog := m.viewForm()

	// Place dialog in bottom right corner
	return placeDialogBottomRight(background, dialog, m.width, m.height)
}

// viewConnectingScreen renders the connecting dialog in bottom right
func (m *ConnectModel) viewConnectingScreen() string {
	// Render the welcome background first
	background := m.viewWelcomeBackground()

	// Render the connecting dialog
	dialog := m.viewConnecting()

	// Place dialog in bottom right corner
	return placeDialogBottomRight(background, dialog, m.width, m.height)
}

// viewWelcomeBackground returns a solid background with logo and help (no animation)
func (m *ConnectModel) viewWelcomeBackground() string {
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

// truncateToWidthNoEllipsis truncates a string to fit within width without adding ellipsis
func truncateToWidthNoEllipsis(s string, width int) string {
	if width < 1 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= width {
		return s
	}
	result := ""
	for _, r := range s {
		runeW := lipgloss.Width(string(r))
		if lipgloss.Width(result)+runeW > width {
			break
		}
		result += string(r)
	}
	return result
}

// spliceLineStyled replaces characters in baseLine starting at column x with overlay content,
// preserving the overlay's ANSI styling.
func spliceLineStyled(baseLine, overlay string, x, totalWidth int) string {
	if x < 0 {
		x = 0
	}
	if x >= totalWidth {
		return baseLine
	}

	overlayW := lipgloss.Width(overlay)
	if overlayW == 0 {
		return baseLine
	}

	// Calculate left padding needed to reach position x
	baseW := lipgloss.Width(baseLine)
	var leftPart string
	if baseW <= x {
		// Base line is shorter than x, pad with spaces
		leftPart = baseLine + strings.Repeat(" ", x-baseW)
	} else {
		// Truncate base line at position x (no ellipsis)
		leftPart = truncateToWidthNoEllipsis(baseLine, x)
	}

	// Calculate right part starting after the overlay
	rightStart := x + overlayW
	var rightPart string
	if rightStart < totalWidth {
		// Get the part of baseLine after the overlay position
		remaining := getFromWidth(baseLine, rightStart)
		// Pad or truncate to fill remaining width (no ellipsis)
		remainingW := lipgloss.Width(remaining)
		needed := totalWidth - rightStart
		if remainingW >= needed {
			rightPart = truncateToWidthNoEllipsis(remaining, needed)
		} else {
			rightPart = remaining + strings.Repeat(" ", needed-remainingW)
		}
	}

	return leftPart + overlay + rightPart
}

// getFromWidth returns the substring of s starting at the specified visual width
func getFromWidth(s string, startWidth int) string {
	if startWidth <= 0 {
		return s
	}
	result := ""
	currentWidth := 0
	for _, r := range s {
		runeW := lipgloss.Width(string(r))
		if currentWidth >= startWidth {
			result += string(r)
		} else if currentWidth+runeW > startWidth {
			// This rune spans the boundary, skip it
			currentWidth += runeW
		} else {
			currentWidth += runeW
		}
	}
	return result
}

func (m *ConnectModel) viewForm() string {
	bodyWidth := 50
	fieldWidth := bodyWidth - 4
	var fields []string

	// Show connection history if available and toggled
	if m.showHistory && len(m.connections) > 0 {
		m.connectionList.SetWidth(fieldWidth)
		m.connectionList.SetHeight(5)
		historySection := m.styles.Subheader.Render("Recent Connections (Ctrl+H)")
		fields = append(fields, historySection)
		fields = append(fields, m.connectionList.View())
		fields = append(fields, "")
	}

	// Path input field - just show the input without extra borders
	pathLabel := m.styles.Muted.Render("Database File:")
	pathValue := m.pathInput.View()
	fields = append(fields, pathLabel+"\n"+pathValue)

	if m.errorMsg != "" {
		fields = append(fields, m.styles.Error.Render(m.truncateError(m.errorMsg, fieldWidth)))
	}

	helpText := "enter Connect • esc Cancel"
	if len(m.connections) > 0 && !m.showHistory {
		helpText += " • ctrl+h History"
	}

	return renderDialogBox("Connect to Database", fields, helpText, bodyWidth)
}

func (m *ConnectModel) viewConnecting() string {
	body := []string{}
	if m.errorMsg != "" {
		body = append(body,
			m.styles.Error.Render("Connection Failed"),
			m.truncateError(m.errorMsg, 46),
			"",
			m.styles.Muted.Render("Press any key to return"),
		)
		return renderDialogBox("Connecting", body, "any key Back", 50)
	}

	frames := []string{"⎛", "⎜", "⎝", "⎞", "⎟", "⎠"}
	frame := frames[m.spinnerFrame%len(frames)]
	body = append(body,
		lipgloss.NewStyle().Foreground(styles.Primary).Render(frame)+"  Connecting to database...",
		"",
		m.styles.Muted.Render("Please wait"),
	)
	return renderDialogBox("Connecting", body, "esc Disabled", 50)
}

// placeDialogBottomRight places a dialog box in the bottom right corner of the screen
func placeDialogBottomRight(background, dialog string, width, height int) string {
	dialogLines := strings.Split(dialog, "\n")
	dialogWidth := 0
	for _, line := range dialogLines {
		if w := lipgloss.Width(line); w > dialogWidth {
			dialogWidth = w
		}
	}
	dialogHeight := len(dialogLines)

	// Position in bottom right with some padding
	padding := 2
	x := width - dialogWidth - padding
	y := height - dialogHeight - padding

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Composite the dialog onto the background at position (x, y)
	baseLines := strings.Split(background, "\n")
	// Pad base to full height with empty lines (no background color)
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	for i, dialogLine := range dialogLines {
		row := y + i
		if row >= len(baseLines) {
			break
		}
		baseLines[row] = spliceLineStyled(baseLines[row], dialogLine, x, width)
	}

	return strings.Join(baseLines, "\n")
}

// spliceStringAt replaces characters in baseLine starting at column x with overlay content
func spliceStringAt(baseLine, overlay string, x, totalWidth int) string {
	if x >= totalWidth {
		return baseLine
	}

	// Pad baseLine to totalWidth if needed
	baseW := lipgloss.Width(baseLine)
	if baseW < totalWidth {
		baseLine += strings.Repeat(" ", totalWidth-baseW)
	}

	overlayW := lipgloss.Width(overlay)

	// Build result: left part of base + overlay + right part of base
	// Simple truncation without ellipsis for left part
	leftPart := ""
	w := 0
	for _, r := range baseLine {
		runeW := lipgloss.Width(string(r))
		if w+runeW > x {
			break
		}
		leftPart += string(r)
		w += runeW
	}
	rightStart := x + overlayW
	rightPart := ""
	if rightStart < totalWidth {
		// Get substring from rightStart to end
		w := 0
		start := 0
		for _, r := range baseLine {
			if w >= rightStart {
				break
			}
			start++
			w += lipgloss.Width(string(r))
		}

		result := ""
		w = 0
		for i, r := range baseLine {
			if i >= start {
				if w >= totalWidth-rightStart {
					break
				}
				result += string(r)
				w += lipgloss.Width(string(r))
			}
		}
		rightPart = result
	}

	return leftPart + overlay + rightPart
}

func (m *ConnectModel) startConnection() tea.Cmd {
	m.state = StateConnecting
	m.errorMsg = ""
	m.spinnerFrame = 0
	m.showHistory = false
	m.pathInput.Blur()

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
	return models.ConnectionConfig{
		Type: "sqlite",
		Path: m.pathInput.Value(),
	}
}

// truncateError truncates error message to fit within width without ellipsis
func (m *ConnectModel) truncateError(s string, width int) string {
	if width < 1 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= width {
		return s
	}
	// Simple truncation without ellipsis
	runes := []rune(s)
	result := ""
	for _, r := range runes {
		if lipgloss.Width(result+string(r)) > width {
			break
		}
		result += string(r)
	}
	return result
}

// Message types
type ConnectSuccessMsg struct {
	DB db.Database
}

type ConnectErrorMsg struct {
	Err string
}
