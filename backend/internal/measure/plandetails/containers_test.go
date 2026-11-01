package plandetails

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/measure"
)

func progress(status, timing string) func(*allwork.Node) {
	return func(n *allwork.Node) { n.Status, n.Timing = status, timing }
}

func containersYear(t *testing.T) *year {
	t.Helper()
	return newYear(t).plan("P-1").plan("P-2").on("P-2", func(n *allwork.Node) { n.Status = "published" }).
		campaign("C-1", "P-1", progress("inProgress", "running"), choice("SA-11", "EU")).
		campaign("C-2", "P-2", progress("allApproved", "finished"), choice("SA-11", "AS")).
		campaign("C-3", "P-2", progress("noActivities", ""), choice("SA-11", "EU")).
		initiative("I-1", "C-1", progress("inProgress", "notStarted")).
		activity("A-1", "I-1", "100.00").activity("A-2", "I-1", "200.00").activity("A-3", "C-2", "50.00")
}

func cluster() allwork.Attribute {
	return atLevel(attribute("SA-11", "Cluster", allwork.TypeSingleChoice, "EU", "AS"), allwork.LevelCampaign)
}

func TestBuiltIn_Containers_FR007(t *testing.T) {
	t.Parallel()
	plans := newView(containersYear(t).matching(), LevelPlan, measure.PeriodMonth).builtIn()
	campaigns := newView(containersYear(t).matching(), LevelCampaign, measure.PeriodMonth).builtIn()

	want := shape{Title: "Plans by status", Keys: []string{"draft", "published"},
		Counts: map[string]int{"draft": 1, "published": 1}, Total: 2}
	if diff := cmp.Diff([]shape{want}, []shape{shapeOf(plans[0])}); diff != "" || len(plans) != 1 {
		t.Errorf("Plan level (-want +got):\n%s", diff)
	}
	byProgress := shape{Title: "Campaigns by approval progress", Keys: []string{"noActivities", "inProgress", "allApproved"},
		Counts: map[string]int{"noActivities": 1, "inProgress": 1, "allApproved": 1}, Total: 3}
	byTiming := shape{Title: "Campaigns by timing", Keys: []string{"running", "finished", KeyNoDates},
		Counts: map[string]int{"running": 1, "finished": 1, KeyNoDates: 1}, Total: 3}
	if diff := cmp.Diff([]shape{byProgress, byTiming}, []shape{shapeOf(campaigns[0]), shapeOf(campaigns[1])}); diff != "" {
		t.Errorf("Campaign level, in lifecycle order (-want +got):\n%s", diff)
	}
	if campaigns[0].Values[1].Label != "Partly approved" || campaigns[0].FilterParam != nil {
		t.Errorf("progress values %+v, want 016's words and no All work link", campaigns[0].Values)
	}
}

func TestCards_PlannedCostFromMatchingActivities_FR005(t *testing.T) {
	t.Parallel()
	cards := newView(containersYear(t).unmatch("A-2").matching(), LevelCampaign, measure.PeriodMonth).cards()

	if cards.Records != 3 || cards.PlannedCost.Amount.String() != "150.00" {
		t.Errorf("cards = %+v, want three campaigns with their matching activities' 150.00", cards)
	}
}

func TestRows_OtherLevels_FR008(t *testing.T) {
	t.Parallel()
	rows := newView(containersYear(t).matching(), LevelCampaign, measure.PeriodMonth).
		rows([]allwork.Attribute{companyCode(), cluster()})

	if len(rows) != 1 || rows[0].AttributeID != "SA-11" {
		t.Fatalf("rows = %+v, want only the campaign attribute Cluster", rows)
	}
	want := shape{Title: "Campaigns by Cluster", Keys: []string{"EU", "AS"}, Counts: map[string]int{"EU": 2, "AS": 1},
		Total: 3}
	if diff := cmp.Diff(want, shapeOf(rows[0].Widgets[0])); diff != "" {
		t.Errorf("Campaigns by Cluster (-want +got):\n%s", diff)
	}
	cost := map[string]string{}
	for _, v := range rows[0].Widgets[1].Values {
		cost[v.Key] = v.PlannedCost.Amount.String()
	}
	if diff := cmp.Diff(map[string]string{"EU": "300.00", "AS": "50.00"}, cost); diff != "" {
		t.Errorf("Planned cost by Cluster from the campaigns' activities (-want +got):\n%s", diff)
	}
}

// One of two plans in scope: the Plan level counts one plan, the others only what lies below it (US2 scenario 7).
func TestLevels_OnlyBelowThePlansInScope_FR003(t *testing.T) {
	t.Parallel()
	y := containersYear(t).unmatch("P-2", "C-2", "C-3", "A-3")

	plans := newView(y.matching(), LevelPlan, measure.PeriodMonth).cards()
	campaigns := newView(y.matching(), LevelCampaign, measure.PeriodMonth).cards()

	if plans.Records != 1 || campaigns.Records != 1 || campaigns.PlannedCost.Amount.String() != "300.00" {
		t.Errorf("plans %+v, campaigns %+v: want P-1 and C-1 with 300.00", plans, campaigns)
	}
}
