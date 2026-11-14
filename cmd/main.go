package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"pomodoro/src/services"
	"pomodoro/src/views"
)

func main() {
	store, err := services.NewDefaultPersistence()
	if err != nil {
		log.Fatalf("persistence: %v", err)
	}

	// NewModel resolves persisted settings from the store and applies them to
	// the service; a nil setting here just means "defaults until loaded".
	timerSvc := services.NewTimerService(nil)

	p := tea.NewProgram(views.NewModel(timerSvc, store), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
