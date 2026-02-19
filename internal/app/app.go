// Package app contains the root application model and screen navigation logic.
//
// This is the heart of your Tornado application. The App struct is the
// top-level Model in the Elm Architecture. It manages:
//   - Which screen is currently active (Connect, Browser, Query, Dashboard)
//   - Global application state (current database connection, settings)
//   - Navigation between screens
//
// The App acts as a "router" - it delegates Update and View calls to the
// active screen, and handles screen transitions via messages.
//
// TODO: Implement the following:
//   - [ ] Define Screen type for type-safe screen identifiers
//   - [ ] Implement screen transition logic in Update
//   - [ ] Create message types for screen changes
//   - [ ] Handle window resize events globally
//   - [ ] Manage global keybindings (quit, help, etc.)
//
// Key Learning - The Elm Architecture:
//
//	The pattern is: Model -> View -> User Action -> Update -> Model (repeat)
//	- Model: Your application state (structs)
//	- View: Renders the model to a string
//	- Update: Takes a message and model, returns new model and commands
//	- Cmd: Asynchronous operations (I/O, timers, etc.)
//
// Key Learning - Composition:
//
//	This app is composed of smaller models (screens). Each screen has its
//	own Init/Update/View methods. The App delegates to the active screen.
//
// References:
//   - https://github.com/charmbracelet/bubbletea#tutorial
//   - https://guide.elm-lang.org/architecture/
package app

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/ui/screens"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// Screen represents the different views/modes of the application.
// TODO: Consider using an int or string type for screen identifiers.
// Using a type alias makes the code more readable and type-safe.
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

}

// View renders the current state of the application.
// It delegates rendering to the active screen.
//
// Key Learning - View Functions:
//   - View should be a pure function of the model
//   - Don't do I/O or calculations in View
//   - Use Lip Gloss for styling and layout
//   - The returned string is printed to the terminal
//
// TODO: Implement View:
//   - [ ] Call active screen's View()
//   - [ ] Add status bar at the bottom
//   - [ ] Handle error display overlay
//   - [ ] Use lipgloss.JoinVertical to compose layout
func (a *App) View() string {
	// TODO: Implement view composition
	//
	// Example structure:
	// 1. Get the active screen's view
	// 2. Add a status bar at the bottom
	// 3. Add an error overlay if there's an error
	//
	// return lipgloss.JoinVertical(
	//     lipgloss.Left,
	//     a.activeScreen().View(),
	//     a.renderStatusBar(),
	// )

	return "Tornado - Database TUI\n\nPress q to quit"
}

// activeScreen returns the model for the currently displayed screen.
// TODO: Implement this helper method
func (a *App) activeScreen() tea.Model {
	// TODO: Return the appropriate screen model based on a.currentScreen
	return nil
}

// ScreenChangeMsg is a message for transitioning between screens.
// TODO: Move this to a separate messages.go file if you have many message types.
type ScreenChangeMsg struct {
	Screen Screen
}

// ConnectSuccessMsg is sent when a database connection is established.
// This triggers a transition from Connect screen to Browser screen.
// TODO: Define this message and handle it in Update
type ConnectSuccessMsg struct {
	DB db.Database
}

// ErrorMsg represents a global error to display to the user.
// TODO: Implement error display in View
type ErrorMsg struct {
	Err error
}
