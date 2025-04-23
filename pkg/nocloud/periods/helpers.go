package periods

import (
	"time"
)

type PeriodKind int

const (
	BillingMonth PeriodKind = iota
)

func GetNextDate(start int64, period PeriodKind, cycleBeginning int64) int64 {
	startTime := time.Unix(start, 0).UTC()
	cycleStart := time.Unix(cycleBeginning, 0).UTC()

	switch period {
	case BillingMonth:
		targetDay := cycleStart.Day()
		year, month, _ := startTime.Date()
		loc := startTime.Location()
		month++
		if month > 12 {
			month = 1
			year++
		}
		_ = time.Date(year, month, 1, startTime.Hour(), startTime.Minute(), startTime.Second(), startTime.Nanosecond(), loc)
		nextMonth := month + 1
		nextYear := year
		if nextMonth > 12 {
			nextMonth = 1
			nextYear++
		}
		firstOfFollowing := time.Date(nextYear, nextMonth, 1, 0, 0, 0, 0, loc)
		daysInNext := int(firstOfFollowing.Sub(time.Date(year, month, 1, 0, 0, 0, 0, loc)).Hours() / 24)
		day := targetDay
		if day > daysInNext {
			day = daysInNext
		}
		result := time.Date(year, month, day, startTime.Hour(), startTime.Minute(), startTime.Second(), startTime.Nanosecond(), loc)
		return result.Unix()
	default:
		return start
	}
}
