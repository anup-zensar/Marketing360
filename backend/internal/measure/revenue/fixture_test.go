package revenue_test

import (
	"context"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/revenue"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
)

var (
	noon  = time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	priya = actor.Actor{ID: "U-2", Name: "Priya Raman"}
)

type fakeTx struct{}

func (fakeTx) Read(_ context.Context, fn func(context.Context) error) error {
	return fn(context.Background())
}

type fakeScoper struct{ scope measure.Scope }

func (f fakeScoper) Scope(context.Context, actor.Actor, measure.Filter, []string) (measure.Scope, error) {
	return f.scope, nil
}

type fakeFacts struct{ rows []revenue.FactTotal }

func (f fakeFacts) ReportTotals(_ context.Context, _ authz.Scope, activityIDs []string, _ int) ([]revenue.FactTotal, error) {
	wanted := map[string]bool{}
	for _, id := range activityIDs {
		wanted[id] = true
	}
	out := []revenue.FactTotal{}
	for _, r := range f.rows {
		if wanted[r.ActivityID] {
			out = append(out, r)
		}
	}
	return out, nil
}

type fakeGoals map[string]revenue.ActivityGoals

func (f fakeGoals) ActivityGoals(_ context.Context, _ authz.Scope, nodes []allwork.Node) (map[string]revenue.ActivityGoals, error) {
	out := map[string]revenue.ActivityGoals{}
	for _, n := range nodes {
		out[n.ID] = f[n.ID]
	}
	return out, nil
}

type fakeCategories []revenue.Category

func (f fakeCategories) ListCategories(context.Context) ([]revenue.Category, error) { return f, nil }

type fakeRates struct{}

func (fakeRates) Rates(context.Context) (money.Rates, error) { return money.Rates{}, nil }

// lubricantsModel mirrors the domain model's own waterfall example (§4.5): 2 %, 5 %, 10,000 USD a deal,
// with 047's velocities, one month from INQ to MQL, two more to a deal.
func lubricantsModel() projection.Model {
	return projection.Model{
		Name:         "Lubricants 1.0",
		InqToMqlRate: money.PercentFromHundredths(200), MqlToDealRate: money.PercentFromHundredths(500),
		AverageDealSize:        money.USD(money.FromCents(1_000_000)),
		InqToMqlVelocityMonths: 1, MqlToDealVelocityMonths: 2,
		ModelChangedAt: time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC),
	}
}

func inqGoal(n int64) *int64 { return &n }

func newFixtureWith(sc measure.Scope, goals fakeGoals, facts []revenue.FactTotal) *revenue.Service {
	return revenue.NewService(fakeTx{}, revenue.Deps{
		Scoper: fakeScoper{scope: sc}, Facts: fakeFacts{rows: facts}, Goals: goals,
		Categories: fakeCategories{}, Rates: fakeRates{},
	}, func() time.Time { return noon })
}
