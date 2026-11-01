package planbudget_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/planbudget"
)

func report(t *testing.T, sc measure.Scope, f measure.Filter) planbudget.Report {
	t.Helper()
	r, err := service(sc).Report(t.Context(), priya, f)
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	return r
}

func percent(p *int) string {
	if p == nil {
		return "—"
	}
	return fmt.Sprint(*p)
}

func TestCards_FR002(t *testing.T) {
	t.Parallel()
	c := report(t, scope(everything, true), measure.Filter{}).Cards

	got := []string{shown(c.Budget), shown(c.Planned), shown(c.Committed), shown(c.Actual), percent(c.BudgetUsed),
		shown(c.LeftToPlan), shown(c.Unspent), percent(c.BudgetSpent), shown(c.NotYetCommitted), percent(c.CommittedShare)}
	want := []string{"1000.00", "500.00", "400.00", "180.00", "50", "500.00", "820.00", "18", "100.00", "80"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("budget, planned, committed, actual, used %%, left, unspent, spent %%, not yet committed, %% committed (-want +got):\n%s", diff)
	}
}

func TestBudget_OncePerBranch_040_FR003(t *testing.T) {
	t.Parallel()
	onlyHelix := allwork.Matched{Plans: []string{"P-1"}, Campaigns: []string{"C-1"}, Initiatives: []string{"I-1"},
		Activities: []string{"A-1"}, LineItems: []string{"L-1"}}

	r := report(t, scope(onlyHelix, true), measure.Filter{})

	if shown(r.Cards.Budget) != "600.00" || shown(r.Notes.NotPhased) != "600.00" {
		t.Errorf("budget %s, not phased %s: with only Helix launch in view its own budget counts, not the plan's",
			shown(r.Cards.Budget), shown(r.Notes.NotPhased))
	}
	if whole := report(t, scope(everything, true), measure.Filter{}); whole.Notes.NotPhased != nil {
		t.Errorf("not phased = %s: the whole plan counts its phased budget", shown(whole.Notes.NotPhased))
	}
}

func TestCards_FinanceOnlyVisiblePlans_FR006(t *testing.T) {
	t.Parallel()
	r := report(t, scope(everything, false), measure.Filter{})

	c := r.Cards
	if c.Committed != nil || c.Actual != nil || c.Unspent != nil || c.BudgetSpent != nil || c.NotYetCommitted != nil {
		t.Errorf("cards %+v: without finance, committed, actual and what comes from them are —", c)
	}
	if shown(c.Planned) != "500.00" || r.Periods[1].Actual != nil {
		t.Errorf("planned %s, February actual %s: planning figures stay, actual is not shown", shown(c.Planned),
			shown(r.Periods[1].Actual))
	}
}

func TestPeriods_FR004_FR007(t *testing.T) {
	t.Parallel()
	r := report(t, scope(everything, true), measure.Filter{})

	var got []string
	for _, p := range r.Periods {
		got = append(got, fmt.Sprintf("%s %s b%s p%s a%s", p.Key, p.State, p.Budget.Amount, p.Planned.Amount, shown(p.Actual)))
	}
	want := []string{"2026-01 actual b500.00 p0.00 a0.00", "2026-02 actual b500.00 p0.00 a100.00",
		"2026-03 actual b0.00 p300.00 a0.00", "2026-04 actual b0.00 p0.00 a0.00", "2026-05 actual b0.00 p0.00 a0.00",
		"2026-06 actual b0.00 p0.00 a0.00", "2026-07 actual b0.00 p0.00 a0.00", "2026-08 actual b0.00 p0.00 a0.00",
		"2026-09 actual b0.00 p0.00 a0.00", "2026-10 inProgress b0.00 p0.00 a80.00", "2026-11 planned b0.00 p100.00 a—",
		"2026-12 planned b0.00 p100.00 a—"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("months (-want +got):\n%s", diff)
	}
	last := r.Periods[11]
	if last.BudgetToDate.Amount.String() != "1000.00" || last.PlannedToDate.Amount.String() != "500.00" {
		t.Errorf("December to date: budget %s, planned %s", last.BudgetToDate.Amount, last.PlannedToDate.Amount)
	}
	quarters := report(t, scope(everything, true), measure.Filter{Period: measure.PeriodQuarter}).Periods
	if len(quarters) != 4 || quarters[3].Key != "2026-Q4" || quarters[3].State != planbudget.StateInProgress ||
		shown(quarters[3].Actual) != "80.00" {
		t.Errorf("quarters = %+v, want four with Q4 in progress and its actual so far", quarters)
	}
}

func TestPeriods_UndatedAndOtherYearsNamed_FR005(t *testing.T) {
	t.Parallel()
	sc := scope(everything, true)
	sc.Nodes = append(sc.Nodes,
		allwork.Node{Level: allwork.LevelActivity, ID: "A-3", ParentID: "I-2", PlanID: "P-1", PlannedCost: dollars(110)},
		allwork.Node{Level: allwork.LevelLineItem, ID: "L-3", ParentID: "A-3", PlanID: "P-1", PlannedCost: dollars(70),
			UnspreadUSD: *dollars(70)},
		allwork.Node{Level: allwork.LevelLineItem, ID: "L-4", ParentID: "A-3", PlanID: "P-1", PlannedCost: dollars(40),
			MonthsUSD: []allwork.MonthAmount{month("2027-01", 40)}})
	sc.Matched.Activities = slices.Concat(everything.Activities, []string{"A-3"})
	sc.Matched.LineItems = slices.Concat(everything.LineItems, []string{"L-3", "L-4"})

	r := report(t, sc, measure.Filter{})

	if shown(r.Notes.NoDates) != "70.00" || shown(r.Notes.OtherYears) != "40.00" {
		t.Errorf("notes = no dates %s, other years %s; want 70.00 and 40.00", shown(r.Notes.NoDates),
			shown(r.Notes.OtherYears))
	}
}

func TestPeriods_TodayAndNotes_FR004_FR005(t *testing.T) {
	t.Parallel()
	r := report(t, scope(everything, true), measure.Filter{})

	if shown(r.Notes.ActualOtherYears) != "50.00" || r.Notes.NoDates != nil || r.Notes.OtherYears != nil {
		t.Errorf("notes = %+v: the December 2025 payment is named, nothing else", r.Notes)
	}
	if r.Periods[10].ActualToDate != nil || shown(r.Periods[9].ActualToDate) != "180.00" {
		t.Errorf("actual to date stops at October: %s then %s", shown(r.Periods[9].ActualToDate),
			shown(r.Periods[10].ActualToDate))
	}
}
