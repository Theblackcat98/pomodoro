package views

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/stopwatch"
	tea "charm.land/bubbletea/v2"
)

// firstEmittedMsg runs cmd and returns the first message it emits. Bubbles'
// stopwatch.Start uses tea.Sequence (whose slice type is unexported), so we
// peek at the first command through reflection rather than sleeping on its
// tick command.
func firstEmittedMsg(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("nil command")
	}
	msg := cmd()
	v := reflect.ValueOf(msg)
	if v.IsValid() && v.Kind() == reflect.Slice && v.Len() > 0 {
		first, ok := v.Index(0).Interface().(tea.Cmd)
		if !ok {
			t.Fatalf("first sequence element is %T, not a command", v.Index(0).Interface())
		}
		return first()
	}
	return msg
}

func driveMsg(m Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

func enterStopwatch(t *testing.T, m Model) Model {
	t.Helper()
	m = driveKey(m, "w")
	if m.tab != TabStopwatch {
		t.Fatalf("tab = %d, want stopwatch (%d)", m.tab, TabStopwatch)
	}
	return m
}

func startStopwatch(t *testing.T, m Model) Model {
	t.Helper()
	cmd := m.stopwatch.engine.Toggle() // stopped -> Start
	return driveMsg(m, firstEmittedMsg(t, cmd))
}

func tickStopwatch(m Model) Model {
	return driveMsg(m, stopwatch.TickMsg{ID: m.stopwatch.engine.ID()})
}

func TestStopwatchElapsedIncreasesWhileRunning(t *testing.T) {
	m := enterStopwatch(t, newBareModel())
	if m.stopwatch.engine.Running() {
		t.Fatal("stopwatch should begin stopped")
	}
	m = startStopwatch(t, m)
	if !m.stopwatch.engine.Running() {
		t.Fatal("stopwatch should be running after toggle")
	}
	before := m.stopwatch.engine.Elapsed()
	m = tickStopwatch(m)
	after := m.stopwatch.engine.Elapsed()
	if after <= before {
		t.Fatalf("elapsed did not increase: before=%s after=%s", before, after)
	}
	if delta := after - before; delta != stopwatchTickInterval {
		t.Fatalf("one tick advanced %s, want %s", delta, stopwatchTickInterval)
	}
}

func TestStopwatchStopHaltsElapsed(t *testing.T) {
	m := startStopwatch(t, enterStopwatch(t, newBareModel()))
	m = tickStopwatch(m)
	// Second toggle returns Stop() (a plain command, no sequence).
	m = driveMsg(m, firstEmittedMsg(t, m.stopwatch.engine.Toggle()))
	if m.stopwatch.engine.Running() {
		t.Fatal("stopwatch should be stopped after second toggle")
	}
	before := m.stopwatch.engine.Elapsed()
	m = tickStopwatch(m)
	if got := m.stopwatch.engine.Elapsed(); got != before {
		t.Fatalf("elapsed changed while stopped: before=%s after=%s", before, got)
	}
}

func TestStopwatchLapsAppendInOrder(t *testing.T) {
	m := startStopwatch(t, enterStopwatch(t, newBareModel()))
	for i := 1; i <= 3; i++ {
		m = tickStopwatch(m)
		m = driveKey(m, "l")
	}
	want := []time.Duration{time.Second, 2 * time.Second, 3 * time.Second}
	if len(m.stopwatch.laps) != len(want) {
		t.Fatalf("got %d laps (%v), want %v", len(m.stopwatch.laps), m.stopwatch.laps, want)
	}
	for i := range want {
		if m.stopwatch.laps[i] != want[i] {
			t.Fatalf("lap %d = %s, want %s", i+1, m.stopwatch.laps[i], want[i])
		}
	}
}

func TestStopwatchViewRendersBoundedLapList(t *testing.T) {
	m := enterStopwatch(t, newBareModel())
	total := stopwatchLapWindow + 3
	for i := 0; i < total; i++ {
		m.stopwatch.addLap() // elapsed is 0 here; lap values need not differ
	}
	view := m.renderStopwatchView()

	if !strings.Contains(view, "LAPS") {
		t.Fatalf("view missing laps section:\n%s", view)
	}
	start, end := m.stopwatch.visibleRange()
	if end-start != stopwatchLapWindow {
		t.Fatalf("visible window = %d laps, want %d", end-start, stopwatchLapWindow)
	}
	if !strings.Contains(view, "scroll") {
		t.Fatalf("overflowing laps should show a scroll hint:\n%s", view)
	}
	// The newest lap is in view; the oldest has scrolled out.
	if !strings.Contains(view, "0"+itoa(total)) {
		t.Fatalf("newest lap %d not rendered:\n%s", total, view)
	}
	if strings.Contains(view, "01 ") {
		t.Fatalf("oldest lap should be scrolled out of the window:\n%s", view)
	}
}

func TestStopwatchScrollWindow(t *testing.T) {
	m := newBareModel()
	for i := 0; i < stopwatchLapWindow+3; i++ {
		m.stopwatch.addLap()
	}
	m.stopwatch.scroll(-100) // clamp at top
	if start, _ := m.stopwatch.visibleRange(); start != 0 {
		t.Fatalf("scrolling up clamped to %d, want 0", start)
	}
	m.stopwatch.scroll(100) // clamp at bottom
	start, end := m.stopwatch.visibleRange()
	if end != len(m.stopwatch.laps) || end-start != stopwatchLapWindow {
		t.Fatalf("bottom window = [%d,%d), want [%d,%d)", start, end, len(m.stopwatch.laps)-stopwatchLapWindow, len(m.stopwatch.laps))
	}
}

func TestStopwatchToggleDoesNotShadowSkip(t *testing.T) {
	m := newBareModel()
	// 's' still belongs to the timer (skip) and must not enter stopwatch.
	m = driveKey(m, "s")
	if m.tab != TabTimer {
		t.Fatalf("'s' changed tab to %d; it must stay on the timer", m.tab)
	}
	// 'w' is the stopwatch toggle and toggles back to the timer.
	m = driveKey(m, "w")
	if m.tab != TabStopwatch {
		t.Fatalf("'w' did not enter stopwatch: tab=%d", m.tab)
	}
	m = driveKey(m, "esc")
	if m.tab != TabTimer {
		t.Fatalf("esc did not leave stopwatch: tab=%d", m.tab)
	}
}

func TestLeavingStopwatchKeepsTimerWorking(t *testing.T) {
	m := startStopwatch(t, enterStopwatch(t, newBareModel()))
	m = tickStopwatch(m)
	m = driveKey(m, "w") // back to timer while the stopwatch keeps counting
	if m.tab != TabTimer {
		t.Fatalf("tab after leaving stopwatch = %d, want timer", m.tab)
	}
	m = driveKey(m, "space")
	if !m.sessionRunning() {
		t.Fatal("timer must still start after visiting the stopwatch")
	}
	m.timer.Shutdown()
}
