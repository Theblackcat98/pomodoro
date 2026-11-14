package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// helpContent builds the scrollable cheat sheet from the key bindings.
func helpContent() string {
	section := func(title string, bs []key.Binding) string {
		lines := []string{lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(title)}
		for _, b := range bs {
			if !b.Enabled() {
				continue
			}
			h := b.Help()
			lines = append(lines, fmt.Sprintf("  %-14s %s",
				keyNameStyle.Render(h.Key), statusHintStyle.Render(h.Desc)))
		}
		return strings.Join(lines, "\n")
	}

	about := statusHintStyle.Render(
		"Sessions follow the classic rhythm: focus, short break, and a long\n" +
			"break every few rounds (configurable under Settings). Completed\n" +
			"sessions, tasks, and settings survive restarts.",
	)

	body := strings.Join([]string{
		section("General", []key.Binding{keys.Tab, keys.BackTab, keys.Help, keys.Quit, keys.ForceQuit}),
		"",
		section("Timer", []key.Binding{keys.StartPause, keys.Reset, keys.Skip, keys.QuickStart}),
		"",
		section("Tasks", []key.Binding{keys.Add, keys.Toggle, keys.Delete, keys.FocusTask, keys.Nav}),
		"",
		section("Settings", []key.Binding{keys.Nav, keys.Adjust, keys.ToggleSet}),
		"",
		section("Help", []key.Binding{keys.Scroll, keys.Cancel}),
		"",
		about,
	}, "\n")

	return panelStyle.Render(body)
}

// helpKeys adapts the shared bindings to bubbles/help's KeyMap interface.
type helpKeys struct{}

func (helpKeys) ShortHelp() []key.Binding {
	return []key.Binding{keys.Scroll, keys.Cancel, keys.Quit}
}

func (helpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{keys.Scroll, keys.Cancel, keys.Quit},
	}
}

// renderHelpView draws the scrollable cheat sheet with a help footer.
func (m Model) renderHelpView() string {
	// Keep the viewport content fresh (width-dependent wrapping).
	m.helpViewport.SetContent(helpContent())

	body := lipgloss.JoinVertical(lipgloss.Center,
		m.helpViewport.View(),
		"",
		m.help.View(helpKeys{}),
	)
	return lipgloss.Place(m.width, max(10, m.height-6), lipgloss.Center, lipgloss.Center, body)
}
