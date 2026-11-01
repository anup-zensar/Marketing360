package customdashboard

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/history"
	"mapp/internal/platform/errs"
)

func TestCreate_OwnedUnsharedLogged_FR008_FR009(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	view, err := f.service.Create(t.Context(), priya, CreateInput{
		Name: "  Lubricants overview ", Description: "For the team", Widgets: []Widget{validBar("w1", 0, 0)},
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got := []any{view.ID, view.Name, view.Owner.Name, view.IsOwner, len(view.SharedWith), view.Version,
		view.CreatedBy, len(view.Widgets), view.Access.Edit.Available, len(view.ShareableTeams)}
	want := []any{"DB-1001", "Lubricants overview", "Priya Raman", true, 0, 1, priya, 1, true, 1}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("created dashboard mismatch (-want +got):\n%s", diff)
	}
	if entries := f.log(t); len(entries) != 1 || entries[0].Action != history.ActionCreated ||
		entries[0].RecordID != "DB-1001" || entries[0].Actor != priya {
		t.Errorf("log = %+v, want one created entry by Priya", entries)
	}
	_, err = f.service.Create(t.Context(), priya, CreateInput{Name: "LUBRICANTS OVERVIEW"})
	assertCode(t, err, http.StatusConflict, errs.CodeNameTaken)
	if _, err = f.service.Create(t.Context(), daniel, CreateInput{Name: "Lubricants overview"}); err != nil {
		t.Errorf("another owner's same name: %v", err)
	}
	_, err = f.service.Create(t.Context(), priya, CreateInput{Name: ""})
	assertCode(t, err, http.StatusBadRequest, errs.CodeValidationFailed)
}

func TestList_MineAndSharedWithMe_FR003_FR005(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.share(t, f.create(t, priya, "Shared overview"), "T-1001")
	f.create(t, priya, "Private notes")
	f.create(t, daniel, "Approvals")

	priyas, err := f.service.List(t.Context(), priya)
	if err != nil {
		t.Fatalf("List(priya) error: %v", err)
	}
	daniels, _ := f.service.List(t.Context(), daniel)
	meis, _ := f.service.List(t.Context(), mei)

	got := []any{names(priyas.Mine), names(priyas.SharedWithMe), names(daniels.Mine), names(daniels.SharedWithMe),
		names(meis.Mine), names(meis.SharedWithMe), daniels.SharedWithMe[0].Access.Edit.Reason}
	want := []any{[]string{"Private notes", "Shared overview"}, []string{}, []string{"Approvals"},
		[]string{"Shared overview"}, []string{}, []string{}, "Only Priya Raman can change this dashboard."}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("lists mismatch (-want +got):\n%s", diff)
	}
}

func names(list []Summary) []string {
	out := []string{}
	for _, s := range list {
		out = append(out, s.Name)
	}
	return out
}

func TestUpdate_NotOwner403_Invisible404_FR006(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	view := f.share(t, f.create(t, priya, "Overview"), "T-1001")
	rename := UpdateInput{Version: &view.Version, Name: ptr("Mine now")}

	_, err := f.service.Update(t.Context(), daniel, view.ID, rename)
	assertCode(t, err, http.StatusForbidden, errs.CodeForbidden)
	_, err = f.service.Update(t.Context(), mei, view.ID, rename)
	assertCode(t, err, http.StatusNotFound, errs.CodeNotFound)
	_, err = f.service.Get(t.Context(), mei, view.ID)
	assertCode(t, err, http.StatusNotFound, errs.CodeNotFound)
	_, err = f.service.Delete(t.Context(), daniel, view.ID)
	assertCode(t, err, http.StatusForbidden, errs.CodeForbidden)
	_, err = f.service.Share(t.Context(), daniel, view.ID, SharingInput{Version: &view.Version, TeamIDs: []string{}})
	assertCode(t, err, http.StatusForbidden, errs.CodeForbidden)
}

func TestUpdate_ChangesAndReplacesWidgets_FR012(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	view := f.create(t, priya, "Overview", validBar("w1", 0, 0))
	f.create(t, priya, "Taken")
	widgets := []Widget{validBar("w2", 0, 4)}

	changed, err := f.service.Update(t.Context(), priya, view.ID, UpdateInput{Version: &view.Version,
		Description: ptr("Figures in USD"), Widgets: &widgets})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if changed.Version != 2 || changed.Name != "Overview" || changed.Description != "Figures in USD" ||
		len(changed.Widgets) != 1 || changed.Widgets[0].ID != "w2" {
		t.Errorf("changed = v%d %q %q %+v", changed.Version, changed.Name, changed.Description, changed.Widgets)
	}
	assertNothingChanged(t, f, view.ID, changed.Version)
	_, err = f.service.Update(t.Context(), priya, view.ID, UpdateInput{Version: &changed.Version, Name: ptr("taken")})
	assertCode(t, err, http.StatusConflict, errs.CodeNameTaken)
	_, err = f.service.Update(t.Context(), priya, view.ID, UpdateInput{Name: ptr("No version")})
	assertCode(t, err, http.StatusBadRequest, errs.CodeValidationFailed)
}

func assertNothingChanged(t *testing.T, f *fixture, id string, version int) {
	t.Helper()
	unchanged, err := f.service.Update(t.Context(), priya, id, UpdateInput{Version: &version, Name: ptr("Overview")})
	if err != nil || unchanged.Version != version || len(f.log(t)) != 3 {
		t.Errorf("a save that changes nothing = v%d, %v, %d log entries; want v%d and no new entry",
			unchanged.Version, err, len(f.log(t)), version)
	}
}

func TestUpdate_VersionConflict_FR012(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	view := f.create(t, priya, "Overview")
	if _, err := f.service.Update(t.Context(), priya, view.ID, UpdateInput{Version: &view.Version, Name: ptr("First")}); err != nil {
		t.Fatalf("first save: %v", err)
	}

	_, err := f.service.Update(t.Context(), priya, view.ID, UpdateInput{Version: &view.Version, Name: ptr("Second")})

	assertCode(t, err, http.StatusConflict, errs.CodeVersionConflict)
}

func TestDelete_ToRecycleBinAndUndo_FR010(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	view := f.share(t, f.create(t, priya, "Overview", validBar("w1", 0, 0)), "T-1001")

	entry, err := f.service.Delete(t.Context(), priya, view.ID)
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	_, err = f.service.Get(t.Context(), daniel, view.ID)
	assertCode(t, err, http.StatusNotFound, errs.CodeNotFound)
	if _, err = f.history.Undo(t.Context(), priya, entry.ID); err != nil {
		t.Fatalf("Undo() error: %v", err)
	}

	back, err := f.service.Get(t.Context(), daniel, view.ID)
	if err != nil {
		t.Fatalf("Get() after Undo error: %v", err)
	}
	got := []any{back.Version, len(back.Widgets), len(back.SharedWith), entry.RecordType, entry.RecordName}
	if diff := cmp.Diff([]any{view.Version, 1, 1, kindDashboard, "Overview"}, got); diff != "" {
		t.Errorf("after Undo mismatch (-want +got):\n%s", diff)
	}
}
