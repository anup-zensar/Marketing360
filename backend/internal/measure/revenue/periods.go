package revenue

import (
	"fmt"

	"mapp/internal/measure"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
)

// window is today and the plan year together, since every month/period calculation needs both.
type window struct {
	today date.Date
	year  int
}

type revenueMonthlySums struct {
	actual, projected, goal [projection.Months]money.Amount
}

// revenueMonths sums every activity's revenue into one twelve-month set, in USD (046-C7).
func revenueMonths(views []activityView, w window) revenueMonthlySums {
	var sums revenueMonthlySums
	for _, v := range views {
		if !v.projection.Present {
			continue
		}
		split := projection.SplitAmounts(v.facts.revenue, v.projection.Revenue, w.today, w.year)
		for i, m := range split.Months {
			if m.Actual != nil {
				sums.actual[i] = sums.actual[i].Add(m.Actual.Amount)
			}
			if m.Projected != nil {
				sums.projected[i] = sums.projected[i].Add(m.Projected.Amount)
			}
		}
		for i, m := range v.projection.Revenue {
			if i < len(sums.goal) {
				sums.goal[i] = sums.goal[i].Add(m.Value.Amount)
			}
		}
	}
	return sums
}

// periodBounds is one period's position in the year: its first month, its size, and where today falls.
type periodBounds struct {
	first, size, todayIndex, year int
}

func (b periodBounds) inProgress() bool {
	return b.todayIndex >= b.first && b.todayIndex < b.first+b.size
}

func periodsOf(w window, period string) []periodBounds {
	size := periodSize(period)
	todayIndex := todayMonthIndex(w.today, w.year)
	out := make([]periodBounds, 0, projection.Months/size)
	for first := 0; first < projection.Months; first += size {
		out = append(out, periodBounds{first: first, size: size, todayIndex: todayIndex, year: w.year})
	}
	return out
}

// revenuePeriods groups the twelve months by month or quarter, split at today (047 FR-010, 049-C4).
func revenuePeriods(sums revenueMonthlySums, period string, w window, display money.Display) []RevenuePeriod {
	bounds := periodsOf(w, period)
	out := make([]RevenuePeriod, 0, len(bounds))
	for _, b := range bounds {
		out = append(out, revenuePeriodOf(sums, b, display))
	}
	return out
}

func revenuePeriodOf(sums revenueMonthlySums, b periodBounds, display money.Display) RevenuePeriod {
	var a, p, g money.Amount
	hasActual, hasProjected := false, false
	for i := b.first; i < b.first+b.size; i++ {
		a, p, g = a.Add(sums.actual[i]), p.Add(sums.projected[i]), g.Add(sums.goal[i])
		if b.todayIndex >= 0 && i <= b.todayIndex {
			hasActual = true
		}
		if b.todayIndex < 0 || i >= b.todayIndex {
			hasProjected = true
		}
	}
	period := RevenuePeriod{Key: periodKey(b.year, b.first, b.size), InProgress: b.inProgress(),
		State: periodState(b.first, b.size, b.todayIndex)}
	if g.Cents() != 0 {
		goalShown := display.Money(money.USD(g))
		period.Goal = &goalShown
	}
	if hasActual {
		shown := display.Money(money.USD(a))
		period.Actual = &shown
	}
	if hasProjected {
		shown := display.Money(money.USD(p))
		period.Projected = &shown
	}
	return period
}

func periodSize(period string) int {
	if period == measure.PeriodQuarter {
		return 3
	}
	return 1
}

func periodKey(year, first, size int) string {
	if size == 3 {
		return fmt.Sprintf("%04d-Q%d", year, first/3+1)
	}
	return fmt.Sprintf("%04d-%02d", year, first+1)
}

func periodState(first, size, todayIndex int) string {
	switch {
	case todayIndex < 0:
		return StateFuture
	case todayIndex >= first+size:
		return StatePast
	case todayIndex >= first:
		return StateInProgress
	}
	return StateFuture
}

// todayMonthIndex is today's place in the plan year, or -1 before it and projection.Months after it.
func todayMonthIndex(today date.Date, year int) int {
	if today.Year() != year {
		if today.Year() < year {
			return -1
		}
		return projection.Months
	}
	return int(today.Month()) - 1
}
