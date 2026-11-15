package models

import "time"

type NotificationType string

const (
	SoundNotification  NotificationType = "sound"
	VisualNotification NotificationType = "visual"
	SilentNotification NotificationType = "silent"
)

type Settings struct {
	ID                 string           `json:"id"`
	FocusDuration      time.Duration    `json:"focus_duration"`       // Default: 25 minutes
	ShortBreakDuration time.Duration    `json:"short_break_duration"` // Default: 5 minutes
	LongBreakDuration  time.Duration    `json:"long_break_duration"`  // Default: 15 minutes
	LongBreakInterval  int              `json:"long_break_interval"`  // After how many focus sessions, default: 4
	AutoStartBreaks    bool             `json:"auto_start_breaks"`    // Default: true
	AutoStartFocus     bool             `json:"auto_start_focus"`     // After breaks, default: false
	NotificationType   NotificationType `json:"notification_type"`    // Default: "sound"
	Theme              string           `json:"theme"`                // Default: "default"
	Mouse              bool             `json:"mouse"`                // Mouse: wheel scroll + tab clicks
	SeenIntro          bool             `json:"seen_intro"`           // First-run hint already shown
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// DefaultSettings returns the documented default settings.
func DefaultSettings() *Settings {
	now := time.Now()
	return &Settings{
		ID:                 "default",
		FocusDuration:      25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
		LongBreakInterval:  4,
		AutoStartBreaks:    true,
		AutoStartFocus:     false,
		NotificationType:   SoundNotification,
		Theme:              "default",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// DurationFor returns the configured duration for a session type.
func (s *Settings) DurationFor(sessionType SessionType) time.Duration {
	switch sessionType {
	case FocusSession:
		return s.FocusDuration
	case ShortBreak:
		return s.ShortBreakDuration
	case LongBreak:
		return s.LongBreakDuration
	default:
		return s.FocusDuration
	}
}
