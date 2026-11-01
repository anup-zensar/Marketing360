package plandetails

import (
	"context"
	"maps"
	"slices"
	"testing"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

const testYearNumber = 2026

// year builds the nodes of an in-memory year for the tests; every node matches unless left out by unmatch.
type year struct {
	t         testing.TB
	nodes     []allwork.Node
	unmatched map[string]bool
}

func newYear(t *testing.T) *year {
	t.Helper()
	return &year{t: t, unmatched: map[string]bool{}}
}

func (y *year) plan(id string, units ...string) *year {
	y.nodes = append(y.nodes, allwork.Node{Level: allwork.LevelPlan, ID: id, PlanID: id, Status: "draft",
		BusinessUnitIDs: units, PlanYear: testYearNumber})
	return y
}

func (y *year) campaign(id, parentID string, changes ...func(*allwork.Node)) *year {
	return y.container(allwork.LevelCampaign, id, parentID, changes)
}

func (y *year) initiative(id, parentID string, changes ...func(*allwork.Node)) *year {
	return y.container(allwork.LevelInitiative, id, parentID, changes)
}

func (y *year) container(level allwork.Level, id, parentID string, changes []func(*allwork.Node)) *year {
	n := allwork.Node{Level: level, ID: id, ParentID: parentID, PlanID: y.planOf(parentID), Status: "noActivities",
		PlanYear: testYearNumber}
	for _, change := range changes {
		change(&n)
	}
	y.nodes = append(y.nodes, n)
	return y
}

// on applies the changes to the node id already added: a plan's status or attributes.
func (y *year) on(id string, changes ...func(*allwork.Node)) *year {
	for i := range y.nodes {
		if y.nodes[i].ID != id {
			continue
		}
		for _, change := range changes {
			change(&y.nodes[i])
		}
	}
	return y
}

// activity adds an activity with a planned cost in USD ("" for none) and applies the changes to it.
func (y *year) activity(id, parentID, cost string, changes ...func(*allwork.Node)) *year {
	n := allwork.Node{Level: allwork.LevelActivity, ID: id, ParentID: parentID, PlanID: y.planOf(parentID),
		Status: "created", PlanYear: testYearNumber}
	if cost != "" {
		n.PlannedCost = amount(y.t, cost)
	}
	for _, change := range changes {
		change(&n)
	}
	y.nodes = append(y.nodes, n)
	return y
}

func (y *year) unmatch(ids ...string) *year {
	for _, id := range ids {
		y.unmatched[id] = true
	}
	return y
}

func (y *year) planOf(id string) string {
	for _, n := range y.nodes {
		if n.ID == id {
			return n.PlanID
		}
	}
	y.t.Fatalf("no parent %s", id)
	return ""
}

func (y *year) matching() Matching {
	matched := map[string]bool{}
	for _, n := range y.nodes {
		if !y.unmatched[n.ID] {
			matched[n.ID] = true
		}
	}
	return Matching{Year: testYearNumber, Nodes: y.nodes, Matched: matched, Currency: allwork.Currency{
		Code: money.BaseCurrency, Rate: money.One,
	}}
}

func amount(tb testing.TB, s string) *money.Amount {
	tb.Helper()
	a, err := money.ParseAmount(s)
	if err != nil {
		tb.Fatalf("ParseAmount(%q) error = %v", s, err)
	}
	return &a
}

func choice(attributeID string, optionIDs ...string) func(*allwork.Node) {
	return func(n *allwork.Node) {
		if n.Attrs == nil {
			n.Attrs = map[string]allwork.AttrValue{}
		}
		n.Attrs[attributeID] = allwork.AttrValue{OptionIDs: optionIDs}
	}
}

// split gives the % split attribute the options with their shares, as "60.00".
func split(t *testing.T, attributeID string, shares map[string]string) func(*allwork.Node) {
	t.Helper()
	return func(n *allwork.Node) {
		v := allwork.AttrValue{Shares: map[string]money.Percent{}}
		for _, id := range slices.Sorted(maps.Keys(shares)) {
			p, err := money.ParsePercent(shares[id])
			if err != nil {
				t.Fatalf("ParsePercent(%q) error = %v", shares[id], err)
			}
			v.OptionIDs = append(v.OptionIDs, id)
			v.Shares[id] = p
		}
		if n.Attrs == nil {
			n.Attrs = map[string]allwork.AttrValue{}
		}
		n.Attrs[attributeID] = v
	}
}

func yesNo(attributeID string, yes bool) func(*allwork.Node) {
	return func(n *allwork.Node) {
		if n.Attrs == nil {
			n.Attrs = map[string]allwork.AttrValue{}
		}
		n.Attrs[attributeID] = allwork.AttrValue{Yes: &yes}
	}
}

func starting(iso string) func(*allwork.Node) {
	return func(n *allwork.Node) {
		d := date.MustParse(iso)
		n.StartDate = &d
	}
}

// attribute is an active activity attribute; atLevel moves it to another level.
func attribute(id, label, kind string, options ...string) allwork.Attribute {
	a := allwork.Attribute{ID: id, Label: label, Level: allwork.LevelActivity, Type: kind, Active: true}
	for _, o := range options {
		a.Options = append(a.Options, allwork.Option{ID: o, Label: o})
	}
	return a
}

func atLevel(a allwork.Attribute, level allwork.Level) allwork.Attribute {
	a.Level = level
	return a
}

type fakeScoped struct {
	matching   Matching
	attributes []allwork.Attribute
	columns    []string
}

func (f *fakeScoped) Matched(_ context.Context, _ actor.Actor, _ measure.Filter, columns []string) (Matching, error) {
	f.columns = columns
	return f.matching, nil
}

func (f *fakeScoped) Attributes(context.Context) ([]allwork.Attribute, error) {
	return f.attributes, nil
}

type fakeTx struct{}

func (fakeTx) Read(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }
