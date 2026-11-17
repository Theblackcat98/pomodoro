package views

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
	"pomodoro/src/models"
)

// Theme names stored in Settings.Theme. "default" is the legacy persisted
// value and resolves to the dark palette so long-standing installs keep
// their look.
const (
	ThemeDefault = "default"
	ThemeDark    = "dark"
	ThemeLight   = "light"
)

// hasDarkBG is resolved once for HasDarkBG() compatibility. Palette choice is
// now driven by the selected theme, not by terminal detection. Detection is
// best-effort: without a terminal we assume dark.
var hasDarkBG = detectDarkBG()

func detectDarkBG() (dark bool) {
	defer func() {
		if recover() != nil {
			dark = true
		}
	}()
	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}

// Palette hex constants — the source of truth for the accents, kept as hex
// strings because the completion pulse blends them with go-colorful.
const (
	hexFocus      = "#E5484D" // tomato red
	hexShortBreak = "#30A46C" // gentle green
	hexLongBreak  = "#0091FF" // calm blue
	hexPaused     = "#FFB224" // amber
)

// palette is the set of colors that vary with the chosen theme. Session
// accents, on-accent and neutral colors are fixed across both themes.
type palette struct {
	dark   bool
	text   color.Color
	dim    color.Color
	faint  color.Color
	accent color.Color
	hint   color.Color
}

var (
	// darkPalette reproduces the original dark-path colors byte-for-byte.
	darkPalette = palette{
		dark:   true,
		text:   lipgloss.Color("#EEEEEE"),
		dim:    lipgloss.Color("#6E6E6E"),
		faint:  lipgloss.Color("#454545"),
		accent: lipgloss.Color("#7D56F4"),
		hint:   lipgloss.Color("#8B8B8B"),
	}
	lightPalette = palette{
		dark:   false,
		text:   lipgloss.Color("#3F3F3F"),
		dim:    lipgloss.Color("#8F8F8F"),
		faint:  lipgloss.Color("#B4B4B4"),
		accent: lipgloss.Color("#874BFD"),
		hint:   lipgloss.Color("#5A5A5A"),
	}
)

// activePalette and the style variables below are rebuilt by buildStyles;
// they are package-level so every view keeps using the same exported names.
var activePalette = darkPalette

// resolvePalette maps a Settings.Theme value to a palette. "default" (and any
// unknown value) falls back to dark, preserving the app's long-standing look.
func resolvePalette(theme string) palette {
	if theme == ThemeLight {
		return lightPalette
	}
	return darkPalette
}

// buildStyles resolves a theme and (re)builds the shared style set. Call it
// at startup from NewModel and again on a theme change so the new palette is
// visible on the very next render — no restart required.
func buildStyles(theme string) {
	activePalette = resolvePalette(theme)

	colorText = activePalette.text
	colorDim = activePalette.dim
	colorFaint = activePalette.faint
	colorAccent = activePalette.accent

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
		Foreground(activePalette.hint)

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

	keyNameStyle = lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true)
}

// Fixed palette tokens. Session accents stay the same in both themes.
var (
	colorFocus      = lipgloss.Color(hexFocus)
	colorShortBreak = lipgloss.Color(hexShortBreak)
	colorLongBreak  = lipgloss.Color(hexLongBreak)
	colorPaused     = lipgloss.Color(hexPaused)
	colorDone       = lipgloss.Color("#46A758")

	colorOnAccent = lipgloss.Color("#FFF7DB")
	colorNeutral  = lipgloss.Color("#8B8B8B")
)

// Mutable palette tokens and shared styles, assigned by buildStyles.
var (
	colorText, colorDim, colorFaint, colorAccent color.Color

	appStyle, headerStyle, brandStyle, tabStyle, activeTabStyle lipgloss.Style
	statusBarStyle, statusChipStyle, statusHintStyle            lipgloss.Style
	primaryButtonStyle, secondaryButtonStyle                    lipgloss.Style
	capsuleStyle, capsuleActiveStyle                            lipgloss.Style
	panelStyle, flashStyle, toastStyle, keyNameStyle            lipgloss.Style
)

// sessionAccentHex is the hex-string twin of sessionAccent, for blending.
func sessionAccentHex(sessionType models.SessionType, status models.SessionStatus) string {
	if status == models.Paused {
		return hexPaused
	}
	switch sessionType {
	case models.FocusSession:
		return hexFocus
	case models.ShortBreak:
		return hexShortBreak
	case models.LongBreak:
		return hexLongBreak
	default:
		return hexFocus
	}
}

// sessionAccent returns the accent color for a session type/status. Paused
// always shows amber; a finished/unknown session shows the focus color.
func sessionAccent(sessionType models.SessionType, status models.SessionStatus) color.Color {
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
