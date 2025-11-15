package models

import "time"

type NotificationType string

const (
	SoundNotification   NotificationType = "sound"
	VibrationNotification NotificationType = "vibration"
	VisualNotification  NotificationType = "visual"
	SilentNotification  NotificationType = "silent"
)

type Settings struct {
	ID                   string           `json:"id"`
	FocusDuration        time.Duration    `json:"focus_duration"`        // Default: 25 minutes
	ShortBreakDuration   time.Duration    `json:"short_break_duration"`  // Default: 5 minutes
	LongBreakDuration    time.Duration    `json:"long_break_duration"`   // Default: 15 minutes
	LongBreakInterval    int              `json:"long_break_interval"`   // After how many focus sessions, default: 4
	AutoStartBreaks      bool             `json:"auto_start_breaks"`     // Default: true
	AutoStartFocus       bool             `json:"auto_start_focus"`      // After breaks, default: false
	NotificationType     NotificationType `json:"notification_type"`     // Default: "sound"
	Theme                string           `json:"theme"`                 // Default: "default"
	CompletedSessions    int              `json:"completed_sessions"`    // Number of completed sessions
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}