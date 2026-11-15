package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderTasksView draws the task list, open/done counts, and the new-task bar.
func (m Model) renderTasksView() string {
	sections := []string{m.taskList.View()}

	if m.adding {
		sections = append(sections, "", m.renderAddBar())
	}

	// Empty state when there is nothing to show.
	if len(m.tasks) == 0 && !m.adding {
		empty := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(colorDim).Render("( ｰｰ)っ🍅"),
			"",
			lipgloss.NewStyle().Foreground(colorDim).Render("No tasks yet — press "+keyNameStyle.Render("a")+" to plant your first tomato"),
		)
		return lipgloss.JoinVertical(lipgloss.Center, empty)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderAddBar draws the slim new-task form.
func (m Model) renderAddBar() string {
	divider := lipgloss.NewStyle().
		Foreground(colorFaint).
		Render("── new task " + strings.Repeat("─", max(10, m.width/2-12)))

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
