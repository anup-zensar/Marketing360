package measure_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/history"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/sheet"
)

type sheetsReport struct{ key string }

func (r sheetsReport) Key() string     { return r.key }
func (r sheetsReport) Available() bool { return true }

func (r sheetsReport) Sheets(context.Context, actor.Actor, measure.Filter) ([]measure.WidgetSheet, error) {
	rows := func(value string) [][]sheet.Cell {
		return [][]sheet.Cell{{sheet.Bold("Value"), sheet.Bold("Activities")}, {sheet.Text(value), sheet.Number(2)}}
	}
	return []measure.WidgetSheet{
		{Key: "statusCount", Sheet: sheet.Sheet{Name: "Activities by status", Rows: rows("Approved")}},
		{Key: "attr12Count", Sheet: sheet.Sheet{Name: "Activities by Brand", Rows: rows("Helix")}},
	}, nil
}

type describedReport struct{ sheetsReport }

func (describedReport) Describe(_ context.Context, _ actor.Actor, own url.Values) ([][2]string, error) {
	return [][2]string{{"Search", own.Get("q")}, {"Status", "Active"}}, nil
}

// exportService serves Plan details as report, or as sheetsReport when none is given.
func exportService(t *testing.T, report ...measure.Report) *measure.Service {
	t.Helper()
	if len(report) == 0 {
		report = []measure.Report{sheetsReport{key: measure.KeyPlanDetails}}
	}
	registry := measure.NewRegistry()
	if err := registry.Register(report[0]); err != nil {
		t.Fatal(err)
	}
	return measure.NewService(history.Tools{Tx: fakeTx{}, Now: func() time.Time { return time.Date(2026, 10, 18, 14, 5, 0, 0, time.UTC) }}, measure.Deps{Planning: fakePlanning{nodes: frameNodes()}, Plans: fakePlans{ids: []string{"P-1"}},
		Currencies: fakeCurrencies{}}, registry)
}

// exportOf exports report for the address values with req's format and widget.
func exportOf(t *testing.T, report string, values url.Values, req measure.ExportRequest) (measure.File, error) {
	t.Helper()
	f, err := measure.ParseFilter(values)
	if err != nil {
		t.Fatal(err)
	}
	req.Filter = f
	return exportService(t).Export(t.Context(), actor.Actor{ID: "U-2", Name: "Priya Raman"}, report, req)
}

func TestExport_Excel_FR022(t *testing.T) {
	t.Parallel()
	file, err := exportOf(t, measure.KeyPlanDetails, url.Values{"owner": {"U-3"}}, measure.ExportRequest{})
	if err != nil {
		t.Fatal(err)
	}
	tables, err := sheet.ReadXLSX(file.Body, 50)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, table := range tables {
		names = append(names, table.Name)
	}
	want := []string{"Activities by status", "Activities by Brand", "In view", "About this export"}
	if diff := cmp.Diff(want, names); diff != "" || file.Name != "SP&A One – Plan details – 2026 – 18 Oct 2026.xlsx" {
		t.Errorf("file %q, sheets in order (-want +got):\n%s", file.Name, diff)
	}
	about := map[string]string{}
	for _, row := range tables[3].Rows {
		about[row.Cells[0]] = row.Cells[1]
	}
	wantAbout := map[string]string{"Report": "Plan details", "Year": "2026", "Filters": "Owner: Daniel Okoro",
		"Period": "Months", "Currency": "GBP", "Exported by": "Priya Raman", "Exported at": "18 Oct 2026 14:05 UTC"}
	if diff := cmp.Diff(wantAbout, about); diff != "" {
		t.Errorf("About this export (-want +got):\n%s", diff)
	}
}

