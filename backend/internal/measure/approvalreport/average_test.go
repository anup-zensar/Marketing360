package approvalreport_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure/approvalreport"
	"mapp/internal/measure/approvalreport/approvalreporttest"
)

func TestAverage_041_FR009(t *testing.T) {
	t.Parallel()

	got := reportOf(t, approvalreporttest.New(t), approvalreport.Query{}).Average

	if got.Days == nil || *got.Days != "1.3" || got.Decisions != 3 {
		t.Errorf("Days, Decisions = %v, %d, want 1.3 over 3: automatic, open, recalled and replaced rounds are left out",
			got.Days, got.Decisions)
	}
}

func TestAverage_NoDecisions_041_FR011(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	f.Scope.Matched.Activities = []string{approvalreporttest.ActivityWaiting}
	f.Deps.Scope = approvalreporttest.Scoper{Result: f.Scope}

	got := reportOf(t, f, approvalreport.Query{}).Average

	if got.Days != nil || got.Decisions != 0 || len(got.Approvers) != 0 {
		t.Errorf("Average = %+v, want no days and no decisions, never zero", got)
	}
}

func TestApproverAverages_041_FR010(t *testing.T) {
	t.Parallel()

	got := reportOf(t, approvalreporttest.New(t), approvalreport.Query{}).Average.Approvers

	want := []approvalreport.ApproverAverage{
		{Person: approvalreport.PersonRef{ID: "p-bob", Name: "Bob Ruiz"}, Days: "1.5", Decisions: 2},
		{Person: approvalreport.PersonRef{ID: "p-erin", Name: "Erin Poe"}, Days: "1.0", Decisions: 1},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Approvers, slowest first, the decider credited and nobody else (-want +got):\n%s", diff)
	}
}
