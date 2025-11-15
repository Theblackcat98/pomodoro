package views

import tea "github.com/charmbracelet/bubbletea"

// App represents the main application model
type App struct {
	mainView *MainView
}

// NewApp creates a new instance of the application
func NewApp() *App {
	return &App{
		mainView: NewMainView(),
	}
}

// Init is the initial command for the application
func (a *App) Init() tea.Cmd {
	return a.mainView.Init()
}

// Update handles messages and updates the application state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return a.mainView.Update(msg)
}

// View returns the current view as a string
func (a *App) View() string {
	return a.mainView.View()
}