package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"pomodoro/src/models"
)

// maxStoredSessions caps the session history so the state file stays small.
const maxStoredSessions = 1000

// JSONPersistence implements PersistenceService with a single JSON file.
// All data lives in memory and is flushed to disk on every mutation; the
// dataset is small (tasks, settings, capped session history), so synchronous
// writes are cheap and keep the file always consistent.
type JSONPersistence struct {
	mu            sync.Mutex
	path          string
	data          AppData
	corruptBackup string // set when the state file was unreadable and moved aside
}

// Compile-time check that JSONPersistence implements PersistenceService.
var _ PersistenceService = (*JSONPersistence)(nil)

// DefaultPersistencePath returns the platform-standard state file location:
// ~/.config/pomodoro/state.json (respects XDG_CONFIG_HOME).
func DefaultPersistencePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(dir, "pomodoro", "state.json"), nil
}

// NewJSONPersistence opens (or creates) the state file at path and loads its
// contents. A missing file is not an error — the app starts fresh.
func NewJSONPersistence(path string) (*JSONPersistence, error) {
	p := &JSONPersistence{
		path: path,
		data: AppData{
			Settings: *models.DefaultSettings(),
			Tasks:    []*models.Task{},
			Sessions: []models.TimerSession{},
		},
	}

	raw, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// First run: persist the defaults so the file exists from the start.
		if err := p.save(); err != nil {
			return nil, err
		}
		return p, nil
	case err != nil:
		return nil, fmt.Errorf("read state file: %w", err)
	}

	if err := json.Unmarshal(raw, &p.data); err != nil {
		// A corrupt state file must not make the app unusable: move it aside
		// (so nothing is silently destroyed) and start from defaults.
		backup := fmt.Sprintf("%s.corrupt-%s", path, time.Now().Format("20060102-150405"))
		renameErr := os.Rename(path, backup)
		if renameErr != nil {
			backup = ""
		}
		if saveErr := p.save(); saveErr != nil {
			return nil, fmt.Errorf("parse state file %s: %w (and rewriting defaults failed: %v / %v)", path, err, renameErr, saveErr)
		}
		p.corruptBackup = backup
	}
	if p.data.Tasks == nil {
		p.data.Tasks = []*models.Task{}
	}
	if p.data.Sessions == nil {
		p.data.Sessions = []models.TimerSession{}
	}
	return p, nil
}

// NewDefaultPersistence opens the state file at DefaultPersistencePath.
func NewDefaultPersistence() (*JSONPersistence, error) {
	path, err := DefaultPersistencePath()
	if err != nil {
		return nil, err
	}
	return NewJSONPersistence(path)
}

