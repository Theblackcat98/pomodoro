package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
	"pomodoro/src/utils"
)

// renderStatsView draws today's numbers, and the recent session history.
// Filled from the persisted session history (Phase 4 adds the 7-day chart).
func (m Model) renderStatsView() string {
	if len(m.sessions) == 0 {
		return lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(colorDim).Render("No sessions yet."),
			lipgloss.NewStyle().Foreground(colorDim).Render(
				"Complete a focus session on the "+keyNameStyle.Render("Timer")+" tab to start your history.",
			),
		)
	}

	// Stat cards: today's pomodoros + focus time today.
	focusToday := time.Duration(0)
	todayPomodoros := 0
	for _, s := range m.sessions {
		if s.Type == models.FocusSession && s.CompletedAt != nil && isToday(*s.CompletedAt) {
			todayPomodoros++
			focusToday += s.Duration
		}
	}

	card := func(title, value string) string {
		return panelStyle.Render(
			lipgloss.NewStyle().Foreground(colorDim).Render(title) + "\n" +
				lipgloss.NewStyle().Bold(true).Render(value),
		)
	}

	cards := lipgloss.JoinHorizontal(lipgloss.Top,
		card("Today", "🍅 × "+itoa(todayPomodoros)),
		"  ",
		card("Focus time", utils.FormatDurationHuman(focusToday)),
	)

	rows := []string{
		cards,
		"",
		lipgloss.NewStyle().Foreground(colorDim).Render("Recent sessions"),
		"",
	}

	for i, s := range m.sessions {
		if i >= 8 {
			break
		}
		glyph := lipgloss.NewStyle().Foreground(sessionAccent(s.Type, s.Status)).Render("●")
		name := utils.FormatSessionTypeShort(s.Type)
		when := ""
		if s.CompletedAt != nil {
			when = s.CompletedAt.Format("15:04")
		}
		rows = append(rows, fmt.Sprintf(
			"%s  %s %-12s %s",
			lipgloss.NewStyle().Foreground(colorDim).Render(when),
			glyph,
			name,
			lipgloss.NewStyle().Foreground(colorDim).Render(utils.FormatDurationHuman(s.Duration)),
		))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}
