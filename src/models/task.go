package models

import "time"

type TaskStatus string

const (
	ToDoStatus     TaskStatus = "todo"
	InProgressStatus TaskStatus = "in_progress"
	DoneStatus     TaskStatus = "done"
)

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