package views

import (
	"strings"
	"testing"
	"time"

	"pomodoro/src/models"
)

// statsModel builds a bare model parked on the Stats tab with the given
// history/tasks, then rebuilds the table exactly as the app would.
func statsModel(sessions []models.TimerSession, tasks ...*models.Task) Model {
	m := newBareModel()
	m.tab = TabStats
	m.sessions = sessions
	m.tasks = tasks
	m.refreshSessionTable()
	return m
}

func sortFilterSession(at time.Time, elapsed time.Duration, typ models.SessionType, taskID string) models.TimerSession {
	done := at
	return models.TimerSession{
		ID:          models.NewID(),
		Type:        typ,
		Duration:    25 * time.Minute,
		Elapsed:     elapsed,
		Status:      models.Completed,
		CompletedAt: &done,
		TaskID:      taskID,
	}
}

// (a) sort toggles direction and orders rows; the header marks the active column.
func TestStatsSortTogglesDirectionAndOrdersRows(t *testing.T) {
	now := time.Now()
	// sessions are stored newest-first (the model's invariant).
	sessions := []models.TimerSession{
		sortFilterSession(now.Add(-1*time.Hour), 20*time.Minute, models.FocusSession, "t3"),
		sortFilterSession(now.Add(-2*time.Hour), 30*time.Minute, models.ShortBreak, "t2"),
		sortFilterSession(now.Add(-3*time.Hour), 10*time.Minute, models.FocusSession, "t1"),
	}
	tasks := []*models.Task{
		{ID: "t1", Title: "Zebra"},
		{ID: "t2", Title: "Alpha"},
		{ID: "t3", Title: "Mango"},
	}
	m := statsModel(sessions, tasks...)

	// No explicit sort: newest-first must be preserved.
	if got := m.sessTable.Rows()[0][3]; got != "Mango" {
		t.Fatalf("default first row = %q, want Mango (newest-first)", got)
	}

	// "3" = LENGTH ascending: 10m, 20m, 30m.
	m = driveKey(m, "3")
	if m.statsSortCol != 2 || !m.statsSortAsc {
		t.Fatalf("after first 3: col=%d asc=%v, want col=2 asc=true", m.statsSortCol, m.statsSortAsc)
	}
	rows := m.sessTable.Rows()
	want := []string{"Zebra", "Mango", "Alpha"}
	for i, w := range want {
		if rows[i][3] != w {
			t.Fatalf("asc row %d task = %q, want %q", i, rows[i][3], w)
		}
	}
	if title := m.sessTable.Columns()[2].Title; !strings.Contains(title, "▲") {
		t.Fatalf("active column header = %q, want an asc marker", title)
	}

	// "3" again toggles to descending: 30m, 20m, 10m.
	m = driveKey(m, "3")
	if m.statsSortCol != 2 || m.statsSortAsc {
		t.Fatalf("after second 3: col=%d asc=%v, want col=2 asc=false", m.statsSortCol, m.statsSortAsc)
	}
	rows = m.sessTable.Rows()
	want = []string{"Alpha", "Mango", "Zebra"}
	for i, w := range want {
		if rows[i][3] != w {
			t.Fatalf("desc row %d task = %q, want %q", i, rows[i][3], w)
		}
	}
	if title := m.sessTable.Columns()[2].Title; !strings.Contains(title, "▼") {
		t.Fatalf("active column header = %q, want a desc marker", title)
	}

	// Switching column resets to ascending on that new column.
	m = driveKey(m, "4")
	if m.statsSortCol != 3 || !m.statsSortAsc {
		t.Fatalf("after 4: col=%d asc=%v, want col=3 asc=true", m.statsSortCol, m.statsSortAsc)
	}
}

