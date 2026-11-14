package services

import (
	"path/filepath"
	"testing"
	"time"

	"pomodoro/src/models"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

func TestPersistenceRoundTrip(t *testing.T) {
	path := tempPath(t)

	p1, err := NewJSONPersistence(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	// First run persists defaults.
	settings, err := p1.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if settings.FocusDuration != 25*time.Minute {
		t.Fatalf("default focus duration = %v, want 25m", settings.FocusDuration)
	}

	task := models.NewTask("Write the plan", "layout doc", 3)
	if err := p1.SaveTask(task); err != nil {
		t.Fatalf("SaveTask: %v", err)
	}

	session := &models.TimerSession{
		ID:        models.NewID(),
		Type:      models.FocusSession,
		Duration:  25 * time.Minute,
		Remaining: 0,
		Status:    models.Completed,
		CreatedAt: time.Now(),
	}
	now := time.Now()
	session.CompletedAt = &now
	if err := p1.SaveSession(session); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	// Mutate in-memory settings and save.
	settings.FocusDuration = 35 * time.Minute
	if err := p1.SaveSettings(settings); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	if err := p1.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Reopen: everything must survive the restart.
	p2, err := NewJSONPersistence(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	tasks, err := p2.GetTasks(TaskFilter{})
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Title != "Write the plan" || tasks[0].EstimatedPomodoros != 3 {
		t.Fatalf("tasks after reload = %+v", tasks)
	}

	sessions, err := p2.GetSessions(SessionFilter{})
	if err != nil {
		t.Fatalf("GetSessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].Type != models.FocusSession {
		t.Fatalf("sessions after reload = %+v", sessions)
	}

	settings2, _ := p2.GetSettings()
	if settings2.FocusDuration != 35*time.Minute {
		t.Fatalf("settings after reload focus = %v, want 35m", settings2.FocusDuration)
	}
}

func TestPersistenceDeleteAndFilter(t *testing.T) {
	path := tempPath(t)
	p, err := NewJSONPersistence(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	a := models.NewTask("one", "", 1)
	b := models.NewTask("two", "", 1)
	done := models.DoneStatus
	b.Status = done
	_ = p.SaveTask(a)
	_ = p.SaveTask(b)

	doneTasks, _ := p.GetTasks(TaskFilter{Status: &done})
	if len(doneTasks) != 1 || doneTasks[0].ID != b.ID {
		t.Fatalf("status filter = %+v", doneTasks)
	}

	if err := p.DeleteTask(a.ID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if _, err := p.GetTask(a.ID); err == nil {
		t.Fatal("GetTask after delete should fail")
	}
	tasks, _ := p.GetTasks(TaskFilter{})
	if len(tasks) != 1 {
		t.Fatalf("tasks after delete = %d, want 1", len(tasks))
	}
}
