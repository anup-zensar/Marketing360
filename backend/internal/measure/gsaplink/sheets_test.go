package gsaplink

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/sheet"
)

// fullWorld has every tab filled: two activities approved, the requests and the orders; the person sees the UK plan's
// financial data only.
func fullWorld(tb testing.TB) *world {
	tb.Helper()
	w := withOrders(withRequests(twoPlans(tb)))
	w.approved[ukActivity], w.approved[deActivity] = now.Add(-10*time.Hour), now.AddDate(0, 0, -11)
	w.choices.FinancePlans = []string{ukPlan}
	return w
}

func sheetsOf(tb testing.TB, w *world, own url.Values) []measure.WidgetSheet {
	tb.Helper()
	sheets, err := w.service().Sheets(tb.Context(), actor.Actor{ID: "U-1002"}, measure.Filter{Own: own})
	if err != nil {
		tb.Fatalf("Sheets() error = %v", err)
	}
	return sheets
}

// firstColumn is the first cell of every row after the heading.
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

func TestSheets_EveryRow_FR019(t *testing.T) {
	t.Parallel()
	sheets := sheetsOf(t, fullWorld(t), nil)

	keys := []string{}
	for _, s := range sheets {
		keys = append(keys, s.Key)
	}
	if diff := cmp.Diff([]string{"swbs", "swbsBands", "waiting", "orders"}, keys); diff != "" {
		t.Fatalf("widgets (-want +got):\n%s", diff)
	}
	want := map[string][]sheet.Cell{
		"swbs":    texts("L-3", "L-1", "L-2"),
		"waiting": texts("PR-1001", "PR-1002"),
		"orders":  append(texts("L-1", "L-2"), sheet.Bold("Total")),
	}
	for _, s := range sheets {
		if rows, listed := want[s.Key]; listed {
			if diff := cmp.Diff(rows, firstColumn(s.Sheet), cellsEqual()); diff != "" {
				t.Errorf("%s rows (-want +got):\n%s", s.Key, diff)
			}
		}
	}
	if amount := sheets[2].Sheet.Rows[2][11]; amount != (sheet.Cell{}) {
		t.Errorf("the hidden amount of PR-1002 = %+v, want an empty cell", amount)
	}
}

// TestSheets_EqualTables_SC005: each sheet holds the rows of its tab's table over every page, in the address's sort.
func TestSheets_EqualTables_SC005(t *testing.T) {
	t.Parallel()
	w := fullWorld(t)
	sheets := sheetsOf(t, w, url.Values{"tab": {"swbs"}, "sort": {"lineItem"}})

	r := w.report(Table{Tab: TabSwbs, Sort: "lineItem", PageSize: MaxPageSize})
	want := []sheet.Cell{}
	for _, row := range r.Swbs.Rows {
		want = append(want, sheet.Text(row.LineItemID))
	}
	if diff := cmp.Diff(want, firstColumn(sheets[0].Sheet), cellsEqual()); diff != "" {
		t.Errorf("SWBS sheet against its table (-want +got):\n%s", diff)
	}
}

func TestSheets_RejectOwnParametersOfAnotherTab(t *testing.T) {
	t.Parallel()
	_, err := fullWorld(t).service().Sheets(t.Context(), actor.Actor{}, measure.Filter{Own: url.Values{"sort": {"vendor"}}})
	if err == nil {
		t.Error("Sheets() with sort=vendor on the SWBS tab: no error, want the parameter refused")
	}
}
