package purchasereport_test

import (
	"net/url"
	"testing"

	"mapp/internal/measure"
	"mapp/internal/measure/purchasereport"
	"mapp/internal/measure/purchasereport/purchasereporttest"
	"mapp/internal/platform/sheet"
)

func sheetsOf(t *testing.T, own url.Values) map[string]sheet.Sheet {
	t.Helper()
	f := purchasereporttest.New(t)
	report := purchasereport.NewReport(purchasereport.NewService(f.DB, f.Deps))
	sheets, err := report.Sheets(t.Context(), purchasereporttest.Alice(), measure.Filter{Own: own})
	if err != nil {
		t.Fatalf("Sheets() error: %v", err)
	}
	byKey := map[string]sheet.Sheet{}
	for i, s := range sheets {
		byKey[s.Key] = s.Sheet
		if want := []string{"cards", "status", "byVendor", "byOwner", "grid"}[i]; s.Key != want {
			t.Errorf("sheet %d = %s, want %s: the five widgets in order", i, s.Key, want)
		}
	}
	return byKey
}

func headingsOf(row []sheet.Cell, want []string) bool {
	if len(row) != len(want) {
		return false
	}
	for i, h := range want {
		if row[i] != sheet.Bold(h) {
			return false
		}
	}
	return true
}

func TestSheets_Widgets_042_FR026(t *testing.T) {
	t.Parallel()
	sheets := sheetsOf(t, nil)

	status := sheets["status"].Rows
	if !headingsOf(status[0], []string{"Status", "Count", "Amount (USD)", "Share (%)"}) || len(status) != 8 {
		t.Errorf("status sheet = %d rows, headings %v, want the six statuses and a total", len(status), status[0])
	}
	vendors := sheets["byVendor"].Rows
	if !headingsOf(vendors[0], []string{"Vendor", "2026 03", "2026 04", "2026 05", "No delivery date", "Total"}) {
		t.Errorf("byVendor headings = %v, want the months, No delivery date and a total", vendors[0])
	}
	if len(vendors) != 5 {
		t.Errorf("byVendor = %d rows, want every vendor (3), the heading and the total row", len(vendors))
	}
}

func TestSheets_Grid_042_FR027(t *testing.T) {
	t.Parallel()

	all := sheetsOf(t, nil)["grid"]
	acme := sheetsOf(t, url.Values{"vendor": {"Acme"}, "columns": {"id,invoiced"}, "pageSize": {"5"}})["grid"]

	if len(all.Rows) != 11 {
		t.Errorf("grid sheet = %d rows, want the heading and every one of the 10 lines", len(all.Rows))
	}
	if !headingsOf(acme.Rows[0], []string{"Vendor", "Purchase request", "Line item", "ID", "Invoiced (USD)"}) {
		t.Errorf("headings = %v, want the chosen columns in order after the three levels", acme.Rows[0])
	}
	if len(acme.Rows) != 5 {
		t.Errorf("filtered grid sheet = %d rows, want the heading and Acme's 4 lines, whatever the page size", len(acme.Rows))
	}
}

func TestSheets_BadQuery_042_C6(t *testing.T) {
	t.Parallel()
	f := purchasereporttest.New(t)
	report := purchasereport.NewReport(purchasereport.NewService(f.DB, f.Deps))

	_, err := report.Sheets(t.Context(), purchasereporttest.Alice(), measure.Filter{Own: url.Values{"columns": {"nonsense"}}})

	if err == nil {
		t.Error("Sheets() with an unknown column = nil error, want a problem")
	}
}
