package main

import (
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	clog "github.com/charmbracelet/log"
	"pomodoro/src/services"
	"pomodoro/src/utils"
	"pomodoro/src/views"
)

func main() {
	setupLogging()

	// A corrupt state file is moved aside by the persistence layer; only real
	// I/O failures abort the launch.
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

// setupLogging routes diagnostics to $XDG_STATE_HOME/pomodoro/pomodoro.log so
// they never disturb the alt-screen render. Best-effort: no log file, no
// problem.
func setupLogging() {
	dir, err := utils.StateDir()
	if err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "pomodoro.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	l := clog.New(f)
	l.SetLevel(clog.InfoLevel)
	l.SetReportTimestamp(false)
	views.SetLogger(l)
}
