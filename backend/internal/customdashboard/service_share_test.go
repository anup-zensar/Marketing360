package customdashboard

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/errs"
)

func TestShare_OwnTeamsOnly_FR046(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	view := f.create(t, priya, "Overview")

	_, err := f.service.Share(t.Context(), priya, view.ID, SharingInput{Version: &view.Version, TeamIDs: []string{"T-1002"}})
	assertCode(t, err, http.StatusBadRequest, errs.CodeValidationFailed)
	shared := f.share(t, view, "T-1001", "T-1001")
	f.write(t, func(ctx context.Context) error { return f.members.Delete(ctx, "TM-1001") })
	kept, err := f.service.Share(t.Context(), priya, view.ID, SharingInput{Version: &shared.Version, TeamIDs: []string{"T-1001"}})
	if err != nil {
		t.Fatalf("keeping a team Priya left: %v", err)
	}
	_, err = f.service.Share(t.Context(), priya, view.ID, SharingInput{Version: &view.Version, TeamIDs: []string{}})
	assertCode(t, err, http.StatusConflict, errs.CodeVersionConflict)

	got := []any{len(shared.SharedWith), shared.SharedWith[0].Name, shared.Version, kept.Version, len(kept.ShareableTeams)}
	if diff := cmp.Diff([]any{1, "Lubricants Europe", 2, 2, 1}, got); diff != "" {
		t.Errorf("sharing mismatch (-want +got):\n%s", diff)
	}
	entries := f.log(t)
	if last := entries[len(entries)-1]; last.Details != "Shared with Lubricants Europe" {
		t.Errorf("log details = %q, want the team shared with", last.Details)
	}
	f.write(t, func(ctx context.Context) error { return f.join(ctx, "TM-1001", "T-1001", priya.ID) })
	if _, err := f.service.Share(t.Context(), priya, view.ID, SharingInput{Version: &kept.Version, TeamIDs: []string{}}); err != nil {
		t.Fatalf("unsharing: %v", err)
	}
	entries = f.log(t)
	if last := entries[len(entries)-1]; last.Details != "No longer shared with Lubricants Europe" {
		t.Errorf("log details = %q", last.Details)
	}
}

func TestShare_FormerMemberLosesSight_US4(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	view := f.share(t, f.create(t, priya, "Overview"), "T-1001")
	if _, err := f.service.Get(t.Context(), daniel, view.ID); err != nil {
		t.Fatalf("Daniel, a member: %v", err)
	}

	f.write(t, func(ctx context.Context) error { return f.members.Delete(ctx, "TM-1002") })

	_, err := f.service.Get(t.Context(), daniel, view.ID)
	assertCode(t, err, http.StatusNotFound, errs.CodeNotFound)
}

func TestCopy_NameAndOwner_FR011(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	long := strings.Repeat("L", 78)
	original := f.share(t, f.create(t, priya, long, validBar("w1", 0, 0)), "T-1001")

	first, err := f.service.Copy(t.Context(), daniel, original.ID, CopyInput{})
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}
	second, err := f.service.Copy(t.Context(), daniel, original.ID, CopyInput{})
	if err != nil {
		t.Fatalf("second Copy() error: %v", err)
	}
	named, err := f.service.Copy(t.Context(), daniel, original.ID, CopyInput{Name: ptr(" Daniel's view ")})
	if err != nil {
		t.Fatalf("named Copy() error: %v", err)
	}

	got := []any{first.Name, second.Name, named.Name, first.Owner.Name, len(first.SharedWith), len(first.Widgets),
		first.Access.Edit.Available}
	want := []any{"Copy of " + strings.Repeat("L", 72), "Copy of " + strings.Repeat("L", 68) + " (2)", "Daniel's view",
		"Daniel Okoro", 0, 1, true}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("copies mismatch (-want +got):\n%s", diff)
	}
	entries := f.log(t)
	if last := entries[len(entries)-1]; last.Details != "Copied from '"+long+"' ("+original.ID+")" || last.Actor != daniel {
		t.Errorf("log = %q by %s", last.Details, last.Actor.Name)
	}
	_, err = f.service.Copy(t.Context(), mei, original.ID, CopyInput{})
	assertCode(t, err, http.StatusNotFound, errs.CodeNotFound)
	_, err = f.service.Copy(t.Context(), daniel, original.ID, CopyInput{Name: ptr("daniel's VIEW")})
	assertCode(t, err, http.StatusConflict, errs.CodeNameTaken)
}

func TestUsingDataset_VisibleDashboardsOnly_FR045(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	reads := validBar("w1", 0, 0)
	reads.Query.Source = "DS-1001"
	f.create(t, priya, "Leads", reads)
	f.create(t, priya, "Costs", validBar("w1", 0, 0))

	mine, err := f.service.UsingDataset(t.Context(), priya, "DS-1001")
	if err != nil {
		t.Fatalf("UsingDataset() error: %v", err)
	}
	others, _ := f.service.UsingDataset(t.Context(), daniel, "DS-1001")
	if diff := cmp.Diff([]any{[]Ref{{ID: "DB-1001", Name: "Leads"}}, []Ref{}}, []any{mine, others}); diff != "" {
		t.Errorf("UsingDataset mismatch (-want +got):\n%s", diff)
	}
}
