package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"pomodoro/src/components"
	"pomodoro/src/models"
	"pomodoro/src/services"
)

// Tab identifies a main view.
type Tab int

const (
	TabTimer Tab = iota
	TabTasks
	TabStats
	TabSettings
	TabHelp
)

// tabCount must match the Tab constants above.
const tabCount = 5

var tabNames = []string{"Timer", "Tasks", "Stats", "Settings", "Help"}

// addField identifies the focused field in the new-task form.
type addField int

const (
	addFieldTitle addField = iota
	addFieldDescription
)

// eventBuffer sizes the channel between the timer service goroutine and the
// bubbletea loop.
const eventBuffer = 32

// flashDuration is how long the completion banner stays on screen before the
// next session (possibly) auto-starts.
const flashDuration = 2500 * time.Millisecond

// toastDuration is how long transient status messages linger.
const toastDuration = 2200 * time.Millisecond

// maxMemorySessions caps the in-memory history used by the stats view.
const maxMemorySessions = 200

// Model is the root bubbletea model; it routes keys to the active view and
// bridges timer-service events into the update loop.
type Model struct {
	settings *models.Settings
	timer    services.TimerService
	store    services.PersistenceService
	events   chan tea.Msg
	notifier *components.Notifier

	tasks         []*models.Task
	currentTaskID string
	sessions      []models.TimerSession // newest first

	tab         Tab
	width       int
	height      int
	clock       time.Time
	initialized bool

	// Tasks view state
	taskList   *components.TaskList
	adding     bool
	addTitle   textinput.Model
	addDesc    textarea.Model
	addEst     int
	addFocus   addField
	deleteArm  bool
	confirmMsg string

	// Transient UI state
	flash       string // completion banner; empty = hidden
	flashSeq    int
	toast       string // transient status message; empty = hidden
	toastSeq    int
	confirmQuit bool
	quitting    bool

	// Animation state
	progress    progress.Model // spring-animated session bar
	spinner     spinner.Model  // shown while a session auto-starts
	spinActive  bool
	checkAnimID string // task row check animation (○→◐→●), empty = off

	// Phase-4 view state
	sessTable      table.Model      // recent sessions (Stats)
	helpViewport   viewport.Model   // scrollable cheat sheet (Help)
	help           help.Model       // short help row (Help footer)
	settingsCursor int              // focused row (Settings)
}

// serviceEvents adapts services.TimerObserver callbacks (invoked on the
// service's ticker goroutine) into bubbletea messages via a channel.
type serviceEvents struct {
	ch chan tea.Msg
}

func newServiceEvents() *serviceEvents {
	return &serviceEvents{ch: make(chan tea.Msg, eventBuffer)}
}

func (s *serviceEvents) push(msg tea.Msg) {
	select {
	case s.ch <- msg:
	default: // drop rather than ever block the ticker goroutine
	}
}

func (s *serviceEvents) OnSessionStarted(session *models.TimerSession) {
	s.push(models.SessionStartedMsg{Session: session})
}
func (s *serviceEvents) OnTimerTick(remaining time.Duration) {
	s.push(models.TimerTickMsg{Remaining: remaining})
}
func (s *serviceEvents) OnSessionPaused(session *models.TimerSession) {
	s.push(models.SessionPausedMsg{Session: session})
}
func (s *serviceEvents) OnSessionResumed(session *models.TimerSession) {
	s.push(models.SessionResumedMsg{Session: session})
}
func (s *serviceEvents) OnSessionCompleted(session *models.TimerSession) {
	s.push(models.SessionCompletedMsg{Session: session})
}

// waitForEvent re-arms the channel read; every Update returns it again.
func waitForEvent(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
}

type clockTickMsg time.Time
type flashExpireMsg struct{ seq int }
type toastExpireMsg struct{ seq int }

func clockTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return clockTickMsg(t) })
}

