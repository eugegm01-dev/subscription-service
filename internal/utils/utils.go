package utils

import "time"

// ParseMonthYear converts "MM-YYYY" to time.Time (first day of month)
func ParseMonthYear(s string) (time.Time, error) {
	return time.Parse("01-2006", s)
}

// LastDayOfMonth returns last day of the month for given time
func LastDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location())
}
