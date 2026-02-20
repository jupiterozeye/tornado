// Package app contains the root application model and screen navigation logic.
// References:
//   - https://github.com/charmbracelet/bubbletea#tutorial
//   - https://guide.elm-lang.org/architecture/
package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jupiterozeye/tornado/internal/db"
	//"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/ui/screens"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// Screen represents the different views/modes of the application.
type Screen int

const (
	// ScreenConnect is the initial screen for database connection
	ScreenConnect Screen = iota
	// ScreenBrowser is for browsing tables and schemas
	ScreenBrowser
	// ScreenQuery is the SQL query editor and results viewer
	ScreenQuery
	// ScreenDashboard shows traffic metrics and charts
	ScreenDashboard
)

// String returns a human-readable name for the screen.
func (s Screen) String() string {
	switch s {
	case ScreenConnect:
		return "Connect"
	case ScreenBrowser:
		return "Browser"
	case ScreenQuery:
		return "Query"
	case ScreenDashboard:
		return "Dashboard"
	default:
		return "Unknown"
	}
}

// App is the root model of the application.
// It manages screen navigation and holds global state.
type App struct {
	currentScreen Screen
	width         int
	height        int
	db            db.Database
	styles        *styles.Styles
	err           error

	// Screen models - each is a separate Bubble Tea model
	connectScreen   *screens.ConnectModel
	browserScreen   *screens.BrowserModel
	queryScreen     *screens.QueryModel
	dashboardScreen *screens.DashboardModel
}

// New creates a new App instance with default values.
func New() *App {
	return &App{
		currentScreen: ScreenConnect,
		styles:        styles.Default(),
		connectScreen: screens.NewConnectModel(),
	}
}

// Init returns the initial command for the application.
func (a *App) Init() tea.Cmd {
	return a.connectScreen.Init()
}

// Update handles incoming messages and updates the model accordingly.
// This is where you handle:
//   - Global keybindings (quit, help, navigation)
//   - Screen transition messages
//   - Window resize events
//   - Messages from child screens
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "tab":
			a.switchToNextScreen()
			return a, nil
		}
	case tea.WindowSizeMsg:
		// Store dimensions for responsive lasyout
		a.width = msg.Width
		a.height = msg.Height
		// Propagate to active screen
		return a.delegateToActiveScreen(msg)

	case ConnectSuccessMsg:
		// Initialize other screens now that we have DB
		a.db = msg.DB
		a.browserScreen = screens.NewBrowserModel(a.db)
		a.queryScreen = screens.NewQueryModel(a.db)
		a.dashboardScreen = screens.NewDashboardModel(a.db)
		a.currentScreen = ScreenBrowser
		return a, a.browserScreen.Init()

	case ScreenChangeMsg:
		// custom message for explicit screen switching
		a.currentScreen = msg.Screen
		return a, a.getActiveScreen().Init()

	case ErrorMsg:
		a.err = msg.Err
		return a, nil

	default:
		// Pass all other messages to current screen
		return a.delegateToActiveScreen(msg)
	}

	return a, nil
}

// Helper methods

func (a *App) switchToNextScreen() {
	switch a.currentScreen {
	case ScreenConnect:
		// Cant switch from connect until connected
		return
	case ScreenBrowser:
		a.currentScreen = ScreenQuery
	case ScreenQuery:
		a.currentScreen = ScreenDashboard
	case ScreenDashboard:
		a.currentScreen = ScreenBrowser
	}
}

func (a *App) getActiveScreen() tea.Model {
	switch a.currentScreen {
	case ScreenConnect:
		return a.connectScreen
	case ScreenQuery:
		return a.queryScreen
	case ScreenDashboard:
		return a.dashboardScreen
	default:
		return a.connectScreen
	}
}

func (a *App) delegateToActiveScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var newModel tea.Model

	switch a.currentScreen {
	case ScreenConnect:
		newModel, cmd = a.connectScreen.Update(msg)
		a.connectScreen = newModel.(*screens.ConnectModel)
	case ScreenBrowser:
		newModel, cmd = a.browserScreen.Update(msg)
		a.browserScreen = newModel.(*screens.BrowserModel)
	case ScreenQuery:
		newModel, cmd = a.queryScreen.Update(msg)
		a.queryScreen = newModel.(*screens.QueryModel)
	case ScreenDashboard:
		newModel, cmd = a.dashboardScreen.Update(msg)
		a.dashboardScreen = newModel.(*screens.DashboardModel)

	}

	return a, cmd

}

// View renders the current state of the application.
// It delegates rendering to the active screen.
func (a *App) View() string {
	// Get active screens content
	content := a.getActiveScreen().View()

	// Render status bar
	statusBar := a.renderStatusBar()

	if a.err != nil {
		// Show error overlay
		return lipgloss.JoinVertical(
			lipgloss.Left,
			a.getActiveScreen().View(),
			a.styles.Error.Render("Error: "+a.err.Error()),
		)

	}
	// Join vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		statusBar,
	)
}

func (a *App) renderStatusBar() string {
	// Show current screen and shortcuts
	screenName := a.currentScreen.String()
	return a.styles.StatusBar.Render(
		fmt.Sprintf(" %s | Tab: Next | q: Quit", screenName),
	)
}

// ScreenChangeMsg is a message for transitioning between screens.
type ScreenChangeMsg struct {
	Screen Screen
}

// ConnectSuccessMsg is sent when a database connection is established.
type ConnectSuccessMsg struct {
	DB db.Database
}

// ErrorMsg represents a global error to display to the user.
type ErrorMsg struct {
	Err error
}
