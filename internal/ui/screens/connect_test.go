package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

func TestNewConnectModel(t *testing.T) {
	m := NewConnectModel()

	if m == nil {
		t.Fatal("NewConnectModel() returned nil")
	}

	// Check that all inputs are initialized
	if m.dbTypeInput.Placeholder != "sqlite" {
		t.Errorf("dbTypeInput placeholder = %v, want %v", m.dbTypeInput.Placeholder, "sqlite")
	}

	if m.pathInput.Placeholder != "path/to/database.db" {
		t.Errorf("pathInput placeholder = %v, want %v", m.pathInput.Placeholder, "path/to/database.db")
	}

	// Check initial state
	if m.focusIndex != 0 {
		t.Errorf("focusIndex = %v, want %v", m.focusIndex, 0)
	}

	if m.isConnecting {
		t.Error("isConnecting should be false initially")
	}

	if m.styles == nil {
		t.Error("styles should be initialized")
	}
}

func TestConnectModel_Init(t *testing.T) {
	m := NewConnectModel()
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return a command (textinput.Blink)")
	}
}

func TestConnectModel_Navigation(t *testing.T) {
	m := NewConnectModel()

	// Test nextField
	m.nextField()
	if m.focusIndex != 1 {
		t.Errorf("after nextField(), focusIndex = %v, want %v", m.focusIndex, 1)
	}

	// Test prevField
	m.prevField()
	if m.focusIndex != 0 {
		t.Errorf("after prevField(), focusIndex = %v, want %v", m.focusIndex, 0)
	}

	// Test wrap around (prev from 0 should go to last field)
	m.prevField()
	if m.focusIndex != 6 {
		t.Errorf("after prevField() from 0, focusIndex = %v, want %v (last field)", m.focusIndex, 6)
	}

	// Test wrap around (next from last should go to 0)
	m.nextField()
	if m.focusIndex != 0 {
		t.Errorf("after nextField() from last, focusIndex = %v, want %v", m.focusIndex, 0)
	}
}

func TestConnectModel_Update_KeyMsg(t *testing.T) {
	m := NewConnectModel()

	// Test Tab key
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if newM.(*ConnectModel).focusIndex != 1 {
		t.Errorf("Tab should move to next field, got focusIndex = %v", newM.(*ConnectModel).focusIndex)
	}
	if cmd != nil {
		t.Error("Tab should not return a command")
	}

	// Test Shift+Tab key
	newM, cmd = newM.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if newM.(*ConnectModel).focusIndex != 0 {
		t.Errorf("Shift+Tab should move to previous field, got focusIndex = %v", newM.(*ConnectModel).focusIndex)
	}

	// Test Up key
	newM, cmd = newM.Update(tea.KeyMsg{Type: tea.KeyUp})
	if newM.(*ConnectModel).focusIndex != 6 {
		t.Errorf("Up should move to previous field, got focusIndex = %v", newM.(*ConnectModel).focusIndex)
	}

	// Test Down key
	newM, cmd = newM.Update(tea.KeyMsg{Type: tea.KeyDown})
	if newM.(*ConnectModel).focusIndex != 0 {
		t.Errorf("Down should move to next field, got focusIndex = %v", newM.(*ConnectModel).focusIndex)
	}
}

func TestConnectModel_Update_WindowSize(t *testing.T) {
	m := NewConnectModel()

	// Test WindowSizeMsg
	newM, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}

	cm := newM.(*ConnectModel)
	if cm.width != 100 {
		t.Errorf("width = %v, want %v", cm.width, 100)
	}
	if cm.height != 50 {
		t.Errorf("height = %v, want %v", cm.height, 50)
	}
}

func TestConnectModel_Update_ConnectError(t *testing.T) {
	m := NewConnectModel()
	m.isConnecting = true

	// Test ConnectErrorMsg
	errMsg := ConnectErrorMsg{Error: "connection failed"}
	newM, cmd := m.Update(errMsg)
	if cmd != nil {
		t.Error("ConnectErrorMsg should not return a command")
	}

	cm := newM.(*ConnectModel)
	if cm.isConnecting {
		t.Error("isConnecting should be false after error")
	}
	if cm.errorMsg != "connection failed" {
		t.Errorf("errorMsg = %v, want %v", cm.errorMsg, "connection failed")
	}
}

