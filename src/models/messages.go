package models

import "time"

// AppMessage interface represents the standardized message types for the Bubble Tea update loop
type AppMessage interface{}

// Timer-related messages
type TimerTickMsg struct {
	Remaining time.Duration
}

type SessionStartedMsg struct {
	Session *TimerSession
}

type SessionPausedMsg struct {
	Session *TimerSession
}

type SessionResumedMsg struct {
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