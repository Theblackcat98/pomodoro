package views

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/lipgloss"
)

// Style definitions based on the lipgloss example
var (
	// General colors
	normal    = lipgloss.Color("#EEEEEE")
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	// Base style
	base = lipgloss.NewStyle().Foreground(normal)

	// Divider
	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	// Tabs
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Padding(0, 1)

	activeTab = tab.Border(activeTabBorder, true)

	tabGap = tab.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	// Title
	titleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(5).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#874BFD")).
			SetString("Pomodoro Timer")

	// Timer display
	timerBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(2).
			Width(30).
			Height(10).
			Margin(1, 2)

	// Button styles
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				MarginRight(2).
				Underline(true)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"}).
			Width(80).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	statusTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})
)

func (m Model) MainView() string {
	// Create tabs
	var timerTab lipgloss.Style
	var tasksTab lipgloss.Style

	if m.CurrentView == "timer" {
		timerTab = activeTab
		tasksTab = tab
	} else {
		timerTab = tab
		tasksTab = activeTab
	}

	tabs := lipgloss.JoinHorizontal(
		lipgloss.Top,
		timerTab.Render("Timer"),
		tasksTab.Render("Tasks"),
	)

	// Add gap to fill the width
	gap := tabGap.Render(strings.Repeat(" ", max(0, m.Width-lipgloss.Width(tabs)-2)))
	tabs = lipgloss.JoinHorizontal(lipgloss.Bottom, tabs, gap)

	// Title
	title := titleStyle.String()

	// Main content area
	var content string
	if m.CurrentView == "timer" {
		content = m.renderTimer()
	} else {
		content = m.renderTasks()
	}

	// Combine header
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		tabs,
		title,
	)

	// Status bar
	var statusText string
	switch m.Session.Status {
	case "running":
		statusText = "RUNNING"
	case "paused":
		statusText = "PAUSED"
	default:
		statusText = "READY"
	}

	status := statusStyle.Render(statusText)
	keyHelp := statusTextStyle.Width(m.Width - lipgloss.Width(status)).Render("[Tab] Switch | [Space] Start/Pause | [Q] Quit")
	bar := lipgloss.JoinHorizontal(lipgloss.Top, status, keyHelp)
	statusBar := statusBarStyle.Width(m.Width).Render(bar)

	// Combine all sections
	doc := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		statusBar,
	)

	return doc
}

func (m Model) renderTimer() string {
	var status string
	var button string

	switch m.Session.Status {
	case "running":
		status = "Focus Time - Running"
		button = activeButtonStyle.Render("PAUSE")
	case "paused":
		status = "Focus Time - Paused"
		button = activeButtonStyle.Render("START")
	default:
		status = "Focus Time - Ready"
		button = activeButtonStyle.Render("START")
	}

	// Format time properly using the session's remaining time
	totalSeconds := int(m.Session.Remaining.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	timerBox := timerBoxStyle.Render(
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
		Margin(1, 2).
		Render("Task List:\n\nNo tasks yet.\nPress 'n' to add a new task.")

	return tasksBox
}

// Helper function for calculating max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}