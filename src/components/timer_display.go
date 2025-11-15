package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
	"pomodoro/src/utils"
)

// TimerDisplay represents the visual timer display with circular progress
type TimerDisplay struct {
	totalDuration time.Duration
	remainingTime time.Duration
	sessionType   models.SessionType
	width         int
	height        int
}

// NewTimerDisplay creates a new timer display
func NewTimerDisplay(totalDuration time.Duration, sessionType models.SessionType) *TimerDisplay {
	return &TimerDisplay{
		totalDuration: totalDuration,
		remainingTime: totalDuration,
		sessionType:   sessionType,
		width:         20,
		height:        10,
	}
}

// SetSize sets the size of the display
func (td *TimerDisplay) SetSize(width, height int) {
	td.width = width
	td.height = height
}

// Update updates the remaining time
func (td *TimerDisplay) Update(remainingTime time.Duration) {
	td.remainingTime = remainingTime
}

// View renders the visual timer display with circular progress
func (td *TimerDisplay) View() string {
	// Calculate the percentage of time remaining
	var percent float64
	if td.totalDuration > 0 {
		percent = float64(td.remainingTime) / float64(td.totalDuration)
	} else {
		percent = 0
	}

	// Create the circular progress visualization
	progressVisual := td.createCircularProgress(percent)

	// Format the time remaining with large digits using lipgloss
	timeStr := td.formatLargeTime(td.remainingTime)

	// Format the session type
	sessionStr := utils.FormatSessionType(td.sessionType)

	// Combine everything with better visual hierarchy
	result := fmt.Sprintf("%s\n%s\n%s",
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#444444")).
			Padding(0, 1).
			Render(sessionStr),
		timeStr,
		progressVisual)

	return result
}

// formatLargeTime creates a large, clear display of the remaining time
func (td *TimerDisplay) formatLargeTime(duration time.Duration) string {
	duration = duration.Round(time.Second)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60

	// Using lipgloss to create larger, more readable digits
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")). // Green color for visibility
		Bold(true).
		Margin(1, 0)

	minutesStr := fmt.Sprintf("%02d", minutes)
	secondsStr := fmt.Sprintf("%02d", seconds)

	return timeStyle.Render(fmt.Sprintf("  %s : %s  ", minutesStr, secondsStr))
}

// createCircularProgress creates a circular visualization of the progress
func (td *TimerDisplay) createCircularProgress(percent float64) string {
	// Calculate how many characters of the "ring" should be filled
	totalSegments := 30
	filledSegments := int(float64(totalSegments) * percent)

	// Using lipgloss for styling the ring
	ringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4CAF50")).Bold(true)

	var sb strings.Builder

	// Create a circular representation using characters
	// This is a simplified representation - a proper implementation would use more sophisticated graphics

	// Create the full ring
	ring := make([]rune, totalSegments)
	for i := 0; i < totalSegments; i++ {
		ring[i] = '○' // Circle character
	}

	// Fill in the appropriate segments
	for i := 0; i < filledSegments; i++ {
		ring[i] = '●' // Filled circle character
	}

	// Create visual representation
	sb.WriteString("\n")

	// Top arc
	sb.WriteString("     ")
	for i := 0; i < 8; i++ {
		if i < filledSegments && filledSegments > 0 {
			sb.WriteString(filledStyle.Render(string(ring[i])))
		} else {
			sb.WriteString(ringStyle.Render(string(ring[i])))
		}
	}
	sb.WriteString("\n")

	// Sides
	for i := 0; i < 3; i++ {
		// Left side
		leftIdx := 7 + i
		if leftIdx < totalSegments {
			if leftIdx < filledSegments {
				sb.WriteString("   " + filledStyle.Render(string(ring[leftIdx])) + "     ")
			} else {
				sb.WriteString("   " + ringStyle.Render(string(ring[leftIdx])) + "     ")
			}
		}

		// Right side
		rightIdx := 17 - i
		if rightIdx < totalSegments {
			if rightIdx < filledSegments {
				sb.WriteString(filledStyle.Render(string(ring[rightIdx])))
			} else {
				sb.WriteString(ringStyle.Render(string(ring[rightIdx])))
			}
		}
		sb.WriteString("\n")
	}

	// Bottom arc
	sb.WriteString("     ")
	for i := 22; i < totalSegments; i++ {
		if i < filledSegments {
			sb.WriteString(filledStyle.Render(string(ring[i])))
		} else {
			sb.WriteString(ringStyle.Render(string(ring[i])))
		}
	}

	return sb.String()
}

// Use the same function from timer.go package


// GetSessionType returns the session type
func (td *TimerDisplay) GetSessionType() models.SessionType {
	return td.sessionType
}

// GetRemainingTime returns the remaining time
func (td *TimerDisplay) GetRemainingTime() time.Duration {
	return td.remainingTime
}

// GetPercent returns the percentage of time remaining
func (td *TimerDisplay) GetPercent() float64 {
	if td.totalDuration > 0 {
		return float64(td.remainingTime) / float64(td.totalDuration)
	}
	return 0
}