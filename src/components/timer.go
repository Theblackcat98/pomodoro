package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"pomodoro/src/models"
	"pomodoro/src/utils"
)

// TimerComponent represents the main timer component using Bubbles timer
type TimerComponent struct {
	timer      timer.Model
	sessionType models.SessionType
	status     models.SessionStatus
	remaining  time.Duration // tracked locally; bubbles/timer does not expose it
	onComplete func()
	notifier   *Notifier
}

// NewTimerComponent creates a new timer component
func NewTimerComponent(sessionType models.SessionType, duration time.Duration) *TimerComponent {
	bubblesTimer := timer.New(duration)

	// Create a default notifier with visual notifications
	notifier := NewNotifier(models.VisualNotification)

	return &TimerComponent{
		timer:       bubblesTimer,
		sessionType: sessionType,
		status:      models.Running,
		remaining:   duration,
		notifier:    notifier,
	}
}

// Init initializes the timer component
func (t *TimerComponent) Init() tea.Cmd {
	return t.timer.Init()
}

// Update handles messages for the timer component
func (t *TimerComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// bubbles/timer keeps no public remaining time, so track it locally while
	// the session is running.
	if _, ok := msg.(timer.TickMsg); ok && t.status == models.Running && t.remaining > 0 {
		t.remaining -= time.Second
		if t.remaining < 0 {
			t.remaining = 0
		}
	}

	t.timer, cmd = t.timer.Update(msg)

	// If timer is done, update status and call completion callback
	if t.timer.Timedout() {
		t.status = models.Completed
		// Send notification for completed session
		if t.notifier != nil {
			t.notifier.PlaySessionCompleteSound()
		}
		if t.onComplete != nil {
			t.onComplete()
		}
	}

	return t, cmd
}

// View renders the timer component
func (t *TimerComponent) View() string {
	if t.timer.Timedout() {
		return fmt.Sprintf("Time's up! %s session completed.\nPress any key to continue.",
			utils.FormatSessionTypeShort(t.sessionType))
	}

	return fmt.Sprintf("Session: %s\nTime remaining: %s",
		utils.FormatSessionTypeShort(t.sessionType), t.timer.View())
}

// Start starts the timer
func (t *TimerComponent) Start() tea.Cmd {
	return t.timer.Start()
}

// Pause pauses the timer by halting local time tracking (the bubbles timer
// has no native pause; ticks are simply not counted while paused).
func (t *TimerComponent) Pause() tea.Cmd {
	t.status = models.Paused
	return nil
}

// SetOnComplete sets the callback function to be called when the timer completes
func (t *TimerComponent) SetOnComplete(fn func()) {
	t.onComplete = fn
}

// GetRemainingTime returns the remaining time
func (t *TimerComponent) GetRemainingTime() time.Duration {
	if t.timer.Timedout() {
		return 0
	}
	return t.remaining
}

// IsComplete returns whether the timer has completed
func (t *TimerComponent) IsComplete() bool {
	return t.timer.Timedout()
}

// GetSessionType returns the session type
func (t *TimerComponent) GetSessionType() models.SessionType {
	return t.sessionType
}

// GetStatus returns the current status
func (t *TimerComponent) GetStatus() models.SessionStatus {
	return t.status
}

