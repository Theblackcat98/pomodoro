package models

import "time"

type SessionType string
type SessionStatus string

const (
	FocusSession SessionType = "focus"
	ShortBreak   SessionType = "short_break"
	LongBreak    SessionType = "long_break"
)

const (
	Running   SessionStatus = "running"
	Paused    SessionStatus = "paused"
	Completed SessionStatus = "completed"
)

type TimerSession struct {
	ID          string        `json:"id"`
	Type        SessionType   `json:"type"`      // Focus, ShortBreak, LongBreak
	Duration    time.Duration `json:"duration"`  // Original duration in seconds
	Remaining   time.Duration `json:"remaining"` // Time remaining in seconds
	Status      SessionStatus `json:"status"`    // Running, Paused, Completed
	TaskID      string        `json:"task_id,omitempty"` // Task the session worked on, if any
	StartTime   *time.Time    `json:"start_time,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}