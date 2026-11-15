package views

import (
	"testing"
	"time"

	"pomodoro/src/models"
)

// focusAt builds a completed focus session at the given time.
func focusAt(at time.Time, elapsed time.Duration) models.TimerSession {
	done := at
	return models.TimerSession{
		ID:          models.NewID(),
		Type:        models.FocusSession,
		Duration:    25 * time.Minute,
		Remaining:   0,
		Elapsed:     elapsed,
		Status:      models.Completed,
		CompletedAt: &done,
	}
}

func TestSessionElapsedPrefersRecordedElapsed(t *testing.T) {
	s := focusAt(time.Now(), 12*time.Minute)
	if got := sessionElapsed(s); got != 12*time.Minute {
		t.Fatalf("elapsed = %v, want 12m", got)
	}
	// Legacy session without Elapsed falls back to Duration.
	s.Elapsed = 0
	if got := sessionElapsed(s); got != 25*time.Minute {
		t.Fatalf("legacy elapsed = %v, want 25m (duration)", got)
	}
}

func TestTodayTotalsUsesElapsedNotDuration(t *testing.T) {
	now := time.Date(2026, 11, 14, 15, 0, 0, 0, time.Local)
	sessions := []models.TimerSession{
		focusAt(now.Add(-2*time.Hour), 20*time.Minute),                         // today, partial
		focusAt(now.Add(-2*time.Hour), 25*time.Minute),                         // today, full
		focusAt(now.AddDate(0, 0, -1), 25*time.Minute),                         // yesterday
		{Type: models.ShortBreak, Status: models.Completed, CompletedAt: &now}, // not focus
	}

	pomodoros, focusTime := todayTotals(sessions, now)
	if pomodoros != 2 {
		t.Fatalf("pomodoros = %d, want 2", pomodoros)
	}
	if focusTime != 45*time.Minute {
		t.Fatalf("focus time = %v, want 45m (20+25, not 25+25)", focusTime)
	}
}

func TestStreaksForgiveBeforeNine(t *testing.T) {
	// Sessions: Mon 10:00, Tue 08:30 (credits MONDAY), Tue 20:00, Wed 10:00.
	mon := time.Date(2026, 11, 9, 10, 0, 0, 0, time.Local)
	tueEarly := time.Date(2026, 11, 10, 8, 30, 0, 0, time.Local)
	tueEvening := time.Date(2026, 11, 10, 20, 0, 0, 0, time.Local)
	wed := time.Date(2026, 11, 11, 10, 0, 0, 0, time.Local)
	sessions := []models.TimerSession{
		focusAt(mon, 25*time.Minute),
		focusAt(tueEarly, 25*time.Minute),   // night owl: Tuesday-before-9am
		focusAt(tueEvening, 25*time.Minute), // real Tuesday session
		focusAt(wed, 25*time.Minute),
	}

	// Wednesday afternoon: Mon, Tue, Wed buckets all non-empty → 3.
	now := time.Date(2026, 11, 11, 15, 0, 0, 0, time.Local)
	current, best := computeStreaks(sessions, now)
	if current != 3 || best != 3 {
		t.Fatalf("current=%d best=%d, want 3/3", current, best)
	}

	// Thursday 08:00, nothing new yet: the current bucket is still
	// WEDNESDAY (before 9am), so the streak is alive without a Thursday
	// session.
	now2 := time.Date(2026, 11, 12, 8, 0, 0, 0, time.Local)
	current, _ = computeStreaks(sessions, now2)
	if current != 3 {
		t.Fatalf("current at Thu 08:00 = %d, want 3 (Wednesday bucket stays open until 9am)", current)
	}

	// Thursday 10:00 with no Thursday session yet: the 3-day run is at risk
	// but still shown as current (the day isn't over).
	now3 := time.Date(2026, 11, 12, 10, 0, 0, 0, time.Local)
	current, _ = computeStreaks(sessions, now3)
	if current != 3 {
		t.Fatalf("current at Thu 10:00 = %d, want 3 (alive until the day ends)", current)
	}

	// Friday 10:00: Thursday passed empty — the streak is genuinely broken.
	now4 := time.Date(2026, 11, 13, 10, 0, 0, 0, time.Local)
	current, _ = computeStreaks(sessions, now4)
	if current != 0 {
		t.Fatalf("current at Fri 10:00 = %d, want 0 (Thursday was empty)", current)
	}
}

func TestStreakSurvivesEmptyToday(t *testing.T) {
	yesterday := time.Date(2026, 11, 13, 10, 0, 0, 0, time.Local)
	sessions := []models.TimerSession{focusAt(yesterday, 25*time.Minute)}

	// Today (Nov 14) has nothing yet: the streak is alive, not broken.
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.Local)
	current, best := computeStreaks(sessions, now)
	if current != 1 || best != 1 {
		t.Fatalf("current=%d best=%d, want 1/1 (today may still happen)", current, best)
	}

	// Two empty days break it.
	now3 := time.Date(2026, 11, 16, 12, 0, 0, 0, time.Local)
	current, _ = computeStreaks(sessions, now3)
	if current != 0 {
		t.Fatalf("current=%d, want 0 after two empty days", current)
	}
}

