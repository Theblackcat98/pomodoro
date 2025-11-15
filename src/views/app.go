package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"pomodoro/src/models"
)

type Model struct {
	Session models.TimerSession
	CurrentView string
	Width int
	Height int
}

func InitialModel() Model {
	return Model{
		Session: models.TimerSession{},
		CurrentView: "timer", // Default to timer view
		Width: 80,
		Height: 24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			// Switch between views
			if m.CurrentView == "timer" {
				m.CurrentView = "tasks"
			} else {
				m.CurrentView = "timer"
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	// Update the model based on session status changes
	// For now, we'll just return the model as is
	return m, nil
}

func (m Model) View() string {
	return m.MainView()
}