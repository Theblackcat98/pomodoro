package components

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	SoundNotification   NotificationType = "sound"
	VisualNotification  NotificationType = "visual"
	SilentNotification  NotificationType = "silent"
)

// Notifier handles different types of notifications
type Notifier struct {
	notificationType NotificationType
}

// NewNotifier creates a new notifier instance
func NewNotifier(notificationType NotificationType) *Notifier {
	return &Notifier{
		notificationType: notificationType,
	}
}

// NotifySessionComplete sends a notification when a session is complete
func (n *Notifier) NotifySessionComplete(sessionType models.SessionType) {
	switch n.notificationType {
	case SoundNotification:
		n.playSoundNotification()
	case VisualNotification:
		n.showVisualNotification(sessionType)
	case SilentNotification:
		// No notification
		return
	}
}

// playSoundNotification plays a sound notification
func (n *Notifier) playSoundNotification() {
	// Cross-platform sound notification
	switch runtime.GOOS {
	case "darwin": // macOS
		// Play a system sound on macOS
		exec.Command("afplay", "/System/Library/Sounds/Ping.aiff").Run()
	case "linux":
		// Use aplay or paplay for Linux
		exec.Command("paplay", "/usr/share/sounds/generic.wav").Run()
		exec.Command("aplay", "/usr/share/sounds/generic.wav").Run() // fallback
	case "windows":
		// Use PowerShell to play a beep on Windows
		exec.Command("powershell", "-c", "(New-Object Media.SoundPlayer 'C:\\Windows\\Media\\Notify.wav').Play(); Start-Sleep -s 1").Run()
	}
	
	// For cross-platform compatibility, we'll also provide visual feedback
	n.showVisualNotification(models.FocusSession) // Using a placeholder
}

// showVisualNotification shows a visual notification
func (n *Notifier) showVisualNotification(sessionType models.SessionType) {
	sessionTypeName := "Focus"
	if sessionType == models.ShortBreak {
		sessionTypeName = "Short Break"
	} else if sessionType == models.LongBreak {
		sessionTypeName = "Long Break"
	}
	
	// Create a visual notification using lipgloss
	notificationStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF0000")).
		Padding(1, 2).
		Background(lipgloss.Color("#FF0000")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	
	notificationText := fmt.Sprintf("Session Complete!\nYour %s session has ended.", sessionTypeName)
	
	fmt.Println(notificationStyle.Render(notificationText))
}

// SetNotificationType sets the notification type
func (n *Notifier) SetNotificationType(notificationType NotificationType) {
	n.notificationType = notificationType
}

// GetNotificationType returns the current notification type
func (n *Notifier) GetNotificationType() NotificationType {
	return n.notificationType
}