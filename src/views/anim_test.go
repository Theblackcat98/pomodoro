package views

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestFlashDigitsMarksChangedPositions(t *testing.T) {
	var a animSet
	a.flashDigits("24:59", "25:00")

	// "24:59" → "25:00": positions 1, 3, 4 change; '2' and ':' do not.
	want := [5]float64{0, 1, 0, 1, 1}
	for i, f := range a.digitFlash {
		if f != want[i] {
			t.Fatalf("digitFlash[%d] = %v, want %v", i, f, want[i])
		}
	}

	// Only the seconds change on a normal tick.
	a2 := animSet{}
	a2.flashDigits("24:58", "24:59")
	if a2.digitFlash[4] != 1 {
		t.Fatalf("ones-second should flash, got %v", a2.digitFlash[4])
	}
	if a2.digitFlash[0] != 0 || a2.digitFlash[1] != 0 || a2.digitFlash[2] != 0 || a2.digitFlash[3] != 0 {
		t.Fatalf("unchanged digits must not flash: %+v", a2.digitFlash)
	}

	// A fresh readout (empty previous) flashes everything.
	a3 := animSet{}
	a3.flashDigits("25:00", "")
	for i, f := range a3.digitFlash {
		if f != 1 {
			t.Fatalf("first readout digitFlash[%d] = %v, want 1", i, f)
		}
	}
}

func TestSpringSettlesTowardTarget(t *testing.T) {
	s := newRingSpring()
	s.snap(0)
	s.setTarget(1)

	// No movement before stepping.
	if s.pos != 0 {
		t.Fatalf("pos before steps = %v, want 0", s.pos)
	}

	// The spring is slightly under-damped: it settles just past 1s.
	for i := 0; i < 2*animFPS && !s.settled(); i++ {
		s.step()
	}
	if !s.settled() {
		t.Fatalf("spring should settle within 2s of frames (pos=%v vel=%v)", s.pos, s.vel)
	}
	if s.pos < 0.95 || s.pos > 1.05 {
		t.Fatalf("settled pos = %v, want ≈1 (small overshoot allowed)", s.pos)
	}
}

func TestDecayValueClampsAtZero(t *testing.T) {
	v := 0.5
	decayValue(&v, 0.2)
	if v != 0.3 {
		t.Fatalf("v = %v, want 0.3", v)
	}
	decayValue(&v, 0.9)
	if v != 0 {
		t.Fatalf("v = %v, want clamped to 0", v)
	}
}

func TestBlendColorsEndpoints(t *testing.T) {
	a := hexFocus
	b := hexShortBreak

	if got := blendColors(a, b, 0); got != lipgloss.Color(a) {
		t.Fatalf("blend t=0 = %v, want a", got)
	}
	if got := blendColors(a, b, 1); got != lipgloss.Color(b) {
		t.Fatalf("blend t=1 = %v, want b", got)
	}
	// Midpoint is a valid hex color distinct from both endpoints.
	mid := blendColors(a, b, 0.5)
	if mid == lipgloss.Color(a) || mid == lipgloss.Color(b) {
		t.Fatalf("blend t=0.5 = %v, want a distinct color", mid)
	}
}

func TestAnimActiveLifecycle(t *testing.T) {
	a := newAnimSet()
	if a.active(false) {
		t.Fatal("fresh animSet must be inactive")
	}

	// Breathing alone keeps the loop alive while paused.
	if !a.active(true) {
		t.Fatal("paused session must keep the loop alive (breathing)")
	}
	if a.active(false) {
		t.Fatal("not paused: breathing alone must not keep the loop alive")
	}

	// A pulse keeps it alive until decayed.
	a.startPulse(hexFocus, hexShortBreak)
	if !a.active(false) {
		t.Fatal("pulse must keep the loop alive")
	}
	for i := 0; i < animFPS; i++ {
		a.step(false)
	}
	if a.pulse != 0 {
		t.Fatalf("pulse after 1s = %v, want 0", a.pulse)
	}
	if a.active(false) {
		t.Fatal("fully decayed animSet must be inactive")
	}
}

func TestTabSlideDirection(t *testing.T) {
	if got := tabSlideDir(TabTimer, TabTasks); got != 1 {
		t.Fatalf("forward slide = %v, want 1", got)
	}
	if got := tabSlideDir(TabHelp, TabTimer); got != -1 {
		t.Fatalf("wrap-back slide = %v, want -1", got)
	}
	if got := tabSlideDir(TabTimer, TabTimer); got != 0 {
		t.Fatalf("same-tab slide = %v, want 0", got)
	}
}

func TestComputeTabHitsCoversAllTabsInOrder(t *testing.T) {
	hits := computeTabHits()
	if len(hits) != tabCount {
		t.Fatalf("hits = %d, want %d", len(hits), tabCount)
	}
	for i, h := range hits {
		if h.tab != Tab(i) {
			t.Fatalf("hit %d tab = %d, want %d", i, h.tab, i)
		}
		if h.endX <= h.startX {
			t.Fatalf("hit %d has empty range %d..%d", i, h.startX, h.endX)
		}
		if i > 0 && hits[i-1].endX >= h.startX {
			t.Fatalf("hit %d overlaps previous (%d..%d vs %d..%d)",
				i, hits[i-1].startX, hits[i-1].endX, h.startX, h.endX)
		}
	}
}
