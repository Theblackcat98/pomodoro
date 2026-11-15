package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"pomodoro/src/models"
	"pomodoro/src/services"
)

// keyPress builds a v2 key press message the way the runtime does.
func keyPress(s string) tea.KeyPressMsg {
	switch s {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	default:
		return tea.KeyPressMsg{Code: []rune(s)[0], Text: s}
	}
}

func driveKey(m Model, s string) Model {
	next, _ := m.Update(keyPress(s))
	return next.(Model)
}

func TestV2TabNavigation(t *testing.T) {
	m := newBareModel()
	if m.tab != TabTimer {
		t.Fatalf("initial tab = %d, want timer", m.tab)
	}
	m = driveKey(m, "tab")
	if m.tab != TabTasks {
		t.Fatalf("after tab = %d, want tasks", m.tab)
	}
	m = driveKey(m, "tab")
	if m.tab != TabStats {
		t.Fatalf("after 2nd tab = %d, want stats", m.tab)
	}
	m = driveKey(m, "shift+tab")
	if m.tab != TabTasks {
		t.Fatalf("after shift+tab = %d, want tasks", m.tab)
	}
	m = driveKey(m, "?")
	if m.tab != TabHelp {
		t.Fatalf("after ? = %d, want help", m.tab)
	}
	m = driveKey(m, "esc")
	if m.tab != TabTimer {
		t.Fatalf("after esc = %d, want timer", m.tab)
	}
}

func TestV2SpaceStartsFocus(t *testing.T) {
	m := newBareModel()
	m = driveKey(m, "space")
	if !m.sessionRunning() {
		t.Fatal("space must start a session")
	}
	s := m.timer.GetCurrentSession()
	if s == nil || s.Type != models.FocusSession {
		t.Fatalf("session after space = %+v", s)
	}
	m.timer.Shutdown()
}

func newBareModel() Model {
	return NewModel(services.NewTimerService(nil), nil)
}
