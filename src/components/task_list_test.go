package components

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"pomodoro/src/models"
)

// pump drains the cmds a list update returns, feeding every resulting msg
// back through Update (mirrors what the bubbletea runtime does). A batch is
// itself a slice of cmds — each gets called, its msg routed back in.
// Filtering is async: list.Update returns a cmd that computes the matches.
//
// Cursor-blink cmds block on real timers; they are abandoned after a short
// wait (their messages are irrelevant to filtering, and the process exits
// with the test anyway).
func pump(tl *TaskList, cmd tea.Cmd) {
	runCmd := func(c tea.Cmd) (tea.Msg, bool) {
		ch := make(chan tea.Msg, 1)
		go func() { ch <- c() }()
		select {
		case msg := <-ch:
			return msg, true
		case <-time.After(80 * time.Millisecond):
			return nil, false // blink or other timer cmd: drop it
		}
	}
	for i := 0; i < 25 && cmd != nil; i++ {
		msg, ok := runCmd(cmd)
		if !ok {
			return
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				if c != nil {
					pump(tl, c)
				}
			}
			return
		}
		cmd = tl.Update(msg)
	}
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func newTestList(t *testing.T) *TaskList {
	t.Helper()
	tl := NewTaskList()
	tl.SetSize(80, 20)
	tl.AddTask(&models.Task{ID: "1", Title: "Write launch report", Status: models.ToDoStatus})
	tl.AddTask(&models.Task{ID: "2", Title: "Read papers", Status: models.DoneStatus})
	tl.AddTask(&models.Task{ID: "3", Title: "Ship newsletter", Status: models.ToDoStatus})
	return tl
}

func TestTaskListFilterLifecycle(t *testing.T) {
	tl := newTestList(t)

	if tl.IsFiltering() {
		t.Fatal("list must start unfiltered")
	}

	// "/" opens the filter; typing narrows; enter accepts.
	pump(tl, tl.Update(keyMsg("/")))
	if !tl.IsFiltering() {
		t.Fatal("/ must start filtering")
	}
	for _, c := range "ship" {
		pump(tl, tl.Update(keyMsg(string(c))))
	}
	pump(tl, tl.Update(keyMsg("enter"))) // accept

	if !tl.IsFiltering() {
		t.Fatal("filter must stay applied after accept (FilterApplied)")
	}
	if got := tl.GetSelected(); got == nil || got.Title != "Ship newsletter" {
		t.Fatalf("selected after filter = %+v, want Ship newsletter", got)
	}

	// esc cancels the filter and shows everything again.
	pump(tl, tl.Update(keyMsg("esc")))
	if tl.IsFiltering() {
		t.Fatal("esc must clear the filter")
	}
}

func TestTaskListArchiveToggle(t *testing.T) {
	tl := newTestList(t)

	tl.HideDone(true)
	if !tl.DoneHidden() {
		t.Fatal("DoneHidden must report true after HideDone(true)")
	}
	if got := tl.GetSelected(); got != nil && got.Title == "Read papers" {
		t.Fatalf("done task must be hidden while archived, got %+v", got)
	}

	tl.HideDone(false)
	if tl.DoneHidden() {
		t.Fatal("DoneHidden must report false after HideDone(false)")
	}
	// Navigate to find the done task to prove it is back.
	found := false
	for i := 0; i < 3; i++ {
		if got := tl.GetSelected(); got != nil && got.Title == "Read papers" {
			found = true
			break
		}
		tl.Update(keyMsg("j"))
	}
	if !found {
		t.Fatal("done task must be visible again after unhide")
	}
}

func TestTaskListHideDoneKeepsData(t *testing.T) {
	tl := newTestList(t)
	tl.HideDone(true)
	tl.HideDone(false)
	// Every original task must still be reachable (data retained).
	seen := map[string]bool{}
	for i := 0; i < 3; i++ {
		if got := tl.GetSelected(); got != nil {
			seen[got.Title] = true
		}
		tl.Update(keyMsg("j"))
	}
	for _, want := range []string{"Write launch report", "Read papers", "Ship newsletter"} {
		if !seen[want] {
			t.Fatalf("task %q not reachable after unhide; seen=%v", want, seen)
		}
	}
}
