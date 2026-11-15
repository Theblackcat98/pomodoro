package models

import (
	"strconv"
	"time"
)

type TaskStatus string

const (
	ToDoStatus       TaskStatus = "todo"
	InProgressStatus TaskStatus = "in_progress"
	DoneStatus       TaskStatus = "done"
)

type Task struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Status             TaskStatus `json:"status"`
	EstimatedPomodoros int        `json:"estimated_pomodoros"` // How many pomodoros user thinks it will take
	CompletedPomodoros int        `json:"completed_pomodoros"` // How many pomodoros actually spent
	CreatedAt          time.Time  `json:"created_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	Tags               []string   `json:"tags,omitempty"`
}

// NewID returns a compact unique identifier based on the current time.
func NewID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// NewTask creates a task with a fresh ID, timestamp, and to-do status.
func NewTask(title, description string, estimatedPomodoros int) *Task {
	if estimatedPomodoros <= 0 {
		estimatedPomodoros = 1
	}
	return &Task{
		ID:                 NewID(),
		Title:              title,
		Description:        description,
		Status:             ToDoStatus,
		EstimatedPomodoros: estimatedPomodoros,
		CreatedAt:          time.Now(),
	}
}
