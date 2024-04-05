package core

import "time"

// scheduleWeekdayTask schedules a task to run on a specific weekday and time.
func (e *Engine)ScheduleWeekdayTask(weekday time.Weekday, hour, min int, task func()) {
	go func() {
		for {
			now := time.Now()
			// Calculate next occurrence of the given weekday and time
			next := nextWeekdayTime(weekday, hour, min)

			// Wait until next scheduled time
			time.Sleep(next.Sub(now))
			task() // Execute the task

			// Wait for one day before recalculating the next execution time
			// This prevents the task from executing multiple times in case it takes less than a day to execute
			time.Sleep(24 * time.Hour)
		}
	}()
}

// nextWeekdayTime calculates the next occurrence of a specific weekday and time.
func nextWeekdayTime(weekday time.Weekday, hour, min int) time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())

	for next.Weekday() != weekday || next.Before(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
