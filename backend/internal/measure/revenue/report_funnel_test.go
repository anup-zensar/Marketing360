package revenue_test

import (
	"testing"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/revenue"
	"mapp/internal/platform/date"
)

func TestFunnelReport_SelectsTheFirstAvailableStageAndCountsUncovered_FR015_049C11(t *testing.T) {
	t.Parallel()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	s := newFixtureWith(sc, goals, nil)

	got, err := s.FunnelReport(t.Context(), priya, measure.Filter{}, "")
	if err != nil {
		t.Fatalf("FunnelReport: %v", err)
	}
	if got.Stage != revenue.StageINQ {
		t.Errorf("stage = %q, want inq, the first the model allows", got.Stage)
	}
	if len(got.AvailableStages) != 2 {
		t.Errorf("availableStages = %v, want inq and mql", got.AvailableStages)
	}
	if got.UncoveredCount != 0 {
		t.Errorf("uncoveredCount = %d, want 0: the one plan in view is covered", got.UncoveredCount)
	}
}

// Proves FR-018: a plan whose activity types only an INQ goal is uncovered for MQL, even while another
// plan's model still offers it.
func TestFunnelReport_AModelThatDoesNotAllowTheStageIsUncovered_FR018(t *testing.T) {
	t.Parallel()
	start, end := date.MustParse("2026-01-01"), date.MustParse("2026-12-31")
	nodes := []allwork.Node{
		{Level: allwork.LevelActivity, ID: "A-1", PlanID: "P-1", Name: "Launch film",
			INQ: inqGoal(365_000), StartDate: &start, EndDate: &end},
		{Level: allwork.LevelActivity, ID: "A-2", PlanID: "P-2", Name: "Direct deal",
			INQ: inqGoal(500), StartDate: &start, EndDate: &end},
	}
	sc := measure.Scope{Year: 2026, Nodes: nodes, Matched: allwork.Matched{Activities: []string{"A-1", "A-2"}},
		Currency: allwork.Currency{Code: "USD"}}
	goals := fakeGoals{
		"A-1": revenue.ActivityGoals{Model: lubricantsModel()},
		"A-2": revenue.ActivityGoals{Model: lubricantsModel(), SetLaterStagesMyself: true},
	}
	s := newFixtureWith(sc, goals, nil)

	got, err := s.FunnelReport(t.Context(), priya, measure.Filter{}, "mql")
	if err != nil {
		t.Fatalf("FunnelReport: %v", err)
	}
	if got.Stage != revenue.StageMQL {
		t.Fatalf("stage = %q, want mql: P-1's model still offers it", got.Stage)
	}
	if got.UncoveredCount != 1 {
		t.Errorf("uncoveredCount = %d, want 1: P-2 typed only an INQ goal", got.UncoveredCount)
	}
}

func TestRankedRevenue_SortsByProjectedByDefault(t *testing.T) {
	t.Parallel()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	s := newFixtureWith(sc, goals, nil)

	got, err := s.RankedRevenue(t.Context(), priya, measure.Filter{Show: "activity"}, revenue.RankedQuery{Sort: revenue.SortProjectedRevenue, Order: revenue.OrderDesc, Size: 10, Page: 1})
	if err != nil {
		t.Fatalf("RankedRevenue: %v", err)
	}
	if got.Total != 1 || len(got.Groups) != 1 || len(got.Groups[0].Entries) != 1 {
		t.Fatalf("ranked = %+v, want the one activity in view", got)
	}
	if got.Groups[0].Entries[0].ID != "A-1" {
		t.Errorf("entry id = %q, want A-1", got.Groups[0].Entries[0].ID)
	}
}

