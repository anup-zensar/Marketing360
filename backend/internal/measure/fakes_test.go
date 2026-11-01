package measure_test

import (
	"context"
	"fmt"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/history"
	"mapp/internal/measure"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// The frame's fixture: 2026 with P-1 (team T-1, in the reports.view scope) and P-2 (team T-2, outside it), each with
// a campaign and two activities; A-1 has two line items, A-2 one. Today is 18 Oct 2026.
func frameNodes() []allwork.Node {
	priya, daniel, wei := &allwork.Owner{ID: "U-2", Name: "Priya Raman"}, &allwork.Owner{ID: "U-3", Name: "Daniel Okoro"},
		&allwork.Owner{ID: "U-9", Name: "Wei Zhang"}
	nodes := []allwork.Node{
		{Level: allwork.LevelPlan, ID: "P-1", PlanID: "P-1", Name: "FY2026 Lubricants", PlanType: "operational",
			TeamIDs: []string{"T-1"}, Owner: priya},
		{Level: allwork.LevelCampaign, ID: "C-1", ParentID: "P-1", PlanID: "P-1", Name: "Helix", Owner: priya,
			Attrs: map[string]allwork.AttrValue{"SA-CL": {OptionIDs: []string{"SO-EU"}}}},
		{Level: allwork.LevelActivity, ID: "A-1", ParentID: "C-1", PlanID: "P-1", Owner: priya, ActivityTypeID: "AT-1",
			Status: "approved", Attrs: map[string]allwork.AttrValue{"SA-B": {OptionIDs: []string{"SO-OLD"}}}},
		{Level: allwork.LevelLineItem, ID: "L-1", ParentID: "A-1", PlanID: "P-1"},
		{Level: allwork.LevelLineItem, ID: "L-2", ParentID: "A-1", PlanID: "P-1"},
		{Level: allwork.LevelActivity, ID: "A-2", ParentID: "C-1", PlanID: "P-1", Owner: daniel, ActivityTypeID: "AT-2",
			Status: "created"},
		{Level: allwork.LevelLineItem, ID: "L-3", ParentID: "A-2", PlanID: "P-1"},
		{Level: allwork.LevelPlan, ID: "P-2", PlanID: "P-2", Name: "FY2026 Mobility China", PlanType: "strategic",
			TeamIDs: []string{"T-2"}, Owner: wei},
		{Level: allwork.LevelCampaign, ID: "C-2", ParentID: "P-2", PlanID: "P-2", Owner: wei},
		{Level: allwork.LevelActivity, ID: "A-3", ParentID: "C-2", PlanID: "P-2", Owner: wei, ActivityTypeID: "AT-3",
			Status: "created"},
	}
	for i := range nodes {
		nodes[i].PlanYear = 2026
	}
	return nodes
}

func frameAttributes() []allwork.Attribute {
	return []allwork.Attribute{
		{ID: "SA-CL", Label: "Cluster", Level: allwork.LevelCampaign, Active: true, Type: allwork.TypeSingleChoice,
			Options: []allwork.Option{{ID: "SO-EU", Label: "Europe"}, {ID: "SO-AS", Label: "Asia"}}},
		{ID: "SA-B", Label: "Brand", Level: allwork.LevelActivity, Active: true, Type: allwork.TypeMultipleChoice,
			Options: []allwork.Option{{ID: "SO-HX", Label: "Helix"}, {ID: "SO-OLD", Label: "Retired brand", Inactive: true},
				{ID: "SO-GONE", Label: "Unused retired brand", Inactive: true}}},
		{ID: "SA-F", Label: "FIT4", Level: allwork.LevelActivity, Active: true, Type: allwork.TypeYesNo},
		{ID: "SA-C2", Label: "Cluster", Level: allwork.LevelActivity, Active: true, Type: allwork.TypeSingleChoice,
			Options: []allwork.Option{{ID: "SO-X", Label: "X"}}},
		{ID: "SA-N", Label: "Expected leads", Level: allwork.LevelActivity, Active: true, Number: true, Type: "number"},
		{ID: "SA-OFF", Label: "Switched off", Level: allwork.LevelActivity, Type: allwork.TypeSingleChoice},
	}
}

type fakePlanning struct {
	nodes       []allwork.Node
	financePlan []string // nil: every plan
	years       []int
}

func (f fakePlanning) Years(context.Context) ([]int, error) { return f.years, nil }

func (f fakePlanning) Year(_ context.Context, year int, _ date.Date, _ []string) (allwork.YearData, error) {
	if year != 2026 {
		return allwork.YearData{}, nil
	}
	return allwork.YearData{Nodes: f.nodes}, nil
}

func (f fakePlanning) Choices(context.Context, string) (allwork.Choices, error) {
	return allwork.Choices{
		Teams:         []allwork.Option{{ID: "T-1", Label: "Lubricants Europe", Mine: true}, {ID: "T-2", Label: "Mobility China"}},
		ActivityTypes: []allwork.Option{{ID: "AT-1", Label: "Event"}, {ID: "AT-2", Label: "Digital"}, {ID: "AT-3", Label: "Fleet"}},
		People: []allwork.Option{{ID: "U-2", Label: "Priya Raman"}, {ID: "U-3", Label: "Daniel Okoro"},
			{ID: "U-9", Label: "Wei Zhang"}},
		SeesFinance: true, FinancePlans: f.financePlan,
	}, nil
}

func (f fakePlanning) Attributes(context.Context) ([]allwork.Attribute, error) {
	return frameAttributes(), nil
}

func (f fakePlanning) Currency(context.Context, string) (allwork.Currency, error) {
	return allwork.Currency{Code: "GBP", Rate: mustRate("0.79")}, nil
}

func (f fakePlanning) RatesUsed(context.Context) ([]string, error) { return nil, nil }

type fakePlans []string

func (f fakePlans) InReportsScope(context.Context) ([]string, error) { return f, nil }

type fakeCurrencies struct{}

func (fakeCurrencies) Active(context.Context) ([]measure.Currency, error) {
	return []measure.Currency{{Code: "EUR", Name: "Euro"}, {Code: "GBP", Name: "Pound sterling"}}, nil
}

func (fakeCurrencies) RateOf(_ context.Context, code string) (money.Rate, error) {
	if code == "EUR" {
		return mustRate("0.92"), nil
	}
	return money.Rate{}, fmt.Errorf("no rate for %s", code)
}

type fakeTx struct{}

func (fakeTx) Read(ctx context.Context, fn func(ctx context.Context) error) error  { return fn(ctx) }
func (fakeTx) Write(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }

func mustRate(text string) money.Rate {
	r, err := money.ParseRate(text)
	if err != nil {
		panic(err)
	}
	return r
}

// frameService serves the fixture to a person who may see reports of P-1 only.
func frameService(planning fakePlanning) *measure.Service {
	if planning.nodes == nil {
		planning.nodes = frameNodes()
	}
	return measure.NewService(history.Tools{Tx: fakeTx{}, Now: func() time.Time { return time.Date(2026, 10, 18, 14, 5, 0, 0, time.UTC) }}, measure.Deps{Planning: planning, Plans: fakePlans{"P-1"}, Currencies: fakeCurrencies{}}, measure.NewRegistry())
}
