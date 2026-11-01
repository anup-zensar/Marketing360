package revenue_test

import (
	"testing"

	"mapp/internal/measure"
	"mapp/internal/measure/revenue"
)

func TestRevenueReport_SheetsExportEveryWidget_FR007(t *testing.T) {
	t.Parallel()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	s := newFixtureWith(sc, goals, nil)

	sheets, err := revenue.NewRevenueReport(s).Sheets(t.Context(), priya, measure.Filter{})
	if err != nil {
		t.Fatalf("Sheets: %v", err)
	}
	var got []string
	for _, sh := range sheets {
		got = append(got, sh.Key+" "+sh.Sheet.Name)
	}
	want := []string{"cards Cards", "periods Periods", "ranked Ranked list"}
	if len(got) != len(want) {
		t.Fatalf("sheets = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("sheets[%d] = %q, want %q", i, got[i], w)
		}
	}
	// Cards: Figure/Value header + goal, actual, projected, performance.
	if len(sheets[0].Sheet.Rows) != 5 {
		t.Errorf("cards rows = %d, want 5 (header + 4 figures)", len(sheets[0].Sheet.Rows))
	}
	// Periods: header + 12 months.
	if len(sheets[1].Sheet.Rows) != 13 {
		t.Errorf("periods rows = %d, want 13 (header + 12 months)", len(sheets[1].Sheet.Rows))
	}
}

func TestFunnelReport_SheetsExportEveryWidget_FR007(t *testing.T) {
	t.Parallel()
	sc := oneActivityScope()
	goals := fakeGoals{"A-1": revenue.ActivityGoals{Model: lubricantsModel()}}
	s := newFixtureWith(sc, goals, nil)

	sheets, err := revenue.NewFunnelReport(s).Sheets(t.Context(), priya, measure.Filter{})
	if err != nil {
		t.Fatalf("Sheets: %v", err)
	}
	var got []string
	for _, sh := range sheets {
		got = append(got, sh.Key+" "+sh.Sheet.Name)
	}
	want := []string{"cards Cards", "periods Periods", "ranked Ranked list"}
	if len(got) != len(want) {
		t.Fatalf("sheets = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("sheets[%d] = %q, want %q", i, got[i], w)
		}
	}
}
