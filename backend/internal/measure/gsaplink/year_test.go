package gsaplink

import (
	"context"
	"slices"
	"testing"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/money"
)

const (
	testYear    = 2026
	companyAttr = "SA-1"
	companyGB01 = "SO-1"
	ukPlan      = "P-1"
	dePlan      = "P-2"
	ukActivity  = "A-1"
	deActivity  = "A-2"
)

// now is Wednesday 2026-03-18, 12:00 UTC.
var now = time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

// world is an in-memory year in the reports.view scope and the fakes of the ports: every node matches unless
// left out by unmatch.
type world struct {
	tb           testing.TB
	nodes        []allwork.Node
	unmatched    map[string]bool
	choices      allwork.Choices
	currency     allwork.Currency
	approved     map[string]time.Time
	requests     []Request
	transactions []Transaction
	asked        [][]string // the line item IDs each OrderTransactions call asked for
}

// twoPlans is the UK plan (A-1 with L-1 and L-2) and the Germany plan (A-2 with L-3), the person seeing both.
func twoPlans(tb testing.TB) *world {
	tb.Helper()
	w := &world{tb: tb, unmatched: map[string]bool{}, choices: allwork.Choices{SeesFinance: true},
		currency: allwork.Currency{Code: money.BaseCurrency, Rate: money.One}, approved: map[string]time.Time{}}
	priya := &allwork.Owner{ID: "U-1002", Name: "Priya Raman"}
	w.add(allwork.LevelPlan, ukPlan, "", "UK plan").add(allwork.LevelCampaign, "C-1", ukPlan, "Launch").
		add(allwork.LevelInitiative, "I-1", "C-1", "Retail").
		add(allwork.LevelActivity, ukActivity, "I-1", "Stand", func(n *allwork.Node) {
			n.Owner, n.Attrs = priya, map[string]allwork.AttrValue{companyAttr: {OptionIDs: []string{companyGB01}}}
		}).
		add(allwork.LevelLineItem, "L-1", ukActivity, "Stand build", ready).
		add(allwork.LevelLineItem, "L-2", ukActivity, "Venue hire", ready)
	w.add(allwork.LevelPlan, dePlan, "", "Germany plan").add(allwork.LevelCampaign, "C-2", dePlan, "Fleet").
		add(allwork.LevelInitiative, "I-2", "C-2", "Trucks").
		add(allwork.LevelActivity, deActivity, "I-2", "Expo").
		add(allwork.LevelLineItem, "L-3", deActivity, "Booth", ready)
	return w
}

func ready(n *allwork.Node) { n.Swbs = swbsReady }

func (w *world) add(level allwork.Level, id, parentID, name string, changes ...func(*allwork.Node)) *world {
	n := allwork.Node{Level: level, ID: id, ParentID: parentID, PlanID: id, Name: name, PlanYear: testYear}
	if parentID != "" {
		n.PlanID = w.node(parentID).PlanID
	}
	for _, change := range changes {
		change(&n)
	}
	w.nodes = append(w.nodes, n)
	return w
}

func (w *world) node(id string) allwork.Node {
	i := slices.IndexFunc(w.nodes, func(n allwork.Node) bool { return n.ID == id })
	if i < 0 {
		w.tb.Fatalf("no node %s", id)
	}
	return w.nodes[i]
}

func (w *world) on(id string, change func(*allwork.Node)) *world {
	for i := range w.nodes {
		if w.nodes[i].ID == id {
			change(&w.nodes[i])
		}
	}
	return w
}

func (w *world) unmatch(ids ...string) *world {
	for _, id := range ids {
		w.unmatched[id] = true
	}
	return w
}

func (w *world) scope() measure.Scope {
	m := allwork.Matched{Plans: []string{}, Campaigns: []string{}, Initiatives: []string{}, Activities: []string{},
		LineItems: []string{}}
	for _, n := range w.nodes {
		if w.unmatched[n.ID] {
			continue
		}
		switch n.Level {
		case allwork.LevelPlan:
			m.Plans = append(m.Plans, n.ID)
		case allwork.LevelActivity:
			m.Activities = append(m.Activities, n.ID)
		case allwork.LevelLineItem:
			m.LineItems = append(m.LineItems, n.ID)
		}
	}
	company := allwork.Attribute{ID: companyAttr, Label: "Company code", Level: allwork.LevelActivity,
		Type: allwork.TypeSingleChoice, Options: []allwork.Option{{ID: companyGB01, Label: "GB01"}}}
	return measure.Scope{Year: testYear, Nodes: w.nodes, Matched: m, Choices: w.choices,
		Attributes: []allwork.Attribute{company}, Currency: w.currency}
}

func (w *world) year() year {
	return newYear(w.scope(), now)
}

func (w *world) service() *Service {
	return NewService(fakeTx{}, Deps{Scoped: w, Store: w, Rates: w}, func() time.Time { return now })
}

func (w *world) report(t Table) Report {
	w.tb.Helper()
	r, err := w.service().Report(w.tb.Context(), actor.Actor{ID: "U-1002"}, measure.Filter{}, t)
	if err != nil {
		w.tb.Fatalf("Report() error = %v", err)
	}
	return r
}

func (w *world) Scope(context.Context, actor.Actor, measure.Filter, []string) (measure.Scope, error) {
	return w.scope(), nil
}

func (w *world) ActivitiesApprovedAt(_ context.Context, activityIDs []string) (map[string]time.Time, error) {
	found := map[string]time.Time{}
	for _, id := range activityIDs {
		if at, ok := w.approved[id]; ok {
			found[id] = at
		}
	}
	return found, nil
}

func (w *world) WaitingForGsap(_ context.Context, lineItemIDs []string) ([]Request, error) {
	return slices.DeleteFunc(slices.Clone(w.requests), func(r Request) bool {
		return !slices.ContainsFunc(r.Lines, func(l RequestLine) bool { return slices.Contains(lineItemIDs, l.LineItemID) })
	}), nil
}

func (w *world) OrderTransactions(_ context.Context, lineItemIDs []string) ([]Transaction, error) {
	w.asked = append(w.asked, lineItemIDs)
	return slices.DeleteFunc(slices.Clone(w.transactions), func(t Transaction) bool {
		return !slices.Contains(lineItemIDs, t.LineItemID)
	}), nil
}

// Rates: a euro is worth 2 US dollars.
func (w *world) Rates(context.Context) (money.Rates, error) {
	return money.Rates{"EUR": rate(w.tb, "0.5"), money.BaseCurrency: money.One}, nil
}

type fakeTx struct{}

func (fakeTx) Read(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }

func rate(tb testing.TB, s string) money.Rate {
	tb.Helper()
	r, err := money.ParseRate(s)
	if err != nil {
		tb.Fatalf("ParseRate(%q) error = %v", s, err)
	}
	return r
}

func usd(tb testing.TB, s string) money.Money {
	tb.Helper()
	return inCode(tb, s, money.BaseCurrency)
}

func inCode(tb testing.TB, s, code string) money.Money {
	tb.Helper()
	a, err := money.ParseAmount(s)
	if err != nil {
		tb.Fatalf("ParseAmount(%q) error = %v", s, err)
	}
	return money.Money{Amount: a, Currency: code}
}
