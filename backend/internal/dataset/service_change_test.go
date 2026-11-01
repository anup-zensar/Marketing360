package dataset

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/history"
	"mapp/internal/platform/errs"
)

func TestUpdate_ColumnKindsNeverChange_FR044(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	created := f.create(t, priya, "Leads")

	view, err := f.service.Update(t.Context(), priya, created.ID, UpdateInput{Version: ptr(1), Name: ptr(" Event leads "),
		Columns: []LabelInput{{Position: 0, Label: "Date"}, {Position: 1, Label: "Event"}}})

	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	labels, kinds := []string{}, []Kind{}
	for _, c := range view.Columns {
		labels, kinds = append(labels, c.Label), append(kinds, c.Kind)
	}
	if diff := cmp.Diff([]string{"Date", "Event", "Leads", "Agency cost"}, labels); diff != "" {
		t.Errorf("labels (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]Kind{KindText, KindDate, KindNumber, KindMoney}, kinds); diff != "" {
		t.Errorf("kinds (-want +got):\n%s", diff)
	}
	if view.Name != "Event leads" || view.Version != 2 {
		t.Errorf("view = %q v%d, want Event leads v2", view.Name, view.Version)
	}
	entries, _, _ := f.history.ListLog(t.Context(), priya, history.LogFilter{RecordID: created.ID, Limit: 10})
	if len(entries) != 2 || len(entries[0].Changes) != 2 || entries[0].Changes[0].NewText != "Date, Event, Leads, Agency cost" {
		t.Errorf("log = %+v, want the name and column names changed", entries[0])
	}
}

func TestUpdate_RulesAndRefusals_FR044(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	created := f.create(t, priya, "Leads")
	f.create(t, priya, "Other")
	tests := []struct {
		name   string
		as     string
		in     UpdateInput
		status int
	}{
		{name: "nothing changed keeps the version", as: priya.ID, in: UpdateInput{Version: ptr(1), Name: ptr("Leads")}, status: 200},
		{name: "a stale version", as: priya.ID, in: UpdateInput{Version: ptr(9), Name: ptr("New")}, status: 409},
		{name: "a name the owner uses", as: priya.ID, in: UpdateInput{Version: ptr(1), Name: ptr("other")}, status: 409},
		{name: "two columns of one name", as: priya.ID, in: UpdateInput{Version: ptr(1),
			Columns: []LabelInput{{Position: 1, Label: "EVENT"}}}, status: 400},
		{name: "no such column", as: priya.ID, in: UpdateInput{Version: ptr(1), Columns: []LabelInput{{Position: 4, Label: "X"}}}, status: 400},
		{name: "no version", as: priya.ID, in: UpdateInput{Name: ptr("New")}, status: 400},
		{name: "not shared with Sara", as: sara.ID, in: UpdateInput{Version: ptr(1), Name: ptr("New")}, status: 404},
	}
	for _, tt := range tests {
		a := priya
		if tt.as == sara.ID {
			a = sara
		}
		view, err := f.service.Update(t.Context(), a, created.ID, tt.in)
		if got := statusOf(err); got != tt.status {
			t.Errorf("%s: Update() = %v, want %d", tt.name, err, tt.status)
		}
		if tt.status == 200 && view.Version != 1 {
			t.Errorf("%s: version = %d, want 1", tt.name, view.Version)
		}
	}
}

func statusOf(err error) int {
	switch codeOf(err) {
	case "":
		return 200
	case errs.CodeValidationFailed:
		return 400
	case errs.CodeForbidden:
		return 403
	case errs.CodeNotFound:
		return 404
	}
	return 409
}

func TestShare_OwnTeamsOnly_FR046(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	created := f.create(t, priya, "Leads")

	_, notHers := f.service.Share(t.Context(), priya, created.ID, SharingInput{Version: ptr(1), TeamIDs: []string{china}})
	shared, err := f.service.Share(t.Context(), priya, created.ID, SharingInput{Version: ptr(1), TeamIDs: []string{europe, europe}})
	if err != nil {
		t.Fatalf("Share() error = %v", err)
	}
	_, bySara := f.service.Share(t.Context(), sara, created.ID, SharingInput{Version: ptr(2), TeamIDs: []string{}})
	_, stale := f.service.Share(t.Context(), priya, created.ID, SharingInput{Version: ptr(1), TeamIDs: []string{}})
	asSara, _ := f.service.Get(t.Context(), sara, created.ID)
	_, asMei := f.service.Get(t.Context(), mei, created.ID)

	if got := []int{statusOf(notHers), statusOf(bySara), statusOf(stale), statusOf(asMei)}; !cmp.Equal(got, []int{400, 403, 409, 404}) {
		t.Errorf("refusals = %v, want 400, 403, 409, 404", got)
	}
	if len(shared.SharedWith) != 1 || shared.SharedWith[0].Name != "Lubricants Europe" || shared.Version != 2 {
		t.Errorf("shared = %+v", shared.SharedWith)
	}
	if asSara.IsOwner || asSara.Access.Edit.Available || len(asSara.ShareableTeams) != 0 {
		t.Errorf("as Sara = %+v, want view-only", asSara)
	}
	entries, _, _ := f.history.ListLog(t.Context(), priya, history.LogFilter{RecordID: created.ID, Limit: 10})
	if entries[0].Details != "Shared with Lubricants Europe" || entries[0].Changes[0].NewText != "Lubricants Europe" {
		t.Errorf("log = %+v", entries[0])
	}
}

func TestShare_FormerTeamStaysRemovable_FR046(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	created := f.create(t, mei, "Leads")
	if _, err := f.service.Share(t.Context(), mei, created.ID, SharingInput{Version: ptr(1), TeamIDs: []string{china}}); err != nil {
		t.Fatalf("Share: %v", err)
	}
	if err := f.db.Write(t.Context(), func(ctx context.Context) error { return f.store.memberships.Delete(ctx, "TM-"+mei.ID+china) }); err != nil {
		t.Fatalf("leave: %v", err)
	}
	view, err := f.service.Get(t.Context(), mei, created.ID)
	if err != nil || len(view.ShareableTeams) != 1 {
		t.Fatalf("Get() = %+v, %v; want the former team still offered", view.ShareableTeams, err)
	}
	unshared, err := f.service.Share(t.Context(), mei, created.ID, SharingInput{Version: ptr(2), TeamIDs: []string{}})
	if err != nil || len(unshared.SharedWith) != 0 {
		t.Errorf("Share() = %+v, %v; want private", unshared.SharedWith, err)
	}
}

func TestDelete_ToRecycleBin_FR045(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	created := f.create(t, priya, "Leads")
	if _, err := f.service.Share(t.Context(), priya, created.ID, SharingInput{Version: ptr(1), TeamIDs: []string{europe}}); err != nil {
		t.Fatalf("Share: %v", err)
	}

	_, bySara := f.service.Delete(t.Context(), sara, created.ID)
	entry, err := f.service.Delete(t.Context(), priya, created.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	_, gone := f.service.Get(t.Context(), priya, created.ID)
	f.create(t, priya, "Leads")
	_, undo := f.history.Undo(t.Context(), priya, entry.ID)
	restored, err := f.history.Restore(t.Context(), priya, entry.ID)

	if statusOf(bySara) != 403 || statusOf(gone) != 404 || codeOf(undo) != errs.CodeNameTaken || err != nil {
		t.Fatalf("by Sara %v, gone %v, undo %v, restore %v", bySara, gone, undo, err)
	}
	back, err := f.service.Get(t.Context(), sara, created.ID)
	rows, _ := f.store.Rows(t.Context(), created.ID)
	if err != nil || back.Name != "Leads (restored)" || restored.RecordName != back.Name || len(rows) != 2 || len(back.SharedWith) != 1 {
		t.Errorf("restored = %+v (%v), %d rows", back, err, len(rows))
	}
}

func TestRestoredName_StaysWithin80Characters_FR043(t *testing.T) {
	t.Parallel()
	long := fmt.Sprintf("%080d", 0)
	got := restoredName([]string{long[:66] + " (restored)"}, long)
	if want := long[:66] + " (restored 2)"; got != want {
		t.Errorf("restoredName() = %q, want %q", got, want)
	}
}
