package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ButtonStatus represents the status of the button
type ButtonStatus string

const (
	ButtonIdle    ButtonStatus = "idle"
	ButtonPressed ButtonStatus = "pressed"
	ButtonHovered ButtonStatus = "hovered"
)

// TimerControl represents the START/PAUSE button control
type TimerControl struct {
	status     ButtonStatus
	isActive   bool
	label      string
	buttonStyle lipgloss.Style
	activeStyle lipgloss.Style
	width      int
	height     int
}

// NewTimerControl creates a new timer control button
func NewTimerControl() *TimerControl {
	buttonStyle := lipgloss.NewStyle().
		Padding(1, 3).
		Background(lipgloss.Color("#626262")).
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true)
	
	activeStyle := lipgloss.NewStyle().
		Padding(1, 3).
		Background(lipgloss.Color("#ff6b6b")).
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true)
	
	return &TimerControl{
		status:      ButtonIdle,
		isActive:    false,
		label:       "START",
		buttonStyle: buttonStyle,
		activeStyle: activeStyle,
		width:       15,
		height:      3,
	}
}

// Init initializes the timer control
func (tc *TimerControl) Init() tea.Cmd {
	return nil
}

// Update handles messages for the timer control. Keys only toggle the button
// while it is focused, so stray Space/Enter presses elsewhere never trigger it.
func (tc *TimerControl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !tc.IsFocused() {
			return tc, nil
		}
		switch msg.Type {
		case tea.KeyEnter, tea.KeySpace:
			tc.toggle()
		}
	}

	return tc, nil
}

// View renders the timer control button
func (tc *TimerControl) View() string {
	var styledButton string
	
	if tc.isActive {
		styledButton = tc.activeStyle.Render(tc.label)
	} else {
		styledButton = tc.buttonStyle.Render(tc.label)
	}
	
	return styledButton
}

// toggle switches the button state between START and PAUSE
func (tc *TimerControl) toggle() {
	tc.isActive = !tc.isActive
	if tc.isActive {
		tc.label = "PAUSE"
	} else {
		tc.label = "START"
	}
}

// IsActive returns whether the timer is currently active (running)
func (tc *TimerControl) IsActive() bool {
	return tc.isActive
}

// SetActive sets the active state of the button
func (tc *TimerControl) SetActive(active bool) {
	tc.isActive = active
	if active {
		tc.label = "PAUSE"
	} else {
		tc.label = "START"
	}
}

// SetLabel updates the button label
func (tc *TimerControl) SetLabel(label string) {
	tc.label = label
}

// Focus handles focus behavior (required by Bubble Tea)
func (tc *TimerControl) Focus() tea.Cmd {
	tc.status = ButtonHovered
	return nil
}

// Blur handles blur behavior (required by Bubble Tea)
func (tc *TimerControl) Blur() {
	tc.status = ButtonIdle
}

// IsFocused returns whether the button is focused
func (tc *TimerControl) IsFocused() bool {
	return tc.status == ButtonHovered
}

// SetSize sets the size of the button
func (tc *TimerControl) SetSize(width, height int) {
	tc.width = width
	tc.height = height
}

// GetPreferredSize returns the preferred size of the button
func (tc *TimerControl) GetPreferredSize() (int, int) {
	return tc.width, tc.height
}

// GetStatus returns the current button status
func (tc *TimerControl) GetStatus() ButtonStatus {
	return tc.status
}

// GetLabel returns the button label
func (tc *TimerControl) GetLabel() string {
	return tc.label
}