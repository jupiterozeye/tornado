package screens

import (
	"testing"

	tea "charm.land/bubbletea/v2"
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
}

func TestConnectModel_Init(t *testing.T) {
	m := NewConnectModel()
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return animation tick command")
	}
}

func TestConnectModel_Update_WelcomeState(t *testing.T) {
	m := NewConnectModel()

	// Test Space key to open form
	newM, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	cm := newM.(*ConnectModel)

	if cm.state != StateForm {
		t.Errorf("Space should switch to StateForm, got state = %v", cm.state)
	}
	if cmd != nil {
		t.Error("Space should not return a command")
	}

	// Test Ctrl+C - this is handled at the app level in app.go
	// The connect screen doesn't handle Ctrl+C directly
	newM, cmd = m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	_ = newM
	_ = cmd
}

func TestConnectModel_Update_FormState(t *testing.T) {
	m := NewConnectModel()
	m.state = StateForm
	m.pathInput.Focus()

	// Test Esc key to go back to welcome
	newM, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	cm := newM.(*ConnectModel)
	if cm.state != StateWelcome {
		t.Errorf("Esc should switch to StateWelcome, got state = %v", cm.state)
	}
	if cmd != nil {
		t.Error("Esc should not return a command")
	}
}

func TestConnectModel_getConfig(t *testing.T) {
	m := NewConnectModel()
	m.pathInput.SetValue("/test/db.sqlite")

	config := m.getConfig()

	if config.Type != "sqlite" {
		t.Errorf("config.Type = %v, want %v", config.Type, "sqlite")
	}
	if config.Path != "/test/db.sqlite" {
		t.Errorf("config.Path = %v, want %v", config.Path, "/test/db.sqlite")
	}
}
