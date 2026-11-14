package components

import (
	"os/exec"
	"runtime"

	"pomodoro/src/models"
)

// Notifier plays audible alerts when sessions complete. Visual notification
// (the flash banner) is rendered by the view layer — nothing here ever writes
// to stdout, which would corrupt the alt-screen TUI.
type Notifier struct {
	notificationType models.NotificationType
}

// NewNotifier creates a new notifier instance
func NewNotifier(notificationType models.NotificationType) *Notifier {
	return &Notifier{
		notificationType: notificationType,
	}
}

// PlaySessionCompleteSound plays the completion sound when sound
// notifications are enabled. Failures are ignored: a missing player or sound
// file must never disturb the app.
func (n *Notifier) PlaySessionCompleteSound() {
	if n.notificationType != models.SoundNotification {
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("afplay", "/System/Library/Sounds/Ping.aiff")
	case "linux":
		// The freedesktop sound set ships with most desktop distributions.
		cmd = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
	case "windows":
		cmd = exec.Command("powershell", "-c",
			"(New-Object Media.SoundPlayer 'C:\\Windows\\Media\\Notify.wav').PlaySync()")
	}
	if cmd != nil {
		_ = cmd.Start() // don't block the caller (ticker goroutine)
	}
}

// SetNotificationType sets the notification type
func (n *Notifier) SetNotificationType(notificationType models.NotificationType) {
	n.notificationType = notificationType
}

// GetNotificationType returns the current notification type
func (n *Notifier) GetNotificationType() models.NotificationType {
	return n.notificationType
}
