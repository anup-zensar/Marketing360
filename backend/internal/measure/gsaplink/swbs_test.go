package gsaplink

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
)

func swbsIDs(rows []SwbsRow) []string {
	ids := []string{}
	for _, r := range rows {
		ids = append(ids, r.LineItemID)
	}
	return ids
}

func swbsOf(w *world) SwbsTab {
	y := w.year()
	ready := y.readyToSend()
	approved, _ := w.ActivitiesApprovedAt(w.tb.Context(), activitiesOf(ready))
	return y.swbsTab(ready, approved)
}

func TestSwbs_ReadyToSendOnly_FR005(t *testing.T) {
	t.Parallel()
	w := twoPlans(t).
		on("L-2", func(n *allwork.Node) { n.Swbs = "received" }).
		on("L-3", func(n *allwork.Node) { n.Swbs = "waitingForApproval" }).
		on(ukActivity, func(n *allwork.Node) { n.Status = "inProgress" }).
		add(allwork.LevelLineItem, "L-4", ukActivity, "Catering").
		on("L-4", func(n *allwork.Node) { n.Swbs = "notRequested" })
	w.approved[ukActivity] = now.Add(-time.Hour)
	if diff := cmp.Diff([]string{"L-1"}, swbsIDs(swbsOf(w).Rows)); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}

func TestSwbs_Row_FR005(t *testing.T) {
	t.Parallel()
	edited := now.Add(-30 * time.Hour)
	w := twoPlans(t).unmatch("L-2").on(deActivity, func(n *allwork.Node) { n.UpdatedAt = edited })
	w.approved[ukActivity] = now.Add(-72 * time.Hour)
	want := []SwbsRow{
		{LineItemID: "L-1", LineItem: "Stand build", Path: Path{PlanID: ukPlan, Plan: "UK plan", Campaign: "Launch",
			Initiative: "Retail", ActivityID: ukActivity, Activity: "Stand"}, Owner: "Priya Raman", CompanyCode: "GB01",
			ApprovedAt: now.Add(-72 * time.Hour), WaitingHours: 72, OverTurnaround: true},
		{LineItemID: "L-3", LineItem: "Booth", Path: Path{PlanID: dePlan, Plan: "Germany plan", Campaign: "Fleet",
			Initiative: "Trucks", ActivityID: deActivity, Activity: "Expo"}, ApprovedAt: edited, WaitingHours: 30},
	}
	if diff := cmp.Diff(want, swbsOf(w).Rows); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}

func TestSwbsBands_FR007(t *testing.T) {
	t.Parallel()
	w := twoPlans(t)
	w.approved[ukActivity], w.approved[deActivity] = now.Add(-10*time.Hour), now.AddDate(0, 0, -11)
	tab := swbsOf(w)
	want := []Band{{Key: BandUnder48h, Label: "Under 48 hours", Count: 2}, {Key: BandDays2to5, Label: "2 – 5 days"},
		{Key: BandDays5to10, Label: "5 – 10 days"}, {Key: BandOver10Days, Label: "Over 10 days", Count: 1}}
	if diff := cmp.Diff(want, tab.Bands); diff != "" {
		t.Errorf("bands (-want +got):\n%s", diff)
	}
	if tab.Waiting != 3 || tab.OverTurnaround != 1 {
		t.Errorf("cards = Waiting %d, Over 48 hours %d; want 3, 1", tab.Waiting, tab.OverTurnaround)
	}
}

// TestSwbs_EqualsAllWork_SC002 matches with All work's own Match: the tab lists exactly its Ready to send line items.
func TestSwbs_EqualsAllWork_SC002(t *testing.T) {
	t.Parallel()
	w := twoPlans(t).on("L-2", func(n *allwork.Node) { n.Swbs = "received" })
	sc := w.scope()
	sc.Matched = allwork.Match(sc.Nodes, allwork.Query{Year: testYear, Plans: []string{ukPlan}}, sc.Attributes, sc.Choices)
	y := newYear(sc, now)
	tab := y.swbsTab(y.readyToSend(), nil)
	if diff := cmp.Diff([]string{"L-1"}, swbsIDs(tab.Rows)); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}
