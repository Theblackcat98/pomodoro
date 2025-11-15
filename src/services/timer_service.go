package services

import (
	"pomodoro/src/models"
	"time"
)

// TimerService interface manages the core timing functionality for Pomodoro sessions
type TimerService interface {
	StartSession(sessionType models.SessionType) error
	PauseSession() error
	ResumeSession() error
	StopSession() error
	GetCurrentSession() *models.TimerSession
	GetRemainingTime() time.Duration
	RegisterObserver(observer TimerObserver)
	UnregisterObserver(observer TimerObserver)
}

// TimerObserver interface defines the observer pattern for timer events
type TimerObserver interface {
	OnTimerTick(remaining time.Duration)
	OnSessionStarted(session *models.TimerSession)
	OnSessionPaused(session *models.TimerSession)
	OnSessionResumed(session *models.TimerSession)
	OnSessionCompleted(session *models.TimerSession)
	OnSessionCancelled(session *models.TimerSession)
}