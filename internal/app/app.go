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
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yourusername/tornado/internal/db"
	"github.com/yourusername/tornado/internal/models"
	"github.com/yourusername/tornado/internal/ui/screens"
	"github.com/yourusername/tornado/internal/ui/styles"
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
// TODO: Implement this for debugging/logging purposes.
func (s Screen) String() string {
	// TODO: Return string representation
	return ""
}

// App is the root model of the application.
// It manages screen navigation and holds global state.
//
// TODO: Add fields for:
//   - Current database connection (db.Database interface)
//   - Current active screen
//   - Screen models (connect, browser, query, dashboard)
//   - Global styles reference
//   - Window dimensions (for responsive layouts)
//   - Error state (for global error display)
type App struct {
	// TODO: Add these fields
	// currentScreen   Screen
	// width           int
	// height          int
	// db              db.Database
	// styles          *styles.Styles
	//
	// Screen models - each is a separate Bubble Tea model
	// connectScreen   *screens.ConnectModel
	// browserScreen   *screens.BrowserModel
	// queryScreen     *screens.QueryModel
	// dashboardScreen *screens.DashboardModel
}

// New creates a new App instance with default values.
// This is the "initial model" in Elm Architecture terms.
//
// TODO: Initialize all screens with their default state
// TODO: Load styles from styles package
// TODO: Accept optional configuration from main()
func New() *App {
	return &App{
		// TODO: Initialize fields
	}
}

// Init returns the initial command for the application.
// This is called once when the program starts.
//
// Common patterns:
//   - Return nil if no initial I/O is needed
//   - Return tea.Batch(screen1.Init(), screen2.Init()) to init multiple things
//   - Return a command that loads config or checks for database files
//
// TODO: Decide what should happen on startup:
//   - Show connect screen immediately? (return nil)
//   - Try to reconnect to last database? (return reconnect command)
//   - Check for update? (return version check command)
func (a *App) Init() tea.Cmd {
	// TODO: Return appropriate initialization commands
	// Example: Initialize the connect screen
	// return a.connectScreen.Init()
	return nil
}

// Update handles incoming messages and updates the model accordingly.
// This is where you handle:
//   - Global keybindings (quit, help, navigation)
//   - Screen transition messages
//   - Window resize events
//   - Messages from child screens
//
// Key Learning - Message Handling:
//
//	Messages can be any type! Common patterns:
//	- tea.KeyMsg: Keyboard input
//	- tea.WindowSizeMsg: Terminal resize
//	- Custom messages: Your own structs for app-specific events
//
// TODO: Implement message handling:
//   - [ ] Handle tea.KeyMsg for global shortcuts (q to quit)
//   - [ ] Handle tea.WindowSizeMsg to update dimensions
//   - [ ] Handle custom ScreenChangeMsg for navigation
//   - [ ] Delegate to active screen's Update for screen-specific messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: Type switch on msg to handle different message types
	//
	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	//     // Handle global keybindings
	//     if msg.String() == "q" || msg.String() == "ctrl+c" {
	//         return a, tea.Quit
	//     }
	//
	// case tea.WindowSizeMsg:
	//     // Update dimensions and propagate to screens
	//     a.width = msg.Width
	//     a.height = msg.Height
	//
	// case ScreenChangeMsg:
	//     // Handle screen transitions
	//     a.currentScreen = msg.Screen
	//     return a, a.activeScreen().Init()
	// }
	//
	// // Delegate to active screen
	// switch a.currentScreen {
	// case ScreenConnect:
	//     return a.connectScreen.Update(msg)
	// case ScreenBrowser:
	//     return a.browserScreen.Update(msg)
	// // ... etc
	// }

	return a, nil
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
