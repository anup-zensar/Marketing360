package revenue_test

import (
	"testing"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/revenue"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// oneActivityScope is one plan with one activity whose goal is 365,000 INQ across the whole of 2026 — the
// domain model's own worked example, extended with 047's velocities (SC-001).
func oneActivityScope() measure.Scope {
	start, end := date.MustParse("2026-01-01"), date.MustParse("2026-12-31")
	node := allwork.Node{Level: allwork.LevelActivity, ID: "A-1", PlanID: "P-1", Name: "Launch film",
		INQ: inqGoal(365_000), StartDate: &start, EndDate: &end}
	return measure.Scope{Year: 2026, Nodes: []allwork.Node{node}, Matched: allwork.Matched{Activities: []string{"A-1"}},
		Currency: allwork.Currency{Code: "USD"}}
}

func revenueReportFixture(t *testing.T) revenue.RevenueReport {
	t.Helper()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	s := newFixtureWith(sc, goals, nil)
	got, err := s.Report(t.Context(), priya, measure.Filter{})
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	return got
}

func TestReport_NamesTheModelAndSplitsAtToday_FR012_FR010(t *testing.T) {
	t.Parallel()
	got := revenueReportFixture(t)

	if got.Projection == nil || got.Projection.ModelName != "Lubricants 1.0" {
		t.Fatalf("stamp = %+v, want the model named so a reviewer can reproduce the figure", got.Projection)
	}
	if len(got.Periods) != 12 {
		t.Fatalf("periods = %d, want 12 months by default", len(got.Periods))
	}
	// The fixed clock sits in July 2026.
	july := got.Periods[6]
	if !july.InProgress {
		t.Errorf("July = %+v, want the month in progress under the fixed clock", july)
	}
}

func TestReport_MatchesTheHandCalculatedTable_SC001(t *testing.T) {
	t.Parallel()
	got := revenueReportFixture(t)

	if got.Cards.Goal.Value == nil {
		t.Fatalf("goal card = %+v, want a value: the activity has a goal", got.Cards.Goal)
	}
	// 273 deals x 10,000 USD = 2,730,000.00 USD, exactly as 047's own table proves.
	wantGoal := money.USD(money.FromCents(273_000_000))
	if got.Cards.Goal.Value.Amount.Cents() != wantGoal.Amount.Cents() {
		t.Errorf("goal = %s, want %s (273 deals x 10,000 USD, 047's own table)", got.Cards.Goal.Value.Amount, wantGoal.Amount)
	}
}

func TestReport_WithoutAnyGoalStatesTheReason_FR009(t *testing.T) {
	t.Parallel()
	sc := measure.Scope{Year: 2026, Currency: allwork.Currency{Code: "USD"}}
	s := newFixtureWith(sc, fakeGoals{}, nil)

	got, err := s.Report(t.Context(), priya, measure.Filter{})
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	if got.Reason != revenue.ReasonNoData {
		t.Errorf("reason = %q, want %q: no activities in view", got.Reason, revenue.ReasonNoData)
	}
}

func TestReport_ActualIsCountedFactsByFactDate_FR021(t *testing.T) {
	t.Parallel()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	facts := []revenue.FactTotal{
		{ActivityID: "A-1", Stage: "revenue", Month: "2026-02", Currency: "USD", AmountMinor: 500_000, FactCount: 1},
	}
	s := newFixtureWith(sc, goals, facts)

	got, err := s.Report(t.Context(), priya, measure.Filter{})
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	february := got.Periods[1]
	if february.Actual == nil || february.Actual.Amount.Cents() != 500_000 {
		t.Fatalf("February actual = %v, want the 5,000.00 USD fact counted in its own month", february.Actual)
	}
	if got.Cards.Actual.Value == nil || got.Cards.Actual.Value.Amount.Cents() != 500_000 {
		t.Errorf("actual card = %v, want the one counted fact", got.Cards.Actual.Value)
	}
}
