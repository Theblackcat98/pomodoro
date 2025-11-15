package services

import (
	"time"
	"pomodoro/src/models"
)

type TimerService struct {
	Session *models.TimerSession
	ticker  *time.Ticker
	done    chan bool
}

func NewTimerService(sessionType models.SessionType, duration time.Duration) *TimerService {
	session := models.TimerSession{
		Type:      sessionType,
		Duration:  duration,
		Remaining: duration,
		Status:    models.Running,
		CreatedAt: time.Now(),
		StartTime: &time.Time{},
	}
	
	return &TimerService{
		Session: &session,
		ticker:  time.NewTicker(1 * time.Second),
		done:    make(chan bool),
	}
}

func (ts *TimerService) Start() {
	*ts.Session.StartTime = time.Now()
	ts.Session.Status = models.Running
}

func (ts *TimerService) Pause() {
	ts.Session.Status = models.Paused
}

func (ts *TimerService) Stop() {
	now := time.Now()
	ts.Session.CompletedAt = &now
	ts.Session.Status = models.Completed
	ts.ticker.Stop()
}

func (ts *TimerService) Update() {
	if ts.Session.Status == models.Running && ts.Session.Remaining > 0 {
		ts.Session.Remaining -= time.Second
		if ts.Session.Remaining <= 0 {
			ts.Session.Remaining = 0
			ts.Stop()
		}
	}
}

func (ts *TimerService) GetSession() *models.TimerSession {
	return ts.Session
}