// TestRankedRevenue_TopLevelTotalsMatchTheSeriesTotals proves SC-002: with no grouping, the ranked list's
// one top-level group totals the same actual, projected and goal as the report's own cards, for the same
// filter.
func TestRankedRevenue_TopLevelTotalsMatchTheSeriesTotals_SC002(t *testing.T) {
	t.Parallel()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	s := newFixtureWith(sc, goals, nil)

	report, err := s.Report(t.Context(), priya, measure.Filter{})
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	ranked, err := s.RankedRevenue(t.Context(), priya, measure.Filter{Show: "activity"},
		revenue.RankedQuery{Sort: revenue.SortProjectedRevenue, Order: revenue.OrderDesc, Size: 10, Page: 1})
	if err != nil {
		t.Fatalf("RankedRevenue: %v", err)
	}
	totals := ranked.Groups[0].Totals
	if totals.Goal == nil || report.Cards.Goal.Value == nil || totals.Goal.Amount.Cents() != report.Cards.Goal.Value.Amount.Cents() {
		t.Errorf("ranked goal total = %v, want the report's own goal card %v", totals.Goal, report.Cards.Goal.Value)
	}
	if totals.Projected == nil || report.Cards.Projected.Value == nil ||
		totals.Projected.Amount.Cents() != report.Cards.Projected.Value.Amount.Cents() {
		t.Errorf("ranked projected total = %v, want the report's own projected card %v", totals.Projected, report.Cards.Projected.Value)
	}
}

// Pins a real defect found while closing T053's SC-002 gap: ranked rows never carried Projected or
// Performance, so every one answered "—" for both regardless of real data.
func TestRankedFunnel_EntriesCarryProjectedAndPerformance_FR012_FR017(t *testing.T) {
	t.Parallel()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	s := newFixtureWith(sc, goals, nil)

	got, err := s.RankedFunnel(t.Context(), priya, measure.Filter{Show: "activity"},
		revenue.RankedQuery{Sort: revenue.SortProjected, Order: revenue.OrderDesc, Size: 10, Page: 1})
	if err != nil {
		t.Fatalf("RankedFunnel: %v", err)
	}
	entry := got.Groups[0].Entries[0]
	if entry.Figures.Projected == nil {
		t.Fatalf("entry projected = %+v, want a value: the activity has a goal and a model", entry.Figures)
	}
	if entry.Figures.Performance == nil {
		t.Errorf("entry performance = %+v, want a value computed from projected ÷ goal", entry.Figures)
	}
}

func TestConversionCard_NoPreviousActualShowsTheReason_FR017(t *testing.T) {
	t.Parallel()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	s := newFixtureWith(sc, goals, nil)

	got, err := s.FunnelReport(t.Context(), priya, measure.Filter{}, "mql")
	if err != nil {
		t.Fatalf("FunnelReport: %v", err)
	}
	if got.Cards.Conversion.Value != nil {
		t.Errorf("conversion = %v, want no value: INQ has no counted facts in this fixture", got.Cards.Conversion.Value)
	}
	if got.Cards.Conversion.Reason != revenue.ReasonNoPreviousStage {
		t.Errorf("conversion reason = %q, want %q", got.Cards.Conversion.Reason, revenue.ReasonNoPreviousStage)
	}
}

func TestConversionCard_WithPreviousActualComputesOneDecimal_FR017(t *testing.T) {
	t.Parallel()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	facts := []revenue.FactTotal{
		{ActivityID: "A-1", Stage: "inq", Month: "2026-02", Counts: 1000, FactCount: 1},
		{ActivityID: "A-1", Stage: "mql", Month: "2026-02", Counts: 25, FactCount: 1},
	}
	s := newFixtureWith(sc, goals, facts)

	got, err := s.FunnelReport(t.Context(), priya, measure.Filter{}, "mql")
	if err != nil {
		t.Fatalf("FunnelReport: %v", err)
	}
	if got.Cards.Conversion.Value == nil {
		t.Fatalf("conversion = %+v, want a value: both stages have counted facts", got.Cards.Conversion)
	}
	if *got.Cards.Conversion.Value != 2.5 {
		t.Errorf("conversion = %v, want 2.5 %% (25 of 1,000, one decimal)", *got.Cards.Conversion.Value)
	}
}
