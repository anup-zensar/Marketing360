package projection

import (
	"time"

	"mapp/internal/platform/date"
)

// window is the part of the plan year the activity covers. An activity without a start runs from the
// year's first day; one that is ongoing or has no end runs to its last (FR-006, 020 FR-049).
func window(in Inputs) (start, end date.Date, covered bool) {
	start, end = date.YearStart(in.PlanYear), date.YearEnd(in.PlanYear)
	if in.StartDate != nil && in.StartDate.After(start) {
		start = *in.StartDate
	}
	if !in.Ongoing && in.EndDate != nil && in.EndDate.Before(end) {
		end = *in.EndDate
	}
	return start, end, !start.After(end)
}

func daysBetween(start, end date.Date) int64 {
	const hoursPerDay = 24
	return int64(end.Time().Sub(start.Time()).Hours()/hoursPerDay) + 1
}

// phase spreads goal evenly across the days of the window and collapses them to months. The running total
// gives every month its exact share: the months add up to the goal with no remainder lost (FR-006).
func phase(goal int64, start, end date.Date) []MonthFigure {
	months := emptyMonths(start.Year())
	totalDays := daysBetween(start, end)
	if goal <= 0 || totalDays <= 0 {
		return months
	}
	var elapsed, placed int64
	for i := range months {
		elapsed += daysOfMonth(start.Year(), time.Month(i+1), start, end)
		share := goal * elapsed / totalDays
		months[i].Value = share - placed
		placed = share
	}
	return months
}

func emptyMonths(year int) []MonthFigure {
	months := make([]MonthFigure, Months)
	for i := range months {
		months[i].Month = monthKey(year, time.Month(i+1))
	}
	return months
}

// daysOfMonth counts the days of the month that fall inside the window.
func daysOfMonth(year int, month time.Month, start, end date.Date) int64 {
	first := date.New(year, month, 1)
	last := first.AddDays(daysIn(year, month) - 1)
	if first.Before(start) {
		first = start
	}
	if last.After(end) {
		last = end
	}
	if first.After(last) {
		return 0
	}
	return daysBetween(first, last)
}

func daysIn(year int, month time.Month) int {
	return date.New(year, month, 1).Time().AddDate(0, 1, -1).Day()
}
