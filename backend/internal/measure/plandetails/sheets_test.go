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
