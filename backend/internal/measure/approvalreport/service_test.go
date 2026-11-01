package approvalreport_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/approval"
	"mapp/internal/measure"
	"mapp/internal/measure/approvalreport"
	"mapp/internal/measure/approvalreport/approvalreporttest"
)

func reportOf(t *testing.T, f approvalreporttest.Fixture, q approvalreport.Query) approvalreport.View {
	t.Helper()
	view, err := approvalreport.NewService(f.DB, f.Deps).Report(t.Context(), approvalreporttest.Alice, measure.Filter{}, q)
	if err != nil {
		t.Fatalf("Report() error: %v", err)
	}
	return view
}

func TestStatusMix_041_FR006_FR007(t *testing.T) {
	t.Parallel()
	mix := reportOf(t, approvalreporttest.New(t), approvalreport.Query{}).StatusMix

	want := []approvalreport.StatusCount{
		{Outcome: approvalreport.OutcomeWaiting, Count: 1, SharePercent: 13},
		{Outcome: approvalreport.OutcomeApproved, Count: 2, SharePercent: 25},
		{Outcome: approvalreport.OutcomeDeclined, Count: 1, SharePercent: 13},
		{Outcome: approvalreport.OutcomeRecalled, Count: 2, SharePercent: 25},
		{Outcome: approvalreport.OutcomeReplaced, Count: 1, SharePercent: 12},
		{Outcome: approvalreport.OutcomeAutoApproved, Count: 1, SharePercent: 12},
	}
	if diff := cmp.Diff(want, mix.Items); diff != "" {
		t.Errorf("Items (-want +got):\n%s", diff)
	}
	if mix.RecordType != approvalreport.RecordActivity || mix.Total != 8 {
		t.Errorf("RecordType, Total = %q, %d, want activity, 8", mix.RecordType, mix.Total)
	}
	wantTypes := []approvalreport.RecordType{approvalreport.RecordActivity, approvalreport.RecordPurchaseRequest}
	if diff := cmp.Diff(wantTypes, mix.RecordTypes); diff != "" {
		t.Errorf("RecordTypes only those with rounds (-want +got):\n%s", diff)
	}
}

func TestStatusMix_UnknownAndEmptyTypes_041_C10(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)

	for _, statusType := range []approvalreport.RecordType{"nonsense", approvalreport.RecordModifiedPurchaseRequest} {
		mix := reportOf(t, f, approvalreport.Query{StatusType: statusType}).StatusMix

		if mix.Total != 0 || mix.Waiting.Count != 0 || mix.Waiting.OldestSince != nil {
			t.Errorf("statusType %q: Total, Waiting = %d, %+v, want empty", statusType, mix.Total, mix.Waiting)
		}
		if len(mix.Items) != 6 || mix.Items[0].SharePercent != 0 {
			t.Errorf("statusType %q: Items = %+v, want six zero counts", statusType, mix.Items)
		}
	}
}

func TestWaiting_ByType_041_FR008(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	sixDaysAgo, twoDaysAgo := f.Now.AddDate(0, 0, -6), f.Now.AddDate(0, 0, -2)

	tests := []struct {
		name  string
		q     approvalreport.Query
		count int
		since time.Time
	}{
		{name: "activities", q: approvalreport.Query{}, count: 1, since: sixDaysAgo},
		{name: "purchase requests", q: approvalreport.Query{StatusType: approvalreport.RecordPurchaseRequest}, count: 1, since: twoDaysAgo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := reportOf(t, f, tt.q).StatusMix.Waiting
			if got.Count != tt.count || got.OldestSince == nil || !got.OldestSince.Equal(tt.since) {
				t.Errorf("Waiting = %+v, want %d since %v", got, tt.count, tt.since)
			}
		})
	}
}

type askedRounds struct {
	approvalreport.Rounds
	subjects []string
	calls    int
}

func (a *askedRounds) RoundsOfSubjects(ctx context.Context, ids []string) ([]approval.Round, error) {
	a.calls++
	a.subjects = append(a.subjects, ids...)
	return a.Rounds.RoundsOfSubjects(ctx, ids)
}

func TestPurchaseRequestsInView_041_C8(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	f.Scope.Matched.LineItems = []string{}
	f.Deps.Scope = approvalreporttest.Scoper{Result: f.Scope}

	mix := reportOf(t, f, approvalreport.Query{StatusType: approvalreport.RecordPurchaseRequest}).StatusMix

	if mix.Total != 0 {
		t.Errorf("Total = %d, want 0: its line item is not in view", mix.Total)
	}
}

func TestScope_PerPersona_041_FR003(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	f.Scope.Matched.Activities = slices.DeleteFunc(slices.Clone(f.Scope.Matched.Activities),
		func(id string) bool { return id == approvalreporttest.ActivitySkippedLevel })
	f.Scope.Matched.Plans = []string{approvalreporttest.PlanWithFinance}
	asked := &askedRounds{Rounds: f.Deps.Rounds}
	f.Deps.Rounds, f.Deps.Scope = asked, approvalreporttest.Scoper{Result: f.Scope}

	view := reportOf(t, f, approvalreport.Query{})

	if slices.Contains(asked.subjects, approvalreporttest.ActivitySkippedLevel) || asked.calls != 1 {
		t.Errorf("rounds asked in %d calls for %v, want one call without the activity outside the scope", asked.calls, asked.subjects)
	}
	if view.StatusMix.Total != 7 {
		t.Errorf("Total = %d, want 7: the round of the other plan is not counted", view.StatusMix.Total)
	}
	if want := (approvalreport.FinanceCovers{Plans: 1, Of: 1}); view.FinanceCovers != want {
		t.Errorf("FinanceCovers = %+v, want %+v", view.FinanceCovers, want)
	}
}

func TestFinanceCovers_041_FR003(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	f.Scope.Choices = allwork.Choices{SeesFinance: false}
	f.Deps.Scope = approvalreporttest.Scoper{Result: f.Scope}

	got := reportOf(t, f, approvalreport.Query{}).FinanceCovers

	if want := (approvalreport.FinanceCovers{Plans: 0, Of: 2}); got != want {
		t.Errorf("FinanceCovers = %+v, want %+v", got, want)
	}
}
