package customdashboard

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/date"
	"mapp/internal/platform/errs"
)

func TestLog_DescribesWidgetChanges_FR009(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	first, second, third := validBar("w1", 0, 0), validBar("w2", 6, 0), validBar("w3", 0, 4)
	second.Title = ""
	view := f.create(t, priya, "Overview", first, second)
	first.Display.Palette = ptr("planning")
	third.Title = "Leads by month"
	widgets := []Widget{first, third}
	filters := Filters{DateFrom: ptr(date.MustParse("2026-07-01")), PlanIDs: []string{"P-1001"}, TeamIDs: []string{"T-1001", "T-1002"}}

	if _, err := f.service.Update(t.Context(), priya, view.ID, UpdateInput{Version: &view.Version, Widgets: &widgets,
		Filters: &filters}); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	entries := f.log(t)
	changes := entries[len(entries)-1].Changes
	got := map[string]string{}
	for _, c := range changes {
		got[c.Label] = c.OldText + " → " + c.NewText
	}
	want := map[string]string{
		labelFilters: "No default filters → From 01 Jul 2026, 1 plan, 2 teams",
		labelWidgets: "2 widgets → 2 widgets; added 'Leads by month'; removed the Bar widget (w2); changed 'Planned cost by type'",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("log changes mismatch (-want +got):\n%s", diff)
	}
}

func TestRestore_RenamesWhenTaken_FR010(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	view := f.create(t, priya, "Overview")
	entry, err := f.service.Delete(t.Context(), priya, view.ID)
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	f.create(t, priya, "overview")

	_, err = f.history.Undo(t.Context(), priya, entry.ID)
	assertCode(t, err, http.StatusConflict, errs.CodeNameTaken)
	result, err := f.history.Restore(t.Context(), priya, entry.ID)
	if err != nil {
		t.Fatalf("Restore() error: %v", err)
	}

	back, err := f.service.Get(t.Context(), priya, view.ID)
	if err != nil || back.Name != "Overview (restored)" || back.Version != 2 {
		t.Errorf("restored = %q v%d, %v; want 'Overview (restored)' v2", back.Name, back.Version, err)
	}
	if want := "Renamed to 'Overview (restored)' because the name was taken."; result.Details != want {
		t.Errorf("restore notes = %v, want %v", result.Details, want)
	}
}
