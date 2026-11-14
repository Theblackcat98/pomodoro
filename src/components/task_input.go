package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
)

// TaskInput represents the task creation UI using Bubbles text input
type TaskInput struct {
	textInput textinput.Model
	title     string
	focused   bool
	err       error
}

// NewTaskInput creates a new task input component
func NewTaskInput() *TaskInput {
	ti := textinput.New()
	ti.Placeholder = "Enter task title..."
	ti.Focus()
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	return &TaskInput{
		textInput: ti,
		focused:   true,
	}
}

// Init initializes the task input
func (ti *TaskInput) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the task input
func (ti *TaskInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	ti.textInput, cmd = ti.textInput.Update(msg)
	return ti, cmd
}

// View renders the task input
func (ti *TaskInput) View() string {
	return ti.textInput.View()
}

// SetValue sets the value of the text input
func (ti *TaskInput) SetValue(value string) {
	ti.textInput.SetValue(value)
	ti.title = value
}

// GetValue returns the current value of the text input
func (ti *TaskInput) GetValue() string {
	return ti.textInput.Value()
}

// GetTask creates and returns a Task model based on the input
func (ti *TaskInput) GetTask() *models.Task {
	title := strings.TrimSpace(ti.textInput.Value())
	if title == "" {
		return nil
	}
	return models.NewTask(title, "", 1)
}

// Focus focuses the text input
func (ti *TaskInput) Focus() tea.Cmd {
	ti.focused = true
	return ti.textInput.Focus()
}

// Blur removes focus from the text input
func (ti *TaskInput) Blur() {
	ti.focused = false
	ti.textInput.Blur()
}

// IsFocused returns whether the text input is focused
func (ti *TaskInput) IsFocused() bool {
	return ti.focused
}

// SetSize sets the size of the text input
func (ti *TaskInput) SetSize(width, height int) {
	// Text input doesn't need explicit size settings
	// We can set a max width if needed
	if width > 0 {
		ti.textInput.Width = width
	}
}

// GetPreferredSize returns the preferred size of the text input
func (ti *TaskInput) GetPreferredSize() (int, int) {
	return 50, 1 // approximate width and height
}

// HandleKeyMsg handles key messages specifically
func (ti *TaskInput) HandleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	
	switch msg.Type {
	case tea.KeyEnter:
		// When user presses enter, we might want to submit the task
		// This would typically happen in the parent component
		ti.title = ti.textInput.Value()
	case tea.KeyTab:
		// Tabbing behavior - blur the input
		ti.Blur()
	}
	
	ti.textInput, cmd = ti.textInput.Update(msg)
	return cmd
}