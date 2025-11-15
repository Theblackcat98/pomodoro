package utils

import (
	"fmt"
	"time"

	"pomodoro/src/models"
)

// FormatSessionType returns a formatted string for the session type
func FormatSessionType(sessionType models.SessionType) string {
	switch sessionType {
	case models.FocusSession:
		return "FOCUS TIME"
	case models.ShortBreak:
		return "SHORT BREAK"
	case models.LongBreak:
		return "LONG BREAK"
	default:
		return "UNKNOWN"
	}
}

// FormatSessionTypeShort returns a short formatted string for the session type
func FormatSessionTypeShort(sessionType models.SessionType) string {
	switch sessionType {
	case models.FocusSession:
		return "Focus"
	case models.ShortBreak:
		return "Short Break"
	case models.LongBreak:
		return "Long Break"
	default:
		return "Unknown"
	}
}

// FormatDuration formats a duration in MM:SS format
func FormatDuration(duration time.Duration) string {
	duration = duration.Round(time.Second)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}