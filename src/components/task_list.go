package components

import (
	"image/color"

	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"pomodoro/src/models"
	"pomodoro/src/utils"
)

// Delegate color tokens (kept local: the views package imports components, so
// it cannot be imported back).
var (
	delegateColorDone  = lipgloss.Color("#46A758")
	delegateColorFocus = lipgloss.Color("#E5484D")
)

// delegateAdaptive picks a light/dark variant (lipgloss v2 removed
// AdaptiveColor; detection duplicated locally since components cannot import
// the views package back).
var delegateAdaptive = func(light, dark string) color.Color {
	return lipgloss.LightDark(detectDarkBG())(lipgloss.Color(light), lipgloss.Color(dark))
}

// detectDarkBG is a best-effort terminal-background probe (assumes dark).
func detectDarkBG() (dark bool) {
	defer func() {
		if recover() != nil {
			dark = true
		}
	}()
	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}

// Item represents a task in the list
type Item struct {
	Task *models.Task
}

// FilterValue returns the value to filter on
func (i Item) FilterValue() string {
	return i.Task.Title
}

// TaskDelegate handles the rendering of list items.
type TaskDelegate struct {
	anim *CheckAnim // shared row-animation state; may be nil
}

// Height returns the height of the delegate: one line per task.
func (d TaskDelegate) Height() int {
	return 1
}

// Spacing returns the spacing between items.
func (d TaskDelegate) Spacing() int {
	return 0
}

// Update handles messages for the delegate.
func (d TaskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render draws one task row:
//
//	❯ ◐ Write the plan doc            🍅🍅🍅 1/3
func (d TaskDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}
	task := i.Task
	selected := index == m.Index()

	// Status glyph: ○ to-do, ◐ in progress, ● done.
	glyph := "○"
	switch task.Status {
	case models.InProgressStatus:
		glyph = "◐"
	case models.DoneStatus:
		glyph = "●"
	}
	// Completion animation: show the intermediate frame right after the check.
	if d.anim != nil && d.anim.ID == task.ID && d.anim.Frame > 0 && task.Status == models.DoneStatus {
		glyph = "◐"
	}

	titleStyle := lipgloss.NewStyle()
	glyphStyle := lipgloss.NewStyle()
	switch task.Status {
	case models.DoneStatus:
		titleStyle = titleStyle.
			Foreground(delegateAdaptive("#9A9A9A", "#5C5C5C")).
			Strikethrough(true)
		glyphStyle = glyphStyle.Foreground(delegateColorDone)
	case models.InProgressStatus:
		glyphStyle = glyphStyle.Foreground(delegateColorFocus)
		fallthrough
	default:
		titleStyle = titleStyle.Foreground(delegateAdaptive("#3F3F3F", "#EEEEEE"))
	}

	title := task.Title
	estimate := ""
	if task.EstimatedPomodoros > 0 {
		tomatoes := task.EstimatedPomodoros
		if tomatoes > 8 {
			tomatoes = 8 // keep rows sane for very large estimates
		}
		estimate = fmt.Sprintf("%s %d/%d",
			strings.Repeat("🍅", tomatoes),
			task.CompletedPomodoros, task.EstimatedPomodoros)
		estimateStyle := lipgloss.NewStyle().Foreground(delegateAdaptive("#8F8F8F", "#6E6E6E"))
		if task.Status == models.DoneStatus {
			estimateStyle = estimateStyle.Foreground(delegateAdaptive("#B4B4B4", "#454545"))
		}
		estimate = estimateStyle.Render(estimate)
	}

	cursor := "  "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(delegateAdaptive("#874BFD", "#7D56F4")).Bold(true).Render("❯ ")
		titleStyle = titleStyle.Bold(true)
	}

	// Pad the title so the estimates right-align across rows.
	availWidth := m.Width() - 6 - lipgloss.Width(estimate)
	if availWidth < 1 {
		availWidth = 1
	}
	padded := titleStyle.Render(fmt.Sprintf("%-*s", lipgloss.Width(utils.StripANSI(title)), utils.StripANSI(title)))
	if pad := availWidth - lipgloss.Width(utils.StripANSI(title)); pad > 0 {
		padded += strings.Repeat(" ", pad)
	}

	row := fmt.Sprintf("%s%s %s %s", cursor, glyphStyle.Render(glyph), padded, estimate)
	fmt.Fprintf(w, "%s", row)
}

// CheckAnim tracks the row completion animation (○→◐→●) for one task.
// An empty ID means "no animation".
type CheckAnim struct {
	ID    string
	Frame int
}

// TaskList represents the task list component using Bubbles list
type TaskList struct {
	list     list.Model
	tasks    []*models.Task
	anim     CheckAnim
	hideDone bool // archive view: completed tasks hidden
}

// NewTaskList creates a new task list component
func NewTaskList() *TaskList {
	// Create default list items from tasks
	var listItems []list.Item

	tl := &TaskList{
		tasks: make([]*models.Task, 0),
	}

	// The delegate animates through tl.anim (pointer keeps them in sync).
	// The list's built-in furniture is disabled: the app renders its own
	// chrome (header, hints, empty state). Filtering (`/`) stays enabled —
	// items expose FilterValue via their title.
	tl.list = list.New(listItems, TaskDelegate{anim: &tl.anim}, 0, 0)
	tl.list.SetShowTitle(false)
	tl.list.SetShowStatusBar(false)
	tl.list.SetShowHelp(false)
	tl.list.SetShowPagination(false)
	tl.list.SetFilteringEnabled(true)
	tl.list.SetStatusBarItemName("task", "tasks")

	return tl
}

// HideDone shows/hides completed tasks (the archive toggle).
func (tl *TaskList) HideDone(hide bool) {
	tl.hideDone = hide
	tl.syncList()
}

// DoneHidden reports whether completed tasks are currently hidden.
func (tl *TaskList) DoneHidden() bool {
	return tl.hideDone
}

// IsFiltering reports whether the list's built-in filter is active (typing
// a `/` query); while it is, keys belong to the list.
func (tl *TaskList) IsFiltering() bool {
	return tl.list.FilterState() != list.Unfiltered
}

// SetCheckAnim sets the row animation target. An empty id clears it.
func (tl *TaskList) SetCheckAnim(id string, frame int) {
	tl.anim = CheckAnim{ID: id, Frame: frame}
}

// Update handles messages for the task list (navigation keys are forwarded
// here by the root model).
func (tl *TaskList) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	tl.list, cmd = tl.list.Update(msg)
	return cmd
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
// removed). Done tasks are dropped while the archive toggle hides them.
func (tl *TaskList) syncList() {
	items := make([]list.Item, 0, len(tl.tasks))
	for _, task := range tl.tasks {
		if tl.hideDone && task.Status == models.DoneStatus {
			continue
		}
		items = append(items, Item{Task: task})
	}
	tl.list.SetItems(items)
}

// GetSelected returns the currently selected task
func (tl *TaskList) GetSelected() *models.Task {
	if item, ok := tl.list.SelectedItem().(Item); ok {
		return item.Task
	}
	return nil
}

// SetSize sets the size of the list
func (tl *TaskList) SetSize(width, height int) {
	tl.list.SetSize(width, height)
}
