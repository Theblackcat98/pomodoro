package views

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
	"pomodoro/src/utils"
)

// Settings rows, in display order.
const (
	setRowFocus = iota
	setRowShortBreak
	setRowLongBreak
	setRowInterval
	setRowAutoBreak
	setRowAutoFocus
	setRowNotify
	setRowCount
)

// renderSettingsView draws the editable settings; changes save instantly.
func (m Model) renderSettingsView() string {
	type row struct{ label, value string }

	boolMark := func(b bool) string {
		if b {
			return lipgloss.NewStyle().Foreground(delegateLikeDone).Render("[✓]")
		}
		return lipgloss.NewStyle().Foreground(colorFaint).Render("[ ]")
	}

	rows := []row{
		{"Focus duration", "‹ " + utils.FormatDurationHuman(m.settings.FocusDuration) + " ›"},
		{"Short break", "‹ " + utils.FormatDurationHuman(m.settings.ShortBreakDuration) + " ›"},
		{"Long break", "‹ " + utils.FormatDurationHuman(m.settings.LongBreakDuration) + " ›"},
		{"Long break every", fmt.Sprintf("‹ %d › focus sessions", m.settings.LongBreakInterval)},
		{"Auto-start breaks", boolMark(m.settings.AutoStartBreaks)},
		{"Auto-start focus", boolMark(m.settings.AutoStartFocus)},
		{"Notification", string(m.settings.NotificationType)},
	}

	lines := []string{
		lipgloss.NewStyle().Foreground(colorDim).Render("Changes apply and save immediately."),
		"",
	}
	for i, r := range rows {
		label := lipgloss.NewStyle().Foreground(colorText).Render(fmt.Sprintf("%-20s", r.label))
		value := lipgloss.NewStyle().Foreground(colorDim).Render(r.value)
		cursor := "  "
		if i == m.settingsCursor {
			cursor = lipgloss.NewStyle().Foreground(colorFocus).Bold(true).Render("❯ ")
			label = lipgloss.NewStyle().Foreground(colorText).Bold(true).Render(fmt.Sprintf("%-20s", r.label))
			value = lipgloss.NewStyle().Foreground(colorText).Render(r.value)
		}
		lines = append(lines, cursor+label+"  "+value)
	}

	lines = append(lines, "", hintJoin(hint("j/k", "select"), hint("+/-", "adjust"), hint("space", "toggle")))

	return lipgloss.Place(m.width, max(10, m.height-6), lipgloss.Center, lipgloss.Center,
		panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...)),
	)
}

// handleSettingsKey navigates and adjusts settings; every change persists.
func (m Model) handleSettingsKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case "down", "j":
		if m.settingsCursor < setRowCount-1 {
			m.settingsCursor++
		}
	case "+", "=", "l", "right":
		return m, tea.Batch(m.adjustSetting(1), waitForEvent(m.events))
	case "-", "_", "h", "left":
		return m, tea.Batch(m.adjustSetting(-1), waitForEvent(m.events))
	case " ", "x", "enter":
		return m, tea.Batch(m.toggleSetting(), waitForEvent(m.events))
	}
	return m, waitForEvent(m.events)
}

// adjustSetting steps the focused numeric setting by dir.
func (m *Model) adjustSetting(dir int) tea.Cmd {
	s := m.settings
	switch m.settingsCursor {
	case setRowFocus:
		s.FocusDuration = clampDuration(s.FocusDuration, time.Duration(dir)*5*time.Minute, 5*time.Minute, 90*time.Minute)
	case setRowShortBreak:
		s.ShortBreakDuration = clampDuration(s.ShortBreakDuration, time.Duration(dir)*time.Minute, time.Minute, 30*time.Minute)
	case setRowLongBreak:
		s.LongBreakDuration = clampDuration(s.LongBreakDuration, time.Duration(dir)*5*time.Minute, 5*time.Minute, 60*time.Minute)
	case setRowInterval:
		s.LongBreakInterval = clampInt(s.LongBreakInterval+dir, 1, 12)
	default:
		return nil
	}
	return m.saveSettings()
}

// toggleSetting flips the focused boolean or cycles the notification mode.
func (m *Model) toggleSetting() tea.Cmd {
	switch m.settingsCursor {
	case setRowAutoBreak:
		m.settings.AutoStartBreaks = !m.settings.AutoStartBreaks
	case setRowAutoFocus:
		m.settings.AutoStartFocus = !m.settings.AutoStartFocus
	case setRowNotify:
		switch m.settings.NotificationType {
		case models.SoundNotification:
			m.settings.NotificationType = models.VisualNotification
		case models.VisualNotification:
			m.settings.NotificationType = models.SilentNotification
		default:
			m.settings.NotificationType = models.SoundNotification
		}
	default:
		return nil
	}
	return m.saveSettings()
}

// saveSettings applies settings to the service/notifier and persists them.
func (m *Model) saveSettings() tea.Cmd {
	m.settings.UpdatedAt = time.Now()
	m.timer.SetSettings(m.settings)
	m.notifier.SetNotificationType(m.settings.NotificationType)
	if m.store != nil {
		return m.errToast("save settings", m.store.SaveSettings(m.settings))
	}
	return nil
}

func clampDuration(current, delta, min, max time.Duration) time.Duration {
	next := current + delta
	if next < min {
		return min
	}
	if next > max {
		return max
	}
	return next
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// delegateLikeDone mirrors the delegate's done color for settings marks.
var delegateLikeDone = lipgloss.Color("#46A758")