func TestExport_CSV_FR022(t *testing.T) {
	t.Parallel()
	file, err := exportOf(t, measure.KeyPlanDetails, url.Values{},
		measure.ExportRequest{Format: measure.FormatCSV, Widget: "attr12Count"})
	if err != nil {
		t.Fatal(err)
	}

	text := string(file.Body)
	if file.Name != "SP&A One – Plan details – Activities by Brand – 2026 – 18 Oct 2026.csv" ||
		strings.Index(text, "Report") > strings.Index(text, "Helix") || strings.Contains(text, "Approved") {
		t.Errorf("file %q = %s, want the notes, then Activities by Brand only", file.Name, text)
	}
}

func TestExport_Refusals_FR022(t *testing.T) {
	t.Parallel()
	for name, req := range map[string]measure.ExportRequest{
		"CSV without a widget": {Format: measure.FormatCSV},
		"unknown widget":       {Widget: "nope"},
		"unknown format":       {Format: "pdf", Widget: "attr12Count"},
	} {
		_, err := exportOf(t, measure.KeyPlanDetails, url.Values{}, req)
		var e *errs.Error
		if !errors.As(err, &e) || e.Status != 400 {
			t.Errorf("%s: error = %v, want 400", name, err)
		}
	}
}

// A report not built yet exports In view and About this export only.
func TestExport_StubReport_FR022(t *testing.T) {
	t.Parallel()
	file, err := exportOf(t, measure.KeyApprovals, url.Values{}, measure.ExportRequest{})
	if err != nil {
		t.Fatal(err)
	}
	tables, err := sheet.ReadXLSX(file.Body, 50)
	if err != nil || len(tables) != 2 || tables[0].Name != "In view" {
		t.Errorf("sheets = %+v, %v; want In view and About this export", tables, err)
	}
}

func TestExport_ReportSaysItsOwnParameters_FR022(t *testing.T) {
	t.Parallel()
	service := exportService(t, describedReport{sheetsReport{key: measure.KeyPlanDetails}})
	f, err := measure.ParseFilter(url.Values{"q": {"planner"}})
	if err != nil {
		t.Fatal(err)
	}
	file, err := service.Export(t.Context(), actor.Actor{ID: "U-2", Name: "Priya Raman"}, measure.KeyPlanDetails,
		measure.ExportRequest{Filter: f})
	if err != nil {
		t.Fatal(err)
	}
	tables, err := sheet.ReadXLSX(file.Body, 50)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, row := range tables[len(tables)-1].Rows[:5] {
		lines = append(lines, row.Cells[0]+": "+row.Cells[1])
	}
	want := []string{"Report: Plan details", "Year: 2026", "Filters: None", "Search: planner", "Status: Active"}
	if diff := cmp.Diff(want, lines); diff != "" {
		t.Errorf("About this export, first lines (-want +got):\n%s", diff)
	}
}

func TestExport_LeftOutValuesNotNamed_FR015(t *testing.T) {
	t.Parallel()
	file, err := exportOf(t, measure.KeyPlanDetails, url.Values{"owner": {"U-9"}, "team": {"T-2"}}, measure.ExportRequest{})
	if err != nil {
		t.Fatal(err)
	}
	tables, err := sheet.ReadXLSX(file.Body, 50)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range tables[len(tables)-1].Rows {
		if row.Cells[0] == "Filters" && row.Cells[1] != "None" {
			t.Errorf("Filters = %q, want None: Wei Zhang and Mobility China are outside Priya's reports scope", row.Cells[1])
		}
	}
}

func TestExport_FiltersInWords_FR022(t *testing.T) {
	t.Parallel()
	file, err := exportOf(t, measure.KeyPlanDetails, url.Values{"planType": {"crossMarket"}, "activityStatus": {"inProgress"}},
		measure.ExportRequest{})
	if err != nil {
		t.Fatal(err)
	}
	tables, err := sheet.ReadXLSX(file.Body, 50)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range tables[len(tables)-1].Rows {
		if row.Cells[0] == "Filters" && row.Cells[1] != "Plan type: Cross-market; Activity status: In progress" {
			t.Errorf("Filters = %q, want the plan type and status as the screen words them", row.Cells[1])
		}
	}
}
