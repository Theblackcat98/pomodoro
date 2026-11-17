package views

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	setRowTheme
	setRowMouse
	setRowHookFocus
	setRowHookBreak
	setRowHookComplete
	setRowCount
)

// themeLabel is the effective palette name shown in the settings row. The
// legacy "default" value reads as dark, matching what the user sees.
func themeLabel(theme string) string {
	if resolvePalette(theme).dark {
		return ThemeDark
	}
	return ThemeLight
}

// renderSettingsView draws the editable settings; changes save instantly.
func (m Model) renderSettingsView() string {
	type row struct{ label, value string }

	boolMark := func(b bool) string {
		if b {
			return lipgloss.NewStyle().Foreground(delegateLikeDone).Render("[✓]")
		}
		return lipgloss.NewStyle().Foreground(colorFaint).Render("[ ]")
	}
	hookValue := func(v string) string {
		if v == "" {
			return lipgloss.NewStyle().Foreground(colorFaint).Render("(none)")
		}
		return truncate(v, 34)
	}

	rows := []row{
		{"Focus duration", "‹ " + utils.FormatDurationHuman(m.settings.FocusDuration) + " ›"},
		{"Short break", "‹ " + utils.FormatDurationHuman(m.settings.ShortBreakDuration) + " ›"},
		{"Long break", "‹ " + utils.FormatDurationHuman(m.settings.LongBreakDuration) + " ›"},
		{"Long break every", fmt.Sprintf("‹ %d › focus sessions", m.settings.LongBreakInterval)},
		{"Auto-start breaks", boolMark(m.settings.AutoStartBreaks)},
		{"Auto-start focus", boolMark(m.settings.AutoStartFocus)},
		{"Notification", string(m.settings.NotificationType)},
		{"Theme", "‹ " + themeLabel(m.settings.Theme) + " ›"},
		{"Mouse", boolMark(m.settings.Mouse)},
		{"Hook: focus start", hookValue(m.settings.HookFocusStart)},
		{"Hook: break start", hookValue(m.settings.HookBreakStart)},
		{"Hook: complete", hookValue(m.settings.HookComplete)},
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
			// The value just changed: flash it bright and let it settle.
			if i == m.anim.flashRow && m.anim.flashValue > 0 {
				value = lipgloss.NewStyle().
					Foreground(blendAdaptiveHex("#8F8F8F", "#6E6E6E", "#3F3F3F", "#EEEEEE", m.anim.flashValue)).
					Bold(true).
					Render(r.value)
			} else {
				value = lipgloss.NewStyle().Foreground(colorText).Render(r.value)
			}
		}
		lines = append(lines, cursor+label+"  "+value)
	}

	// Inline editor for the focused hook row.
	if m.editingHook >= 0 {
		field := lipgloss.NewStyle().Foreground(colorFocus).Bold(true).Render("edit")
		lines = append(lines, "", field+"  "+m.hookInput.View(),
			lipgloss.NewStyle().Foreground(colorDim).Render("enter saves · esc cancels · runs with POMODORO_STATE set"))
	}

	lines = append(lines, "", hintJoin(hint("j/k", "select"), hint("+/-", "adjust"), hint("space", "toggle"), hint("enter", "edit hook")))

	return lipgloss.Place(m.width, max(10, m.height-6), lipgloss.Center, lipgloss.Center,
		panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...)),
	)
}

// handleSettingsKey navigates and adjusts settings; every change persists.
// While a hook is being edited, all keys go to the input.
func (m Model) handleSettingsKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	if m.editingHook >= 0 {
		return m.handleHookEditKey(msg)
	}
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
	case "space", "x", "enter":
		return m, tea.Batch(m.toggleSetting(), waitForEvent(m.events))
	}
	return m, waitForEvent(m.events)
}

// handleHookEditKey drives the inline hook editor: enter saves, esc cancels,
// everything else feeds the input.
func (m Model) handleHookEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		value := m.hookInput.Value()
		switch m.editingHook {
		case setRowHookFocus:
			m.settings.HookFocusStart = value
		case setRowHookBreak:
			m.settings.HookBreakStart = value
		case setRowHookComplete:
			m.settings.HookComplete = value
		}
		m.editingHook = -1
		m.hookInput.Blur()
		return m, tea.Batch(m.flashChangedRowCmd(), m.saveSettings(), m.toastf("Hook saved"), waitForEvent(m.events))
	case "esc":
		m.editingHook = -1
		m.hookInput.Blur()
		return m, waitForEvent(m.events)
	}
	var cmd tea.Cmd
	m.hookInput, cmd = m.hookInput.Update(msg)
	return m, tea.Batch(cmd, waitForEvent(m.events))
}

// openHookEditor starts editing the focused hook row (or toggles booleans
// for the other row kinds via enter).
func (m *Model) openHookEditor() tea.Cmd {
	switch m.settingsCursor {
	case setRowHookFocus:
		m.editingHook = setRowHookFocus
		m.hookInput.SetValue(m.settings.HookFocusStart)
	case setRowHookBreak:
		m.editingHook = setRowHookBreak
		m.hookInput.SetValue(m.settings.HookBreakStart)
	case setRowHookComplete:
		m.editingHook = setRowHookComplete
		m.hookInput.SetValue(m.settings.HookComplete)
	default:
		return nil
	}
	return m.hookInput.Focus()
}

// flashChangedRowCmd is the cmd-returning form of flashChangedRow.
func (m *Model) flashChangedRowCmd() tea.Cmd {
	m.flashChangedRow()
	return m.kickAnim()
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
	m.flashChangedRow()
	return tea.Batch(m.saveSettings(), m.kickAnim())
}

// toggleSetting flips the focused boolean, cycles the notification mode, or
// opens the inline editor for a hook row.
func (m *Model) toggleSetting() tea.Cmd {
	var extra tea.Cmd
	switch m.settingsCursor {
	case setRowAutoBreak:
		m.settings.AutoStartBreaks = !m.settings.AutoStartBreaks
	case setRowAutoFocus:
		m.settings.AutoStartFocus = !m.settings.AutoStartFocus
	case setRowNotify:
		// Cycle through the declared modes in order.
		next := 0
		for i, mode := range models.NotificationModes {
			if mode == m.settings.NotificationType {
				next = (i + 1) % len(models.NotificationModes)
				break
			}
		}
		m.settings.NotificationType = models.NotificationModes[next]
	case setRowTheme:
		// Toggle dark ↔ light; the legacy "default" value cycles to light.
		if m.settings.Theme == ThemeLight {
			m.settings.Theme = ThemeDark
		} else {
			m.settings.Theme = ThemeLight
		}
		// Rebuild the shared styles now so the change shows without restart;
		// saveSettings below persists the chosen value.
		buildStyles(m.settings.Theme)
	case setRowMouse:
		m.settings.Mouse = !m.settings.Mouse
		// The terminal mouse mode follows the View declaration on the next
		// render — no imperative command needed in v2.
		extra = nil
	case setRowHookFocus, setRowHookBreak, setRowHookComplete:
		return tea.Batch(m.openHookEditor(), waitForEvent(m.events))
	default:
		return nil
	}
	m.flashChangedRow()
	return tea.Batch(m.saveSettings(), extra, m.kickAnim())
}

// flashChangedRow arms the value flash for the focused settings row.
func (m *Model) flashChangedRow() {
	m.anim.flashRow = m.settingsCursor
	m.anim.flashValue = 1
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
