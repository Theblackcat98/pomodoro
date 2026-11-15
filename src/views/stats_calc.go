package views

import (
	"sort"
	"time"

	"pomodoro/src/models"
)

// Pure statistics over the session history. Everything here is a function of
// (sessions, now) so it is trivially testable and the view layer stays thin.

// streakHour is the "day boundary" for streaks: a focus session completed
// before 09:00 counts for the previous calendar day (night-owl forgiveness).
const streakHour = 9

// sessionElapsed returns the time actually spent in a session (pause-aware).
// Sessions recorded before the Elapsed field existed fall back to Duration.
func sessionElapsed(s models.TimerSession) time.Duration {
	if s.Elapsed > 0 {
		return s.Elapsed
	}
	return s.Duration
}

// isCompletedFocus reports whether a session counts for stats: focus type,
// completed, with a completion timestamp.
func isCompletedFocus(s models.TimerSession) bool {
	return s.Type == models.FocusSession && s.CompletedAt != nil
}

// streakBucket maps a completion time to its streak "day": completions before
// 09:00 belong to the previous day.
func streakBucket(t time.Time) time.Time {
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	if t.Hour() < streakHour {
		day = day.AddDate(0, 0, -1)
	}
	return day
}

// dayBucket maps a completion time to its plain calendar day (no shift) —
// used for the heatmap and weekly chart, where the 9am rule would be
// confusing.
func dayBucket(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// focusCountsByDay buckets completed focus sessions by calendar day.
func focusCountsByDay(sessions []models.TimerSession) map[time.Time]int {
	counts := make(map[time.Time]int)
	for _, s := range sessions {
		if isCompletedFocus(s) {
			counts[dayBucket(*s.CompletedAt)]++
		}
	}
	return counts
}

// todayTotals returns today's completed pomodoro count and the time actually
// spent focusing today (pause-aware elapsed, not configured duration).
func todayTotals(sessions []models.TimerSession, now time.Time) (pomodoros int, focusTime time.Duration) {
	today := dayBucket(now)
	for _, s := range sessions {
		if isCompletedFocus(s) && dayBucket(*s.CompletedAt).Equal(today) {
			pomodoros++
			focusTime += sessionElapsed(s)
		}
	}
	return pomodoros, focusTime
}

// computeStreaks returns the current and best daily streaks of completed
// focus sessions. A day counts when it has at least one completed focus
// session; completions before 09:00 credit the previous day, so yesterday's
// bucket stays open until 09:00 today. The current streak stays alive while
// today's bucket is still empty (the day isn't over).
func computeStreaks(sessions []models.TimerSession, now time.Time) (current, best int) {
	buckets := make(map[time.Time]int)
	for _, s := range sessions {
		if isCompletedFocus(s) {
			buckets[streakBucket(*s.CompletedAt)]++
		}
	}
	if len(buckets) == 0 {
		return 0, 0
	}

	// Best: longest consecutive run anywhere in the history.
	days := make([]time.Time, 0, len(buckets))
	for d := range buckets {
		days = append(days, d)
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })

	run := 1
	best = 1
	for i := 1; i < len(days); i++ {
		if days[i].Equal(days[i-1].AddDate(0, 0, 1)) {
			run++
		} else {
			run = 1
		}
		if run > best {
			best = run
		}
	}

	// Current: walk back from today; an empty today doesn't break the streak
	// (the day isn't over), an empty earlier day does.
	today := streakBucket(now)
	start := today
	if buckets[start] == 0 {
		start = today.AddDate(0, 0, -1)
	}
	if buckets[start] == 0 {
		return 0, best
	}
	for d := start; buckets[d] > 0; d = d.AddDate(0, 0, -1) {
		current++
	}
	return current, best
}

// heatCell is one day in the activity heatmap.
type heatCell struct {
	Day   time.Time
	Count int
}

// heatmapWeeks builds the GitHub-style grid: weeks columns × 7 weekday rows
// (Monday first), ending with the current week. Cells outside history or in
// the future are zero.
func heatmapWeeks(sessions []models.TimerSession, now time.Time, weeks int) [][]heatCell {
	counts := focusCountsByDay(sessions)

	// Monday of the current week.
	today := dayBucket(now)
	weekday := int(today.Weekday()) // Sunday = 0
	monday := today.AddDate(0, 0, -((weekday + 6) % 7))

	grid := make([][]heatCell, weeks)
	for w := 0; w < weeks; w++ {
		col := make([]heatCell, 7)
		for d := 0; d < 7; d++ {
			day := monday.AddDate(0, 0, -(weeks-1-w)*7+d)
			col[d] = heatCell{Day: day, Count: counts[day]}
		}
		grid[w] = col
	}
	return grid
}

// heatLevel maps a day's count to an intensity level 0..4 relative to the
// grid's busiest day.
func heatLevel(count, maxCount int) int {
	if count == 0 || maxCount == 0 {
		return 0
	}
	level := (count * 4) / maxCount
	if level < 1 {
		level = 1
	}
	if level > 4 {
		level = 4
	}
	return level
}

// taskStat aggregates focus sessions per task.
type taskStat struct {
	TaskID    string
	Title     string
	Pomodoros int
	FocusTime time.Duration
}

// perTaskStats groups completed focus sessions by task (unattributed sessions
// land under "—") and returns them sorted by focus time, busiest first.
func perTaskStats(sessions []models.TimerSession, tasks []*models.Task, limit int) []taskStat {
	titles := make(map[string]string, len(tasks))
	for _, t := range tasks {
		titles[t.ID] = t.Title
	}

	byID := make(map[string]*taskStat)
	order := make([]string, 0)
	for _, s := range sessions {
		if !isCompletedFocus(s) {
			continue
		}
		id := s.TaskID
		if id == "" {
			id = "—"
		}
		stat, ok := byID[id]
		if !ok {
			title := "Unattributed"
			if id != "—" {
				title = titles[id]
			}
			if title == "" {
				title = "(deleted task)"
			}
			stat = &taskStat{TaskID: id, Title: title}
			byID[id] = stat
			order = append(order, id)
		}
		stat.Pomodoros++
		stat.FocusTime += sessionElapsed(s)
	}

	stats := make([]taskStat, 0, len(byID))
	for _, id := range order {
		stats = append(stats, *byID[id])
	}
	sort.SliceStable(stats, func(i, j int) bool {
		if stats[i].FocusTime != stats[j].FocusTime {
			return stats[i].FocusTime > stats[j].FocusTime
		}
		return stats[i].Pomodoros > stats[j].Pomodoros
	})
	if limit > 0 && len(stats) > limit {
		stats = stats[:limit]
	}
	return stats
}
