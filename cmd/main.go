package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"pomodoro/src/views"
)

func main() {
	p := tea.NewProgram(views.NewApp(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}