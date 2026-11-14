package services

import (
	"errors"
	"sync"
	"time"

	"pomodoro/src/models"
)

// TimerObserver receives lifecycle events from the timer service.
// Callbacks are invoked from the service's internal ticker goroutine, so
// implementations must hand events to the UI loop safely (e.g. via a channel).
type TimerObserver interface {
	OnSessionStarted(session *models.TimerSession)
	OnTimerTick(remaining time.Duration)
	OnSessionPaused(session *models.TimerSession)
	OnSessionResumed(session *models.TimerSession)
	OnSessionCompleted(session *models.TimerSession)
}

// TimerService drives Pomodoro sessions: start/pause/resume/stop, per-second
// ticking, and lifecycle notifications for observers.
type TimerService interface {
	StartSession(sessionType models.SessionType) error
	StartSessionWithDuration(sessionType models.SessionType, duration time.Duration) error
	PauseSession() error
	ResumeSession() error
	StopSession() error
	GetCurrentSession() *models.TimerSession
	GetRemainingTime() time.Duration
	FocusSessionsCompleted() int
	SuggestNext() models.SessionType
	SetSettings(settings *models.Settings)
	RegisterObserver(observer TimerObserver)
	UnregisterObserver(observer TimerObserver)
	Shutdown()
}

// ConcreteTimerService is the concrete implementation of TimerService.
// A single ticker goroutine runs for the lifetime of the service (between the
// first StartSession and Shutdown); pause/resume only flips the session status.
type ConcreteTimerService struct {
	mu               sync.Mutex
	settings         *models.Settings
	currentSession   *models.TimerSession
	focusCompleted   int
	observers        []TimerObserver
	running          bool // goroutine lifecycle
	done             chan struct{}
	once             sync.Once
}

// Compile-time check that ConcreteTimerService implements TimerService.
var _ TimerService = (*ConcreteTimerService)(nil)

// NewTimerService creates a new timer service using the given settings for
// session durations. Passing nil falls back to models.DefaultSettings().
func NewTimerService(settings *models.Settings) TimerService {
	if settings == nil {
		settings = models.DefaultSettings()
	}
	return &ConcreteTimerService{
		settings:  settings,
		observers: make([]TimerObserver, 0),
		done:      make(chan struct{}),
	}
}

// StartSession starts a new Pomodoro session using the configured duration.
func (t *ConcreteTimerService) StartSession(sessionType models.SessionType) error {
	return t.StartSessionWithDuration(sessionType, 0)
}

// StartSessionWithDuration starts a session with an explicit duration
// (used by the quick-pick capsules). A duration of 0 uses the configured one.
func (t *ConcreteTimerService) StartSessionWithDuration(sessionType models.SessionType, duration time.Duration) error {
	t.mu.Lock()
	if t.currentSession != nil && t.currentSession.Status == models.Running {
		t.mu.Unlock()
		return errors.New("a session is already running")
	}
	if duration <= 0 {
		duration = t.settings.DurationFor(sessionType)
	}

	now := time.Now()
	session := &models.TimerSession{
		ID:        models.NewID(),
		Type:      sessionType,
		Duration:  duration,
		Remaining: duration,
		Status:    models.Running,
		StartTime: &now,
		CreatedAt: now,
	}
	t.currentSession = session
	t.ensureTicker()

	observers := append([]TimerObserver(nil), t.observers...)
	t.mu.Unlock()

	notify(observers, func(o TimerObserver) { o.OnSessionStarted(session) })
	return nil
}

// PauseSession pauses the current session.
func (t *ConcreteTimerService) PauseSession() error {
	t.mu.Lock()
	if t.currentSession == nil {
		t.mu.Unlock()
		return errors.New("no active session to pause")
	}
	if t.currentSession.Status != models.Running {
		t.mu.Unlock()
		return errors.New("session is not running")
	}
	t.currentSession.Status = models.Paused
	session := t.currentSession
	observers := append([]TimerObserver(nil), t.observers...)
	t.mu.Unlock()

	notify(observers, func(o TimerObserver) { o.OnSessionPaused(session) })
	return nil
}

// ResumeSession resumes a paused session.
func (t *ConcreteTimerService) ResumeSession() error {
	t.mu.Lock()
	if t.currentSession == nil {
		t.mu.Unlock()
		return errors.New("no session to resume")
	}
	if t.currentSession.Status != models.Paused {
		t.mu.Unlock()
		return errors.New("session is not paused")
	}
	t.currentSession.Status = models.Running
	session := t.currentSession
	observers := append([]TimerObserver(nil), t.observers...)
	t.mu.Unlock()

	notify(observers, func(o TimerObserver) { o.OnSessionResumed(session) })
	return nil
}

