package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	tableTitle := m.statsHeader()

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
					Foreground(blendAdaptiveHex("#B4B4B4", "#454545", hexFocus, hexFocus, 0.25+0.75*intensity)).
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

// statsSortColumns are the table columns in the order the 1-4 keys address.
var statsSortColumns = []string{"WHEN", "TYPE", "LENGTH", "TASK"}

// statsColumns returns the table columns with a marker on the active sort
// column so the current ordering is visible in the header.
func (m Model) statsColumns() []table.Column {
	cols := []table.Column{
		{Title: "WHEN", Width: 7},
		{Title: "TYPE", Width: 15},
		{Title: "LENGTH", Width: 9},
		{Title: "TASK", Width: 32},
	}
	if m.statsSortCol >= 0 && m.statsSortCol < len(cols) {
		marker := " ▼"
		if m.statsSortAsc {
			marker = " ▲"
		}
		cols[m.statsSortCol].Title += marker
	}
	return cols
}

// statsHeader renders the table title plus the active sort/filter indicators.
func (m Model) statsHeader() string {
	title := lipgloss.NewStyle().Foreground(colorDim).Render("Recent sessions")
	if m.statsSortCol >= 0 && m.statsSortCol < len(statsSortColumns) {
		arrow := "▼"
		if m.statsSortAsc {
			arrow = "▲"
		}
		title += lipgloss.NewStyle().Foreground(colorFaint).Render(
			"  sort " + statsSortColumns[m.statsSortCol] + " " + arrow)
	}
	if m.statsFilter != "" {
		title += lipgloss.NewStyle().Foreground(colorFocus).Render("  filter: " + m.statsFilter)
	}
	if m.statsFiltering {
		title += "  " + m.statsFilterInput.View()
	}
	return title
}

// statsSessionTitle returns the task title displayed for a session.
func (m Model) statsSessionTitle(s models.TimerSession) string {
	if s.TaskID != "" {
		if t := m.taskByID(s.TaskID); t != nil {
			return t.Title
		}
	}
	return "—"
}

// sortedFilteredSessions applies the active filter and sort to the FULL
// history slice. The 50-row cap is a render concern applied later in
// refreshSessionTable, so filtering/sorting never run on a truncated set.
func (m Model) sortedFilteredSessions() []models.TimerSession {
	list := make([]models.TimerSession, 0, len(m.sessions))
	needle := strings.ToLower(m.statsFilter)
	for _, s := range m.sessions {
		if needle != "" {
			task := strings.ToLower(m.statsSessionTitle(s))
			kind := strings.ToLower(string(s.Type))
			label := strings.ToLower(utils.FormatSessionTypeShort(s.Type))
			if !strings.Contains(task, needle) &&
				!strings.Contains(kind, needle) &&
				!strings.Contains(label, needle) {
				continue
			}
		}
		list = append(list, s)
	}

	// No explicit sort: keep the stored newest-first order.
	if m.statsSortCol < 0 || m.statsSortCol >= len(statsSortColumns) {
		return list
	}

	col := m.statsSortCol
	cmp := func(a, b models.TimerSession) int {
		switch col {
		case 0: // WHEN
			at, bt := time.Time{}, time.Time{}
			if a.CompletedAt != nil {
				at = *a.CompletedAt
			}
			if b.CompletedAt != nil {
				bt = *b.CompletedAt
			}
			if at.Before(bt) {
				return -1
			}
			if at.After(bt) {
				return 1
			}
		case 1: // TYPE
			return strings.Compare(strings.ToLower(string(a.Type)), strings.ToLower(string(b.Type)))
		case 2: // LENGTH
			ae, be := sessionElapsed(a), sessionElapsed(b)
			if ae < be {
				return -1
			}
			if ae > be {
				return 1
			}
		case 3: // TASK
			return strings.Compare(strings.ToLower(m.statsSessionTitle(a)), strings.ToLower(m.statsSessionTitle(b)))
		}
		return 0
	}
	sort.SliceStable(list, func(i, j int) bool {
		c := cmp(list[i], list[j])
		if m.statsSortAsc {
			return c < 0
		}
		return c > 0
	})
	return list
}

// refreshSessionTable rebuilds the recent-sessions table rows from history.
// Filtering and sorting run on all sessions; only the rendered rows are
// capped at 50.
func (m *Model) refreshSessionTable() {
	list := m.sortedFilteredSessions()
	rows := make([]table.Row, 0, min(50, len(list)))
	for i, s := range list {
		if i >= 50 {
			break
		}
		when := ""
		if s.CompletedAt != nil {
			when = s.CompletedAt.Format("15:04")
		}
		rows = append(rows, table.Row{
			when,
			utils.FormatSessionTypeShort(s.Type),
			utils.FormatDurationHuman(sessionElapsed(s)),
			m.statsSessionTitle(s),
		})
	}
	m.sessTable.SetColumns(m.statsColumns())
	m.sessTable.SetRows(rows)
}

// handleStatsFilterKey drives the inline history filter: enter applies,
// esc cancels and clears, everything else feeds the input.
func (m Model) handleStatsFilterKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.statsFilter = strings.TrimSpace(m.statsFilterInput.Value())
		m.statsFiltering = false
		m.statsFilterInput.Blur()
		m.refreshSessionTable()
	case "esc":
		m.statsFilter = ""
		m.statsFilterInput.SetValue("")
		m.statsFiltering = false
		m.statsFilterInput.Blur()
		m.refreshSessionTable()
	default:
		var cmd tea.Cmd
		m.statsFilterInput, cmd = m.statsFilterInput.Update(msg)
		return m, tea.Batch(cmd, waitForEvent(m.events))
	}
	return m, waitForEvent(m.events)
}

func (m Model) taskByID(id string) *models.Task {
	for _, t := range m.tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}
