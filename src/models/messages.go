package models

import "time"

// Timer-related messages bridged from the timer service into the update loop.
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

type SessionAbandonedMsg struct {
	Session *TimerSession
}
