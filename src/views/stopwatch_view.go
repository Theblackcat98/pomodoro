package views

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/stopwatch"
	"charm.land/lipgloss/v2"
	"pomodoro/src/utils"
)

// stopwatchTickInterval drives the Bubbles stopwatch engine. Bubbles'
// stopwatch.New does not set a default interval, so pass it explicitly.
const stopwatchTickInterval = time.Second

// stopwatchLapWindow is the maximum number of laps drawn at once. The list is
// intentionally bounded: older laps scroll out instead of growing the column
// without limit (docs/layout.md "scrollable list").
const stopwatchLapWindow = 5

// stopwatchState owns the Bubbles stopwatch engine and the recorded laps.
// engine is the single source of truth for elapsed time and running state.
type stopwatchState struct {
	engine stopwatch.Model
	laps   []time.Duration
	top    int // index of the first visible lap
}

func newStopwatchState() stopwatchState {
	return stopwatchState{
		engine: stopwatch.New(stopwatch.WithInterval(stopwatchTickInterval)),
	}
}

// addLap records the current elapsed value, keeping the newest lap in view.
func (s *stopwatchState) addLap() {
	s.laps = append(s.laps, s.engine.Elapsed())
	if len(s.laps) > s.top+stopwatchLapWindow {
		s.top = len(s.laps) - stopwatchLapWindow
	}
}

// clear resets laps and their scroll position (the engine's elapsed time is
// reset separately by a Bubbles ResetMsg).
func (s *stopwatchState) clear() {
	s.laps = nil
	s.top = 0
}

// visibleRange returns the half-open [start,end) slice of laps to draw. It is
// always a bounded window and is safe for empty/short lap lists.
func (s stopwatchState) visibleRange() (start, end int) {
	window := stopwatchLapWindow
	if window > len(s.laps) {
		window = len(s.laps)
	}
	start = s.top
	if start < 0 {
		start = 0
	}
	if start > len(s.laps)-window {
		start = len(s.laps) - window
	}
	return start, start + window
}

// scroll moves the laps window by delta rows (+1 down / -1 up), clamped to the
// available laps.
func (s *stopwatchState) scroll(delta int) {
	maxTop := len(s.laps) - stopwatchLapWindow
	if maxTop < 0 {
		maxTop = 0
	}
	s.top += delta
	if s.top < 0 {
		s.top = 0
	}
	if s.top > maxTop {
		s.top = maxTop
	}
}

// renderStopwatchView draws the same ring as the timer, but as a plain upward
// counter: the digits grow from zero, a lap button sits below, and earlier
// laps stack beneath in a bounded, scrollable list.
func (m Model) renderStopwatchView() string {
	elapsed := m.stopwatch.engine.Elapsed()
	running := m.stopwatch.engine.Running()

	centerSub := "READY"
	if running {
		centerSub = "COUNTING"
	} else if elapsed > 0 {
		centerSub = "PAUSED"
	}
	accent := colorFocus

	// The arc sweeps once per minute ("time flows upward"), rather than
	// counting down to a fixed duration like the timer.
	fraction := float64(elapsed%time.Minute) / float64(time.Minute)

	radius := defaultRingRadius
	if m.compact() {
		radius = 3
	}
	ring := renderRingSized(fraction,
		lipgloss.NewStyle().Foreground(accent),
		lipgloss.NewStyle().Foreground(colorFaint),
		[]string{
			lipgloss.NewStyle().Foreground(accent).Bold(true).Render("STOPWATCH"),
			lipgloss.NewStyle().Bold(true).Render(spacedTime(utils.FormatDuration(elapsed))),
			lipgloss.NewStyle().Foreground(colorDim).Render(centerSub),
		},
		radius,
	)

	return lipgloss.JoinVertical(lipgloss.Center,
		ring,
		"",
		m.renderStopwatchControls(),
		"",
		m.renderStopwatchLaps(),
	)
}

// renderStopwatchControls draws the primary start/stop button alongside lap
// and reset buttons.
func (m Model) renderStopwatchControls() string {
	primaryLabel := "▶ START"
	if m.stopwatch.engine.Running() {
		primaryLabel = "⏸ STOP"
	}
	main := primaryButtonStyle.Background(colorFocus).Render(primaryLabel)
	return lipgloss.JoinHorizontal(lipgloss.Center,
		main, "  ",
		secondaryButtonStyle.Render("LAP"), "  ",
		secondaryButtonStyle.Render("RESET"),
	)
}

// renderStopwatchLaps renders only the bounded visible window, plus a scroll
// hint when laps overflow it.
func (m Model) renderStopwatchLaps() string {
	if len(m.stopwatch.laps) == 0 {
		return lipgloss.NewStyle().Foreground(colorDim).Render("No laps yet · l to record")
	}

	start, end := m.stopwatch.visibleRange()
	lines := []string{
		lipgloss.NewStyle().Foreground(colorDim).Bold(true).Render("LAPS"),
	}
	for i := start; i < end; i++ {
		marker := "  "
		if i == end-1 {
			marker = "› "
		}
		lines = append(lines, fmt.Sprintf("%s%s  %s",
			marker,
			lipgloss.NewStyle().Foreground(colorFocus).Render(fmt.Sprintf("%02d", i+1)),
			utils.FormatDuration(m.stopwatch.laps[i]),
		))
	}
	if start > 0 || end < len(m.stopwatch.laps) {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render(
			fmt.Sprintf("↑↓ scroll · %d/%d", end, len(m.stopwatch.laps)),
		))
	}
	return strings.Join(lines, "\n")
}
