package views

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"pomodoro/src/components"
	"pomodoro/src/models"
	"pomodoro/src/services"
)

// MainView represents the main application view combining all components
type MainView struct {
	timerComponent *components.TimerComponent
	timerDisplay   *components.TimerDisplay
	timerControl   *components.TimerControl
	timerService   services.TimerService
	currentSession *models.TimerSession
	sessionType    models.SessionType
}

// NewMainView creates a new main view
func NewMainView() *MainView {
	// Initialize timer service
	timerService := services.NewTimerService()

	// Default to focus session
	sessionType := models.FocusSession
	defaultDuration := 25 * time.Minute

	// Create components
	timerComp := components.NewTimerComponent(sessionType, defaultDuration)
	timerDisplay := components.NewTimerDisplay(defaultDuration, sessionType)
	timerControl := components.NewTimerControl()

	mainView := &MainView{
		timerComponent: timerComp,
		timerDisplay:   timerDisplay,
		timerControl:   timerControl,
		timerService:   timerService,
		sessionType:    sessionType,
	}

	// Set up completion callback
	timerComp.SetOnComplete(func() {
		// Handle timer completion
		timerControl.SetActive(false)
	})

	return mainView
}

// Init initializes the main view
func (mv *MainView) Init() tea.Cmd {
	return tea.Batch(
		mv.timerComponent.Init(),
	)
}

// Update handles messages for the main view
func (mv *MainView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeySpace:
			// Toggle start/pause based on control state
			if mv.timerControl.IsActive() {
				// Currently running, so pause
				mv.timerControl.SetActive(false)
				err := mv.timerService.(interface{ PauseSession() error }).(services.TimerService).PauseSession()
				if err != nil {
					// Handle error as needed
					return mv, tea.Quit
				}
			} else {
				// Currently paused or not running, so start
				if mv.currentSession == nil || mv.currentSession.Status == models.Completed {
					// Start a new session
					err := mv.timerService.StartSession(mv.sessionType)
					if err != nil {
						// Handle error as needed
						return mv, tea.Quit
					}
					mv.currentSession = mv.timerService.GetCurrentSession()
				} else if mv.currentSession.Status == models.Paused {
					// Resume existing session
					err := mv.timerService.ResumeSession()
					if err != nil {
						// Handle error as needed
						return mv, tea.Quit
					}
				}
				mv.timerControl.SetActive(true)
			}
		case tea.KeyRunes:
			if string(msg.Runes) == "q" {
				return mv, tea.Quit
			}
		}
	case models.TimerTickMsg:
		// Update the display with new remaining time
		mv.timerDisplay.Update(msg.Remaining)
	}

	// Update individual components
	_, cmd := mv.timerComponent.Update(msg)
	// Since we already have a reference to the component, no type assertion needed
	// We just need to use it as is

	return mv, cmd
}

// View renders the main view
func (mv *MainView) View() string {
	if mv.currentSession == nil {
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\nPress SPACE or ENTER to start the timer\nPress Q to quit",
			mv.timerDisplay.View(),
			"Session: Focus Time",
			mv.timerControl.View(),
		)
	}

	// Update display based on current session
	sessionTypeStr := "Focus Time"
	if mv.currentSession.Type == models.ShortBreak {
		sessionTypeStr = "Short Break"
	} else if mv.currentSession.Type == models.LongBreak {
		sessionTypeStr = "Long Break"
	}

	remainingTime := mv.timerService.GetRemainingTime()
	mv.timerDisplay.Update(remainingTime)
	mv.timerDisplay = components.NewTimerDisplay(mv.currentSession.Duration, mv.currentSession.Type)
	mv.timerDisplay.Update(remainingTime)

	return fmt.Sprintf(
		"%s\n\nSession: %s\n\n%s\n\nPress SPACE or ENTER to %s\nPress Q to quit",
		mv.timerDisplay.View(),
		sessionTypeStr,
		mv.timerControl.View(),
		func() string {
			if mv.timerControl.IsActive() {
				return "PAUSE"
			}
			return "START"
		}(),
	)
}