// NewModel builds the root model. It loads persisted tasks/sessions/settings
// and resolves settings fallbacks.
func NewModel(timer services.TimerService, store services.PersistenceService) Model {
	events := newServiceEvents()
	timer.RegisterObserver(events)

	m := Model{
		timer:    timer,
		store:    store,
		events:   events.ch,
		tab:      TabTimer,
		taskList: components.NewTaskList(),
		addEst:   1,
		clock:    time.Now(),
		progress: progress.New(progress.WithDefaultGradient(), progress.WithWidth(40)),
		spinner:  spinner.New(spinner.WithSpinner(spinner.MiniDot)),
		sessTable: table.New(
			table.WithColumns([]table.Column{
				{Title: "WHEN", Width: 7},
				{Title: "TYPE", Width: 15},
				{Title: "LENGTH", Width: 9},
				{Title: "TASK", Width: 32},
			}),
			table.WithFocused(true),
			table.WithHeight(8),
		),
		help:         help.New(),
		helpViewport: viewport.New(0, 10),
	}

	titleInput := textinput.New()
	titleInput.Placeholder = "Task title…"
	titleInput.CharLimit = 120
	m.addTitle = titleInput

	m.addDesc = textarea.New()
	m.addDesc.Placeholder = "Description (optional)…"
	m.addDesc.SetHeight(2)
	m.addDesc.ShowLineNumbers = false

	m.notifier = components.NewNotifier(models.DefaultSettings().NotificationType)

	if store != nil {
		if data, err := store.GetAllData(); err == nil {
			m.tasks = data.Tasks
			m.sessions = data.Sessions
			if len(m.sessions) > maxMemorySessions {
				m.sessions = m.sessions[:maxMemorySessions]
			}
			m.settings = &data.Settings
			for _, t := range m.tasks {
				m.taskList.AddTask(t)
			}
		}
	}
	if m.settings == nil || m.settings.ID == "" {
		m.settings = models.DefaultSettings()
		if store != nil {
			_ = store.SaveSettings(m.settings)
		}
	}
	m.timer.SetSettings(m.settings)
	m.notifier = components.NewNotifier(m.settings.NotificationType)
	m.helpViewport.SetContent(helpContent())
	m.refreshSessionTable()

	return m
}

// Init starts the event bridge and the wall clock.
func (m Model) Init() tea.Cmd {
	return tea.Batch(waitForEvent(m.events), clockTick())
}

// Update is the central message router.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.initialized = true
		m.taskList.SetSize(max(20, m.width-6), max(6, m.height-10))
		// The stats view puts chart and table side by side on wide terminals.
		if m.width >= 100 {
			m.sessTable.SetWidth(max(20, m.width-36))
		} else {
			m.sessTable.SetWidth(max(20, m.width-8))
		}
		m.sessTable.SetHeight(max(4, min(9, m.height-14)))
		m.helpViewport.Width = max(20, m.width-8)
		m.helpViewport.Height = max(6, m.height-8)
		if m.width < 80 || m.height < 20 {
			m.progress.Width = 24
		} else {
			m.progress.Width = 40
		}
		return m, nil

	case clockTickMsg:
		m.clock = time.Time(msg)
		return m, clockTick()

	case models.SessionStartedMsg:
		m.flash = ""
		m.spinActive = false
		return m, tea.Batch(m.progress.SetPercent(0), waitForEvent(m.events))

	case models.SessionPausedMsg, models.SessionResumedMsg:
		m.flash = ""

	case models.SessionCompletedMsg:
		return m.onSessionCompleted(msg.Session)

	case models.TimerTickMsg:
		// The service already advanced the session; ease the bar toward the
		// new fraction.
		if cmd := m.updateProgress(); cmd != nil {
			return m, tea.Batch(cmd, waitForEvent(m.events))
		}

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, tea.Batch(cmd, waitForEvent(m.events))

	case spinner.TickMsg:
		if m.spinActive {
			// spinner.Update already returns the next-tick command; spawning
			// another spinner.Tick here would multiply the tick chains.
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, tea.Batch(cmd, waitForEvent(m.events))
		}

	case checkAnimMsg:
		return m.onCheckAnimTick(msg)

	case flashExpireMsg:
		return m.onFlashExpired(msg)

	case toastExpireMsg:
		if msg.seq == m.toastSeq {
			m.toast = ""
		}
	}

	return m, waitForEvent(m.events)
}

