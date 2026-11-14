package views

import (
	"github.com/charmbracelet/lipgloss"
)

// renderHelpView draws the keybinding cheat sheet (upgraded to a bubbles/help
// + viewport overlay in Phase 4).
func (m Model) renderHelpView() string {
	group := func(title string, lines []string) string {
		body := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(title) + "\n" +
			stringsJoin(lines)
		return panelStyle.Render(body)
	}

	general := group("General", []string{
		"tab / shift+tab    switch views",
		"?                  open this help",
		"q                  quit (confirms while a session runs)",
		"ctrl+c             quit immediately",
	})
	timer := group("Timer", []string{
		"space / enter      start, pause, resume",
		"r                  reset to ready",
		"s                  skip to the next session",
		"1 / 2 / 3          start 15 / 25 / 35 minute focus",
	})
	tasks := group("Tasks", []string{
		"a                  add a task",
		"x                  toggle done",
		"d                  delete (y to confirm)",
		"enter              focus the selected task",
		"j / k              move the cursor",
	})

	grid := lipgloss.JoinHorizontal(lipgloss.Top, general, "  ", timer)
	return lipgloss.JoinVertical(lipgloss.Center, grid, "  ", tasks)
}

func stringsJoin(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n"
		}
		out += l
	}
	return out
}
