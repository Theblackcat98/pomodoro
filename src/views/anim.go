package views

import (
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// Shared animation engine.
//
// One 30fps tick loop drives every animated value in the app. It runs only
// while something is actually animating: handlers start animations via
// kickAnim (which starts the chain exactly once), and the animTickMsg handler
// is the only place the chain is re-armed — the same discipline that fixed
// the spinner tick storm. Renderers read animSet values; they never step
// them, so rendering stays a pure function of state.

// animFPS is the shared loop's frame rate. 30fps keeps the v1 renderer
// comfortable; revisit 60fps after the v2 (Cursed renderer) migration.
const animFPS = 30

const animFrame = time.Second / animFPS

// Decay rates, in value-units per frame at animFPS.
const (
	pulseDecayPerFrame = 1.0 / (0.8 * animFPS)   // completion flash ≈ 0.8s
	digitDecayPerFrame = 1.0 / (0.45 * animFPS)  // digit settle ≈ 0.45s
	toastDecayPerFrame = 1.0 / (0.18 * animFPS)  // toast slide-in ≈ 0.18s
	tabDecayPerFrame   = 1.0 / (0.16 * animFPS)  // tab body slide ≈ 0.16s
	flashDecayPerFrame = 1.0 / (0.35 * animFPS)  // settings value flash ≈ 0.35s
	breathePeriod      = 2600 * time.Millisecond // paused "breathing" cycle
)

// animTickMsg advances all running animations by one frame.
type animTickMsg time.Time

func animTickCmd() tea.Cmd {
	return tea.Tick(animFrame, func(t time.Time) tea.Msg { return animTickMsg(t) })
}

// springValue is a harmonica spring around a scalar in [0,1].
type springValue struct {
	spring harmonica.Spring
	pos    float64
	vel    float64
	target float64
}

func newRingSpring() springValue {
	// Slightly under-damped: a hint of overshoot makes the fill/empty sweeps
	// feel alive without wobbling.
	return springValue{spring: harmonica.NewSpring(harmonica.FPS(animFPS), 6.0, 0.8)}
}

// snap teleports the value (used when a new session starts).
func (s *springValue) snap(v float64) {
	s.pos, s.vel, s.target = clamp01(v), 0, clamp01(v)
}

// setTarget eases toward v.
func (s *springValue) setTarget(v float64) { s.target = clamp01(v) }

func (s *springValue) settled() bool {
	return math.Abs(s.pos-s.target) < 0.002 && math.Abs(s.vel) < 0.002
}

func (s *springValue) step() {
	s.pos, s.vel = s.spring.Update(s.pos, s.vel, s.target)
	if s.pos < 0 {
		s.pos = 0
	}
}

// animSet bundles every animated value. Stepped only by the anim loop.
type animSet struct {
	// Ring: spring-eased elapsed fraction. Springs the fill in on start,
	// sweeps to full on completion, empties on reset.
	ring springValue

	// Completion pulse: one-shot 1→0 decay. While >0 the ring accent lerps
	// from pulseFrom to pulseTo, flashing the state transition.
	pulse              float64
	pulseFrom, pulseTo lipgloss.Color

	// Breathe: [0,1] phase of the paused ring track. Only advances while a
	// session is paused.
	breathe float64

	// digitFlash[i]: per-character brightness of the "MM:SS" readout; the
	// digits that just changed flash and settle back.
	digitFlash [5]float64

	// toastSlide: 1→0 slide-in offset for the status-bar toast.
	toastSlide float64

	// tabSlide: signed 1→0 body slide on tab switches (+1 forward, -1 back).
	tabSlide float64

	// flashRow/flashValue: the settings row whose value just changed.
	flashRow   int
	flashValue float64
}

func newAnimSet() animSet {
	return animSet{ring: newRingSpring(), flashRow: -1}
}

// step advances every animation one frame. paused reports whether the current
// session is paused (breathing only runs then).
func (a *animSet) step(paused bool) {
	a.ring.step()
	decayValue(&a.pulse, pulseDecayPerFrame)
	if paused {
		a.breathe += animFrame.Seconds() / breathePeriod.Seconds()
		a.breathe -= math.Floor(a.breathe)
	} else {
		a.breathe = 0
	}
	for i := range a.digitFlash {
		decayValue(&a.digitFlash[i], digitDecayPerFrame)
	}
	decayValue(&a.toastSlide, toastDecayPerFrame)
	// Slide offsets ease toward zero without crossing it.
	if a.tabSlide > 0 {
		decayValue(&a.tabSlide, tabDecayPerFrame)
	} else if a.tabSlide < 0 {
		a.tabSlide = math.Min(0, a.tabSlide+tabDecayPerFrame)
	}
	decayValue(&a.flashValue, flashDecayPerFrame)
}

// active reports whether the loop must keep running.
func (a *animSet) active(paused bool) bool {
	if !a.ring.settled() || a.pulse > 0 || a.toastSlide > 0 || a.tabSlide != 0 || a.flashValue > 0 {
		return true
	}
	if paused {
		return true // breathing runs for as long as the pause does
	}
	for _, f := range a.digitFlash {
		if f > 0 {
			return true
		}
	}
	return false
}

// flashDigits sets the per-character flash for the digits that changed.
func (a *animSet) flashDigits(digits, prev string) {
	d, p := []rune(digits), []rune(prev)
	for i := 0; i < len(a.digitFlash) && i < len(d); i++ {
		if i >= len(p) || d[i] != p[i] {
			a.digitFlash[i] = 1
		}
	}
}

// startPulse arms the completion flash, morphing from one accent to the next.
func (a *animSet) startPulse(from, to lipgloss.Color) {
	a.pulse, a.pulseFrom, a.pulseTo = 1, from, to
}

// decayValue steps v down by rate; clamps at zero.
func decayValue(v *float64, rate float64) {
	*v -= rate
	if *v < 0 {
		*v = 0
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// --- color blending ---

// blendColors lerps two hex colors; t=0 → a, t=1 → b. Used for the completion
// morph and the paused breathing (no terminal alpha, so we blend the palette).
func blendColors(a, b lipgloss.Color, t float64) lipgloss.Color {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	ca, errA := colorful.Hex(string(a))
	cb, errB := colorful.Hex(string(b))
	if errA != nil || errB != nil {
		return a
	}
	return lipgloss.Color(ca.BlendLuv(cb, t).Clamped().Hex())
}

// blendAdaptive resolves an AdaptiveColor for the current terminal and blends.
func blendAdaptive(a, b lipgloss.AdaptiveColor, t float64) lipgloss.Color {
	return blendColors(resolveAdaptive(a), resolveAdaptive(b), t)
}

// resolveAdaptive picks the light/dark variant for this terminal.
func resolveAdaptive(c lipgloss.AdaptiveColor) lipgloss.Color {
	if lipgloss.HasDarkBackground() {
		return lipgloss.Color(c.Dark)
	}
	return lipgloss.Color(c.Light)
}
