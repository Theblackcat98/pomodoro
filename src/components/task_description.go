package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TaskDescription represents the task description editor using Bubbles text area
type TaskDescription struct {
	textarea textarea.Model
	focused  bool
}

// NewTaskDescription creates a new task description component
func NewTaskDescription() *TaskDescription {
	ta := textarea.New()
	ta.Placeholder = "Enter task description..."
	ta.Focus()
	
	// Style the text area
	ta.Prompt = "┃ "
	ta.Blur()
	
	// Styling when focused
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	// Styling when not focused
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("247"))
	
	return &TaskDescription{
		textarea: ta,
		focused:  false,
	}
}

// Init initializes the task description
func (td *TaskDescription) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the task description
func (td *TaskDescription) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	td.textarea, cmd = td.textarea.Update(msg)
	return td, cmd
}

// View renders the task description
func (td *TaskDescription) View() string {
	return td.textarea.View()
}

// SetValue sets the value of the text area
func (td *TaskDescription) SetValue(value string) {
	td.textarea.SetValue(value)
}

// GetValue returns the current value of the text area
func (td *TaskDescription) GetValue() string {
	return td.textarea.Value()
}

// Focus focuses the text area
func (td *TaskDescription) Focus() tea.Cmd {
	td.focused = true
	return td.textarea.Focus()
}

// Blur removes focus from the text area
func (td *TaskDescription) Blur() {
	td.focused = false
	td.textarea.Blur()
}

// IsFocused returns whether the text area is focused
func (td *TaskDescription) IsFocused() bool {
	return td.focused
}

// SetSize sets the size of the text area
func (td *TaskDescription) SetSize(width, height int) {
	if width > 0 {
		td.textarea.SetWidth(width)
	}
	if height > 0 {
		td.textarea.SetHeight(height)
	}
}

// GetPreferredSize returns the preferred size of the text area
func (td *TaskDescription) GetPreferredSize() (int, int) {
	return 60, 5 // width and height
}

// HandleKeyMsg handles key messages specifically
func (td *TaskDescription) HandleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	
	switch msg.Type {
	case tea.KeyEnter:
		// In text area, enter creates a new line
		// We might use a different key combination for submission
	case tea.KeyTab:
		// Tabbing behavior - blur the input
		td.Blur()
	}
	
	td.textarea, cmd = td.textarea.Update(msg)
	return cmd
}