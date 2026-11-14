package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
	"pomodoro/src/utils"
)

// renderStatsView draws today's numbers, the 7-day rhythm chart, and the
// recent-sessions table.
func (m Model) renderStatsView() string {
	if len(m.sessions) == 0 {
		return lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(colorDim).Render("No sessions yet."),
			lipgloss.NewStyle().Foreground(colorDim).Render(
				"Complete a focus session on the "+keyNameStyle.Render("Timer")+" tab to start your history.",
			),
		)
	}

	cards := m.statCards()
	chart := m.weekChart()
	tableTitle := lipgloss.NewStyle().Foreground(colorDim).Render("Recent sessions")

	// Side by side when the terminal is wide enough; the table is tall, so
	// stacking would push the cards off screen.
	var lower string
	if m.width >= 100 {
		lower = lipgloss.JoinHorizontal(lipgloss.Top, chart, "    ", tableTitle+"\n"+m.sessTable.View())
	} else {
		lower = lipgloss.JoinVertical(lipgloss.Left, chart, "", tableTitle, m.sessTable.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, cards, "", lower)
}

// statCards renders the today card row.
func (m Model) statCards() string {
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

	return lipgloss.JoinHorizontal(lipgloss.Top,
		card("Today", "🍅 × "+itoa(todayPomodoros)),
		"  ",
		card("Focus time", utils.FormatDurationHuman(focusToday)),
	)
}

// weekChart renders one bar per day for the last 7 days.
func (m Model) weekChart() string {
	type dayStat struct {
		label string
		count int
		today bool
	}

	now := time.Now()
	days := make([]dayStat, 7)
	maxCount := 1
	for i := range days {
		day := now.AddDate(0, 0, -(6 - i))
		days[i] = dayStat{label: day.Format("Mon"), today: isToday(day)}
		for _, s := range m.sessions {
			if s.Type == models.FocusSession && s.CompletedAt != nil && sameDay(*s.CompletedAt, day) {
				days[i].count++
			}
		}
		if days[i].count > maxCount {
			maxCount = days[i].count
		}
	}

	title := lipgloss.NewStyle().Foreground(colorDim).Render("Last 7 days")
	lines := []string{title, ""}

	// Scale bars to a max height of 6 blocks.
	for _, d := range days {
		blocks := 0
		if d.count > 0 {
			blocks = d.count * 6 / maxCount
			if blocks < 1 {
				blocks = 1
			}
		}
		bar := ""
		if blocks > 0 {
			bar = lipgloss.NewStyle().Foreground(colorFocus).Render(strings.Repeat("▆", blocks))
		} else {
			bar = lipgloss.NewStyle().Foreground(colorFaint).Render("·")
		}
		label := d.label
		style := lipgloss.NewStyle().Foreground(colorDim)
		if d.today {
			label += "*"
			style = lipgloss.NewStyle().Foreground(colorText).Bold(true)
		}
		lines = append(lines, fmt.Sprintf("%s  %s %s", style.Render(fmt.Sprintf("%-5s", label)), bar, lipgloss.NewStyle().Foreground(colorDim).Render(itoa(d.count))))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// refreshSessionTable rebuilds the recent-sessions table rows from history.
func (m *Model) refreshSessionTable() {
	rows := make([]table.Row, 0, len(m.sessions))
	for i, s := range m.sessions {
		if i >= 50 {
			break
		}
		when := ""
		if s.CompletedAt != nil {
			when = s.CompletedAt.Format("15:04")
		}
		task := "—"
		if s.TaskID != "" {
			if t := m.taskByID(s.TaskID); t != nil {
				task = t.Title
			}
		}
		rows = append(rows, table.Row{
			when,
			utils.FormatSessionTypeShort(s.Type),
			utils.FormatDurationHuman(s.Duration),
			task,
		})
	}
	m.sessTable.SetRows(rows)
}

func (m Model) taskByID(id string) *models.Task {
	for _, t := range m.tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}
