package helpers

import "time"

func MonthsBetween(date time.Time, filterYear int, filterMonth int) int {
	yearDiff := filterYear - date.Year()
	monthDiff := filterMonth - int(date.Month())
	totalMonths := yearDiff*12 + monthDiff
	if totalMonths < 0 {
		return 0
	}
	return totalMonths
}