// (b) filter narrows rows case-insensitively; esc cancels and clears it.
func TestStatsFilterNarrowsCaseInsensitiveAndEscClears(t *testing.T) {
	now := time.Now()
	sessions := []models.TimerSession{
		sortFilterSession(now.Add(-1*time.Hour), 25*time.Minute, models.FocusSession, ""),
		sortFilterSession(now.Add(-2*time.Hour), 5*time.Minute, models.ShortBreak, "t2"),
		sortFilterSession(now.Add(-3*time.Hour), 25*time.Minute, models.FocusSession, "t1"),
	}
	tasks := []*models.Task{
		{ID: "t1", Title: "Write report"},
		{ID: "t2", Title: "Review code"},
	}
	m := statsModel(sessions, tasks...)
	before := len(m.sessTable.Rows())

	m = driveKey(m, "/")
	if !m.statsFiltering {
		t.Fatal("/ must enter filter-entry mode")
	}
	for _, c := range []string{"W", "R", "I", "T", "E"} {
		m = driveKey(m, c)
	}
	m = driveKey(m, "enter")
	if m.statsFiltering {
		t.Fatal("enter must leave filter-entry mode")
	}
	if m.statsFilter != "WRITE" {
		t.Fatalf("applied filter = %q, want WRITE", m.statsFilter)
	}
	rows := m.sessTable.Rows()
	if len(rows) != 1 || rows[0][3] != "Write report" {
		t.Fatalf("filtered rows = %v, want only Write report", rows)
	}

	// Entering the filter again then esc clears it entirely.
	m = driveKey(m, "/")
	m = driveKey(m, "esc")
	if m.statsFiltering || m.statsFilter != "" {
		t.Fatalf("after esc: filtering=%v filter=%q, want false and empty", m.statsFiltering, m.statsFilter)
	}
	if got := len(m.sessTable.Rows()); got != before {
		t.Fatalf("rows after esc = %d, want %d", got, before)
	}

	// Filtering also matches TYPE (case-insensitively).
	m = driveKey(m, "/")
	for _, c := range []string{"b", "r", "e", "a", "k"} {
		m = driveKey(m, c)
	}
	m = driveKey(m, "enter")
	rows = m.sessTable.Rows()
	if len(rows) != 1 || rows[0][3] != "Review code" {
		t.Fatalf("type-filter rows = %v, want only Review code (Short Break)", rows)
	}
}

// (c) sort and filter apply to the full slice before the 50-row render cap.
func TestStatsSortAndFilterApplyBeforeCap(t *testing.T) {
	now := time.Now()
	const total = 55
	sessions := make([]models.TimerSession, 0, total)
	tasks := make([]*models.Task, 0, total)
	for i := 0; i < total; i++ {
		at := now.Add(-time.Duration(i) * time.Hour) // newest first in the slice
		id := "t" + itoa(i)
		elapsed := time.Duration(total-i) * time.Minute // longest first, shortest last
		title := "Task " + itoa(i)
		if i == total-1 {
			title = "SHORTEST"
		}
		sessions = append(sessions, sortFilterSession(at, elapsed, models.FocusSession, id))
		tasks = append(tasks, &models.Task{ID: id, Title: title})
	}
	m := statsModel(sessions, tasks...)

	// The render cap is 50 rows.
	m = driveKey(m, "3") // LENGTH ascending
	rows := m.sessTable.Rows()
	if len(rows) != 50 {
		t.Fatalf("rendered rows = %d, want the 50-row cap", len(rows))
	}
	// The shortest session was last in the original slice (index 54), so it
	// would be dropped by a cap-first implementation. Sort-before-cap must
	// bring it to the top.
	if rows[0][3] != "SHORTEST" {
		t.Fatalf("first row after LENGTH asc = %q, want SHORTEST (index 54)", rows[0][3])
	}

	// A filter matching only a session past the cap must still surface it.
	m = driveKey(m, "/")
	for _, c := range []string{"S", "H", "O", "R", "T", "E", "S", "T"} {
		m = driveKey(m, c)
	}
	m = driveKey(m, "enter")
	rows = m.sessTable.Rows()
	if len(rows) != 1 || rows[0][3] != "SHORTEST" {
		t.Fatalf("filtered rows = %v, want only SHORTEST (index 54)", rows)
	}
}
