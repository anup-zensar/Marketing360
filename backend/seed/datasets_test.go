package seed_test

import (
	"encoding/base64"
	"io/fs"
	"net/http"
	"os"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/dataset"
	"mapp/internal/dataset/datasettest"
	"mapp/internal/platform/apitest"
	"mapp/seed"
)

func TestLoad_EventLeadsDataset_D376(t *testing.T) {
	t.Parallel()
	store := datasettest.NewStore(loadedStore(t))
	d, found, err := store.Get(t.Context(), "DS-1001")
	if err != nil || !found {
		t.Fatalf("DS-1001: %v, %v", found, err)
	}
	rows, err := store.Rows(t.Context(), d.ID)
	if err != nil {
		t.Fatalf("Rows: %v", err)
	}
	kinds := []dataset.Kind{}
	for _, c := range d.Columns {
		kinds = append(kinds, c.Kind)
	}
	want := []any{"Event leads 2026", "U-1002", 0, 12, 12, "EUR", int64(1850000),
		[]dataset.Kind{dataset.KindText, dataset.KindDate, dataset.KindNumber, dataset.KindMoney}}
	got := []any{d.Name, d.OwnerID, len(d.SharedWith), d.RowCount, len(rows), *d.Columns[3].Currency, *rows[0][3].Minor, kinds}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("DS-1001 (-want +got):\n%s", diff)
	}
}

func TestEventLeadsFixtures_MatchTheSeed_D376(t *testing.T) {
	t.Parallel()
	priya := apitest.New(t).SignIn(t, "priya.raman@shell.com")
	for _, name := range []string{"event-leads.csv", "event-leads.xlsx"} {
		data, err := fs.ReadFile(os.DirFS("../../frontend/e2e/fixtures"), name)
		if err != nil {
			t.Fatalf("fixture: %v", err)
		}
		r := priya.Do(t, http.MethodPost, "/api/v1/datasets/preview",
			map[string]any{"fileName": name, "content": base64.StdEncoding.EncodeToString(data)})
		var p dataset.Preview
		r.Decode(t, &p)
		kinds := []dataset.Kind{}
		for _, c := range p.Columns {
			kinds = append(kinds, c.ProposedKind)
		}
		want := []dataset.Kind{dataset.KindText, dataset.KindDate, dataset.KindNumber, dataset.KindMoney}
		leads := seed.EventLeads()
		if r.Status != http.StatusOK || p.RowCount != len(leads) || !slices.Equal(kinds, want) ||
			p.Rows[2][0] != leads[2].Event || p.Rows[2][1] != leads[2].Date {
			t.Errorf("%s: %d %+v", name, r.Status, p)
		}
	}
}