// onSessionCompleted records the finished session, links pomodoros to the
// current task, and raises the completion banner.
func (m Model) onSessionCompleted(session *models.TimerSession) (tea.Model, tea.Cmd) {
	if m.store != nil {
		_ = m.store.SaveSession(session)
	}
	m.sessions = append([]models.TimerSession{*session}, m.sessions...)
	if len(m.sessions) > maxMemorySessions {
		m.sessions = m.sessions[:maxMemorySessions]
	}
	m.refreshSessionTable()

	if session.Type == models.FocusSession {
		if task := m.currentTask(); task != nil {
			task.CompletedPomodoros++
			if m.store != nil {
				_ = m.store.SaveTask(task)
			}
		}
		m.flash = "🍅 Focus complete — time for a break"
	} else {
		m.flash = "☕ Break over — back to focus"
	}
	m.notifier.PlaySessionCompleteSound()

	// Show the transition spinner only when a session will auto-start.
	autoStart := (session.Type == models.FocusSession && m.settings.AutoStartBreaks) ||
		(session.Type != models.FocusSession && m.settings.AutoStartFocus)
	m.spinActive = autoStart

	m.flashSeq++
	cmds := []tea.Cmd{flashCmd(m.flashSeq), waitForEvent(m.events)}
	if autoStart {
		cmds = append(cmds, spinner.Tick)
	}
	return m, tea.Batch(cmds...)
}

// onFlashExpired hides the banner and applies auto-start settings.
func (m Model) onFlashExpired(msg flashExpireMsg) (tea.Model, tea.Cmd) {
	if msg.seq == m.flashSeq {
		m.flash = ""
		if done := m.timer.GetCurrentSession(); done != nil && done.Status == models.Completed {
			autoStart := (done.Type == models.FocusSession && m.settings.AutoStartBreaks) ||
				(done.Type != models.FocusSession && m.settings.AutoStartFocus)
			if autoStart {
				if err := m.timer.StartSession(m.timer.SuggestNext()); err == nil {
					// Keep attributing auto-started sessions to the current task.
					if s := m.timer.GetCurrentSession(); s != nil {
						s.TaskID = m.currentTaskID
					}
				}
			}
		}
	}
	return m, waitForEvent(m.events)
}

func flashCmd(seq int) tea.Cmd {
	return tea.Tick(flashDuration, func(time.Time) tea.Msg { return flashExpireMsg{seq} })
}

// checkAnimMsg drives the task-row completion animation (○→◐→●).
type checkAnimMsg struct{ frame int }

func checkAnimCmd(frame int) tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg { return checkAnimMsg{frame} })
}

// onCheckAnimTick advances the row animation one frame.
func (m Model) onCheckAnimTick(msg checkAnimMsg) (tea.Model, tea.Cmd) {
	switch msg.frame {
	case 1:
		m.taskList.SetCheckAnim(m.checkAnimID, 2)
		return m, tea.Batch(checkAnimCmd(2), waitForEvent(m.events))
	default:
		m.checkAnimID = ""
		m.taskList.SetCheckAnim("", 0)
	}
	return m, waitForEvent(m.events)
}

// updateProgress eases the bar toward the current session's elapsed fraction.
func (m *Model) updateProgress() tea.Cmd {
	s := m.timer.GetCurrentSession()
	if s == nil || s.Duration <= 0 || (s.Status != models.Running && s.Status != models.Paused) {
		return nil
	}
	fraction := 1 - float64(s.Remaining)/float64(s.Duration)
	return m.progress.SetPercent(fraction)
}

func (m *Model) toastf(format string, args ...interface{}) {
	m.toast = fmt.Sprintf(format, args...)
	m.toastSeq++
	go func(seq int) {
		time.Sleep(toastDuration)
		m.events <- toastExpireMsg{seq}
	}(m.toastSeq)
}

// --- session helpers ---

func (m Model) currentTask() *models.Task {
	if m.currentTaskID == "" {
		return nil
	}
	for _, t := range m.tasks {
		if t.ID == m.currentTaskID {
			return t
		}
	}
	return nil
}

func (m Model) todayCount() int {
	today := time.Now()
	n := 0
	for _, s := range m.sessions {
		if s.Type != models.FocusSession || s.CompletedAt == nil {
			continue
		}
		c := *s.CompletedAt
		if c.Year() == today.Year() && c.YearDay() == today.YearDay() {
			n++
		}
	}
	return n
}

// startFocus starts a focus session, optionally overriding the duration.
// The session is linked to the current task so stats can attribute it.
func (m *Model) startFocus(duration time.Duration) {
	m.flash = ""
	if duration > 0 {
		_ = m.timer.StartSessionWithDuration(models.FocusSession, duration)
	} else {
		_ = m.timer.StartSession(models.FocusSession)
	}
	if s := m.timer.GetCurrentSession(); s != nil {
		s.TaskID = m.currentTaskID
	}
}

