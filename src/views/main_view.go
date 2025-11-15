package views

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#2580B3")).
		Padding(0, 1)

	descStyle = lipgloss.NewStyle().
		Margin(1, 0, 2, 0)

	buttonStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#336AE2")).
		Padding(0, 3).
		Margin(1, 1)

	inactiveButtonStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#969B86")).
		Background(lipgloss.Color("#E6E7E8")).
		Padding(0, 3).
		Margin(1, 1)
)

func (m Model) MainView() string {
	title := titleStyle.Render("Pomodoro Timer")
	desc := descStyle.Render("Focus on your tasks with the Pomodoro Technique")

	var timerDisplay string
	if m.CurrentView == "timer" {
		timerDisplay = m.renderTimer()
	} else {
		timerDisplay = m.renderTasks()
	}

	help := "\n\n[Tab] Switch View | [Space] Start/Pause | [Q] Quit"

	return strings.Join([]string{
		title,
		desc,
		timerDisplay,
		help,
	}, "\n")
}

func (m Model) renderTimer() string {
	var status string
	var button string

	switch m.Session.Status {
	case "running":
		status = "Focus Time"
		button = buttonStyle.Render("PAUSE")
	case "paused":
		status = "Paused"
		button = buttonStyle.Render("START")
	default:
		status = "Focus Time"
		button = buttonStyle.Render("START")
	}

	timeStr := fmt.Sprintf("%02d:%02d", 25, 0) // Placeholder for actual time remaining

	timerBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#336AE2")).
		Padding(2).
		Width(30).
		Height(10).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				status,
				timeStr,
				button,
			),
		)

	return timerBox
}

func (m Model) renderTasks() string {
	tasksBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6124DC")).
		Padding(2).
		Width(50).
		Height(15).
		Render("Task List:\n\nNo tasks yet.\nPress 'n' to add a new task.")

	return tasksBox
}