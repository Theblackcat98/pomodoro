package services

import (
	"testing"
	"time"

	"pomodoro/src/models"
)

// recordingObserver captures every service event for assertions.
type recordingObserver struct {
	ticks     []time.Duration
	started   int
	paused    int
	resumed   int
	completed int
}

func (r *recordingObserver) OnSessionStarted(session *models.TimerSession) { r.started++ }
func (r *recordingObserver) OnTimerTick(remaining time.Duration)           { r.ticks = append(r.ticks, remaining) }
func (r *recordingObserver) OnSessionPaused(session *models.TimerSession)  { r.paused++ }
func (r *recordingObserver) OnSessionResumed(session *models.TimerSession) { r.resumed++ }
func (r *recordingObserver) OnSessionCompleted(session *models.TimerSession) {
	r.completed++
}

func newTestService() (*ConcreteTimerService, *recordingObserver) {
	s := NewTimerService(models.DefaultSettings()).(*ConcreteTimerService)
	obs := &recordingObserver{}
	s.RegisterObserver(obs)
	return s, obs
}

func TestStartSessionUsesSettingsDuration(t *testing.T) {
	s, obs := newTestService()

	if err := s.StartSession(models.FocusSession); err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if got := s.GetRemainingTime(); got != 25*time.Minute {
		t.Fatalf("remaining = %v, want 25m", got)
	}
	if obs.started != 1 {
		t.Fatalf("started events = %d, want 1", obs.started)
	}
}

func TestTickDecrementsOnlyWhileRunning(t *testing.T) {
	s, obs := newTestService()
	if err := s.StartSessionWithDuration(models.FocusSession, 10*time.Second); err != nil {
		t.Fatalf("StartSessionWithDuration: %v", err)
	}

	s.tickOnce()
	s.tickOnce()
	if got := s.GetRemainingTime(); got != 8*time.Second {
		t.Fatalf("remaining after 2 ticks = %v, want 8s", got)
	}

	if err := s.PauseSession(); err != nil {
		t.Fatalf("PauseSession: %v", err)
	}
	s.tickOnce()
	s.tickOnce()
	if got := s.GetRemainingTime(); got != 8*time.Second {
		t.Fatalf("remaining after paused ticks = %v, want 8s (ticks must not run while paused)", got)
	}
	if obs.paused != 1 {
		t.Fatalf("paused events = %d, want 1", obs.paused)
	}

	if err := s.ResumeSession(); err != nil {
		t.Fatalf("ResumeSession: %v", err)
	}
	s.tickOnce()
	if got := s.GetRemainingTime(); got != 7*time.Second {
		t.Fatalf("remaining after resume tick = %v, want 7s", got)
	}
}

func TestSessionCompletesAndCountsFocus(t *testing.T) {
	s, obs := newTestService()
	if err := s.StartSessionWithDuration(models.FocusSession, 3*time.Second); err != nil {
		t.Fatalf("StartSessionWithDuration: %v", err)
	}

	s.tickOnce()
	s.tickOnce()
	s.tickOnce()

	session := s.GetCurrentSession()
	if session.Status != models.Completed {
		t.Fatalf("status = %q, want completed", session.Status)
	}
	if s.GetRemainingTime() != 0 {
		t.Fatalf("remaining = %v, want 0", s.GetRemainingTime())
	}
	if s.FocusSessionsCompleted() != 1 {
		t.Fatalf("focusCompleted = %d, want 1", s.FocusSessionsCompleted())
	}
	if obs.completed != 1 {
		t.Fatalf("completed events = %d, want 1", obs.completed)
	}
}

func TestSuggestNextFollowsBreakRhythm(t *testing.T) {
	settings := models.DefaultSettings()
	settings.LongBreakInterval = 4
	s := NewTimerService(settings).(*ConcreteTimerService)

	if got := s.SuggestNext(); got != models.FocusSession {
		t.Fatalf("fresh suggestion = %q, want focus", got)
	}

	// Complete LongBreakInterval focus sessions; a long break follows the last.
	for i := 0; i < settings.LongBreakInterval; i++ {
		if err := s.StartSessionWithDuration(models.FocusSession, time.Second); err != nil {
			t.Fatalf("start focus %d: %v", i+1, err)
		}
		s.tickOnce()
	}

	want := models.LongBreak
	if got := s.SuggestNext(); got != want {
		t.Fatalf("after %d focuses suggestion = %q, want %q", settings.LongBreakInterval, got, want)
	}

	// Complete the long break; focus comes next.
	if err := s.StartSessionWithDuration(models.LongBreak, time.Second); err != nil {
		t.Fatalf("start long break: %v", err)
	}
	s.tickOnce()
	if got := s.SuggestNext(); got != models.FocusSession {
		t.Fatalf("after break suggestion = %q, want focus", got)
	}
}

func TestSecondFocusSuggestsShortBreak(t *testing.T) {
	s, _ := newTestService() // interval 4
	for i := 0; i < 3; i++ {
		if err := s.StartSessionWithDuration(models.FocusSession, time.Second); err != nil {
			t.Fatalf("start focus %d: %v", i+1, err)
		}
		s.tickOnce()
		if got := s.SuggestNext(); got != models.ShortBreak {
			t.Fatalf("after %d focuses suggestion = %q, want short break", i+1, got)
		}
		// Consume the suggested break so the next focus starts fresh.
		if err := s.StartSessionWithDuration(models.ShortBreak, time.Second); err != nil {
			t.Fatalf("start short break: %v", err)
		}
		s.tickOnce()
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	s, _ := newTestService()
	if err := s.StartSession(models.FocusSession); err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	s.Shutdown()
	s.Shutdown() // must not panic on double close
}