func TestStreakBestRunInHistory(t *testing.T) {
	// 5-day run last month, nothing since.
	sessions := make([]models.TimerSession, 5)
	for i := range sessions {
		sessions[i] = focusAt(time.Date(2026, 10, 5+i, 10, 0, 0, 0, time.Local), 25*time.Minute)
	}
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.Local)
	current, best := computeStreaks(sessions, now)
	if current != 0 {
		t.Fatalf("current = %d, want 0", current)
	}
	if best != 5 {
		t.Fatalf("best = %d, want 5", best)
	}
}

func TestHeatmapWeeksGridShape(t *testing.T) {
	// Wednesday, Nov 12 2026? Use a known Monday: 2026-11-09 is a Monday.
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.Local) // Thursday
	sessions := []models.TimerSession{
		focusAt(time.Date(2026, 11, 12, 9, 30, 0, 0, time.Local), 25*time.Minute),
		focusAt(time.Date(2026, 11, 12, 11, 0, 0, 0, time.Local), 25*time.Minute),
	}

	grid := heatmapWeeks(sessions, now, 13)
	if len(grid) != 13 {
		t.Fatalf("weeks = %d, want 13", len(grid))
	}
	last := grid[12]
	if len(last) != 7 {
		t.Fatalf("weekdays = %d, want 7", len(last))
	}
	// The final column is the current week; its Thursday (index 3, Mon-first)
	// carries today's two sessions.
	if last[3].Count != 2 {
		t.Fatalf("thursday count = %d, want 2", last[3].Count)
	}
	// First column starts 12 weeks before the current Monday (Nov 9 2026
	// minus 84 days = Aug 17 2026, itself a Monday).
	wantStart := time.Date(2026, 8, 17, 0, 0, 0, 0, time.Local)
	if !grid[0][0].Day.Equal(wantStart) {
		t.Fatalf("grid start = %v, want %v", grid[0][0].Day, wantStart)
	}
}

func TestHeatLevelScaling(t *testing.T) {
	if heatLevel(0, 8) != 0 {
		t.Fatal("zero count must be level 0")
	}
	if heatLevel(1, 8) != 1 {
		t.Fatalf("1/8 = %d, want 1", heatLevel(1, 8))
	}
	if heatLevel(4, 8) != 2 {
		t.Fatalf("4/8 = %d, want 2", heatLevel(4, 8))
	}
	if heatLevel(8, 8) != 4 {
		t.Fatalf("8/8 = %d, want 4", heatLevel(8, 8))
	}
	if heatLevel(99, 8) != 4 {
		t.Fatalf("overflow = %d, want 4", heatLevel(99, 8))
	}
}

func TestPerTaskStatsOrdersByFocusTime(t *testing.T) {
	now := time.Now()
	tasks := []*models.Task{
		{ID: "a", Title: "Alpha"},
		{ID: "b", Title: "Beta"},
	}
	sessions := []models.TimerSession{
		{Type: models.FocusSession, Status: models.Completed, CompletedAt: &now, TaskID: "b", Elapsed: 10 * time.Minute},
		{Type: models.FocusSession, Status: models.Completed, CompletedAt: &now, TaskID: "b", Elapsed: 10 * time.Minute},
		{Type: models.FocusSession, Status: models.Completed, CompletedAt: &now, TaskID: "a", Elapsed: 25 * time.Minute},
		{Type: models.FocusSession, Status: models.Completed, CompletedAt: &now, TaskID: ""},   // unattributed
		{Type: models.FocusSession, Status: models.Completed, CompletedAt: &now, TaskID: "zz"}, // deleted task
		{Type: models.ShortBreak, Status: models.Completed, CompletedAt: &now, TaskID: "a"},    // not focus
	}

	stats := perTaskStats(sessions, tasks, 0)
	if len(stats) != 4 {
		t.Fatalf("stats = %d rows, want 4", len(stats))
	}
	// Beta 20m > Alpha 25m? No: Alpha 25m is the most.
	if stats[0].Title != "Alpha" || stats[0].FocusTime != 25*time.Minute {
		t.Fatalf("top = %+v, want Alpha 25m", stats[0])
	}
	if stats[1].Title != "Beta" || stats[1].Pomodoros != 2 {
		t.Fatalf("second = %+v, want Beta ×2", stats[1])
	}
	if stats[2].Title != "Unattributed" || stats[3].Title != "(deleted task)" {
		t.Fatalf("tail = %+v", stats[2:])
	}

	// Limit truncates.
	if got := perTaskStats(sessions, tasks, 2); len(got) != 2 {
		t.Fatalf("limit 2 returned %d rows", len(got))
	}
}
