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
