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
	abandoned int
	lastEvent *models.TimerSession
}

func (r *recordingObserver) OnSessionStarted(session *models.TimerSession) {
	r.started++
	r.lastEvent = session
}
func (r *recordingObserver) OnTimerTick(remaining time.Duration) {
	r.ticks = append(r.ticks, remaining)
}
func (r *recordingObserver) OnSessionPaused(session *models.TimerSession)  { r.paused++ }
func (r *recordingObserver) OnSessionResumed(session *models.TimerSession) { r.resumed++ }
func (r *recordingObserver) OnSessionCompleted(session *models.TimerSession) {
	r.completed++
	r.lastEvent = session
}
func (r *recordingObserver) OnSessionAbandoned(session *models.TimerSession) {
	r.abandoned++
	r.lastEvent = session
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

func TestSkipSessionCompletesCreditsAndRecordsElapsed(t *testing.T) {
	s, obs := newTestService()
	if err := s.StartSessionWithDuration(models.FocusSession, 10*time.Minute); err != nil {
		t.Fatalf("StartSessionWithDuration: %v", err)
	}
	// Two minutes pass (ticks are deterministic here).
	for i := 0; i < 120; i++ {
		s.tickOnce()
	}

	if err := s.SkipSession(); err != nil {
		t.Fatalf("SkipSession: %v", err)
	}
	session := s.GetCurrentSession()
	if session.Status != models.Completed {
		t.Fatalf("status = %q, want completed", session.Status)
	}
	if session.Elapsed != 2*time.Minute {
		t.Fatalf("elapsed = %v, want 2m", session.Elapsed)
	}
	if s.FocusSessionsCompleted() != 1 {
		t.Fatalf("focusCompleted = %d, want 1 (skip advances the rhythm)", s.FocusSessionsCompleted())
	}
	if obs.completed != 1 || obs.abandoned != 0 {
		t.Fatalf("events: completed=%d abandoned=%d, want 1/0", obs.completed, obs.abandoned)
	}
	// Skipping a skipped (already completed) session is a no-op.
	if err := s.SkipSession(); err != nil {
		t.Fatalf("SkipSession after completion: %v", err)
	}
	if obs.completed != 1 {
		t.Fatalf("completed events = %d, want still 1", obs.completed)
	}
}

func TestStopSessionAbandonsWithoutCredit(t *testing.T) {
	s, obs := newTestService()
	if err := s.StartSessionWithDuration(models.FocusSession, 10*time.Minute); err != nil {
		t.Fatalf("StartSessionWithDuration: %v", err)
	}
	for i := 0; i < 60; i++ {
		s.tickOnce()
	}

	if err := s.StopSession(); err != nil {
		t.Fatalf("StopSession: %v", err)
	}
	if got := s.GetCurrentSession(); got != nil {
		t.Fatalf("current session = %v, want nil after abandon", got)
	}
	if s.FocusSessionsCompleted() != 0 {
		t.Fatalf("focusCompleted = %d, want 0 (abandon gives no credit)", s.FocusSessionsCompleted())
	}
	if obs.completed != 0 || obs.abandoned != 1 {
		t.Fatalf("events: completed=%d abandoned=%d, want 0/1", obs.completed, obs.abandoned)
	}
	if obs.lastEvent == nil || obs.lastEvent.Elapsed != 0 || obs.lastEvent.CompletedAt != nil {
		t.Fatalf("abandoned session must not look completed: %+v", obs.lastEvent)
	}
	// Ticks after abandon must be harmless no-ops.
	s.tickOnce()
	if got := s.GetRemainingTime(); got != 0 {
		t.Fatalf("remaining after abandon = %v, want 0", got)
	}
}

func TestStartSessionForTaskAttributesFromFirstEvent(t *testing.T) {
	s, obs := newTestService()
	if err := s.StartSessionForTask("task-42", models.FocusSession, 5*time.Minute); err != nil {
		t.Fatalf("StartSessionForTask: %v", err)
	}
	if got := s.GetCurrentSession().TaskID; got != "task-42" {
		t.Fatalf("session TaskID = %q, want task-42", got)
	}
	// The very first observer event must already carry the attribution —
	// completion stats never see an unattributed session.
	if obs.started != 1 || obs.lastEvent.TaskID != "task-42" {
		t.Fatalf("first event attribution: started=%d taskID=%q", obs.started, obs.lastEvent.TaskID)
	}

	// Skipping that session credits and records with attribution intact.
	for i := 0; i < 30; i++ {
		s.tickOnce()
	}
	if err := s.SkipSession(); err != nil {
		t.Fatalf("SkipSession: %v", err)
	}
	if obs.lastEvent.TaskID != "task-42" || obs.lastEvent.Elapsed != 30*time.Second {
		t.Fatalf("completed session attribution/elapsed: %+v", obs.lastEvent)
	}
}
