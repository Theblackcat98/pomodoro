package services

import (
	"pomodoro/src/models"
	"time"
)

// PersistenceService interface handles data persistence for sessions, tasks, and settings
type PersistenceService interface {
	SaveSession(session *models.TimerSession) error
	GetSession(id string) (*models.TimerSession, error)
	GetSessions(filter SessionFilter) ([]*models.TimerSession, error)
	SaveSettings(settings *models.Settings) error
	GetSettings() (*models.Settings, error)
	SaveTask(task *models.Task) error
	GetTask(id string) (*models.Task, error)
	GetTasks(filter TaskFilter) ([]*models.Task, error)
	DeleteTask(id string) error
	GetAllData() (*AppData, error)
	ImportData(data *AppData) error
	Close() error
}

// AppData represents all application data for import/export
type AppData struct {
	Sessions []models.TimerSession `json:"sessions"`
	Settings models.Settings       `json:"settings"`
	Tasks    []models.Task         `json:"tasks"`
}

// SessionFilter defines filters for querying sessions
type SessionFilter struct {
	SessionTypes []models.SessionType   `json:"session_types,omitempty"`
	Statuses     []models.SessionStatus `json:"statuses,omitempty"`
	DateRange    *DateRange             `json:"date_range,omitempty"`
}

// TaskFilter defines filters for querying tasks
type TaskFilter struct {
	Status *models.TaskStatus `json:"status,omitempty"`
	Tag    *string            `json:"tag,omitempty"`
}

// DateRange represents a range of dates
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}