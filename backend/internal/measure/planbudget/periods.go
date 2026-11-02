package planbudget

import (
	"fmt"

	"mapp/internal/measure"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

const monthsInYear = 12

// months are the year's twelve months of budget, planned cost and actual in USD, with what falls outside them.
type months struct {
	budget, planned, actual                       [monthsInYear]money.Amount
	notPhased, noDates, otherYears, actualOutside money.Amount
	showsActual                                   bool
}

func (r *records) months() months {
	var m months
	r.budgetMonths(&m)
	r.plannedMonths(&m)
	r.actualMonths(&m)
	return m
}

func (r *records) budgetMonths(m *months) {
	for _, n := range r.counted() {
		if n.PhasedBudgetUSD == nil {
			m.notPhased = m.notPhased.Add(orZero(n.Budget))
		}
		for _, p := range n.PhasedBudgetUSD {
			if i := r.monthIndex(p.Month); i >= 0 {
				m.budget[i] = m.budget[i].Add(p.USD)
			}
		}
	}
}

func (r *records) plannedMonths(m *months) {
	for _, id := range r.matched.Activities {
		spread := r.planned[id]
		for month, amount := range spread.ByMonth {
			if i := r.monthIndex(month); i >= 0 {
				m.planned[i] = m.planned[i].Add(amount)
			}
		}
		m.noDates, m.otherYears = m.noDates.Add(spread.NoDates), m.otherYears.Add(spread.OtherYears)
	}
}

func (r *records) actualMonths(m *months) {
	for _, id := range r.financedLineItems() {
		m.showsActual = true
		for month, f := range r.spend[id] {
			if i := r.monthIndex(month); i >= 0 {
				m.actual[i] = m.actual[i].Add(f.Paid)
			} else {
				m.actualOutside = m.actualOutside.Add(f.Paid)
			}
		}
	}
}

// monthIndex is the month's place in the report's year, or -1 for a month of another year.
func (r *records) monthIndex(month string) int {
	var year, i int
	if _, err := fmt.Sscanf(month, "%d-%d", &year, &i); err != nil || year != r.year || i < 1 || i > monthsInYear {
		return -1
	}
	return i - 1
}

// periods groups the months by month or quarter; actual stops at today's period (FR-004, FR-007).
func (r *records) periods(period string, today date.Date) ([]Period, Notes) {
	m := r.months()
	size := 1
	if period == measure.PeriodQuarter {
		size = 3
	}
	out := make([]Period, 0, monthsInYear/size)
	var budgetSoFar, plannedSoFar, actualSoFar money.Amount
	for first := 0; first < monthsInYear; first += size {
		var budget, planned, actual money.Amount
		for i := first; i < first+size; i++ {
			budget, planned, actual = budget.Add(m.budget[i]), planned.Add(m.planned[i]), actual.Add(m.actual[i])
		}
		budgetSoFar, plannedSoFar, actualSoFar = budgetSoFar.Add(budget), plannedSoFar.Add(planned), actualSoFar.Add(actual)
		p := Period{Key: r.periodKey(first, size), State: r.stateOf(first, size, today),
			Budget: r.shown(budget), Planned: r.shown(planned), BudgetToDate: r.shown(budgetSoFar),
			PlannedToDate: r.shown(plannedSoFar)}
		if m.showsActual && p.State != StateFuture {
			a, toDate := r.shown(actual), r.shown(actualSoFar)
			p.Actual, p.ActualToDate = &a, &toDate
		}
		out = append(out, p)
	}
	return out, Notes{NotPhased: r.noted(m.notPhased), NoDates: r.noted(m.noDates), OtherYears: r.noted(m.otherYears),
		ActualOtherYears: r.noted(m.actualOutside)}
}

func (r *records) periodKey(first, size int) string {
	if size == 3 {
		return fmt.Sprintf("%04d-Q%d", r.year, first/3+1)
	}
	return fmt.Sprintf("%04d-%02d", r.year, first+1)
}

// stateOf is past, today's or future for the period starting at month index first (039 FR-019).
func (r *records) stateOf(first, size int, today date.Date) string {
	switch {
	case r.year < today.Year():
		return StatePast
	case r.year > today.Year():
		return StateFuture
	}
	current := int(today.Month()) - 1
	switch {
	case current >= first+size:
		return StatePast
	case current >= first:
		return StateInProgress
	}
	return StateFuture
}

func (r *records) shown(a money.Amount) money.Money {
	return r.display.Money(money.USD(a))
}

func (r *records) noted(a money.Amount) *money.Money {
	if a.IsZero() {
		return nil
	}
	shown := r.shown(a)
	return &shown
}
