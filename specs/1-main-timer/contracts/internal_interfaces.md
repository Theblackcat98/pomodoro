# Internal Interface Contracts: Pomodoro TUI App

## Overview

This document defines the internal interfaces and contracts for the Pomodoro TUI application components. These contracts ensure consistent interaction between different parts of the application.

## Timer Service Interface

### Purpose
Manage the core timing functionality for Pomodoro sessions

### Contract

```go
type TimerService interface {
    StartSession(sessionType SessionType) error
    PauseSession() error
    ResumeSession() error
    StopSession() error
    GetCurrentSession() *TimerSession
    GetRemainingTime() time.Duration
    RegisterObserver(observer TimerObserver)
    UnregisterObserver(observer TimerObserver)
}

type TimerObserver interface {
    OnTimerTick(remaining time.Duration)
    OnSessionStarted(session *TimerSession)
    OnSessionPaused(session *TimerSession)
    OnSessionResumed(session *TimerSession)
    OnSessionCompleted(session *TimerSession)
    OnSessionCancelled(session *TimerSession)
}
```

### Validations
- Session cannot be started if another session is already running
- Pausing requires an active session
- Remaining time must always be >= 0

## Settings Service Interface

### Purpose
Manage user preferences and application settings

### Contract

```go
type SettingsService interface {
    GetSettings() *Settings
    UpdateSettings(updates SettingsUpdate) error
    ResetToDefaults() error
    ExportSettings() ([]byte, error)
    ImportSettings(data []byte) error
}

type SettingsUpdate struct {
    FocusDuration      *time.Duration   `json:"focus_duration,omitempty"`
    ShortBreakDuration *time.Duration   `json:"short_break_duration,omitempty"`
    LongBreakDuration  *time.Duration   `json:"long_break_duration,omitempty"`
    LongBreakInterval  *int             `json:"long_break_interval,omitempty"`
    AutoStartBreaks    *bool            `json:"auto_start_breaks,omitempty"`
    AutoStartFocus     *bool            `json:"auto_start_focus,omitempty"`
    NotificationType   *NotificationType `json:"notification_type,omitempty"`
    Theme              *string          `json:"theme,omitempty"`
}
```

### Validations
- Duration values must be positive
- LongBreakInterval must be at least 1
- Theme must be a valid theme identifier

## Task Service Interface

### Purpose
Manage user tasks and their association with Pomodoro sessions

### Contract

```go
type TaskService interface {
    CreateTask(task TaskCreateRequest) (*Task, error)
    GetTask(id string) (*Task, error)
    GetTasks(filter TaskFilter) ([]*Task, error)
    UpdateTask(id string, update TaskUpdateRequest) (*Task, error)
    DeleteTask(id string) error
    CompleteTask(id string) error
    StartTask(id string) error
}

type TaskCreateRequest struct {
    Title              string   `json:"title"`
    Description        string   `json:"description"`
    EstimatedPomodoros *int     `json:"estimated_pomodoros,omitempty"`
    Tags               []string `json:"tags,omitempty"`
}

type TaskUpdateRequest struct {
    Title              *string   `json:"title,omitempty"`
    Description        *string   `json:"description,omitempty"`
    Status             *TaskStatus `json:"status,omitempty"`
    EstimatedPomodoros *int      `json:"estimated_pomodoros,omitempty"`
    CompletedPomodoros *int      `json:"completed_pomodoros,omitempty"`
    Tags               *[]string `json:"tags,omitempty"`
}

type TaskFilter struct {
    Status *TaskStatus `json:"status,omitempty"`
    Tag    *string     `json:"tag,omitempty"`
}
```

### Validations
- Task title cannot be empty
- EstimatedPomodoros must be non-negative
- Tags cannot contain invalid characters

## Persistence Interface

### Purpose
Handle data persistence for sessions, tasks, and settings

### Contract

```go
type PersistenceService interface {
    SaveSession(session *TimerSession) error
    GetSession(id string) (*TimerSession, error)
    GetSessions(filter SessionFilter) ([]*TimerSession, error)
    SaveSettings(settings *Settings) error
    GetSettings() (*Settings, error)
    SaveTask(task *Task) error
    GetTask(id string) (*Task, error)
    GetTasks(filter TaskFilter) ([]*Task, error)
    DeleteTask(id string) error
    GetAllData() (*AppData, error)
    ImportData(data *AppData) error
    Close() error
}

type AppData struct {
    Sessions []TimerSession `json:"sessions"`
    Settings Settings      `json:"settings"`
    Tasks    []Task        `json:"tasks"`
}

type SessionFilter struct {
    SessionTypes []SessionType   `json:"session_types,omitempty"`
    Statuses     []SessionStatus `json:"statuses,omitempty"`
    DateRange    *DateRange      `json:"date_range,omitempty"`
}

type DateRange struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}
```

### Validations
- Data must be valid according to individual entity contracts
- File operations must handle concurrent access safely

## UI Component Interface

### Purpose
Standardized interface for all UI components to ensure consistency

### Contract

```go
// This interface represents the Bubble Tea Model interface that all components follow
type UIComponent interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (tea.Model, tea.Cmd)
    View() string
}

// Additional standardized methods for all components
type ComponentWithFocus interface {
    Focus() tea.Cmd
    Blur()
    IsFocused() bool
}

type ComponentWithSize interface {
    SetSize(width, height int)
    GetPreferredSize() (int, int)
}

type ComponentWithHelp interface {
    Help() []key.Binding
}
```

## Event System Contract

### Purpose
Standardized event communication between components

### Contract

```go
// Standardized message types for the Bubble Tea update loop
type AppMessage interface{}

// Timer-related messages
type TimerTickMsg struct {
    Remaining time.Duration
}

type SessionStartedMsg struct {
    Session *TimerSession
}

type SessionCompletedMsg struct {
    Session *TimerSession
}

// Task-related messages
type TaskCreatedMsg struct {
    Task *Task
}

type TaskUpdatedMsg struct {
    Task *Task
}

type TaskDeletedMsg struct {
    TaskID string
}

// Settings-related messages
type SettingsUpdatedMsg struct {
    Settings *Settings
}

// Internal command messages
type ShowErrorMessage struct {
    Error error
}

type ShowSuccessMessage struct {
    Message string
}
```