// save flushes the in-memory data to disk atomically. Caller holds the lock.
func (p *JSONPersistence) save() error {
	if err := os.MkdirAll(filepath.Dir(p.path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	raw, err := json.MarshalIndent(p.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	tmp := p.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}
	if err := os.Rename(tmp, p.path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}

// SaveSession appends or updates a session in the history.
func (p *JSONPersistence) SaveSession(session *models.TimerSession) error {
	if session == nil {
		return errors.New("session is nil")
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	replaced := false
	for i := range p.data.Sessions {
		if p.data.Sessions[i].ID == session.ID {
			p.data.Sessions[i] = *session
			replaced = true
			break
		}
	}
	if !replaced {
		p.data.Sessions = append(p.data.Sessions, *session)
	}
	// Keep the newest sessions; sort descending by completion/creation time.
	sort.Slice(p.data.Sessions, func(i, j int) bool {
		ti := sessionEnd(p.data.Sessions[i])
		tj := sessionEnd(p.data.Sessions[j])
		return ti.After(tj)
	})
	if len(p.data.Sessions) > maxStoredSessions {
		p.data.Sessions = p.data.Sessions[:maxStoredSessions]
	}
	return p.save()
}

// sessionEnd is the best timestamp for ordering a session.
func sessionEnd(s models.TimerSession) time.Time {
	if s.CompletedAt != nil {
		return *s.CompletedAt
	}
	if s.StartTime != nil {
		return *s.StartTime
	}
	return s.CreatedAt
}

// GetSession returns a session by ID.
func (p *JSONPersistence) GetSession(id string) (*models.TimerSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := range p.data.Sessions {
		if p.data.Sessions[i].ID == id {
			s := p.data.Sessions[i]
			return &s, nil
		}
	}
	return nil, errors.New("session not found")
}

// GetSessions returns sessions matching the filter, newest first.
func (p *JSONPersistence) GetSessions(filter SessionFilter) ([]*models.TimerSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]*models.TimerSession, 0, len(p.data.Sessions))
	for i := range p.data.Sessions {
		s := p.data.Sessions[i]
		if matchesSessionFilter(s, filter) {
			copy := s
			result = append(result, &copy)
		}
	}
	return result, nil
}

func matchesSessionFilter(s models.TimerSession, filter SessionFilter) bool {
	if len(filter.SessionTypes) > 0 {
		found := false
		for _, t := range filter.SessionTypes {
			if s.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(filter.Statuses) > 0 {
		found := false
		for _, st := range filter.Statuses {
			if s.Status == st {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.DateRange != nil {
		t := sessionEnd(s)
		if t.Before(filter.DateRange.Start) || t.After(filter.DateRange.End) {
			return false
		}
	}
	return true
}

// SaveSettings replaces the stored settings.
func (p *JSONPersistence) SaveSettings(settings *models.Settings) error {
	if settings == nil {
		return errors.New("settings is nil")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	settings.UpdatedAt = time.Now()
	p.data.Settings = *settings
	return p.save()
}

// GetSettings returns the stored settings.
func (p *JSONPersistence) GetSettings() (*models.Settings, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	s := p.data.Settings
	return &s, nil
}

// SaveTask appends or updates a task.
func (p *JSONPersistence) SaveTask(task *models.Task) error {
	if task == nil {
		return errors.New("task is nil")
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	replaced := false
	for i, t := range p.data.Tasks {
		if t.ID == task.ID {
			p.data.Tasks[i] = task
			replaced = true
			break
		}
	}
	if !replaced {
		p.data.Tasks = append(p.data.Tasks, task)
	}
	return p.save()
}

// GetTask returns a task by ID.
func (p *JSONPersistence) GetTask(id string) (*models.Task, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, t := range p.data.Tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, errors.New("task not found")
}

// GetTasks returns tasks matching the filter.
func (p *JSONPersistence) GetTasks(filter TaskFilter) ([]*models.Task, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]*models.Task, 0, len(p.data.Tasks))
	for _, t := range p.data.Tasks {
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		if filter.Tag != nil && !taskHasTag(t, *filter.Tag) {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func taskHasTag(t *models.Task, tag string) bool {
	for _, tg := range t.Tags {
		if tg == tag {
			return true
		}
	}
	return false
}

// DeleteTask removes a task by ID.
func (p *JSONPersistence) DeleteTask(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	kept := p.data.Tasks[:0]
	for _, t := range p.data.Tasks {
		if t.ID != id {
			kept = append(kept, t)
		}
	}
	p.data.Tasks = kept
	return p.save()
}

// GetAllData returns everything stored (settings, tasks, session history).
func (p *JSONPersistence) GetAllData() (*AppData, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	data := p.data // shallow copy is fine; slices are shared but treated read-only by callers
	tasksCopy := make([]*models.Task, len(p.data.Tasks))
	copy(tasksCopy, p.data.Tasks)
	data.Tasks = tasksCopy

	sessionsCopy := make([]models.TimerSession, len(p.data.Sessions))
	copy(sessionsCopy, p.data.Sessions)
	data.Sessions = sessionsCopy

	return &data, nil
}

// ImportData replaces all stored data (used by the future export/import flow).
func (p *JSONPersistence) ImportData(data *AppData) error {
	if data == nil {
		return errors.New("data is nil")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data = *data
	if p.data.Tasks == nil {
		p.data.Tasks = []*models.Task{}
	}
	if p.data.Sessions == nil {
		p.data.Sessions = []models.TimerSession{}
	}
	return p.save()
}

// CorruptBackup returns the path the previous state file was moved to when it
// could not be parsed (empty when the last load was clean). The UI shows it
// once so the user knows where the old data lives.
func (p *JSONPersistence) CorruptBackup() string {
	return p.corruptBackup
}

// Close flushes any pending state. Data is saved eagerly, so this only
// re-writes the file for safety.
func (p *JSONPersistence) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.save()
}
