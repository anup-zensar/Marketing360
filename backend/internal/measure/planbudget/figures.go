package planbudget

import (
	"cmp"
	"slices"
	"strconv"

	"mapp/internal/allwork"
	"mapp/internal/financials"
	"mapp/internal/measure"
	"mapp/internal/platform/money"
	"mapp/internal/rollup"
)

// records indexes one year's records in view with what the report adds up (data-model).
type records struct {
	year       int
	byID       map[string]allwork.Node
	below      map[string][]string
	inView     map[string]bool
	matched    allwork.Matched
	financed   map[string]bool
	spend      map[string]map[string]financials.Facts
	planned    map[string]allwork.Months
	display    money.Display
	attributes map[string]allwork.Attribute
	memo       map[string]sums
}

type sums struct {
	budget                     *money.Amount
	planned, committed, actual money.Amount
	financed                   bool
	finBudget                  *money.Amount // the budget and planned cost of plans whose finance the person sees
	finPlanned                 money.Amount
}

func newRecords(sc measure.Scope) *records {
	r := &records{year: sc.Year, byID: map[string]allwork.Node{}, below: map[string][]string{}, inView: map[string]bool{},
		matched: sc.Matched, financed: map[string]bool{}, attributes: map[string]allwork.Attribute{}, memo: map[string]sums{}}
	for _, a := range sc.Attributes {
		r.attributes[a.ID] = a
	}
	nodes := slices.Clone(sc.Nodes)
	slices.SortStableFunc(nodes, func(x, y allwork.Node) int { return cmp.Compare(x.Order, y.Order) })
	for _, n := range nodes {
		r.byID[n.ID] = n
		if n.ParentID != "" {
			r.below[n.ParentID] = append(r.below[n.ParentID], n.ID)
		}
	}
	m := sc.Matched
	for _, ids := range [][]string{m.Plans, m.Campaigns, m.Initiatives, m.Activities, m.LineItems} {
		for _, id := range ids {
			r.inView[id] = true
		}
	}
	for _, id := range m.Plans {
		r.financed[id] = sc.Choices.SeesFinance && (sc.Choices.FinancePlans == nil || slices.Contains(sc.Choices.FinancePlans, id))
	}
	return r
}

// financedLineItems are the line items in view whose financial data the person may see.
func (r *records) financedLineItems() []string {
	var ids []string
	for _, id := range r.matched.LineItems {
		if r.financed[r.byID[id].PlanID] {
			ids = append(ids, id)
		}
	}
	return ids
}

func (r *records) childrenInView(id string) []string {
	return slices.DeleteFunc(slices.Clone(r.below[id]), func(child string) bool { return !r.inView[child] })
}

// sumsOf adds up a record in view: its own budget, and the planned cost, committed and actual of what is in view below.
func (r *records) sumsOf(id string) sums {
	if s, done := r.memo[id]; done {
		return s
	}
	s := r.addUp(id)
	if r.financed[r.byID[id].PlanID] {
		s.finBudget, s.finPlanned = s.budget, s.planned
	}
	r.memo[id] = s
	return s
}

func (r *records) addUp(id string) sums {
	n := r.byID[id]
	var s sums
	if n.Level == allwork.LevelLineItem {
		s.planned = orZero(n.PlannedCost)
		if r.financed[n.PlanID] {
			s.committed, s.actual, _ = r.spendOf(id)
			s.financed = true
		}
		return s
	}
	if n.Level == allwork.LevelActivity {
		s.planned = orZero(n.PlannedCost)
	} else {
		s.budget = n.Budget
	}
	for _, child := range r.childrenInView(id) {
		c := r.sumsOf(child)
		if n.Level != allwork.LevelActivity {
			s.planned = s.planned.Add(c.planned)
		}
		s.committed, s.actual = s.committed.Add(c.committed), s.actual.Add(c.actual)
		s.financed = s.financed || c.financed
	}
	return s
}

// spendOf is a line item's committed and actual posted in the report's year, and its actual of other years.
func (r *records) spendOf(lineItemID string) (committed, actual, otherYears money.Amount) {
	prefix := strconv.Itoa(r.year) + "-"
	for month, f := range r.spend[lineItemID] {
		if len(month) >= len(prefix) && month[:len(prefix)] == prefix {
			committed, actual = committed.Add(f.Committed), actual.Add(f.Paid)
		} else {
			otherYears = otherYears.Add(f.Paid)
		}
	}
	return committed, actual, otherYears
}

// counted are the containers whose budget counts, once per branch (040-C2, research R3).
func (r *records) counted() []allwork.Node {
	var out []allwork.Node
	for _, id := range r.matched.Plans {
		out = append(out, r.budgetBranch(id)...)
	}
	return out
}

func (r *records) budgetBranch(id string) []allwork.Node {
	n := r.byID[id]
	if n.Level == allwork.LevelInitiative || (n.Budget != nil && r.whole(id)) {
		if n.Budget == nil {
			return nil
		}
		return []allwork.Node{n}
	}
	var out []allwork.Node
	for _, child := range r.containersBelow(id) {
		if r.inView[child] {
			out = append(out, r.budgetBranch(child)...)
		}
	}
	return out
}

// whole holds when a container and every container below it are in view.
func (r *records) whole(id string) bool {
	if !r.inView[id] {
		return false
	}
	return !slices.ContainsFunc(r.containersBelow(id), func(child string) bool { return !r.whole(child) })
}

func (r *records) containersBelow(id string) []string {
	return slices.DeleteFunc(slices.Clone(r.below[id]), func(child string) bool {
		level := r.byID[child].Level
		return level != allwork.LevelCampaign && level != allwork.LevelInitiative
	})
}

// figures converts sums once to the display currency, with 036's rules for left to plan and over budget.
func (r *records) figures(s sums) Figures {
	d := r.display
	planned := money.USD(s.planned)
	budget := usdOf(s.budget)
	f := Figures{Budget: d.Optional(budget), Planned: d.Optional(&planned),
		LeftToPlan: d.Optional(rollup.LeftToPlan(budget, &planned)), OverAllocated: rollup.OverBudget(&planned, budget)}
	if s.financed {
		committed, actual := money.USD(s.committed), money.USD(s.actual)
		f.Committed, f.Actual = d.Optional(&committed), d.Optional(&actual)
		if s.finBudget != nil {
			unspent := money.USD(s.finBudget.Sub(s.actual))
			f.Unspent = d.Optional(&unspent)
		}
	}
	return f
}

func addAmount(a, b *money.Amount) *money.Amount {
	if b == nil {
		return a
	}
	sum := orZero(a).Add(*b)
	return &sum
}

func usdOf(a *money.Amount) *money.Money {
	if a == nil {
		return nil
	}
	m := money.USD(*a)
	return &m
}

func orZero(a *money.Amount) money.Amount {
	if a == nil {
		return money.Amount{}
	}
	return *a
}
