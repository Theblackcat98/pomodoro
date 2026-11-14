package components

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pomodoro/src/models"
)

// Item represents a task in the list
type Item struct {
	Task *models.Task
}

// FilterValue returns the value to filter on
func (i Item) FilterValue() string {
	return i.Task.Title
}

// TaskDelegate handles the rendering of list items
type TaskDelegate struct{}

// Height returns the height of the delegate
func (d TaskDelegate) Height() int {
	return 2 // Height for each task item
}

// Spacing returns the spacing between items
func (d TaskDelegate) Spacing() int {
	return 1
}

// Update handles messages for the delegate
func (d TaskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders the delegate
func (d TaskDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	task := i.Task

	// Check if this item is selected
	taskStyle := lipgloss.NewStyle().Padding(0, 1)
	if index == m.Index() {
		taskStyle = taskStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("62")).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57"))
	}

	statusSymbol := "○" // Uncompleted
	if task.Status == models.DoneStatus {
		statusSymbol = "●" // Completed
	}

	title := task.Title
	if task.Status == models.DoneStatus {
		title = fmt.Sprintf("[DONE] %s", task.Title)
	}

	fmt.Fprintf(w, taskStyle.Render(fmt.Sprintf("%s %s\n%s", 
		statusSymbol, 
		title,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(task.Description))))
}

// TaskList represents the task list component using Bubbles list
type TaskList struct {
	list     list.Model
	tasks    []*models.Task
	selected *models.Task
}

// NewTaskList creates a new task list component
func NewTaskList() *TaskList {
	// Create default list items from tasks
	var listItems []list.Item
	
	delegate := TaskDelegate{}
	
	// Create the list model with default settings
	listModel := list.New(listItems, delegate, 0, 0)
	listModel.Title = "Tasks"
	listModel.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	
	return &TaskList{
		list:  listModel,
		tasks: make([]*models.Task, 0),
	}
}

// Init initializes the task list
func (tl *TaskList) Init() tea.Cmd {
	return nil
}

// Update handles messages for the task list
func (tl *TaskList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	tl.list, cmd = tl.list.Update(msg)
	
	// Check if an item was selected
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyEnter:
			// Get the selected task
			if item, ok := tl.list.SelectedItem().(Item); ok {
				tl.selected = item.Task
			}
		}
	}
	
	return tl, cmd
}

// View renders the task list
func (tl *TaskList) View() string {
	return tl.list.View()
}

// AddTask adds a task to the list
func (tl *TaskList) AddTask(task *models.Task) {
	tl.tasks = append(tl.tasks, task)
	tl.syncList()
}

// RemoveTask removes a task from the list by ID
func (tl *TaskList) RemoveTask(taskID string) {
	kept := tl.tasks[:0]
	for _, task := range tl.tasks {
		if task.ID != taskID {
			kept = append(kept, task)
		}
	}
	tl.tasks = kept
	tl.syncList()
}

// UpdateTask updates a task in the list
func (tl *TaskList) UpdateTask(updatedTask *models.Task) {
	for i, task := range tl.tasks {
		if task.ID == updatedTask.ID {
			tl.tasks[i] = updatedTask
			break
		}
	}
	tl.syncList()
}

// syncList rebuilds the bubbles list items from tl.tasks so the two views of
// the data can never drift apart (the old per-index copy did when tasks were
// removed).
func (tl *TaskList) syncList() {
	items := make([]list.Item, 0, len(tl.tasks))
	for _, task := range tl.tasks {
		items = append(items, Item{Task: task})
	}
	tl.list.SetItems(items)
}

// GetTasks returns all tasks in the list
func (tl *TaskList) GetTasks() []*models.Task {
	return tl.tasks
}

// GetSelected returns the currently selected task
func (tl *TaskList) GetSelected() *models.Task {
	if item, ok := tl.list.SelectedItem().(Item); ok {
		return item.Task
	}
	return nil
}

// ClearSelection clears the selection
func (tl *TaskList) ClearSelection() {
	tl.selected = nil
}

// SetSize sets the size of the list
func (tl *TaskList) SetSize(width, height int) {
	tl.list.SetSize(width, height)
}

// Focus focuses the list
func (tl *TaskList) Focus() tea.Cmd {
	tl.list.Select(0)
	return nil
}

// Blur blurs the list
func (tl *TaskList) Blur() {
	// The list model doesn't have a blur method, so we do nothing
}

// IsFocused returns whether the list is focused
func (tl *TaskList) IsFocused() bool {
	// The list model doesn't have a IsFocused method, so we'll track it ourselves
	// For now, we'll just return a default value
	return true // This should be managed by the parent component or application state
}