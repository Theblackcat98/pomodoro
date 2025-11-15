package models

import (
	"time"
)

// SessionType represents the type of Pomodoro session
type SessionType string

const (
	FocusSession SessionType = "focus"
	ShortBreak   SessionType = "short_break"
	LongBreak    SessionType = "long_break"
)

// SessionStatus represents the current status of a session
type SessionStatus string

const (
	Running   SessionStatus = "running"
	Paused    SessionStatus = "paused"
	Completed SessionStatus = "completed"
)

// TimerSession represents a single Pomodoro session
type TimerSession struct {
	ID          string        `json:"id"`
	Type        SessionType   `json:"type"`      // Focus, ShortBreak, LongBreak
	Duration    time.Duration `json:"duration"`  // Original duration in seconds
	Remaining   time.Duration `json:"remaining"` // Time remaining in seconds
	Status      SessionStatus `json:"status"`    // Running, Paused, Completed
	StartTime   *time.Time    `json:"start_time,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}