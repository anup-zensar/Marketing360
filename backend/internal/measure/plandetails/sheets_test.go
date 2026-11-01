package plandetails

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/sheet"
)

func csvOf(t *testing.T, rows [][]sheet.Cell) []string {
	t.Helper()
	var b bytes.Buffer
	if err := sheet.WriteCSV(&b, nil, rows); err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}
	text := strings.TrimPrefix(b.String(), string(rune(0xFEFF))) // the byte order mark Excel needs
	return strings.Split(strings.TrimSpace(strings.ReplaceAll(text, "\r\n", "\n")), "\n")
}

// Each measure's columns as the table view: Value, the records named for the level, and the planned cost (US3).
func TestSheets_ColumnsPerMeasure_FR017(t *testing.T) {
	t.Parallel()
	s := NewService(fakeTx{}, &fakeScoped{matching: companyCodeYear(t).matching(),
		attributes: []allwork.Attribute{companyCode()}})

	sheets, err := s.Sheets(t.Context(), actor.Actor{ID: "U-1002"}, measure.Filter{Show: measure.ShowActivity})

	if err != nil {
		t.Fatalf("Sheets() error = %v", err)
	}
	got := map[string][]string{}
	for _, ws := range sheets[len(sheets)-3:] {
		got[ws.Key] = csvOf(t, ws.Sheet.Rows)
	}
	want := map[string][]string{
		"attr1Count":       {"Value,Activities", "GB01,12", "CN04,10", "Not filled in,6", "Total,28"},
		"attr1PlannedCost": {"Value,Activities,Planned cost (USD)", "GB01,12,1200.00", "CN04,10,1000.00", "Not filled in,6,600.00", "Total,28,2800.00"},
		"attr1FilledIn":    {"Value,Activities", "Filled in,22", "Not filled in,6", "Total,28"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("sheets (-want +got):\n%s", diff)
	}
}

// A widget with nothing to count still exports its header and total row (US3, scenario 4's file side).
func TestSheets_EmptyWidget_HeaderAndTotal_FR017(t *testing.T) {
	t.Parallel()
	s := NewService(fakeTx{}, &fakeScoped{matching: companyCodeYear(t).unmatch("A-1", "A-2", "A-3", "A-4", "A-5", "A-6",
		"A-7", "A-8", "A-9", "A-10", "A-11", "A-12", "A-13", "A-14", "A-15", "A-16", "A-17", "A-18", "A-19", "A-20",
		"A-21", "A-22", "A-23", "A-24", "A-25", "A-26", "A-27", "A-28").matching(), attributes: []allwork.Attribute{companyCode()}})

	sheets, err := s.Sheets(t.Context(), actor.Actor{ID: "U-1002"}, measure.Filter{})

	if err != nil || len(sheets) != 3 {
		t.Fatalf("Sheets() = %d sheets, %v; want the three of the attribute row", len(sheets), err)
	}
	if diff := cmp.Diff([]string{"Value,Activities", "Total,0"}, csvOf(t, sheets[0].Sheet.Rows)); diff != "" {
		t.Errorf("empty count widget (-want +got):\n%s", diff)
	}
}

func TestSheets_EveryValueAndTotal_FR017(t *testing.T) {
	t.Parallel()
	y := newYear(t).plan("P-1").campaign("C-1", "P-1")
	options := make([]string, 12)
	for i := range options {
		options[i] = fmt.Sprintf("O%02d", i)
		y.activity(fmt.Sprintf("A-%d", i), "C-1", "10.00", choice("SA-1", options[i]))
	}
	code := attribute("SA-1", "Company code", allwork.TypeSingleChoice, options...)
	s := NewService(fakeTx{}, &fakeScoped{matching: y.matching(), attributes: []allwork.Attribute{code}})

	sheets, err := s.Sheets(t.Context(), actor.Actor{ID: "U-1002"}, measure.Filter{})

	if err != nil {
		t.Fatalf("Sheets() error = %v", err)
	}
	cost := sheets[len(sheets)-2]
	lines := csvOf(t, cost.Sheet.Rows)
	if cost.Key != "attr1PlannedCost" || cost.Sheet.Name != "Planned cost by Company code" || len(lines) != 14 ||
		lines[0] != "Value,Activities,Planned cost (USD)" || lines[13] != "Total,12,120.00" {
		t.Errorf("sheet %s %q = %q, want the header, all twelve values and the total", cost.Key, cost.Sheet.Name, lines)
	}
}

func TestSheets_KeysOfLevel_FR001(t *testing.T) {
	t.Parallel()
	s := NewService(fakeTx{}, &fakeScoped{matching: companyCodeYear(t).matching(),
		attributes: []allwork.Attribute{companyCode()}})

	sheets, err := s.Sheets(t.Context(), actor.Actor{ID: "U-1002"}, measure.Filter{})

	if err != nil {
		t.Fatalf("Sheets() error = %v", err)
	}
	var got []string
	for _, ws := range sheets {
		got = append(got, ws.Key)
	}
	want := []string{"activityTypeCount", "activityTypePlannedCost", "statusCount", "startPeriodCount",
		"plannedCostEntered", "goalEntered", "attr1Count", "attr1PlannedCost", "attr1FilledIn"}
	if diff := cmp.Diff(want, got); diff != "" || s.Key() != measure.KeyPlanDetails || !s.Available() {
		t.Errorf("sheet keys in screen order (-want +got):\n%s", diff)
	}
}
