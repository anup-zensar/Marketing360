package approvalreport_test

import (
	"testing"

	"mapp/internal/measure"
	"mapp/internal/measure/approvalreport"
	"mapp/internal/measure/approvalreport/approvalreporttest"
	"mapp/internal/platform/sheet"
)

func TestSheets_041_FR023(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	report := approvalreport.NewReport(approvalreport.NewService(f.DB, f.Deps))

	sheets, err := report.Sheets(t.Context(), approvalreporttest.Alice, measure.Filter{})
	if err != nil {
		t.Fatalf("Sheets() error: %v", err)
	}

	if !report.Available() || report.Key() != measure.KeyApprovals {
		t.Errorf("Key, Available = %q, %v, want %q, true", report.Key(), report.Available(), measure.KeyApprovals)
	}
	if len(sheets) != 2 || sheets[0].Key != "status" || sheets[1].Key != "average" {
		t.Fatalf("sheet keys = %v, want status, average in this order", sheets)
	}
	status, average := sheets[0].Sheet, sheets[1].Sheet
	if status.Name != "Current approval status" || average.Name != "Average days to a decision" {
		t.Errorf("sheet names = %q, %q", status.Name, average.Name)
	}
	if got := status.Rows[0][0]; got != sheet.Bold("Activity") {
		t.Errorf("status first block = %v, want the Activity heading", got)
	}
	heading := []sheet.Cell{sheet.Bold("Approver"), sheet.Bold("Days"), sheet.Bold("Decisions")}
	if got := average.Rows[0]; !equalRows(got, heading) {
		t.Errorf("average headings = %v, want %v", got, heading)
	}
	all := []sheet.Cell{sheet.Text("All approvers"), sheet.Text("1.3"), sheet.Number(3)}
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
