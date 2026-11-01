package app

import (
	"context"
	"fmt"
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/financials"
	"mapp/internal/plan"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/clock"
	"mapp/internal/platform/date"
	"mapp/internal/platform/memo"
	"mapp/internal/platform/money"
	"mapp/internal/stores"
	"mapp/internal/widgetdata"
)

// widgetPlanning reads the plans in the person's scope with everything in them, as All work does (010), each
// year once per request, with the GSAP figures of the line items whose financial data the person sees only.
type widgetPlanning struct {
	years   dashboardPlanning
	plans   stores.PlanStore
	people  stores.PeopleStore
	spend   *financials.Service
	finance financials.Access
	now     clock.Clock
}

type planningData struct {
	nodes   []*allwork.Node
	byID    map[string]*allwork.Node
	years   map[int][]allwork.Node
	teams   map[string]string
	finance map[string]bool
	sums    map[string]*sums
	notes   []string
}

// readPlanning reads the years the date range overlaps when dated, else every year with plans in scope.
func (p widgetPlanning) read(ctx context.Context, a actor.Actor, f widgetdata.Prefilter, dated bool) (*planningData, error) {
	visible, err := p.plans.ListInScope(ctx, authz.ScopeOf(ctx))
	if err != nil {
		return nil, err
	}
	d := &planningData{byID: map[string]*allwork.Node{}, years: map[int][]allwork.Node{}, finance: map[string]bool{},
		sums: map[string]*sums{}}
	keep := keeper(f, visible)
	for _, year := range yearsOf(visible, f, dated) {
		nodes, readErr := p.year(ctx, year)
		if readErr != nil {
			return nil, readErr
		}
		d.years[year] = nodes
		d.add(nodes, keep)
	}
	if d.teams, err = p.teamNames(ctx); err != nil {
		return nil, err
	}
	return d, p.addFinance(ctx, a, d)
}

func (p widgetPlanning) year(ctx context.Context, year int) ([]allwork.Node, error) {
	return memo.Get(ctx, fmt.Sprintf("%s:widgetdata-year-%d", memo.Derived, year), func(ctx context.Context) ([]allwork.Node, error) {
		data, err := p.years.Year(ctx, year, date.Today(p.now()), nil)
		return data.Nodes, err
	})
}

func (p widgetPlanning) teamNames(ctx context.Context) (map[string]string, error) {
	teams, err := p.people.ListTeams(ctx)
	names := make(map[string]string, len(teams))
	for _, t := range teams {
		names[t.ID] = t.Name
	}
	return names, err
}

// addFinance counts the purchase orders and payments of the line items whose plan's financial data a sees.
func (p widgetPlanning) addFinance(ctx context.Context, a actor.Actor, d *planningData) error {
	sees := p.finance.SeeFinance(ctx, a) == nil
	financePlans, err := p.finance.VisiblePlanIDs(ctx, a)
	if err != nil {
		return err
	}
	var plans, counted int
	for _, n := range d.nodes {
		if n.Level == allwork.LevelPlan {
			d.finance[n.ID] = sees && (financePlans == nil || slices.Contains(financePlans, n.ID))
			plans, counted = plans+1, counted+boolCount(d.finance[n.ID])
		}
	}
	if counted > 0 && counted < plans {
		d.notes = append(d.notes, widgetdata.FinanceNote(counted, plans))
	}
	ids := d.financeLineItems()
	if len(ids) == 0 {
		return nil
	}
	spend, err := p.spend.SpendByMonth(ctx, a, ids)
	d.addSpend(spend)
	return err
}

func (d *planningData) add(nodes []allwork.Node, keep func(n *allwork.Node) bool) {
	kept := map[string]bool{}
	for i := range nodes {
		if n := &nodes[i]; n.Level == allwork.LevelPlan && keep(n) {
			kept[n.ID] = true
		}
	}
	var added []*allwork.Node
	for i := range nodes {
		if n := &nodes[i]; kept[n.PlanID] {
			added = append(added, n)
			d.byID[n.ID] = n
			d.sums[n.ID] = &sums{}
		}
	}
	d.nodes = append(d.nodes, added...)
	d.addPlanned(added)
}

// addPlanned gives every record above an activity the activity's planned cost.
func (d *planningData) addPlanned(nodes []*allwork.Node) {
	for _, n := range nodes {
		if n.Level == allwork.LevelActivity && n.PlannedCost != nil {
			d.upward(n.ParentID, func(s *sums) { s.planned = s.planned.Add(*n.PlannedCost) })
		}
	}
}

func (d *planningData) financeLineItems() []string {
	var ids []string
	for _, n := range d.nodes {
		if n.Level == allwork.LevelLineItem && d.finance[n.PlanID] {
			ids = append(ids, n.ID)
		}
	}
	return ids
}

func (d *planningData) addSpend(spend map[string]map[string]financials.Facts) {
	for lineItemID, months := range spend {
		var facts financials.Facts
		for _, f := range months {
			facts.Committed, facts.Paid = facts.Committed.Add(f.Committed), facts.Paid.Add(f.Paid)
		}
		d.upward(lineItemID, func(s *sums) {
			s.committed, s.actual = s.committed.Add(facts.Committed), s.actual.Add(facts.Paid)
		})
	}
}

// upward applies add to the record id and every record above it.
func (d *planningData) upward(id string, add func(s *sums)) {
	for n := d.byID[id]; n != nil; n = d.byID[n.ParentID] {
		add(d.sums[n.ID])
		if n.ParentID == "" && n.Level != allwork.LevelPlan {
			add(d.sums[n.PlanID])
		}
	}
}

// sums are a record's planned cost (its activities') and its line items' committed and actual, in USD.
type sums struct {
	planned, committed, actual money.Amount
}

func keeper(f widgetdata.Prefilter, visible []plan.Plan) func(n *allwork.Node) bool {
	var planIDs []string
	for _, id := range f.PlanIDs {
		if slices.ContainsFunc(visible, func(p plan.Plan) bool { return p.ID == id }) {
			planIDs = append(planIDs, id)
		}
	}
	return func(n *allwork.Node) bool {
		planOK := len(planIDs) == 0 || slices.Contains(planIDs, n.ID)
		teamOK := len(f.TeamIDs) == 0 || slices.ContainsFunc(n.TeamIDs, func(id string) bool { return slices.Contains(f.TeamIDs, id) })
		return planOK && teamOK
	}
}

func yearsOf(visible []plan.Plan, f widgetdata.Prefilter, dated bool) []int {
	var years []int
	for _, p := range visible {
		inRange := !dated || ((f.From == nil || p.Year >= f.From.Year()) && (f.To == nil || p.Year <= f.To.Year()))
		if inRange && !slices.Contains(years, p.Year) {
			years = append(years, p.Year)
		}
	}
	slices.Sort(years)
	return years
}

func boolCount(b bool) int {
	if b {
		return 1
	}
	return 0
}
