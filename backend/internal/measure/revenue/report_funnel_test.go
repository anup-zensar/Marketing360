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

// TestFunnelReport_AModelThatDoesNotAllowTheStageIsUncovered proves FR-018: a plan whose activity sets its
// own later stages, and never typed an MQL goal, shows "—" for MQL and is counted among the uncovered
// plans — even while another plan's model covers that stage and the selector keeps it available.
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
