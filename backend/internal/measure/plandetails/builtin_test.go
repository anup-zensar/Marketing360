package plandetails

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/measure"
)

func widgetNamed(t *testing.T, widgets []Widget, key string) Widget {
	t.Helper()
	for _, w := range widgets {
		if w.Key == key {
			return w
		}
	}
	t.Fatalf("no widget %s among %d", key, len(widgets))
	return Widget{}
}

func typed(id, status string) func(*allwork.Node) {
	return func(n *allwork.Node) { n.ActivityTypeID, n.Status = id, status }
}

func TestBuiltIn_Activity_FR006(t *testing.T) {
	t.Parallel()
	y := newYear(t).plan("P-1").campaign("C-1", "P-1").
		activities(1, 2, "C-1", typed("AT-1", "created")).activities(3, 1, "C-1", typed("AT-1", "approved")).
		activity("A-4", "C-1", "500.00", typed("AT-2", "approved"))
	m := y.matching()
	m.ActivityTypes = []allwork.Option{{ID: "AT-1", Label: "Event"}, {ID: "AT-2", Label: "Digital"}}

	widgets := newView(m, LevelActivity, measure.PeriodMonth).builtIn()

	byType := widgetNamed(t, widgets, "activityTypeCount")
	if diff := cmp.Diff(map[string]int{"AT-1": 3, "AT-2": 1}, counts(byType)); diff != "" || byType.Values[0].Label != "Event" {
		t.Errorf("Activities by activity type (-want +got):\n%s first %+v", diff, byType.Values[0])
	}
	cost := widgetNamed(t, widgets, "activityTypePlannedCost")
	if diff := cmp.Diff([]string{"AT-2", "AT-1"}, keys(cost)); diff != "" || cost.Total.PlannedCost.Amount.String() != "800.00" {
		t.Errorf("Planned cost by activity type, largest first (-want +got):\n%s total %+v", diff, cost.Total.PlannedCost)
	}
	status := widgetNamed(t, widgets, "statusCount")
	if diff := cmp.Diff(map[string]int{"created": 2, "approved": 2}, counts(status)); diff != "" {
		t.Errorf("Activities by status (-want +got):\n%s", diff)
	}
	if status.FilterParam == nil || *status.FilterParam != FilterActivityStatus || !status.Values[0].Linkable {
		t.Errorf("status values open All work by activity status, got %+v", status)
	}
}

func TestStartPeriods_MonthsQuarters_FR006(t *testing.T) {
	t.Parallel()
	y := newYear(t).plan("P-1").campaign("C-1", "P-1").
		activity("A-1", "C-1", "1.00", starting("2026-04-10")).activity("A-2", "C-1", "1.00", starting("2026-05-01")).
		activity("A-3", "C-1", "1.00", starting("2026-11-30")).activity("A-4", "C-1", "1.00", starting("2025-12-01")).
		activity("A-5", "C-1", "1.00")

	months := widgetNamed(t, newView(y.matching(), LevelActivity, measure.PeriodMonth).builtIn(), "startPeriodCount")
	quarters := widgetNamed(t, newView(y.matching(), LevelActivity, measure.PeriodQuarter).builtIn(), "startPeriodCount")

	wantMonths := map[string]int{KeyOutsideYear: 1, KeyNoDates: 1}
	for m := 1; m <= 12; m++ {
		wantMonths[fmt.Sprintf("2026-%02d", m)] = 0
	}
	wantMonths["2026-04"], wantMonths["2026-05"], wantMonths["2026-11"] = 1, 1, 1
	if diff := cmp.Diff(wantMonths, counts(months)); diff != "" || months.Values[0].Label != "Jan 2026" ||
		months.Total.Count != 5 {
		t.Errorf("Activities by start month, first %q, total %d (-want +got):\n%s", months.Values[0].Label,
			months.Total.Count, diff)
	}
	wantQuarters := []string{"2026-Q1", "2026-Q2", "2026-Q3", "2026-Q4", KeyOutsideYear, KeyNoDates}
	if diff := cmp.Diff(wantQuarters, keys(quarters)); diff != "" || quarters.Title != "Activities by start quarter" {
		t.Errorf("Activities by start quarter %q (-want +got):\n%s", quarters.Title, diff)
	}
}

func TestPlannedCostEntered_FR006(t *testing.T) {
	t.Parallel()
	y := newYear(t).plan("P-1").campaign("C-1", "P-1").activities(1, 26, "C-1").
		activity("A-27", "C-1", "").activity("A-28", "C-1", "0.00")

	entered := widgetNamed(t, newView(y.matching(), LevelActivity, measure.PeriodMonth).builtIn(), "plannedCostEntered")

	if diff := cmp.Diff(map[string]int{KeyYes: 26, KeyNo: 2}, counts(entered)); diff != "" {
		t.Errorf("Planned cost entered (-want +got):\n%s", diff)
	}
}

func TestGoalEntered_FR006(t *testing.T) {
	t.Parallel()
	five, three := int64(5), int64(3)
	y := newYear(t).plan("P-1").campaign("C-1", "P-1").
		activity("A-1", "C-1", "", func(n *allwork.Node) { n.INQ = &five }).
		activity("A-2", "C-1", "", func(n *allwork.Node) { n.MQL = &three }).
		activity("A-3", "C-1", "", func(n *allwork.Node) { n.Revenue = amount(t, "10.00") }).
		activity("A-4", "C-1", "").activity("A-5", "C-1", "", func(n *allwork.Node) { n.Revenue = amount(t, "0.00") })

	goal := widgetNamed(t, newView(y.matching(), LevelActivity, measure.PeriodMonth).builtIn(), "goalEntered")

	if diff := cmp.Diff(map[string]int{KeyYes: 3, KeyNo: 2}, counts(goal)); diff != "" {
		t.Errorf("Goal entered (-want +got):\n%s", diff)
	}
}
