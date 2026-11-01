package planbudget_test

import (
	"context"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/financials"
	"mapp/internal/measure"
	"mapp/internal/measure/planbudget"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/money"
)

var (
	today = time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC)
	priya = actor.Actor{ID: "U-2", Name: "Priya Raman"}
)

type fakeTx struct{}

func (fakeTx) Read(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }

type fakeScoper struct{ scope measure.Scope }

func (f fakeScoper) Scope(context.Context, actor.Actor, measure.Filter, []string) (measure.Scope, error) {
	return f.scope, nil
}

type fakeSpending map[string]map[string]financials.Facts

func (f fakeSpending) SpendByMonth(_ context.Context, _ actor.Actor, ids []string) (map[string]map[string]financials.Facts, error) {
	out := map[string]map[string]financials.Facts{}
	for _, id := range ids {
		out[id] = f[id]
	}
	return out, nil
}

func dollars(n int64) *money.Amount {
	a := money.FromCents(n * 100)
	return &a
}

func month(m string, n int64) allwork.MonthAmount {
	return allwork.MonthAmount{Month: m, USD: *dollars(n)}
}

// fixtureNodes: P-1 (1,000, phased Jan–Feb) › C-1 Helix launch (600) › I-1 (400) › A-1 (300) › L-1 (300 in March);
// P-1 › C-2 Rimula fleet (300, Daniel) › I-2 (100, below its planned cost) › A-2 (200) › L-2 (100 in Nov, 100 in Dec).
func fixtureNodes() []allwork.Node {
	daniel := &allwork.Owner{ID: "U-3", Name: "Daniel Okoro"}
	split := allwork.AttrValue{OptionIDs: []string{"SO-HX", "SO-PZ"},
		Shares: map[string]money.Percent{"SO-HX": money.PercentFromHundredths(6000), "SO-PZ": money.PercentFromHundredths(4000)}}
	return []allwork.Node{
		{Level: allwork.LevelPlan, ID: "P-1", PlanID: "P-1", Name: "FY2026 Lubricants", Budget: dollars(1000),
			PhasedBudgetUSD: []allwork.MonthAmount{month("2026-01", 500), month("2026-02", 500)}},
		{Level: allwork.LevelCampaign, ID: "C-1", ParentID: "P-1", PlanID: "P-1", Order: 1, Name: "Helix launch",
			Budget: dollars(600), Owner: &allwork.Owner{ID: "U-2", Name: "Priya Raman"}, Attrs: map[string]allwork.AttrValue{"SA-1022": split}},
		{Level: allwork.LevelCampaign, ID: "C-2", ParentID: "P-1", PlanID: "P-1", Order: 2, Name: "Rimula fleet",
			Budget: dollars(300), Owner: daniel},
		{Level: allwork.LevelInitiative, ID: "I-1", ParentID: "C-1", PlanID: "P-1", Name: "Launch events", Budget: dollars(400)},
		{Level: allwork.LevelInitiative, ID: "I-2", ParentID: "C-2", PlanID: "P-1", Name: "Fleet days", Budget: dollars(100)},
		{Level: allwork.LevelActivity, ID: "A-1", ParentID: "I-1", PlanID: "P-1", Name: "Trade show", PlannedCost: dollars(300)},
		{Level: allwork.LevelActivity, ID: "A-2", ParentID: "I-2", PlanID: "P-1", Name: "Fleet day", PlannedCost: dollars(200)},
		{Level: allwork.LevelLineItem, ID: "L-1", ParentID: "A-1", PlanID: "P-1", Name: "Stand", PlannedCost: dollars(300),
			MonthsUSD: []allwork.MonthAmount{month("2026-03", 300)}},
		{Level: allwork.LevelLineItem, ID: "L-2", ParentID: "A-2", PlanID: "P-1", Name: "Venue", PlannedCost: dollars(200),
			MonthsUSD: []allwork.MonthAmount{month("2026-11", 100), month("2026-12", 100)}},
	}
}

var everything = allwork.Matched{Plans: []string{"P-1"}, Campaigns: []string{"C-1", "C-2"}, Initiatives: []string{"I-1", "I-2"},
	Activities: []string{"A-1", "A-2"}, LineItems: []string{"L-1", "L-2"}}

var spend = fakeSpending{
	"L-1": {"2026-02": {Committed: *dollars(250), Paid: *dollars(100)}, "2025-12": {Paid: *dollars(50)}},
	"L-2": {"2026-10": {Committed: *dollars(150), Paid: *dollars(80)}},
}

func scope(matched allwork.Matched, seesFinance bool) measure.Scope {
	return measure.Scope{Year: 2026, Nodes: fixtureNodes(), Matched: matched,
		Choices:    allwork.Choices{SeesFinance: seesFinance},
		Attributes: []allwork.Attribute{{ID: "SA-1022", Label: "Brand", Options: []allwork.Option{{ID: "SO-HX", Label: "Helix"}, {ID: "SO-PZ", Label: "Pennzoil"}}}},
		Currency:   allwork.Currency{Code: "USD", Rate: money.One}}
}

func service(sc measure.Scope) *planbudget.Service {
	return planbudget.NewService(fakeTx{}, planbudget.Deps{Scoper: fakeScoper{sc}, Spending: spend},
		func() time.Time { return today })
}

func shown(m *money.Money) string {
	if m == nil {
		return "—"
	}
	return m.Amount.String()
}
