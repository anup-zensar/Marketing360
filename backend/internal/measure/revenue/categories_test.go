package revenue_test

import (
	"net/url"
	"testing"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/revenue"
)

func queryWith(key, value string) url.Values {
	return url.Values{key: []string{value}}
}

// twoCategoryScope has two activities of different types: A-1 (AT-Asset) and A-2 (AT-Promo, left out of
// every category).
func twoCategoryScope() measure.Scope {
	nodes := []allwork.Node{
		{Level: allwork.LevelActivity, ID: "A-1", PlanID: "P-1", Name: "Trade show", INQ: inqGoal(1000), ActivityTypeID: "AT-Asset"},
		{Level: allwork.LevelActivity, ID: "A-2", PlanID: "P-1", Name: "Webinar", INQ: inqGoal(500), ActivityTypeID: "AT-Promo"},
	}
	return measure.Scope{Year: 2026, Nodes: nodes,
		Matched:  allwork.Matched{Activities: []string{"A-1", "A-2"}},
		Currency: allwork.Currency{Code: "USD"}}
}

func newFixtureWithCategories(sc measure.Scope, goals fakeGoals, categories fakeCategories) *revenue.Service {
	return revenue.NewService(fakeTx{}, revenue.Deps{
		Scoper: fakeScoper{scope: sc}, Facts: fakeFacts{}, Goals: goals,
		Categories: categories, Rates: fakeRates{},
	}, func() time.Time { return noon })
}

func TestRankedRevenue_CategoryNarrowsToItsOwnTypes_049C10(t *testing.T) {
	t.Parallel()
	sc := twoCategoryScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}, "A-2": revenue.ActivityGoals{Model: lubricantsModel()}}
	categories := fakeCategories{{ID: "CAT-1", Name: "Asset", ActivityTypeIDs: []string{"AT-Asset"}}}
	s := newFixtureWithCategories(sc, goals, categories)

	got, err := s.RankedRevenue(t.Context(), priya, measure.Filter{Show: "activity", Own: queryWith("category", "CAT-1")},
		revenue.RankedQuery{Sort: revenue.SortProjectedRevenue, Order: revenue.OrderDesc, Size: 10, Page: 1})
	if err != nil {
		t.Fatalf("RankedRevenue: %v", err)
	}
	if got.Total != 1 || got.Groups[0].Entries[0].ID != "A-1" {
		t.Fatalf("ranked = %+v, want only A-1, the one activity in the Asset category", got)
	}
}

func TestRankedRevenue_UncategorisedCountsTypesInNoCategory_049C10(t *testing.T) {
	t.Parallel()
	sc := twoCategoryScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}, "A-2": revenue.ActivityGoals{Model: lubricantsModel()}}
	categories := fakeCategories{{ID: "CAT-1", Name: "Asset", ActivityTypeIDs: []string{"AT-Asset"}}}
	s := newFixtureWithCategories(sc, goals, categories)

	got, err := s.RankedRevenue(t.Context(), priya, measure.Filter{Show: "activity", Own: queryWith("category", "uncategorised")},
		revenue.RankedQuery{Sort: revenue.SortProjectedRevenue, Order: revenue.OrderDesc, Size: 10, Page: 1})
	if err != nil {
		t.Fatalf("RankedRevenue: %v", err)
	}
	if got.Total != 1 || got.Groups[0].Entries[0].ID != "A-2" {
		t.Fatalf("ranked = %+v, want only A-2: its type is in no category", got)
	}
}

func TestRankedRevenue_AllCategoriesCountsEveryActivityOnce_049C10(t *testing.T) {
	t.Parallel()
	sc := twoCategoryScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}, "A-2": revenue.ActivityGoals{Model: lubricantsModel()}}
	categories := fakeCategories{{ID: "CAT-1", Name: "Asset", ActivityTypeIDs: []string{"AT-Asset"}}}
	s := newFixtureWithCategories(sc, goals, categories)

	got, err := s.RankedRevenue(t.Context(), priya, measure.Filter{Show: "activity"},
		revenue.RankedQuery{Sort: revenue.SortProjectedRevenue, Order: revenue.OrderDesc, Size: 10, Page: 1})
	if err != nil {
		t.Fatalf("RankedRevenue: %v", err)
	}
	if got.Total != 2 {
		t.Errorf("total = %d, want 2: with no category chosen, every activity counts exactly once", got.Total)
	}
}
