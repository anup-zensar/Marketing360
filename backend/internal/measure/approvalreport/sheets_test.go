package approvalreport_test

import (
	"net/url"
	"testing"

	"mapp/internal/measure"
	"mapp/internal/measure/approvalreport"
	"mapp/internal/measure/approvalreport/approvalreporttest"
	"mapp/internal/platform/sheet"
)

func sheetsOf(t *testing.T) (*approvalreport.Report, []measure.WidgetSheet) {
	t.Helper()
	f := approvalreporttest.New(t)
	report := approvalreport.NewReport(approvalreport.NewService(f.DB, f.Deps))
	sheets, err := report.Sheets(t.Context(), approvalreporttest.Alice(), measure.Filter{})
	if err != nil {
		t.Fatalf("Sheets() error: %v", err)
	}
	return report, sheets
}

func TestSheets_Order_041_FR023(t *testing.T) {
	t.Parallel()
	report, sheets := sheetsOf(t)

	if !report.Available() || report.Key() != measure.KeyApprovals {
		t.Errorf("Key, Available = %q, %v, want %q, true", report.Key(), report.Available(), measure.KeyApprovals)
	}
	if len(sheets) != 3 || sheets[0].Key != "status" || sheets[1].Key != "average" {
		t.Fatalf("sheet keys = %v, want status, average, approvals in this order", sheets)
	}
	if sheets[0].Sheet.Name != "Current approval status" || sheets[1].Sheet.Name != "Average days to a decision" {
		t.Errorf("sheet names = %q, %q", sheets[0].Sheet.Name, sheets[1].Sheet.Name)
	}
}

func TestSheets_Status_041_FR023(t *testing.T) {
	t.Parallel()
	_, sheets := sheetsOf(t)

	if got := sheets[0].Sheet.Rows[0][0]; got != sheet.Bold("Activity") {
		t.Errorf("status first block = %v, want the Activity heading", got)
	}
}

func TestSheets_Average_041_FR023(t *testing.T) {
	t.Parallel()
	_, sheets := sheetsOf(t)
	average := sheets[1].Sheet

	heading := []sheet.Cell{sheet.Bold("Approver"), sheet.Bold("Days"), sheet.Bold("Decisions")}
	all := []sheet.Cell{sheet.Text("All approvers"), sheet.Text("1.3"), sheet.Number(3)}
	if got := average.Rows[0]; !equalRows(got, heading) {
		t.Errorf("average headings = %v, want %v", got, heading)
	}
	if got := average.Rows[1]; !equalRows(got, all) {
		t.Errorf("average first row = %v, want %v", got, all)
	}
}

func equalRows(a, b []sheet.Cell) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func headingsMatch(row []sheet.Cell, want []string) bool {
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

func TestSheets_Grid_041_FR023(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	report := approvalreport.NewReport(approvalreport.NewService(f.DB, f.Deps))

	sheets, err := report.Sheets(t.Context(), approvalreporttest.Alice(), measure.Filter{Own: url.Values{"pageSize": {"5"}, "outcome": {"waiting"}}})
	if err != nil {
		t.Fatalf("Sheets() error: %v", err)
	}

	if len(sheets) != 3 || sheets[2].Key != "approvals" || sheets[2].Sheet.Name != "Approvals" {
		t.Fatalf("sheets = %v, want approvals third", sheets)
	}
	rows := sheets[2].Sheet.Rows
	wantHead := []string{"Item ID", "Name", "Record type", "Approver", "Decided on", "Status", "Submitted on", "Level",
		"Purchase order", "Amount (USD)", "Owner", "Comments"}
	if !headingsMatch(rows[0], wantHead) {
		t.Errorf("headings = %v, want %v", rows[0], wantHead)
	}
	if len(rows) != 3 {
		t.Errorf("rows = %d, want the heading and every one of the 2 waiting rows, whatever the page size", len(rows))
	}
}

func TestSheets_BadQuery_041_C10(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	report := approvalreport.NewReport(approvalreport.NewService(f.DB, f.Deps))

	_, err := report.Sheets(t.Context(), approvalreporttest.Alice(), measure.Filter{Own: url.Values{"sort": {"nope"}}})

	if err == nil {
		t.Error("Sheets() with a malformed sort = nil error, want a problem")
	}
}
