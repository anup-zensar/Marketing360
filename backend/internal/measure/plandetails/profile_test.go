package plandetails

import (
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/measure"
)

func counts(w Widget) map[string]int {
	got := map[string]int{}
	for _, v := range w.Values {
		got[v.Key] = v.Count
	}
	return got
}

func keys(w Widget) []string {
	var got []string
	for _, v := range w.Values {
		got = append(got, v.Key)
	}
	return got
}

func widgetKeys(r AttributeRow) []string {
	var got []string
	for _, w := range r.Widgets {
		got = append(got, w.Key)
	}
	return got
}

// activities adds n activities under parent, numbered from first, each with the changes.
func (y *year) activities(first, n int, parent string, changes ...func(*allwork.Node)) *year {
	for i := first; i < first+n; i++ {
		y.activity(fmt.Sprintf("A-%d", i), parent, "100.00", changes...)
	}
	return y
}

func companyCodeYear(t *testing.T) *year {
	t.Helper()
	return newYear(t).plan("P-1").campaign("C-1", "P-1").
		activities(1, 12, "C-1", choice("SA-1", "GB01")).
		activities(13, 10, "C-1", choice("SA-1", "CN04")).
		activities(23, 6, "C-1")
}

func companyCode() allwork.Attribute {
	return attribute("SA-1", "Company code", allwork.TypeSingleChoice, "CN04", "GB01")
}

func TestRows_NotFilledIn_FR012(t *testing.T) {
	t.Parallel()
	v := newView(companyCodeYear(t).matching(), LevelActivity, measure.PeriodMonth)

	rows := v.rows([]allwork.Attribute{companyCode()})

	if len(rows) != 1 || len(rows[0].Widgets) != 3 {
		t.Fatalf("rows = %+v, want one row of three widgets", rows)
	}
	want := shape{Title: "Activities by Company code", Keys: []string{"GB01", "CN04", KeyNotFilledIn},
		Counts: map[string]int{"GB01": 12, "CN04": 10, KeyNotFilledIn: 6}, Total: 28}
	if diff := cmp.Diff(want, shapeOf(rows[0].Widgets[0])); diff != "" {
		t.Errorf("Activities by Company code, largest first, then Not filled in (-want +got):\n%s", diff)
	}
	want = shape{Title: "Company code filled in", Keys: []string{KeyFilled, KeyNotFilledIn},
		Counts: map[string]int{KeyFilled: 22, KeyNotFilledIn: 6}, Total: 28, Percent: "79"}
	if diff := cmp.Diff(want, shapeOf(rows[0].Widgets[2])); diff != "" {
		t.Errorf("Company code filled in (-want +got):\n%s", diff)
	}
}

// shape is what the tests compare of a widget.
type shape struct {
	Title   string
	Keys    []string
	Counts  map[string]int
	Total   int
	Percent string
}

func shapeOf(w Widget) shape {
	s := shape{Title: w.Title, Keys: keys(w), Counts: counts(w), Total: w.Total.Count}
	if w.FilledInPercent != nil {
		s.Percent = fmt.Sprint(*w.FilledInPercent)
	}
	return s
}

func TestFilledInPercent_LeavesOutDoesNotApply_FR012(t *testing.T) {
	t.Parallel()
	y := newYear(t).plan("P-1", "BU-LUB").campaign("C-1", "P-1").
		activities(1, 15, "C-1", choice("SA-5", "NEW")).activities(16, 5, "C-1").
		plan("P-2", "BU-MOB").campaign("C-2", "P-2").activities(21, 8, "C-2", choice("SA-5", "NEW"))
	launch := attribute("SA-5", "Product launch", allwork.TypeSingleChoice, "NEW")
	launch.Condition = []string{"BU-LUB"}

	rows := newView(y.matching(), LevelActivity, measure.PeriodMonth).rows([]allwork.Attribute{launch})

	count, filled := rows[0].Widgets[0], rows[0].Widgets[2]
	if diff := cmp.Diff(map[string]int{"NEW": 15, KeyNotFilledIn: 5, KeyDoesNotApply: 8}, counts(count)); diff != "" {
		t.Errorf("Activities by Product launch (-want +got):\n%s", diff)
	}
	want := map[string]int{KeyFilled: 15, KeyNotFilledIn: 5, KeyDoesNotApply: 8}
	if diff := cmp.Diff(want, counts(filled)); diff != "" || *filled.FilledInPercent != 75 {
		t.Errorf("Product launch filled in (-want +got):\n%s per cent %d, want 75", diff, *filled.FilledInPercent)
	}
}