// --- key handling ---

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m.quitNow()
	}

	// New-task form captures all keys except esc/ctrl+c.
	if m.adding {
		return m.handleAddFormKey(msg)
	}

	switch key {
	case "esc":
		switch {
		case m.confirmQuit:
			m.confirmQuit = false
		case m.deleteArm:
			m.deleteArm = false
		case m.tab == TabHelp:
			m.tab = TabTimer
		}
		return m, nil

	case "?":
		m.tab = TabHelp
		return m, nil

	case "tab":
		m.tab = (m.tab + 1) % tabCount
		return m, nil
	case "shift+tab":
		m.tab = (m.tab + tabCount - 1) % tabCount
		return m, nil

	case "q":
		if m.sessionRunning() {
			m.confirmQuit = true
			return m, nil
		}
		return m.quitNow()
	}

	if m.confirmQuit {
		switch key {
		case "y", "Y":
			return m.quitNow()
		case "n", "N":
			m.confirmQuit = false
			return m, nil
		}
		return m, nil
	}

	if m.deleteArm {
		switch key {
		case "y", "Y":
			return m.confirmDeleteTask()
		default:
			m.deleteArm = false
			return m, nil
		}
	}

	switch m.tab {
	case TabTimer:
		return m.handleTimerKey(key)
	case TabTasks:
		return m.handleTasksKey(msg, key)
	case TabStats:
		return m.handleStatsKey(msg, key)
	case TabSettings:
		return m.handleSettingsKey(key)
	case TabHelp:
		return m.handleHelpKey(msg, key)
	}

	return m, waitForEvent(m.events)
}

func (m Model) handleTimerKey(key string) (tea.Model, tea.Cmd) {
	session := m.timer.GetCurrentSession()

	switch key {
	case " ", "space", "enter":
		switch {
		case session == nil || session.Status == models.Completed:
			m.startFocus(0)
		case session.Status == models.Running:
			_ = m.timer.PauseSession()
		case session.Status == models.Paused:
			_ = m.timer.ResumeSession()
		}

	case "r":
		// Reset back to the ready state (abandons the session, no credit).
		if session != nil {
			_ = m.timer.StopSession()
			m.flash = ""
			m.spinActive = false
			return m, tea.Batch(m.progress.SetPercent(0), waitForEvent(m.events))
		}

	case "s":
		// Skip to the next session in the rhythm.
		if session != nil {
			_ = m.timer.StopSession()
		}
		if err := m.timer.StartSession(m.timer.SuggestNext()); err == nil {
			m.flash = ""
		}

	case "1":
		m.startFocus(15 * time.Minute)
	case "2":
		m.startFocus(25 * time.Minute)
	case "3":
		m.startFocus(35 * time.Minute)
	}

	return m, waitForEvent(m.events)
}

// handleTasksKey handles the tasks tab keys; nav keys are forwarded to the
// bubbles list.
func (m Model) handleTasksKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch key {
	case "a":
		m.adding = true
		m.addEst = 1
		m.addFocus = addFieldTitle
		m.addTitle.Reset()
		m.addDesc.Reset()
		return m, m.addTitle.Focus()

	case "x":
		if task := m.taskList.GetSelected(); task != nil {
			wasDone := task.Status == models.DoneStatus
			m.toggleTaskDone(task)
			if !wasDone {
				// Animate the check mark: ○ → ◐ → ●
				m.checkAnimID = task.ID
				m.taskList.SetCheckAnim(task.ID, 1)
				return m, tea.Batch(checkAnimCmd(1), waitForEvent(m.events))
			}
			m.taskList.SetCheckAnim("", 0)
		}
		return m, nil

	case "d":
		if task := m.taskList.GetSelected(); task != nil {
			m.deleteArm = true
			m.confirmMsg = task.Title
			m.toastf("Delete %q? y to confirm", task.Title)
			return m, waitForEvent(m.events)
		}
		return m, nil

	case "enter":
		if task := m.taskList.GetSelected(); task != nil {
			m.currentTaskID = task.ID
			// Working on it now: mark in progress (unless already done).
			if task.Status == models.ToDoStatus {
				task.Status = models.InProgressStatus
				m.taskList.UpdateTask(task)
				if m.store != nil {
					_ = m.store.SaveTask(task)
				}
			}
			m.toastf("Focusing on: %s", task.Title)
		}
		return m, nil
	}

	// Navigation goes to the bubbles list.
	switch key {
	case "up", "down", "j", "k", "home", "end", "pgup", "pgdown", "g", "G":
		updated, c := m.taskList.Update(msg)
		m.taskList = updated.(*components.TaskList)
		cmd = c
	}

	return m, tea.Batch(cmd, waitForEvent(m.events))
}

