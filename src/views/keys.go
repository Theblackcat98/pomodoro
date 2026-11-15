package views

import "github.com/charmbracelet/bubbles/key"

// keyMap is the single source of truth for the app's keybindings. The help
// view renders it; the status-bar hints mirror it per context.
type keyMap struct {
	Tab        key.Binding
	BackTab    key.Binding
	Help       key.Binding
	Quit       key.Binding
	ForceQuit  key.Binding
	StartPause key.Binding
	Reset      key.Binding
	Resume     key.Binding
	Skip       key.Binding
	QuickStart key.Binding
	Add        key.Binding
	Edit       key.Binding
	Toggle     key.Binding
	Delete     key.Binding
	Archive    key.Binding
	Filter     key.Binding
	FocusTask  key.Binding
	Nav        key.Binding
	Adjust     key.Binding
	ToggleSet  key.Binding
	Scroll     key.Binding
	Export     key.Binding
	Cancel     key.Binding
}

var keys = keyMap{
	Tab:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch views")),
	BackTab:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "previous view")),
	Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:       key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	ForceQuit:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "force quit")),
	StartPause: key.NewBinding(key.WithKeys(" ", "enter"), key.WithHelp("space", "start/pause")),
	Reset:      key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reset")),
	Resume:     key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "resume recovered session")),
	Skip:       key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "skip")),
	QuickStart: key.NewBinding(key.WithKeys("1", "2", "3"), key.WithHelp("1/2/3", "15/25/35 min")),
	Add:        key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	Edit:       key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Toggle:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "toggle done")),
	Delete:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Archive:    key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "hide/show done")),
	Filter:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter tasks")),
	FocusTask:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "focus task")),
	Nav:        key.NewBinding(key.WithKeys("j", "k", "up", "down"), key.WithHelp("j/k", "move")),
	Adjust:     key.NewBinding(key.WithKeys("+", "-", "="), key.WithHelp("+/-", "adjust")),
	ToggleSet:  key.NewBinding(key.WithKeys(" ", "x", "enter"), key.WithHelp("space", "toggle")),
	Scroll:     key.NewBinding(key.WithKeys("j", "k", "up", "down"), key.WithHelp("j/k", "scroll")),
	Export:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "export JSON (Stats)")),
	Cancel:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

// bindings renders a "key action · key action" row from the given bindings.
func bindings(bs ...key.Binding) string {
	out := ""
	for i, b := range bs {
		if i > 0 {
			out += statusHintStyle.Render(" · ")
		}
		out += hint(b.Help().Key, b.Help().Desc)
	}
	return out
}
