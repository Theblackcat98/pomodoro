package views

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// sizedModel drives a real Model through a WindowSizeMsg so the responsive
// layout branch in Update runs exactly as it does at runtime.
func sizedModel(width, height int) Model {
	m := newBareModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return next.(Model)
}

// TestLayoutCompactWindowSize covers the compact branch: a window smaller
// than 80x20 shrinks the progress bar and uses the narrow session split.
func TestLayoutCompactWindowSize(t *testing.T) {
	m := sizedModel(79, 19)

	if m.width != 79 || m.height != 19 {
		t.Fatalf("window size = %dx%d, want 79x19", m.width, m.height)
	}
	if !m.compact() {
		t.Fatal("compact() = false at 79x19, want true")
	}
	if got := m.progress.Width(); got != 24 {
		t.Fatalf("progress width = %d at 79x19, want 24 (compact branch)", got)
	}
	// Narrow split is max(20, width-8); the wide split would be max(20, width-36).
	wantSplit := max(20, 79-8)
	if got := m.sessTable.Width(); got != wantSplit {
		t.Fatalf("session table width = %d at 79x19, want %d (narrow split)", got, wantSplit)
	}
	if got := m.helpViewport.Width(); got != max(20, 79-8) {
		t.Fatalf("help viewport width = %d at 79x19, want %d", got, max(20, 79-8))
	}
}

// TestLayoutNormalWindowSize covers the normal branch: a roomy window keeps
// the full-width progress bar and the wide (chart + table) session split.
func TestLayoutNormalWindowSize(t *testing.T) {
	m := sizedModel(120, 40)

	if m.width != 120 || m.height != 40 {
		t.Fatalf("window size = %dx%d, want 120x40", m.width, m.height)
	}
	if m.compact() {
		t.Fatal("compact() = true at 120x40, want false")
	}
	if got := m.progress.Width(); got != 40 {
		t.Fatalf("progress width = %d at 120x40, want 40 (normal branch)", got)
	}
	// Wide split is max(20, width-36); the narrow split would be max(20, width-8).
	wantSplit := max(20, 120-36)
	if got := m.sessTable.Width(); got != wantSplit {
		t.Fatalf("session table width = %d at 120x40, want %d (wide split)", got, wantSplit)
	}
	if got := m.helpViewport.Width(); got != max(20, 120-8) {
		t.Fatalf("help viewport width = %d at 120x40, want %d", got, max(20, 120-8))
	}
}

// TestLayoutThresholdBoundaries pins the exact side of each documented
// threshold. compact() is true only when width < 80 OR height < 20, so the
// 80x20 corner is NOT compact; the session split flips at width >= 100, so
// 100 columns takes the wide split and 99 the narrow one.
func TestLayoutThresholdBoundaries(t *testing.T) {
	cases := []struct {
		name         string
		width        int
		height       int
		compact      bool
		progressW    int
		sessionW     int
		sessionSplit string
	}{
		{"width_below_80", 79, 20, true, 24, max(20, 79-8), "narrow"},
		{"height_below_20", 80, 19, true, 24, max(20, 80-8), "narrow"},
		{"corner_80x20_not_compact", 80, 20, false, 40, max(20, 80-8), "narrow"},
		{"columns_99_narrow_split", 99, 20, false, 40, max(20, 99-8), "narrow"},
		{"columns_100_wide_split", 100, 20, false, 40, max(20, 100-36), "wide"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := sizedModel(tc.width, tc.height)

			if got := m.compact(); got != tc.compact {
				t.Fatalf("compact() at %dx%d = %v, want %v", tc.width, tc.height, got, tc.compact)
			}
			if got := m.progress.Width(); got != tc.progressW {
				t.Fatalf("progress width at %dx%d = %d, want %d", tc.width, tc.height, got, tc.progressW)
			}
			if got := m.sessTable.Width(); got != tc.sessionW {
				t.Fatalf("session table width at %dx%d = %d, want %d (%s split)",
					tc.width, tc.height, got, tc.sessionW, tc.sessionSplit)
			}
		})
	}
}

// TestLayoutRenderWidth ensures the responsive sizing keeps every rendered
// row within the terminal width. lipgloss.Width counts visible cells, so the
// ANSI/style sequences embedded in the view do not skew the measurement.
func TestLayoutRenderWidth(t *testing.T) {
	for _, sz := range []struct{ width, height int }{{79, 19}, {120, 40}, {80, 20}, {100, 20}} {
		m := sizedModel(sz.width, sz.height)
		view := m.View().Content

		for i, line := range strings.Split(view, "\n") {
			if w := lipgloss.Width(line); w > m.width {
				t.Fatalf("rendered line %d at %dx%d has visible width %d, exceeds m.width %d",
					i, sz.width, sz.height, w, m.width)
			}
		}
	}
}
