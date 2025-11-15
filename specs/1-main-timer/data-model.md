# Data Model: Complete Bubbles Component Integration

## Core Entities

### TimerSession
Represents a single Pomodoro session

```go
type TimerSession struct {
    ID          string        `json:"id"`
    Type        SessionType   `json:"type"`      // Focus, ShortBreak, LongBreak
    Duration    time.Duration `json:"duration"`  // Original duration in seconds
    Remaining   time.Duration `json:"remaining"` // Time remaining in seconds
    Status      SessionStatus `json:"status"`    // Running, Paused, Completed
    StartTime   *time.Time    `json:"start_time,omitempty"`
    CompletedAt *time.Time    `json:"completed_at,omitempty"`
    CreatedAt   time.Time     `json:"created_at"`
}
```

**Fields:**
- `ID`: Unique identifier for the session
- `Type`: The type of session (focus, short break, long break)
- `Duration`: Original session duration
- `Remaining`: Remaining time in the session
- `Status`: Current state of the session
- `StartTime`: When the session was started
- `CompletedAt`: When the session was completed
- `CreatedAt`: When the record was created

**Validation Rules:**
- Duration must be positive
- Status must be one of the allowed values
- Start time must be before completion time if both exist

**State Transitions:**
- Created → Running (when timer starts)
- Running → Paused (when paused)
- Paused → Running (when resumed)
- Running → Completed (when timer finishes)
- Paused → Completed (when manually completed)

### SessionType
Enumeration for session types

```go
type SessionType string
const (
    FocusSession SessionType = "focus"
    ShortBreak   SessionType = "short_break" 
    LongBreak    SessionType = "long_break"
)
```

### SessionStatus
Enumeration for session states

```go
type SessionStatus string
const (
    Running   SessionStatus = "running"
    Paused    SessionStatus = "paused"
    Completed SessionStatus = "completed"
)
```

### Settings
User preferences for the Pomodoro experience

```go
type Settings struct {
    ID                   string           `json:"id"`
    FocusDuration        time.Duration    `json:"focus_duration"`        // Default: 25 minutes
    ShortBreakDuration   time.Duration    `json:"short_break_duration"`  // Default: 5 minutes
    LongBreakDuration    time.Duration    `json:"long_break_duration"`   // Default: 15 minutes
    LongBreakInterval    int              `json:"long_break_interval"`   // After how many focus sessions, default: 4
    AutoStartBreaks      bool             `json:"auto_start_breaks"`     // Default: true
    AutoStartFocus       bool             `json:"auto_start_focus"`      // After breaks, default: false
    NotificationType     NotificationType `json:"notification_type"`     // Default: "sound"
    Theme                string           `json:"theme"`                 // Default: "default"
    CompletedSessions    int              `json:"completed_sessions"`    // Number of completed sessions
    CreatedAt            time.Time        `json:"created_at"`
    UpdatedAt            time.Time        `json:"updated_at"`
}
```

**Fields:**
- `FocusDuration`: Duration of focus sessions
- `ShortBreakDuration`: Duration of short breaks
- `LongBreakDuration`: Duration of long breaks
- `LongBreakInterval`: How many focus sessions before a long break
- `AutoStartBreaks`: Whether to automatically start breaks
- `AutoStartFocus`: Whether to automatically start focus after breaks
- `NotificationType`: Type of notification (sound, vibration, visual)
- `Theme`: Selected application theme
- `CompletedSessions`: Running count of completed sessions

### NotificationType
Enumeration for notification types

```go
type NotificationType string
const (
    SoundNotification   NotificationType = "sound"
    VibrationNotification NotificationType = "vibration"
    VisualNotification  NotificationType = "visual"
    SilentNotification  NotificationType = "silent"
)
```

### Task
Represents a task that can be associated with Pomodoro sessions

```go
type Task struct {
    ID                   string         `json:"id"`
    Title                string         `json:"title"`
    Description          string         `json:"description"`
    Status               TaskStatus     `json:"status"`
    EstimatedPomodoros   int            `json:"estimated_pomodoros"`    // How many pomodoros user thinks it will take
    CompletedPomodoros   int            `json:"completed_pomodoros"`    // How many pomodoros actually spent
    CreatedAt            time.Time      `json:"created_at"`
    CompletedAt          *time.Time     `json:"completed_at,omitempty"`
    Tags                 []string       `json:"tags,omitempty"`
}
```

**Fields:**
- `Title`: Task name/title
- `Description`: Detailed task description
- `Status`: Current status of the task
- `EstimatedPomodoros`: How many pomodoros user estimates it will take
- `CompletedPomodoros`: How many pomodoros actually spent
- `Tags`: Optional tags for categorizing tasks

### TaskStatus
Enumeration for task states

```go
type TaskStatus string
const (
    ToDoStatus     TaskStatus = "todo"
    InProgressStatus TaskStatus = "in_progress"
    DoneStatus     TaskStatus = "done"
)
```

## Relationships

- Settings: One-to-many with TimerSession (one user settings, many sessions)
- Task: One-to-many with TimerSession (one task can have multiple session attempts)

## Validation Rules

1. **TimerSession**:
   - Duration must be positive
   - Status must be one of valid values
   - If StartTime and CompletedAt exist, CompletedAt must be after StartTime

2. **Settings**:
   - Duration values must be positive
   - LongBreakInterval must be positive

3. **Task**:
   - Title cannot be empty
   - EstimatedPomodoros must be non-negative
   - CompletedPomodoros must be non-negative
   - CompletedPomodoros cannot exceed actual sessions spent (domain rule)

## State Transitions

### TimerSession State Machine
```
Created → Running (start timer)
Running → Paused (pause timer)  
Paused → Running (resume timer)
Running → Completed (timer finishes)
Paused → Completed (user completes manually)
```

### Task Status Flow
```
ToDo → InProgress (user starts working on task)
InProgress → Done (user completes task)
Done → InProgress (user reopens task)
```