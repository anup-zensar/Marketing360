package dataset

import (
	"context"
	"fmt"
	"testing"

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
	same(t, "labels", []string{"Date", "Event", "Leads", "Agency cost"}, labels)
	same(t, "kinds", []Kind{KindText, KindDate, KindNumber, KindMoney}, kinds)
	same(t, "name and version", []any{"Event leads", 2}, []any{view.Name, view.Version})
	entries, _, _ := f.history.ListLog(t.Context(), priya, history.LogFilter{RecordID: created.ID, Limit: 10})
	same(t, "logged changes", []string{"Date, Event, Leads, Agency cost", "Event leads"},
		[]string{entries[0].Changes[0].NewText, entries[0].Changes[1].NewText})
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

	same(t, "refusals", []int{400, 403, 409, 404}, []int{statusOf(notHers), statusOf(bySara), statusOf(stale), statusOf(asMei)})
	same(t, "shared", []any{1, "Lubricants Europe", 2}, []any{len(shared.SharedWith), shared.SharedWith[0].Name, shared.Version})
	same(t, "as Sara", []any{false, false, 0}, []any{asSara.IsOwner, asSara.Access.Edit.Available, len(asSara.ShareableTeams)})
	entries, _, _ := f.history.ListLog(t.Context(), priya, history.LogFilter{RecordID: created.ID, Limit: 10})
	same(t, "log", []string{"Shared with Lubricants Europe", "Lubricants Europe"},
		[]string{entries[0].Details, entries[0].Changes[0].NewText})
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
	restored, restoreErr := f.history.Restore(t.Context(), priya, entry.ID)
	back, _ := f.service.Get(t.Context(), sara, created.ID)
	rows, _ := f.store.Rows(t.Context(), created.ID)

	same(t, "refusals", []any{403, 404, errs.CodeNameTaken, true}, []any{statusOf(bySara), statusOf(gone), codeOf(undo), restoreErr == nil})
	same(t, "restored", []any{"Leads (restored)", "Leads (restored)", 2, 1},
		[]any{back.Name, restored.RecordName, len(rows), len(back.SharedWith)})
}

func TestRestoredName_StaysWithin80Characters_FR043(t *testing.T) {
	t.Parallel()
	long := fmt.Sprintf("%080d", 0)
	got := restoredName([]string{long[:66] + " (restored)"}, long)
	if want := long[:66] + " (restored 2)"; got != want {
		t.Errorf("restoredName() = %q, want %q", got, want)
	}
}
