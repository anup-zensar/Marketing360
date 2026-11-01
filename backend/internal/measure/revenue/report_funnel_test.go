package revenue_test

import (
	"testing"

	"mapp/internal/measure"
	"mapp/internal/measure/revenue"
)

func TestFunnelReport_SelectsTheFirstAvailableStageAndCountsUncovered_049C11(t *testing.T) {
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
