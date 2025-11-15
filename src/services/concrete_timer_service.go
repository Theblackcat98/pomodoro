package services

import (
	"errors"
	"time"
	
	"pomodoro/src/models"
)

// ConcreteTimerService is the concrete implementation of TimerService
type ConcreteTimerService struct {
	currentSession *models.TimerSession
	observers      []TimerObserver
	ticker         *time.Ticker
	done           chan bool
}

// NewTimerService creates a new instance of TimerService
func NewTimerService() TimerService {
	return &ConcreteTimerService{
		observers: make([]TimerObserver, 0),
		done:      make(chan bool),
	}
}

// StartSession starts a new Pomodoro session
func (t *ConcreteTimerService) StartSession(sessionType models.SessionType) error {
	if t.currentSession != nil && t.currentSession.Status == models.Running {
		return errors.New("a session is already running")
	}

	now := time.Now()
	session := &models.TimerSession{
		Type:      sessionType,
		Status:    models.Running,
		CreatedAt: now,
		StartTime: &now,
	}

	// Set default durations based on session type
	switch sessionType {
	case models.FocusSession:
		session.Duration = 25 * time.Minute
		session.Remaining = 25 * time.Minute
	case models.ShortBreak:
		session.Duration = 5 * time.Minute
		session.Remaining = 5 * time.Minute
	case models.LongBreak:
		session.Duration = 15 * time.Minute
		session.Remaining = 15 * time.Minute
	}

	t.currentSession = session

	// Notify observers
	for _, observer := range t.observers {
		observer.OnSessionStarted(session)
	}

	// Start the countdown ticker
	t.ticker = time.NewTicker(1 * time.Second)
	go t.runTicker()

	return nil
}

// runTicker runs the ticker in a separate goroutine
func (t *ConcreteTimerService) runTicker() {
	for {
		select {
		case <-t.ticker.C:
			if t.currentSession != nil && t.currentSession.Status == models.Running {
				t.currentSession.Remaining -= time.Second

				// Notify observers of timer tick
				for _, observer := range t.observers {
					observer.OnTimerTick(t.currentSession.Remaining)
				}

				// Check if session is complete
				if t.currentSession.Remaining <= 0 {
					t.completeSession()
				}
			}
		case <-t.done:
			return
		}
	}
}

// completeSession marks the current session as completed
func (t *ConcreteTimerService) completeSession() {
	if t.currentSession != nil {
		now := time.Now()
		t.currentSession.CompletedAt = &now
		t.currentSession.Status = models.Completed

		// Stop the ticker
		if t.ticker != nil {
			t.ticker.Stop()
		}

		// Notify observers
		for _, observer := range t.observers {
			observer.OnSessionCompleted(t.currentSession)
		}
	}
}

// PauseSession pauses the current session
func (t *ConcreteTimerService) PauseSession() error {
	if t.currentSession == nil {
		return errors.New("no active session to pause")
	}

	if t.currentSession.Status != models.Running {
		return errors.New("session is not running")
	}

	t.currentSession.Status = models.Paused

	// Stop the ticker
	if t.ticker != nil {
		t.ticker.Stop()
	}

	// Notify observers
	for _, observer := range t.observers {
		observer.OnSessionPaused(t.currentSession)
	}

	return nil
}

// ResumeSession resumes a paused session
func (t *ConcreteTimerService) ResumeSession() error {
	if t.currentSession == nil {
		return errors.New("no session to resume")
	}

	if t.currentSession.Status != models.Paused {
		return errors.New("session is not paused")
	}

	t.currentSession.Status = models.Running

	// Restart the ticker
	t.ticker = time.NewTicker(1 * time.Second)
	go t.runTicker()

	// Notify observers
	for _, observer := range t.observers {
		observer.OnSessionResumed(t.currentSession)
	}

	return nil
}

// StopSession stops the current session
func (t *ConcreteTimerService) StopSession() error {
	if t.currentSession == nil {
		return errors.New("no active session to stop")
	}

	// Stop the ticker
	if t.ticker != nil {
		t.ticker.Stop()
	}

	// Mark as completed
	now := time.Now()
	t.currentSession.CompletedAt = &now
	t.currentSession.Status = models.Completed

	// Notify observers
	for _, observer := range t.observers {
		observer.OnSessionCompleted(t.currentSession)
	}

	return nil
}

// GetCurrentSession returns the current session
func (t *ConcreteTimerService) GetCurrentSession() *models.TimerSession {
	return t.currentSession
}

// GetRemainingTime returns the remaining time in the current session
func (t *ConcreteTimerService) GetRemainingTime() time.Duration {
	if t.currentSession == nil {
		return 0
	}
	return t.currentSession.Remaining
}

// RegisterObserver adds an observer to the service
func (t *ConcreteTimerService) RegisterObserver(observer TimerObserver) {
	t.observers = append(t.observers, observer)
}

// UnregisterObserver removes an observer from the service
func (t *ConcreteTimerService) UnregisterObserver(observer TimerObserver) {
	for i, obs := range t.observers {
		if obs == observer {
			t.observers = append(t.observers[:i], t.observers[i+1:]...)
			break
		}
	}
}