func TestRows_ActiveInDisplayOrder_FR008(t *testing.T) {
	t.Parallel()
	retired := attribute("SA-9", "Retired", allwork.TypeSingleChoice, "X")
	retired.Active = false
	cluster := atLevel(attribute("SA-8", "Cluster", allwork.TypeSingleChoice, "EU"), allwork.LevelCampaign)
	region := attribute("SA-3", "Region", allwork.TypeSingleChoice, "EMEA")
	v := newView(companyCodeYear(t).matching(), LevelActivity, measure.PeriodMonth)

	rows := v.rows([]allwork.Attribute{region, retired, cluster, companyCode()})

	var got []string
	for _, r := range rows {
		got = append(got, r.AttributeID)
	}
	if diff := cmp.Diff([]string{"SA-3", "SA-1"}, got); diff != "" {
		t.Errorf("the level's active attributes in Settings' order (-want +got):\n%s", diff)
	}
}

func TestRows_InactiveOption_FR008(t *testing.T) {
	t.Parallel()
	code := companyCode()
	code.Options[0].Inactive = true // CN04
	v := newView(companyCodeYear(t).matching(), LevelActivity, measure.PeriodMonth)

	count := v.rows([]allwork.Attribute{code})[0].Widgets[0]

	for _, value := range count.Values {
		if value.Key == "CN04" && (!value.Inactive || value.Label != "CN04 (inactive)") {
			t.Errorf("CN04 = %+v, want inactive and labelled (inactive)", value)
		}
	}
}

func TestRows_WidgetsByType_FR009_FR010(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind string
		want []string
	}{
		{allwork.TypeSingleChoice, []string{"attr7Count", "attr7PlannedCost", "attr7FilledIn"}},
		{allwork.TypeMultipleChoice, []string{"attr7Count", "attr7PlannedCost", "attr7FilledIn"}},
		{allwork.TypeMultipleChoiceSplit, []string{"attr7Count", "attr7PlannedCost", "attr7FilledIn"}},
		{allwork.TypeYesNo, []string{"attr7Count", "attr7PlannedCost", "attr7FilledIn"}},
		{"text", []string{"attr7FilledIn"}},
		{"longText", []string{"attr7FilledIn"}},
		{"number", []string{"attr7FilledIn"}},
	}
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			t.Parallel()
			v := newView(companyCodeYear(t).matching(), LevelActivity, measure.PeriodMonth)
			rows := v.rows([]allwork.Attribute{attribute("SA-7", "FIT4 ID", tt.kind, "X")})
			if diff := cmp.Diff(tt.want, widgetKeys(rows[0])); diff != "" {
				t.Errorf("widgets (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRows_TextFilledIn_FR010(t *testing.T) {
	t.Parallel()
	fit4 := func(n *allwork.Node) { n.Fields = map[string]allwork.Value{"attr.SA-7": {Text: "F-123"}} }
	y := newYear(t).plan("P-1").campaign("C-1", "P-1").
		activities(1, 3, "C-1", fit4).activities(4, 1, "C-1")

	rows := newView(y.matching(), LevelActivity, measure.PeriodMonth).
		rows([]allwork.Attribute{attribute("SA-7", "FIT4 ID", "text")})

	filled := rows[0].Widgets[0]
	if got := counts(filled); got[KeyFilled] != 3 || got[KeyNotFilledIn] != 1 || filled.Title != "FIT4 ID filled in" {
		t.Errorf("FIT4 ID filled in = %q %v, want 3 · 1", filled.Title, got)
	}
	if filled.FilterParam != nil || slices.ContainsFunc(filled.Values, func(v Value) bool { return v.Linkable }) {
		t.Errorf("FIT4 ID filled in links to All work, which cannot filter a text attribute")
	}
}

func TestRows_NoAttributesOfTheLevel_FR016(t *testing.T) {
	t.Parallel()
	cluster := atLevel(attribute("SA-8", "Cluster", allwork.TypeSingleChoice, "EU"), allwork.LevelCampaign)

	rows := newView(companyCodeYear(t).matching(), LevelActivity, measure.PeriodMonth).rows([]allwork.Attribute{cluster})

	if len(rows) != 0 {
		t.Errorf("rows = %+v, want none: no activity attribute is set up", rows)
	}
}

// TestRows_NotFilledInAtZero_FR009: every value widget lists Not filled in, also when every record has a value.
func TestRows_NotFilledInAtZero_FR009(t *testing.T) {
	t.Parallel()
	code := func(n *allwork.Node) { n.Attrs = map[string]allwork.AttrValue{"SA-1": {OptionIDs: []string{"GB01"}}} }
	y := newYear(t).plan("P-1").campaign("C-1", "P-1").activities(1, 2, "C-1", code)

	rows := newView(y.matching(), LevelActivity, measure.PeriodMonth).rows([]allwork.Attribute{companyCode()})

	for _, w := range rows[0].Widgets[:2] {
		if got, listed := counts(w)[KeyNotFilledIn]; !listed || got != 0 {
			t.Errorf("%s: Not filled in = %d, listed %v; want 0, listed", w.Title, got, listed)
		}
	}
}
