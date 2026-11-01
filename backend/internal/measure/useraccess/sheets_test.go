package useraccess

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/sheet"
)

func sheetsOf(tb testing.TB, own url.Values) []measure.WidgetSheet {
	tb.Helper()
	o := newOffice()
	sheets, err := NewService(fakeTx{}, o, o, func() time.Time { return now }).Sheets(tb.Context(), actor.Actor{},
		measure.Filter{Own: own})
	if err != nil {
		tb.Fatalf("Sheets() error = %v", err)
	}
	return sheets
}

func firstColumn(s sheet.Sheet) []sheet.Cell {
	cells := []sheet.Cell{}
	for _, row := range s.Rows[1:] {
		cells = append(cells, row[0])
	}
	return cells
}

func texts(values ...string) []sheet.Cell {
	cells := []sheet.Cell{}
	for _, v := range values {
		cells = append(cells, sheet.Text(v))
	}
	return cells
}

func cellsEqual() cmp.Option {
	return cmp.Comparer(func(x, y sheet.Cell) bool { return x == y })
}

func TestSheets_EveryRow_FR013(t *testing.T) {
	t.Parallel()
	sheets := sheetsOf(t, url.Values{"status": {"all"}, "sort": {"-lastSignedIn"}, "pageSize": {"5"}, "page": {"2"}})

	if len(sheets) != 2 || sheets[0].Key != "people" || sheets[1].Key != "signInBands" {
		t.Fatalf("widgets = %v, want people and signInBands", sheets)
	}
	want := texts("Priya Raman", "Carlos Mendes", "Alex Morgan", "Daniel Okoro", "Mei Lin")
	if diff := cmp.Diff(want, firstColumn(sheets[0].Sheet), cellsEqual()); diff != "" {
		t.Errorf("every person, not a page, latest sign-in first (-want +got):\n%s", diff)
	}
	mei := sheets[0].Sheet.Rows[5]
	if mei[3] != sheet.Text("No role") || mei[6] != (sheet.Cell{}) {
		t.Errorf("Mei's roles %v and last sign-in %v; want No role and an empty cell for Never", mei[3], mei[6])
	}
	if last := sheets[1].Sheet.Rows[len(sheets[1].Sheet.Rows)-1]; last[1] != sheet.Number(5) {
		t.Errorf("bands total = %v, want 5", last[1])
	}
}

// TestSheets_EqualTable_SC005: the export holds the rows of the table, in its sort, over every page.
func TestSheets_EqualTable_SC005(t *testing.T) {
	t.Parallel()
	own := url.Values{"q": {"a"}, "team": {"T-1002"}, "sort": {"-name"}}
	sheets := sheetsOf(t, own)

	q, err := ParseQuery(own)
	if err != nil {
		t.Fatal(err)
	}
	q.PageSize = MaxPageSize
	want := []sheet.Cell{}
	for _, r := range newOffice().report(t, q).Rows {
		want = append(want, sheet.Text(r.Person.Name))
	}
	if diff := cmp.Diff(want, firstColumn(sheets[0].Sheet), cellsEqual()); diff != "" || len(want) == 0 {
		t.Errorf("People sheet against the table (-table +sheet):\n%s", diff)
	}
}

func TestSheets_RejectsUnknownParameters(t *testing.T) {
	t.Parallel()
	o := newOffice()
	_, err := NewService(fakeTx{}, o, o, func() time.Time { return now }).Sheets(t.Context(), actor.Actor{},
		measure.Filter{Own: url.Values{"status": {"gone"}}})
	if err == nil {
		t.Error("Sheets() with status=gone: no error, want the parameter refused")
	}
}
