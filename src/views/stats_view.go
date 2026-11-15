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

// renderStatsView draws today's numbers with streaks, the 13-week activity
// heatmap, per-task focus, and the recent-sessions table.
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
	tableTitle := lipgloss.NewStyle().Foreground(colorDim).Render("Recent sessions")

	// Compact terminals keep only the essentials.
	if m.compact() {
		return lipgloss.JoinVertical(lipgloss.Left, cards, "", tableTitle, m.sessTable.View())
	}

	// Wide: heatmap and per-task breakdown side by side; narrow: stacked.
	var middle string
	if m.width >= 100 {
		middle = lipgloss.JoinHorizontal(lipgloss.Top, m.renderHeatmap(), "    ", m.renderPerTask())
	} else {
		middle = lipgloss.JoinVertical(lipgloss.Left, m.renderHeatmap(), "", m.renderPerTask())
	}

	var sections []string
	sections = append(sections, cards, "", middle)
	if m.width >= 100 {
		sections = append(sections, "", tableTitle, m.sessTable.View())
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// statCards renders the today row: pomodoros, true focused time, and the
// current/best streak.
func (m Model) statCards() string {
	pomodoros, focusTime := todayTotals(m.sessions, time.Now())
	current, best := computeStreaks(m.sessions, time.Now())

	card := func(title, value string) string {
		return panelStyle.Render(
			lipgloss.NewStyle().Foreground(colorDim).Render(title) + "\n" +
				lipgloss.NewStyle().Bold(true).Render(value),
		)
	}

	streak := itoa(current) + " day"
	if current != 1 {
		streak += "s"
	}
	streak += lipgloss.NewStyle().Foreground(colorDim).Render("  ·  best " + itoa(best))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		card("Today", "🍅 × "+itoa(pomodoros)),
		"  ",
		card("Focus time", utils.FormatDurationHuman(focusTime)),
		"  ",
		card("Streak", streak),
	)
}

// renderHeatmap draws the GitHub-style 13-week focus grid: weekday rows
// (Mon-first), one 2-column cell per day, five intensity levels.
func (m Model) renderHeatmap() string {
	const weeks = 13
	now := time.Now()
	grid := heatmapWeeks(m.sessions, now, weeks)

	maxCount := 1
	for _, col := range grid {
		for _, cell := range col {
			if cell.Count > maxCount {
				maxCount = cell.Count
			}
		}
	}

	title := lipgloss.NewStyle().Foreground(colorDim).Render("Last " + itoa(weeks) + " weeks")

	// Month markers: the initial of a month is printed above the first week
	// that intersects it.
	monthRow := make([]string, weeks)
	lastMonth := -1
	for w, col := range grid {
		month := int(col[0].Day.Month())
		if month != lastMonth {
			monthRow[w] = col[0].Day.Format("Jan")[:1]
			lastMonth = month
		}
	}

	weekdayLabels := [7]string{"Mo", "  ", "We", "  ", "Fr", "  ", "  "}

	lines := []string{title, ""}
	header := "   "
	for w := 0; w < weeks; w++ {
		cell := monthRow[w]
		if cell == "" {
			cell = " "
		}
		header += cell + " "
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render(header))

	for d := 0; d < 7; d++ {
		row := lipgloss.NewStyle().Foreground(colorDim).Render(weekdayLabels[d]) + " "
		for w := 0; w < weeks; w++ {
			cell := grid[w][d]
			future := cell.Day.After(dayBucket(now))
			switch {
			case future:
				row += lipgloss.NewStyle().Foreground(colorFaint).Render(" ")
			case cell.Count == 0:
				row += lipgloss.NewStyle().Foreground(colorFaint).Render("·")
			default:
				intensity := float64(heatLevel(cell.Count, maxCount)) / 4
				row += lipgloss.NewStyle().
					Foreground(blendColors(resolveAdaptive(colorFaint), colorFocus, 0.25+0.75*intensity)).
					Render("■")
			}
			row += " "
		}
		lines = append(lines, row)
	}

	lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).
		Render("less "+lipgloss.NewStyle().Foreground(colorFaint).Render("·")+
			" ▁ ▂ ▃ "+lipgloss.NewStyle().Foreground(colorFocus).Render("■")+" more"))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderPerTask lists the tasks that received the most focus recently.
func (m Model) renderPerTask() string {
	title := lipgloss.NewStyle().Foreground(colorDim).Render("Focus by task")

	stats := perTaskStats(m.sessions, m.tasks, 5)
	if len(stats) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title,
			lipgloss.NewStyle().Foreground(colorDim).Render("no attributed sessions yet"))
	}

	lines := []string{title, ""}
	for i, st := range stats {
		name := st.Title
		if lipgloss.Width(name) > 22 {
			name = truncate(name, 21) + "…"
		}
		if pad := 22 - lipgloss.Width(name); pad > 0 {
			name += strings.Repeat(" ", pad)
		}
		label := lipgloss.NewStyle().Foreground(colorText).Render(name)
		count := lipgloss.NewStyle().Foreground(colorDim).Render("🍅 × " + itoa(st.Pomodoros))
		spent := utils.FormatDurationHuman(st.FocusTime)
		if pad := 7 - lipgloss.Width(spent); pad > 0 {
			spent = strings.Repeat(" ", pad) + spent
		}
		lines = append(lines, fmt.Sprintf("%s %s  %s  %s",
			lipgloss.NewStyle().Foreground(colorFaint).Render(itoa(i+1)+"."), label, count,
			lipgloss.NewStyle().Foreground(colorDim).Render(spent)))
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
			utils.FormatDurationHuman(sessionElapsed(s)),
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
