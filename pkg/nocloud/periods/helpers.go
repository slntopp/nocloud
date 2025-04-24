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
		cycleStartDay := cycleStart.Day()
		startDay := startTime.Day()

		year, month, _ := startTime.Date()
		nextMonth := month + 1
		nextYear := year
		if nextMonth > 12 {
			nextMonth = 1
			nextYear++
		}

		firstOfMonthAfterNext := time.Date(nextYear, nextMonth+1, 1, 0, 0, 0, 0, time.UTC)
		lastOfNextMonth := firstOfMonthAfterNext.Add(-time.Hour * 24)
		daysInNextMonth := lastOfNextMonth.Day()

		var targetDay int
		if cycleStartDay > 28 && startDay > 27 {
			if cycleStartDay <= daysInNextMonth {
				targetDay = cycleStartDay
			} else {
				targetDay = daysInNextMonth
			}
		} else {
			if startDay <= daysInNextMonth {
				targetDay = startDay
			} else {
				targetDay = daysInNextMonth
			}
		}

		nextDate := time.Date(nextYear, nextMonth, targetDay,
			startTime.Hour(), startTime.Minute(), startTime.Second(),
			startTime.Nanosecond(), time.UTC)

		return nextDate.Unix()
	}

	return start
}
