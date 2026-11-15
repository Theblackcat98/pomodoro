package views

import (
	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
)

// Palette tokens. The look is intentionally "monastic" (docs/layout.md): a
// tomato accent for focus, gentle green for short breaks, blue for long
// breaks, amber for paused, and muted text everywhere else.
var (
	// Session accents
	colorFocus      = lipgloss.Color("#E5484D") // tomato red
	colorShortBreak = lipgloss.Color("#30A46C") // gentle green
	colorLongBreak  = lipgloss.Color("#0091FF") // calm blue
	colorPaused     = lipgloss.Color("#FFB224") // amber
	colorDone       = lipgloss.Color("#46A758")

	// Text
	colorText  = lipgloss.AdaptiveColor{Light: "#3F3F3F", Dark: "#EEEEEE"}
	colorDim   = lipgloss.AdaptiveColor{Light: "#8F8F8F", Dark: "#6E6E6E"}
	colorFaint = lipgloss.AdaptiveColor{Light: "#B4B4B4", Dark: "#454545"}

	// Brand accent (tabs, selection)
	colorAccent = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	colorOnAccent = lipgloss.Color("#FFF7DB")
	colorNeutral  = lipgloss.Color("#8B8B8B")
)

// sessionAccent returns the accent color for a session type/status. Paused
// always shows amber; a finished/unknown session shows the dim text color.
func sessionAccent(sessionType models.SessionType, status models.SessionStatus) lipgloss.Color {
	if status == models.Paused {
		return colorPaused
	}
	switch sessionType {
	case models.FocusSession:
		return colorFocus
	case models.ShortBreak:
		return colorShortBreak
	case models.LongBreak:
		return colorLongBreak
	default:
		return colorFocus
	}
}

// Shared style building blocks. Kept in one place so every view stays
// consistent (specs/2-UI-Layout: centralized styles).
var (
	appStyle = lipgloss.NewStyle().Foreground(colorText)

	// Header
	headerStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 1)

	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorFocus)

	// Tabs
	tabStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true).
			Underline(true).
			Padding(0, 1)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Padding(0, 1)

	statusChipStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	statusHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#5A5A5A", Dark: "#8B8B8B"})

	// Buttons
	primaryButtonStyle = lipgloss.NewStyle().
				Foreground(colorOnAccent).
				Bold(true).
				Padding(0, 3)

	secondaryButtonStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colorFaint).
				Padding(0, 2)

	capsuleStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(colorFaint).
			Padding(0, 1)

	capsuleActiveStyle = capsuleStyle.
				Bold(true)

	// Panels
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(colorFaint).
			Padding(1, 2)

	flashStyle = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.RoundedBorder(), true).
			Padding(0, 2)

	// Toast (transient status message)
	toastStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)
)

// hintStyle renders a single key hint like "space start".
var keyNameStyle = lipgloss.NewStyle().
	Foreground(colorText).
	Bold(true)

// hint composes a key hint with a dim action label, e.g. "space pause".
func hint(key, action string) string {
	return keyNameStyle.Render(key) + statusHintStyle.Render(" "+action)
}

// hintJoin joins hints with separators.
func hintJoin(hints ...string) string {
	out := ""
	for i, h := range hints {
		if i > 0 {
			out += statusHintStyle.Render(" · ")
		}
		out += h
	}
	return out
}
