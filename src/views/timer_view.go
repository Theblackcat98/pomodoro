package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
	"pomodoro/src/utils"
)

// renderTimerView draws the hero screen: ring with big digits, mode wordmark,
// current task, controls, and a progress bar.
func (m Model) renderTimerView() string {
	session := m.timer.GetCurrentSession()
	active := session != nil && (session.Status == models.Running || session.Status == models.Paused)

	var (
		sessionType models.SessionType
		wordmark    string
		centerSub   string
		digits      string
		fraction    float64
		accent      lipgloss.Color
	)

	if active {
		sessionType = session.Type
		wordmark = utils.FormatSessionType(sessionType)
		digits = spacedTime(utils.FormatDuration(session.Remaining))
		switch session.Status {
		case models.Running:
			centerSub = "RUNNING"
		case models.Paused:
			centerSub = "PAUSED"
		}
		if session.Duration > 0 {
			fraction = 1 - float64(session.Remaining)/float64(session.Duration)
		}
		accent = sessionAccent(sessionType, session.Status)
	} else {
		// Ready frame: show the next session in the rhythm.
		next := m.timer.SuggestNext()
		sessionType = next
		wordmark = utils.FormatSessionType(next)
		digits = spacedTime(utils.FormatDuration(m.settings.DurationFor(next)))
		centerSub = "READY"
		fraction = 0
		accent = sessionAccent(next, "")
	}

	trackStyle := lipgloss.NewStyle().Foreground(colorFaint)
	// Paused "breathing": the track gently brightens and dims.
	if active && session.Status == models.Paused && m.clock.Second()%2 == 1 {
		trackStyle = lipgloss.NewStyle().Foreground(colorDim)
	}

	compact := m.width < 80 || m.height < 20
	radius := defaultRingRadius
	if compact {
		radius = 3
	}
	ring := renderRingSized(fraction, lipgloss.NewStyle().Foreground(accent), trackStyle, []string{
		lipgloss.NewStyle().Foreground(accent).Bold(true).Render(wordmark),
		lipgloss.NewStyle().Bold(true).Render(digits),
		lipgloss.NewStyle().Foreground(colorDim).Render(centerSub),
	}, radius)

	sections := []string{ring}

	// Current task line (when one is picked).
	if task := m.currentTask(); task != nil {
		line := lipgloss.NewStyle().Foreground(colorDim).Render(
			"🍅 " + task.Title + " · " + itoa(task.CompletedPomodoros) + "/" + itoa(task.EstimatedPomodoros),
		)
		sections = append(sections, line)
	}

	// Controls row.
	sections = append(sections, "", m.renderTimerControls(session))

	// Break suggestion line while on a break.
	if active && sessionType != models.FocusSession {
		sections = append(sections, lipgloss.NewStyle().Foreground(colorDim).Italic(true).Render(breakSuggestion(sessionType)))
	}

	// Progress bar while a session is on screen (spring-eased per tick).
	if active || session != nil && session.Status == models.Completed {
		sections = append(sections, m.progress.View())
	}

	return lipgloss.JoinVertical(lipgloss.Center, sections...)
}

// renderTimerControls draws the primary button plus reset/skip, or the
// duration capsules when idle.
func (m Model) renderTimerControls(session *models.TimerSession) string {
	primary := func(label string, color lipgloss.Color) string {
		return primaryButtonStyle.Background(color).Render(label)
	}
	secondary := func(label string) string {
		return secondaryButtonStyle.Render(label)
	}

	// While the completion banner is up and a session will auto-start, show
	// the transition spinner instead of the controls.
	if m.spinActive && m.flash != "" {
		next := m.timer.SuggestNext()
		label := "starting focus…"
		if next != models.FocusSession {
			label = "starting break…"
		}
		return primary(m.spinner.View()+" "+label, sessionAccent(next, ""))
	}

	if session != nil && (session.Status == models.Running || session.Status == models.Paused) {
		var main string
		if session.Status == models.Running {
			main = primary("⏸ PAUSE", sessionAccent(session.Type, session.Status))
		} else {
			main = primary("▶ RESUME", colorPaused)
		}
		return lipgloss.JoinHorizontal(lipgloss.Center,
			main, "  ", secondary("RESET"), "  ", secondary("SKIP"),
		)
	}

	// Idle: start + duration capsules (capsules are 3 rows tall; Center keeps
	// the 1-row START button vertically aligned with them). Compact layouts
	// drop the capsules to keep the frame tight.
	if m.compact() {
		return primary("▶ START", colorFocus)
	}

	start := primary("▶ START", colorFocus)
	capsules := []string{}
	for _, c := range []struct {
		key     string
		minutes int
	}{{"15m", 15}, {"25m", 25}, {"35m", 35}} {
		style := capsuleStyle
		if c.minutes == int(m.settings.FocusDuration.Minutes()) {
			style = capsuleActiveStyle.Foreground(colorFocus).BorderForeground(colorFocus)
		}
		capsules = append(capsules, style.Render(c.key))
	}
	return lipgloss.JoinHorizontal(lipgloss.Center,
		append([]string{start, "   "}, capsules...)...,
	)
}

// compact reports whether the terminal is too small for the full layout.
func (m Model) compact() bool {
	return m.width < 80 || m.height < 20
}

// progressBar renders the linear progress bar under the timer.
func progressBar(fraction float64, width int, color lipgloss.Color) string {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	filled := int(fraction * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	styled := "▕" + lipgloss.NewStyle().Foreground(color).Render(bar) + "▏"
	pct := itoa(int(fraction*100)) + "%"
	return styled + " " + lipgloss.NewStyle().Foreground(colorDim).Render(pct)
}

// breakSuggestion returns the rotating break activity hint.
func breakSuggestion(sessionType models.SessionType) string {
	if sessionType == models.LongBreak {
		return "step away · walk · breathe"
	}
	return "stretch · water · look away"
}
