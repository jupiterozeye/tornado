package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewConnectModel(t *testing.T) {
	m := NewConnectModel()

	if m == nil {
		t.Fatal("NewConnectModel() returned nil")
	}

	// Check initial state
	if m.state != StateWelcome {
		t.Errorf("state = %v, want %v", m.state, StateWelcome)
	}

	if m.focusIndex != 0 {
		t.Errorf("focusIndex = %v, want %v", m.focusIndex, 0)
	}

	if m.selectedDb != "SQLite" {
		t.Errorf("selectedDb = %v, want %v", m.selectedDb, "SQLite")
	}
}

func TestConnectModel_Init(t *testing.T) {
	m := NewConnectModel()
	cmd := m.Init()

	// Init should return nil now
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestConnectModel_Navigation(t *testing.T) {
	m := NewConnectModel()
	m.state = StateForm
	m.showDbList = false

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
}

func TestConnectModel_Update_WelcomeState(t *testing.T) {
	m := NewConnectModel()

	// Test Space key to open form
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	cm := newM.(*ConnectModel)

	if cm.state != StateForm {
		t.Errorf("Space should switch to StateForm, got state = %v", cm.state)
	}
	if cmd != nil {
		t.Error("Space should not return a command")
	}

	// Test Ctrl+C - this is handled at the app level in app.go
	// The connect screen doesn't handle Ctrl+C directly
	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	_ = newM
	_ = cmd
}

func TestConnectModel_Update_FormState(t *testing.T) {
	m := NewConnectModel()
	m.state = StateForm
	m.showDbList = false

	// Test Tab key
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	cm := newM.(*ConnectModel)
	if cm.focusIndex != 1 {
		t.Errorf("Tab should move to next field, got focusIndex = %v", cm.focusIndex)
	}
	if cmd != nil {
		t.Error("Tab should not return a command")
	}

	// Test Esc key to go back to welcome
	newM, cmd = cm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	cm = newM.(*ConnectModel)
	if cm.state != StateWelcome {
		t.Errorf("Esc should switch to StateWelcome, got state = %v", cm.state)
	}
}

func TestConnectModel_isSQLite(t *testing.T) {
	tests := []struct {
		name       string
		selectedDb string
		want       bool
	}{
		{"SQLite selected", "SQLite", true},
		{"sqlite lowercase", "sqlite", true},
		{"PostgreSQL selected", "PostgreSQL", false},
		{"postgres lowercase", "postgres", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewConnectModel()
			m.selectedDb = tt.selectedDb
			if got := m.isSQLite(); got != tt.want {
				t.Errorf("isSQLite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnectModel_getMaxFieldIndex(t *testing.T) {
	tests := []struct {
		name       string
		selectedDb string
		want       int
	}{
		{"SQLite", "SQLite", 1},
		{"PostgreSQL", "PostgreSQL", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewConnectModel()
			m.selectedDb = tt.selectedDb
			if got := m.getMaxFieldIndex(); got != tt.want {
				t.Errorf("getMaxFieldIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnectModel_getConfig(t *testing.T) {
	m := NewConnectModel()
	m.selectedDb = "SQLite"
	m.pathInput.SetValue("/test/db.sqlite")

	config := m.getConfig()

	if config.Type != "sqlite" {
		t.Errorf("config.Type = %v, want %v", config.Type, "sqlite")
	}
	if config.Path != "/test/db.sqlite" {
		t.Errorf("config.Path = %v, want %v", config.Path, "/test/db.sqlite")
	}

	// Test PostgreSQL config
	m.selectedDb = "PostgreSQL"
	m.hostInput.SetValue("localhost")
	m.portInput.SetValue("5432")
	m.userInput.SetValue("admin")
	m.passwordInput.SetValue("secret")
	m.databaseInput.SetValue("mydb")

	config = m.getConfig()

	if config.Type != "postgres" {
		t.Errorf("config.Type = %v, want %v", config.Type, "postgres")
	}
	if config.Host != "localhost" {
		t.Errorf("config.Host = %v, want %v", config.Host, "localhost")
	}
	if config.Port != 5432 {
		t.Errorf("config.Port = %v, want %v", config.Port, 5432)
	}
	if config.User != "admin" {
		t.Errorf("config.User = %v, want %v", config.User, "admin")
	}
	if config.Password != "secret" {
		t.Errorf("config.Password = %v, want %v", config.Password, "secret")
	}
	if config.Database != "mydb" {
		t.Errorf("config.Database = %v, want %v", config.Database, "mydb")
	}
}
