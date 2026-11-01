package approvalreport_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure/approvalreport"
	"mapp/internal/measure/approvalreport/approvalreporttest"
)

func allRows(t *testing.T, f approvalreporttest.Fixture, q approvalreport.Query) []approvalreport.Row {
	t.Helper()
	q.PageSize = 100
	return reportOf(t, f, q).Grid.Items
}

func keyOf(r approvalreport.Row) string {
	who := "-"
	if r.Approver != nil {
		who = r.Approver.ID
	}
	return fmt.Sprintf("%s|%s|%s|%d", r.ItemID, who, r.Outcome, r.Level)
}

func TestRows_041_FR012_C2_C3(t *testing.T) {
	t.Parallel()

	rows := allRows(t, approvalreporttest.New(t), approvalreport.Query{})

	got := make([]string, 0, len(rows))
	for _, r := range rows {
		got = append(got, keyOf(r))
	}
	slices.Sort(got)
	want := []string{
		"act-approved-any|p-bob|approved|1",
		"act-approved-any|p-carol|notNeeded|1",
		"act-automatic|-|autoApproved|1",
		"act-cancelled|p-alice|recalled|1",
		"act-declined-all|p-dan|notNeeded|1",
		"act-declined-all|p-erin|declined|1",
		"act-recalled|p-alice|recalled|1",
		"act-replaced|p-alice|replaced|1",
		"act-skipped-level|p-bob|approved|2",
		"act-waiting|p-alice|waiting|1",
		"pr-waiting|p-carol|waiting|1",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("rows: one per named approver per level, no skipped level, none of another kind (-want +got):\n%s", diff)
	}
}

func rowsByKey(t *testing.T) map[string]approvalreport.Row {
	t.Helper()
	byKey := map[string]approvalreport.Row{}
	for _, r := range allRows(t, approvalreporttest.New(t), approvalreport.Query{}) {
		byKey[keyOf(r)] = r
	}
	return byKey
}

func TestRows_DecidedRow_041_FR012(t *testing.T) {
	t.Parallel()

	approved := rowsByKey(t)["act-approved-any|p-bob|approved|1"]

	if approved.DecidedOn == nil || approved.DecidedOn.String() != "2026-03-12" || approved.Comments != "Fine by me" {
		t.Errorf("decided row = %+v, want decided on 2026-03-12 with its comment", approved)
	}
	if approved.SubmittedOn.String() != "2026-03-10" || approved.LevelName != "Level 1" {
		t.Errorf("SubmittedOn, LevelName = %s, %s", approved.SubmittedOn, approved.LevelName)
	}
}

func TestRows_NoDecision_041_FR019(t *testing.T) {
	t.Parallel()
	byKey := rowsByKey(t)

	for _, key := range []string{"act-waiting|p-alice|waiting|1", "act-approved-any|p-carol|notNeeded|1"} {
		if byKey[key].DecidedOn != nil {
			t.Errorf("%s decided on %v, want none: shown as a dash, never a date", key, byKey[key].DecidedOn)
		}
	}
}

func TestRows_AutomaticAndDeclined_041_C2_C3(t *testing.T) {
	t.Parallel()
	byKey := rowsByKey(t)

	automatic := byKey["act-automatic|-|autoApproved|1"]
	declined := byKey["act-declined-all|p-erin|declined|1"]

	if automatic.Approver != nil || automatic.DecidedOn == nil || automatic.DecidedOn.String() != "2026-03-16" {
		t.Errorf("automatic row = %+v, want no approver, decided on the submission date", automatic)
	}
	if declined.Comments != "Out of budget" {
		t.Errorf("declined row comments = %q, want the full comment of the decider", declined.Comments)
	}
}

func TestRows_PurchaseRequest_041_C11(t *testing.T) {
	t.Parallel()
	byKey := map[string]approvalreport.Row{}
	for _, r := range allRows(t, approvalreporttest.New(t), approvalreport.Query{}) {
		byKey[keyOf(r)] = r
	}

	request := byKey["pr-waiting|p-carol|waiting|1"]

	if request.PurchaseOrderNumber == nil || *request.PurchaseOrderNumber != "4500001, 4500002" {
		t.Errorf("PurchaseOrderNumber = %v, want the distinct numbers joined in line order", request.PurchaseOrderNumber)
	}
	if request.ItemName != "Event stand" || request.Owner.Name != "Erin Poe" || request.RecordType != approvalreport.RecordPurchaseRequest {
		t.Errorf("request row = %+v, want its name, owner and type", request)
	}
	if got := byKey["act-waiting|p-alice|waiting|1"].PurchaseOrderNumber; got != nil {
		t.Errorf("an activity has no purchase order, got %v", got)
	}
}
