package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderTasksView draws the task list (or the empty state) and the new-task
// bar. The custom empty state is used whenever there are no tasks, including
// while the add form is open — the bubbles list's default "No items." never
// leaks through.
func (m Model) renderTasksView() string {
	var sections []string

	if len(m.tasks) == 0 {
		empty := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(colorDim).Render("( ｰｰ)っ🍅"),
			"",
			lipgloss.NewStyle().Foreground(colorDim).Render("No tasks yet — press "+keyNameStyle.Render("a")+" to plant your first tomato"),
		)
		sections = append(sections, empty)
	} else {
		sections = append(sections, m.taskList.View())
	}

	if m.adding {
		sections = append(sections, "", m.renderAddBar())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderAddBar draws the slim new-task form (reused for editing).
func (m Model) renderAddBar() string {
	heading := "new task"
	if m.editingID != "" {
		heading = "edit task"
	}
	divider := lipgloss.NewStyle().
		Foreground(colorFaint).
		Render("── " + heading + " " + strings.Repeat("─", max(10, m.width/2-12)))

	estimate := lipgloss.NewStyle().Foreground(colorDim).Render(fmt.Sprintf("🍅 × %d", m.addEst))

	titlePrompt := "Title"
	descPrompt := "Desc"
	if m.addFocus == addFieldTitle {
		titlePrompt = lipgloss.NewStyle().Foreground(colorFocus).Bold(true).Render("Title")
	} else {
		descPrompt = lipgloss.NewStyle().Foreground(colorFocus).Bold(true).Render("Desc")
	}

	rows := []string{
		divider,
		titlePrompt + "  " + m.addTitle.View(),
		descPrompt + "   " + m.addDesc.View(),
		estimate + "  " + statusHintStyle.Render("[tab] field · [enter] save · [esc] cancel"),
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
