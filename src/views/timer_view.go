package views

import (
	"image/color"

	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"pomodoro/src/models"
	"pomodoro/src/utils"
)

// styledDigits renders "MM:SS" letter-spaced; digits that changed in the last
// tick flash in the session accent and settle back as the flash decays.
func (m Model) styledDigits(digits string, accent color.Color) string {
	runes := []rune(digits)
	parts := make([]string, 0, len(runes)*2)
	for i, r := range runes {
		style := lipgloss.NewStyle().Bold(true)
		if i < len(m.anim.digitFlash) && m.anim.digitFlash[i] > 0.05 {
			style = style.Foreground(accent)
		}
		if i > 0 {
			parts = append(parts, " ")
		}
		parts = append(parts, style.Render(string(r)))
	}
	return strings.Join(parts, "")
}

// renderTimerView draws the hero screen: ring with big digits, mode wordmark,
// current task, controls, and a progress bar. Motion (spring-eased fill,
// completion pulse, paused breathing, digit flash) comes from anim.go.
func (m Model) renderTimerView() string {
	session := m.timer.GetCurrentSession()
	active := session != nil && (session.Status == models.Running || session.Status == models.Paused)

	var (
		sessionType models.SessionType
		wordmark    string
		centerSub   string
		digits      string
		accent      color.Color
	)

	if active {
		sessionType = session.Type
		wordmark = utils.FormatSessionType(sessionType)
		centerSub = "RUNNING"
		if session.Status == models.Paused {
			centerSub = "PAUSED"
		}
		accent = sessionAccent(sessionType, session.Status)
	} else {
		// Ready frame: show the next session in the rhythm.
		next := m.timer.SuggestNext()
		sessionType = next
		wordmark = utils.FormatSessionType(next)
		centerSub = "READY"
		accent = sessionAccent(next, "")
	}

	// Completion pulse: the ring accent lerps from the finished session's
	// color to the next one's while the flash decays.
	if m.anim.pulse > 0 {
		accent = blendColors(m.anim.pulseFrom, m.anim.pulseTo, 1-m.anim.pulse)
	}

	// The elapsed arc is spring-eased: it sweeps full on completion and
	// empties on reset instead of snapping.
	fraction := m.anim.ring.pos

	// Paused "breathing": the track smoothly brightens and dims.
	trackStyle := lipgloss.NewStyle().Foreground(colorFaint)
	if active && session.Status == models.Paused {
		breathe := 0.5 + 0.5*float64(math.Sin(2*math.Pi*m.anim.breathe))
		trackStyle = lipgloss.NewStyle().Foreground(
			blendAdaptiveHex("#B4B4B4", "#454545", "#8F8F8F", "#6E6E6E", breathe),
		)
	}

	if active {
		digits = m.styledDigits(utils.FormatDuration(session.Remaining), accent)
	} else {
		digits = spacedTime(utils.FormatDuration(m.settings.DurationFor(sessionType)))
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
	primary := func(label string, color color.Color) string {
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

// breakSuggestion returns the rotating break activity hint.
func breakSuggestion(sessionType models.SessionType) string {
	if sessionType == models.LongBreak {
		return "step away · walk · breathe"
	}
	return "stretch · water · look away"
}
