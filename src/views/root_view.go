package views

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
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
	case TabSettings:
		body = m.renderSettingsView()
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

	// Tab-switch slide: the body glides a few columns in from the direction
	// of travel and settles to center.
	if m.anim.tabSlide != 0 {
		pad := int(math.Abs(m.anim.tabSlide) * 5)
		if pad > 0 {
			if m.anim.tabSlide > 0 {
				body = lipgloss.NewStyle().PaddingLeft(pad).Render(body)
			} else {
				body = lipgloss.NewStyle().PaddingRight(pad).Render(body)
			}
		}
	}

	// Confirmation modals replace the body (keys are trapped by handleKey).
	switch {
	case m.confirmQuit:
		body = m.renderConfirmModal(
			"Session still running",
			"Quit anyway? The session will not be recorded.",
			colorPaused,
		)
	case m.deleteArm:
		body = m.renderConfirmModal(
			"Delete task",
			fmt.Sprintf("%q will be removed permanently.", m.confirmMsg),
			colorFocus,
		)
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
	for i, name := range tabNames {
		style := tabStyle
		if Tab(i) == m.tab {
			style = activeTabStyle.Foreground(m.currentAccent())
		}
		tabs += style.Render(name)
	}

	right := ""
	if !m.compact() {
		right = toastStyle.Render("🍅 " + itoa(m.todayCount()) + " today")
	}

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
// A visible toast temporarily replaces the hints (both would never fit) and
// is truncated to the free space so the bar always stays a single row.
func (m Model) renderStatusBar() string {
	chipText, chipColor := m.statusChip()
	chip := statusChipStyle.
		Background(chipColor).
		Foreground(lipgloss.Color("#FFF7DB")).
		Render(chipText)

	clock := statusHintStyle.Render(m.clock.Format("15:04"))

	middle := m.contextualHints()
	if m.toast != "" {
		// Toast slides in from the right (anim.go) before settling.
		toast := m.toast
		if off := int(m.anim.toastSlide * 10); off > 0 {
			toast = strings.Repeat(" ", off) + toast
		}
		avail := m.width - lipgloss.Width(chip) - lipgloss.Width(clock) - 6
		if avail < 8 {
			avail = 8
		}
		if w := lipgloss.Width(toast); w > avail {
			toast = truncate(toast, avail-1) + "…"
		}
		middle = toastStyle.Render(toast)
	}

	// Account for the bar's own horizontal padding (1 column each side) so
	// the clock never wraps to the next row.
	used := lipgloss.Width(chip) + 1 + lipgloss.Width(middle) + 1 + lipgloss.Width(clock)
	gap := max(1, m.width-used-2)

	bar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		chip,
		" ",
		middle,
		strings.Repeat(" ", gap),
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
		// Blink the dot while running so the chip feels alive.
		glyph := "●"
		if m.clock.Second()%2 == 1 {
			glyph = "◦"
		}
		return glyph + " RUNNING", colorFocus
	default:
		return "☕ ON BREAK", sessionAccent(session.Type, session.Status)
	}
}

// contextualHints returns the key hints valid for the current screen and mode.
// Compact terminals only get the two most important hints.
func (m Model) contextualHints() string {
	if m.compact() {
		switch {
		case m.confirmQuit:
			return hintJoin(hint("y", "quit"), hint("n", "stay"))
		case m.adding:
			return hintJoin(hint("enter", "save"), hint("esc", "cancel"))
		}
		switch m.tab {
		case TabTimer:
			return hintJoin(hint("space", "start/pause"), hint("?", "help"))
		case TabTasks:
			return hintJoin(hint("a", "add"), hint("?", "help"))
		default:
			return hintJoin(hint("?", "help"))
		}
	}
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
		return hintJoin(hint("j/k", "select"), hint("e", "export JSON"), hint("?", "help"))
	case TabSettings:
		return hintJoin(hint("j/k", "select"), hint("+/-", "adjust"), hint("space", "toggle"), hint("?", "help"))
	default:
		return hintJoin(hint("j/k", "scroll"), hint("esc", "back"), hint("?", "help"))
	}
}

// computeTabHits derives the header tabs' column ranges deterministically
// from the same layout math renderHeader uses (View copies the model, so
// positions captured during rendering would be lost). Mouse clicks are
// matched against these.
func computeTabHits() []tabHit {
	hits := make([]tabHit, 0, tabCount)
	x := 1 + // headerStyle left padding
		lipgloss.Width("🍅 Pomodoro") + // brand
		2 // gap between brand and tabs
	for i, name := range tabNames {
		w := lipgloss.Width(name) + 2 // tabStyle horizontal padding
		hits = append(hits, tabHit{tab: Tab(i), startX: x, endX: x + w - 1})
		x += w
	}
	return hits
}

// renderConfirmModal draws the shared centered confirmation panel used for
// quit-while-running and task deletion. The safe answer is always offered
// first and Esc backs out.
func (m Model) renderConfirmModal(title, message string, accent lipgloss.Color) string {
	panel := panelStyle.
		BorderForeground(accent).
		Width(46).
		Align(lipgloss.Center).
		Render(
			lipgloss.NewStyle().Bold(true).Foreground(accent).Render(title) + "\n" +
				message + "\n\n" +
				hintJoin(hint("y", "confirm"), hint("n/esc", "cancel")),
		)
	return lipgloss.Place(m.width, max(1, m.height-4), lipgloss.Center, lipgloss.Center, panel)
}

// itoa is a tiny wrapper so view code reads cleanly.
func itoa(n int) string {
	return strconv.Itoa(n)
}

// truncate cuts a (possibly ANSI-styled) string to at most width printable
// columns, preserving escape sequences.
func truncate(s string, width int) string {
	if width < 1 {
		return ""
	}
	var out strings.Builder
	cols := 0
	inEscape := false
	for _, r := range s {
		if inEscape {
			out.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		if r == '\x1b' {
			inEscape = true
			out.WriteRune(r)
			continue
		}
		w := runewidth.RuneWidth(r)
		if cols+w > width {
			break
		}
		cols += w
		out.WriteRune(r)
	}
	return out.String()
}
