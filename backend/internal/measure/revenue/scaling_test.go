package revenue_test

import (
	"context"
	"testing"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/revenue"
	"mapp/internal/platform/authz"
)

// countedFacts counts every call to ReportTotals, so a test can pin that the statement count does not grow
// with the number of activities in view (research R2, SC-003).
type countedFacts struct {
	calls int
}

func (f *countedFacts) ReportTotals(context.Context, authz.Scope, []string, int) ([]revenue.FactTotal, error) {
	f.calls++
	return nil, nil
}

func manyActivityScope(n int) measure.Scope {
	nodes := make([]allwork.Node, 0, n)
	ids := make([]string, 0, n)
	for i := range n {
		id := "A-" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		nodes = append(nodes, allwork.Node{Level: allwork.LevelActivity, ID: id, PlanID: "P-1",
			Name: "Activity", INQ: inqGoal(1000)})
		ids = append(ids, id)
	}
	return measure.Scope{Year: 2026, Nodes: nodes, Matched: allwork.Matched{Activities: ids},
		Currency: allwork.Currency{Code: "USD"}}
}

func manyGoals(n int) fakeGoals {
	goals := make(fakeGoals, n)
	sc := manyActivityScope(n)
	for _, node := range sc.Nodes {
		goals[node.ID] = revenue.ActivityGoals{Model: lubricantsModel()}
	}
	return goals
}

func TestReport_StatementCountStaysConstantWhateverTheActivityCount_SC003(t *testing.T) {
	t.Parallel()
	facts := &countedFacts{}

	for _, n := range []int{5, 500} {
		facts.calls = 0
		sc := manyActivityScope(n)
		s := revenue.NewService(fakeTx{}, revenue.Deps{
			Scoper: fakeScoper{scope: sc}, Facts: facts, Goals: manyGoals(n),
			Categories: fakeCategories{}, Rates: fakeRates{},
		}, func() time.Time { return noon })
		if _, err := s.Report(t.Context(), priya, measure.Filter{}); err != nil {
			t.Fatalf("Report() with %d activities: %v", n, err)
		}
		if facts.calls != 1 {
			t.Errorf("ReportTotals called %d times for %d activities, want exactly 1: one statement whatever the view (SC-003)",
				facts.calls, n)
		}
	}
}
