package planbudget

import (
	"mapp/internal/platform/money"
	"mapp/internal/rollup"
)

// total adds up the plans in view, with the budget counted once per branch (FR-002, FR-003).
func (r *records) total() sums {
	var t sums
	for _, id := range r.matched.Plans {
		s := r.sumsOf(id)
		t.planned, t.committed, t.actual = t.planned.Add(s.planned), t.committed.Add(s.committed), t.actual.Add(s.actual)
		t.financed = t.financed || s.financed
	}
	for _, n := range r.counted() {
		budget := orZero(n.Budget)
		if t.budget != nil {
			budget = budget.Add(*t.budget)
		}
		t.budget = &budget
	}
	return t
}

// cards are measure-metrics §1's figures; the per cents use 036's whole-per-cent rule and "—" on a zero base.
func (r *records) cards() Cards {
	t := r.total()
	c := Cards{Figures: r.figures(t)}
	planned, budget := money.USD(t.planned), usdOf(t.budget)
	c.BudgetUsed = rollup.BudgetUsedPercent(&planned, budget)
	if t.financed {
		actual, committed := money.USD(t.actual), money.USD(t.committed)
		c.BudgetSpent = rollup.BudgetUsedPercent(&actual, budget)
		c.CommittedShare = rollup.BudgetUsedPercent(&committed, &planned)
		notYet := money.USD(t.planned.Sub(t.committed))
		c.NotYetCommitted = r.display.Optional(&notYet)
	}
	return c
}