// StopSession stops (abandons) the current session without counting it as a
// completed Pomodoro.
func (t *ConcreteTimerService) StopSession() error {
	t.mu.Lock()
	if t.currentSession == nil {
		t.mu.Unlock()
		return errors.New("no active session to stop")
	}
	if t.currentSession.Status == models.Completed {
		t.mu.Unlock()
		return nil
	}
	now := time.Now()
	t.currentSession.CompletedAt = &now
	t.currentSession.Status = models.Completed
	session := t.currentSession
	observers := append([]TimerObserver(nil), t.observers...)
	t.mu.Unlock()

	notify(observers, func(o TimerObserver) { o.OnSessionCompleted(session) })
	return nil
}

// GetCurrentSession returns the current (or most recent) session.
func (t *ConcreteTimerService) GetCurrentSession() *models.TimerSession {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.currentSession
}

// GetRemainingTime returns the remaining time in the current session.
func (t *ConcreteTimerService) GetRemainingTime() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.currentSession == nil {
		return 0
	}
	return t.currentSession.Remaining
}

// FocusSessionsCompleted returns how many focus sessions have fully completed,
// which drives the short/long break rhythm.
func (t *ConcreteTimerService) FocusSessionsCompleted() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.focusCompleted
}

// SuggestNext suggests the session type that should follow the current one:
// focus -> short/long break (per LongBreakInterval), break -> focus.
func (t *ConcreteTimerService) SuggestNext() models.SessionType {
	t.mu.Lock()
	defer t.mu.Unlock()

	var lastType models.SessionType
	completed := true
	if t.currentSession != nil {
		lastType = t.currentSession.Type
		completed = t.currentSession.Status == models.Completed
	}
	if !completed {
		// Session still in progress; nothing to suggest yet.
		if lastType == models.FocusSession {
			return models.FocusSession
		}
	}

	if lastType == models.FocusSession {
		if t.focusCompleted > 0 && t.focusCompleted%t.settings.LongBreakInterval == 0 {
			return models.LongBreak
		}
		return models.ShortBreak
	}
	if lastType == models.ShortBreak || lastType == models.LongBreak {
		return models.FocusSession
	}
	return models.FocusSession
}

// SetSettings swaps the settings used for durations and the break interval.
func (t *ConcreteTimerService) SetSettings(settings *models.Settings) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if settings != nil {
		t.settings = settings
	}
}

// RegisterObserver adds an observer to the service.
func (t *ConcreteTimerService) RegisterObserver(observer TimerObserver) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.observers = append(t.observers, observer)
}

// UnregisterObserver removes an observer from the service.
func (t *ConcreteTimerService) UnregisterObserver(observer TimerObserver) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, obs := range t.observers {
		if obs == observer {
			t.observers = append(t.observers[:i], t.observers[i+1:]...)
			break
		}
	}
}

// Shutdown stops the ticker goroutine. Safe to call multiple times.
func (t *ConcreteTimerService) Shutdown() {
	t.once.Do(func() {
		close(t.done)
	})
}

// ensureTicker starts the ticker goroutine on first use. Caller holds the lock.
func (t *ConcreteTimerService) ensureTicker() {
	if t.running {
		return
	}
	t.running = true
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.tickOnce()
			case <-t.done:
				return
			}
		}
	}()
}

// tickOnce advances the session by one second. It is separated from the
// goroutine so tests can drive time deterministically.
func (t *ConcreteTimerService) tickOnce() {
	t.mu.Lock()
	if t.currentSession == nil || t.currentSession.Status != models.Running {
		t.mu.Unlock()
		return
	}
	t.currentSession.Remaining -= time.Second
	timedOut := t.currentSession.Remaining <= 0
	if timedOut {
		t.currentSession.Remaining = 0
		now := time.Now()
		t.currentSession.CompletedAt = &now
		t.currentSession.Status = models.Completed
		if t.currentSession.Type == models.FocusSession {
			t.focusCompleted++
		}
	}
	remaining := t.currentSession.Remaining
	session := t.currentSession
	observers := append([]TimerObserver(nil), t.observers...)
	t.mu.Unlock()

	for _, o := range observers {
		o.OnTimerTick(remaining)
	}
	if timedOut {
		notify(observers, func(o TimerObserver) { o.OnSessionCompleted(session) })
	}
}

// notify invokes fn on every observer snapshot. Caller must not hold the lock.
func notify(observers []TimerObserver, fn func(o TimerObserver)) {
	for _, o := range observers {
		fn(o)
	}
}
