package views

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
)

// View composes the whole UI: header (brand + tabs + today count), the active
// view's content centered in the remaining space, and the status bar.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	header := m.renderHeader()
	status := m.renderStatusBar()
	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(status)
	if contentHeight < 1 {
		contentHeight = 1
	}

	var body string
	switch m.tab {
	case TabTasks:
		body = m.renderTasksView()
	case TabStats:
		body = m.renderStatsView()
	case TabHelp:
		body = m.renderHelpView()
	default:
		body = m.renderTimerView()
	}

	// Completion flash banner rides above the content.
	if m.flash != "" {
		accent := m.currentAccent()
		banner := flashStyle.Foreground(accent).BorderForeground(accent).Render(m.flash)
		body = lipgloss.JoinVertical(lipgloss.Center, banner, "", body)
	}

	if m.width > 0 {
		body = lipgloss.Place(m.width, contentHeight, lipgloss.Center, lipgloss.Center, body)
	}

	if m.confirmQuit {
		body = m.renderQuitConfirm()
	}

	doc := lipgloss.JoinVertical(lipgloss.Left, header, body, status)
	return appStyle.Render(doc)
}

// currentAccent is the accent color for the session currently on screen.
func (m Model) currentAccent() lipgloss.Color {
	session := m.timer.GetCurrentSession()
	sessionType := models.FocusSession
	status := models.SessionStatus("")
	if session != nil {
		sessionType = session.Type
		if session.Status == models.Running || session.Status == models.Paused {
			status = session.Status
		}
	}
	return sessionAccent(sessionType, status)
}

// renderHeader draws brand, tabs, and today's pomodoro count.
func (m Model) renderHeader() string {
	brand := brandStyle.Render("🍅 Pomodoro")

	var tabs string
	for i, name := range []string{"Timer", "Tasks", "Stats", "Help"} {
		style := tabStyle
		if Tab(i) == m.tab {
			style = activeTabStyle.Foreground(m.currentAccent())
		}
		tabs += style.Render(name)
	}

	right := toastStyle.Render("🍅 " + itoa(m.todayCount()) + " today")

	brandWidth := lipgloss.Width(brand)
	tabsWidth := lipgloss.Width(tabs)
	rightWidth := lipgloss.Width(right)
	gap := max(1, m.width-brandWidth-tabsWidth-rightWidth-4)

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		brand,
		strings.Repeat(" ", 2),
		tabs,
		strings.Repeat(" ", gap),
		right,
	)
	return headerStyle.Render(header)
}

// renderStatusBar draws the state chip, contextual key hints, and the clock.
func (m Model) renderStatusBar() string {
	chipText, chipColor := m.statusChip()
	chip := statusChipStyle.
		Background(chipColor).
		Foreground(lipgloss.Color("#FFF7DB")).
		Render(chipText)

	hints := m.contextualHints()

	clock := statusHintStyle.Render(m.clock.Format("15:04"))

	used := lipgloss.Width(chip) + 1 + lipgloss.Width(hints) + lipgloss.Width(m.toast) + 1 + lipgloss.Width(clock)
	gap := max(1, m.width-used)

	bar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		chip,
		" ",
		hints,
		strings.Repeat(" ", gap),
		m.toast,
		" ",
		clock,
	)
	// Pin to the terminal width so the padding is included in the line
	// (otherwise the last characters wrap to the next row).
	return statusBarStyle.Width(m.width).Render(bar)
}

// statusChip returns the label + color for the left status chip.
func (m Model) statusChip() (string, lipgloss.Color) {
	session := m.timer.GetCurrentSession()
	switch {
	case session == nil || session.Status == models.Completed:
		return "○ READY", colorNeutral
	case session.Status == models.Paused:
		return "‖ PAUSED", colorPaused
	case session.Type == models.FocusSession:
		return "● RUNNING", colorFocus
	default:
		return "☕ ON BREAK", sessionAccent(session.Type, session.Status)
	}
}

// contextualHints returns the key hints valid for the current screen and mode.
func (m Model) contextualHints() string {
	switch {
	case m.confirmQuit:
		return hintJoin(hint("y", "quit"), hint("n/esc", "stay"))
	case m.deleteArm:
		return hintJoin(hint("y", "delete"), hint("n/esc", "cancel"))
	case m.adding:
		return hintJoin(hint("tab", "field"), hint("+/-", "estimate"), hint("enter", "save"), hint("esc", "cancel"))
	}

	switch m.tab {
	case TabTimer:
		hints := []string{hint("space", "start/pause"), hint("r", "reset"), hint("s", "skip"), hint("1/2/3", "quick-start"), hint("?", "help")}
		return hintJoin(hints...)
	case TabTasks:
		return hintJoin(hint("a", "add"), hint("x", "toggle"), hint("d", "delete"), hint("enter", "focus"), hint("j/k", "move"), hint("?", "help"))
	case TabStats:
		return hintJoin(hint("?", "help"))
	default:
		return hintJoin(hint("esc", "back"), hint("?", "help"))
	}
}

// renderQuitConfirm shows the quit-while-running confirmation modal.
func (m Model) renderQuitConfirm() string {
	panel := panelStyle.
		BorderForeground(colorPaused).
		Width(44).
		Align(lipgloss.Center).
		Render(
			"Session still running — quit anyway?\n\n" +
				hintJoin(hint("y", "quit"), hint("n/esc", "keep going")),
		)
	return lipgloss.Place(m.width, max(1, m.height-4), lipgloss.Center, lipgloss.Center, panel)
}

// itoa is a tiny wrapper so view code reads cleanly.
func itoa(n int) string {
	return strconv.Itoa(n)
}
