package planbudget

import (
	"mapp/internal/platform/money"
	"mapp/internal/rollup"
)

// total adds up the plans in view, with the budget counted once per branch (FR-002, FR-003).
func (r *records) total() sums {
	var t sums
	for _, id := range r.matched.Plans {
		p := r.sumsOf(id)
		p.budget = nil
		for _, n := range r.budgetBranch(id) {
			p.budget = addAmount(p.budget, n.Budget)
		}
		if r.financed[id] {
			p.finBudget = p.budget
		}
		t = t.plus(p)
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
		finPlanned := money.USD(t.finPlanned)
		c.BudgetSpent = rollup.BudgetUsedPercent(&actual, usdOf(t.finBudget))
		c.CommittedShare = rollup.BudgetUsedPercent(&committed, &finPlanned)
		notYet := money.USD(t.finPlanned.Sub(t.committed))
		c.NotYetCommitted = r.display.Optional(&notYet)
	}
	return c
}
