package main

import (
	"log"

	"github.com/charmbracelet/bubbletea"
	"pomodoro/src/views"
)

func main() {
	p := tea.NewProgram(views.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}