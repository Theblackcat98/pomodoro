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

// FormatDurationHuman renders durations compactly for labels: "25m",
// "1h 05m", "45s".
func FormatDurationHuman(duration time.Duration) string {
	duration = duration.Round(time.Second)
	switch {
	case duration >= time.Hour:
		return fmt.Sprintf("%dh %02dm", int(duration.Hours()), int(duration.Minutes())%60)
	case duration >= time.Minute:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	default:
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
}