func (m *Model) toggleTaskDone(task *models.Task) {
	if task.Status == models.DoneStatus {
		task.Status = models.ToDoStatus
		task.CompletedAt = nil
	} else {
		task.Status = models.DoneStatus
		now := time.Now()
		task.CompletedAt = &now
	}
	m.taskList.UpdateTask(task)
	if m.store != nil {
		_ = m.store.SaveTask(task)
	}
}

func (m *Model) confirmDeleteTask() (tea.Model, tea.Cmd) {
	if task := m.taskList.GetSelected(); task != nil {
		id := task.ID
		m.taskList.RemoveTask(id)
		for i, t := range m.tasks {
			if t.ID == id {
				m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
				break
			}
		}
		if m.store != nil {
			_ = m.store.DeleteTask(id)
		}
		if m.currentTaskID == id {
			m.currentTaskID = ""
		}
		m.toastf("Deleted task")
	}
	m.deleteArm = false
	return m, waitForEvent(m.events)
}

// handleAddFormKey routes keys while the new-task form is open.
func (m Model) handleAddFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.adding = false
		m.addTitle.Blur()
		m.addDesc.Blur()
		return m, nil

	case "enter":
		title := strings.TrimSpace(m.addTitle.Value())
		if title == "" {
			m.toastf("Give the task a title first")
			return m, waitForEvent(m.events)
		}
		task := models.NewTask(title, m.addDesc.Value(), m.addEst)
		m.tasks = append(m.tasks, task)
		m.taskList.AddTask(task)
		if m.store != nil {
			_ = m.store.SaveTask(task)
		}
		m.adding = false
		m.addTitle.Blur()
		m.addDesc.Blur()
		m.toastf("Added: %s", task.Title)
		return m, waitForEvent(m.events)

	case "tab":
		if m.addFocus == addFieldTitle {
			m.addFocus = addFieldDescription
			m.addTitle.Blur()
			cmd = m.addDesc.Focus()
		} else {
			m.addFocus = addFieldTitle
			m.addDesc.Blur()
			cmd = m.addTitle.Focus()
		}
		return m, tea.Batch(cmd, waitForEvent(m.events))

	case "+", "=":
		if m.addEst < 16 {
			m.addEst++
		}
		return m, waitForEvent(m.events)

	case "-", "_":
		if m.addEst > 1 {
			m.addEst--
		}
		return m, waitForEvent(m.events)
	}

	// Everything else feeds the focused input.
	if m.addFocus == addFieldTitle {
		var c tea.Cmd
		m.addTitle, c = m.addTitle.Update(msg)
		cmd = c
	} else {
		var c tea.Cmd
		m.addDesc, c = m.addDesc.Update(msg)
		cmd = c
	}
	return m, tea.Batch(cmd, waitForEvent(m.events))
}

// handleStatsKey scrolls the recent-sessions table.
func (m Model) handleStatsKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch key {
	case "up", "down", "j", "k", "pgup", "pgdown", "home", "end":
		var updated table.Model
		updated, cmd = m.sessTable.Update(msg)
		m.sessTable = updated
	}
	return m, tea.Batch(cmd, waitForEvent(m.events))
}

// handleHelpKey scrolls the cheat sheet.
func (m Model) handleHelpKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch key {
	case "up", "down", "j", "k", "pgup", "pgdown":
		m.helpViewport, cmd = m.helpViewport.Update(msg)
	}
	return m, tea.Batch(cmd, waitForEvent(m.events))
}

// --- quit ---

func (m Model) sessionRunning() bool {
	s := m.timer.GetCurrentSession()
	return s != nil && s.Status == models.Running
}

func (m Model) quitNow() (tea.Model, tea.Cmd) {
	m.timer.Shutdown()
	if m.store != nil {
		_ = m.store.Close()
	}
	m.quitting = true
	return m, tea.Quit
}
