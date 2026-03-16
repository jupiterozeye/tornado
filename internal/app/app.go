// Package app contains the root application model and screen navigation logic.
// References:
//   - https://charm.land/bubbletea/v2#tutorial
//   - https://guide.elm-lang.org/architecture/
package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jupiterozeye/tornado/internal/config"
	"github.com/jupiterozeye/tornado/internal/db"
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
)

// String returns a human-readable name for the screen.
func (s Screen) String() string {
	switch s {
	case ScreenConnect:
		return "Connect"
	case ScreenBrowser:
		return "Browser"
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
	connectScreen *screens.ConnectModel
	browserScreen *screens.BrowserModel
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
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		default:
			// Pass all other keys to the active screen
			return a.delegateToActiveScreen(msg)
		}
	case tea.WindowSizeMsg:
		// Store dimensions for responsive layout
		a.width = msg.Width
		a.height = msg.Height
		// Propagate to active screen
		return a.delegateToActiveScreen(msg)

	case screens.ConnectSuccessMsg:
		// Initialize browser screen now that we have DB
		a.db = msg.DB
		a.browserScreen = screens.NewBrowserModel(a.db)

		// Ensure the browser gets current dimensions immediately.
		// Without this it can stay in a "Loading..." state waiting for a resize.
		w := a.width
		h := a.height
		if w == 0 {
			w = 80
		}
		if h == 0 {
			h = 24
		}
		resizedModel, _ := a.browserScreen.Update(tea.WindowSizeMsg{Width: w, Height: h})
		a.browserScreen = resizedModel.(*screens.BrowserModel)

		a.currentScreen = ScreenBrowser
		return a, a.browserScreen.Init()

	case ScreenChangeMsg:
		// custom message for explicit screen switching
		a.currentScreen = msg.Screen
		return a, a.getActiveScreen().Init()

	case screens.RequestConnectMsg:
		// Cancel background operations and grab DB reference before nulling it
		if a.browserScreen != nil {
			a.browserScreen.Cleanup()
		}
		oldDB := a.db
		a.db = nil
		a.browserScreen = nil

		// Fire-and-forget disconnect outside Bubble Tea's command pipeline.
		if oldDB != nil {
			go func() { _ = oldDB.Disconnect() }()
		}

		// Load connections snapshot RIGHT NOW on the event-loop goroutine.
		// This is safe because AddQuery's Lock is only held for microseconds,
		// and we are not inside any lock ourselves here.
		// Using NewConnectModelWithConnections avoids calling GetConnections
		// later when a background goroutine might hold the write lock.
		var connections []config.ConnectionEntry
		if cfg := config.Get(); cfg != nil {
			connections = cfg.GetConnections()
		}
		a.connectScreen = screens.NewConnectModelWithConnections(connections)

		// Pass current window size to the new connect screen
		if a.width > 0 && a.height > 0 {
			a.connectScreen.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
		}
		a.currentScreen = ScreenConnect
		return a, a.connectScreen.Init()

	case ErrorMsg:
		a.err = msg.Err
		return a, nil

	default:
		// Pass all other messages to current screen
		return a.delegateToActiveScreen(msg)
	}
}

// Helper methods

func (a *App) getActiveScreen() tea.Model {
	switch a.currentScreen {
	case ScreenBrowser:
		if a.browserScreen != nil {
			return a.browserScreen
		}
		return a.connectScreen
	default:
		return a.connectScreen
	}
}

func (a *App) delegateToActiveScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var newModel tea.Model

	switch a.currentScreen {
	case ScreenConnect:
		if a.connectScreen == nil {
			return a, nil
		}
		newModel, cmd = a.connectScreen.Update(msg)
		a.connectScreen = newModel.(*screens.ConnectModel)
	case ScreenBrowser:
		if a.browserScreen == nil {
			return a, nil
		}
		newModel, cmd = a.browserScreen.Update(msg)
		a.browserScreen = newModel.(*screens.BrowserModel)
	}

	return a, cmd
}

// View renders the current state of the application.
func (a *App) View() tea.View {
	// Get active screen's view
	v := a.getActiveScreen().View()

	if a.err != nil {
		// Show error overlay
		v.Content = lipgloss.JoinVertical(
			lipgloss.Left,
			v.Content,
			a.styles.Error.Render("Error: "+a.err.Error()),
		)
	}

	return v
}

// ScreenChangeMsg is a message for transitioning between screens.
type ScreenChangeMsg struct {
	Screen Screen
}

// ErrorMsg represents a global error to display to the user.
type ErrorMsg struct {
	Err error
}