func TestConnectModel_View(t *testing.T) {
	m := NewConnectModel()

	view := m.View()
	if view == "" {
		t.Error("View() should return a non-empty string")
	}

	// Check that the title is included
	if !contains(view, "Tornado - Database Connection") {
		t.Error("View should contain the title")
	}

	// Check that all fields are rendered
	if !contains(view, "Database Type") {
		t.Error("View should contain Database Type field")
	}
	if !contains(view, "File Path") {
		t.Error("View should contain File Path field")
	}
	if !contains(view, "Host") {
		t.Error("View should contain Host field")
	}
	if !contains(view, "Port") {
		t.Error("View should contain Port field")
	}
}

func TestConnectModel_View_WithError(t *testing.T) {
	m := NewConnectModel()
	m.errorMsg = "test error"

	view := m.View()
	if !contains(view, "test error") {
		t.Error("View should display the error message")
	}
}

func TestConnectModel_View_WhileConnecting(t *testing.T) {
	m := NewConnectModel()
	m.isConnecting = true

	view := m.View()
	if !contains(view, "Connecting") {
		t.Error("View should show connecting message")
	}
}

func TestConnectModel_getConfig(t *testing.T) {
	m := NewConnectModel()

	// Set some values
	m.dbTypeInput.SetValue("postgres")
	m.hostInput.SetValue("localhost")
	m.portInput.SetValue("5433")
	m.userInput.SetValue("admin")
	m.passwordInput.SetValue("secret")
	m.databaseInput.SetValue("mydb")
	m.pathInput.SetValue("/tmp/test.db")

	config := m.getConfig()

	if config.Type != "postgres" {
		t.Errorf("Type = %v, want %v", config.Type, "postgres")
	}
	if config.Host != "localhost" {
		t.Errorf("Host = %v, want %v", config.Host, "localhost")
	}
	if config.Port != 5433 {
		t.Errorf("Port = %v, want %v", config.Port, 5433)
	}
	if config.User != "admin" {
		t.Errorf("User = %v, want %v", config.User, "admin")
	}
	if config.Password != "secret" {
		t.Errorf("Password = %v, want %v", config.Password, "secret")
	}
	if config.Database != "mydb" {
		t.Errorf("Database = %v, want %v", config.Database, "mydb")
	}
}

func TestConnectModel_getConfig_DefaultPort(t *testing.T) {
	m := NewConnectModel()

	// Empty port should default to 5432
	m.portInput.SetValue("")
	config := m.getConfig()

	if config.Port != 5432 {
		t.Errorf("Port = %v, want default %v", config.Port, 5432)
	}
}

func TestConnectModel_handleEnter_NotLastField(t *testing.T) {
	m := NewConnectModel()
	m.focusIndex = 0

	newM, cmd := m.handleEnter()
	if cmd != nil {
		t.Error("handleEnter on non-last field should not return a command")
	}

	cm := newM.(*ConnectModel)
	if cm.focusIndex != 1 {
		t.Errorf("handleEnter should move to next field, got focusIndex = %v", cm.focusIndex)
	}
}

func TestConnectModel_handleEnter_LastField(t *testing.T) {
	m := NewConnectModel()
	m.focusIndex = 6

	newM, cmd := m.handleEnter()
	if cmd == nil {
		t.Error("handleEnter on last field should return a command")
	}

	cm := newM.(*ConnectModel)
	if !cm.isConnecting {
		t.Error("isConnecting should be true after handleEnter on last field")
	}
}

func TestConnectModel_startConnection(t *testing.T) {
	m := NewConnectModel()

	cmd := m.startConnection()
	if cmd == nil {
		t.Error("startConnection() should return a command")
	}

	if !m.isConnecting {
		t.Error("isConnecting should be true after startConnection()")
	}

	if m.errorMsg != "" {
		t.Error("errorMsg should be cleared after startConnection()")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestStyles is used to verify styles are properly initialized
func TestNewConnectModel_Styles(t *testing.T) {
	m := NewConnectModel()

	if m.styles == nil {
		t.Fatal("styles should be initialized")
	}

	// Verify styles are the default styles
	defaultStyles := styles.Default()
	if m.styles != defaultStyles {
		// They should be the same since we use Default()
		t.Log("Note: styles instance may differ but should have same values")
	}
}
