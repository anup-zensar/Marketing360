package customdashboard

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/date"
	"mapp/internal/platform/errs"
)

func ptr[T any](v T) *T { return &v }

func validBar(id string, x, y int) Widget {
	return Widget{ID: id, Type: typeBar, Title: "Planned cost by type", Layout: Layout{X: x, Y: y, W: 6, H: 4},
		Query: &Query{Source: "activities", Dimension: &GroupBy{Field: "activityType"},
			Measures: []Measure{{Aggregation: "sum", Field: ptr("plannedCost")}}, Filters: []Filter{},
			Sort: Sort{By: "measure", Direction: "desc"}, Limit: 10}}
}

func validDashboard(widgets ...Widget) Dashboard {
	return Dashboard{Name: "Lubricants overview", OwnerID: "U-1002", SharedWith: []string{},
		Filters: Filters{PlanIDs: []string{}, TeamIDs: []string{}}, Widgets: widgets}
}

// problemsOf gives each field's messages joined by " | ", so a test sees every problem at once.
func problemsOf(d Dashboard) map[string]string {
	got := map[string]string{}
	for _, p := range validate(d) {
		if got[p.Field] != "" {
			got[p.Field] += " | "
		}
		got[p.Field] += p.Message
	}
	return got
}

func assertProblems(t *testing.T, d Dashboard, want map[string]string) {
	t.Helper()
	if diff := cmp.Diff(want, problemsOf(d)); diff != "" {
		t.Errorf("validate() problems mismatch (-want +got):\n%s", diff)
	}
}

func TestValidateDashboard_NameAndDescription_FR008(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		dashboard   string
		description string
		want        map[string]string
	}{
		{name: "a name of 80 characters and a description of 500", dashboard: strings.Repeat("n", 80),
			description: strings.Repeat("d", 500), want: map[string]string{}},
		{name: "no name", dashboard: "", want: map[string]string{nameField: msgEnterName}},
		{name: "81 characters", dashboard: strings.Repeat("é", 81), want: map[string]string{nameField: msgNameTooLong}},
		{name: "a line break", dashboard: "Two\nlines",
			want: map[string]string{nameField: "Names can't contain tabs, line breaks or text-direction characters."}},
		{name: "501 characters of description", dashboard: "Fine", description: strings.Repeat("d", 501),
			want: map[string]string{descriptionField: msgDescriptionTooLong}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := validDashboard()
			d.Name, d.Description = tt.dashboard, tt.description
			assertProblems(t, d, tt.want)
		})
	}
}

func TestValidateDashboard_Filters_FR008(t *testing.T) {
	t.Parallel()
	d := validDashboard()
	d.Filters = Filters{DateFrom: ptr(date.MustParse("2026-12-31")), DateTo: ptr(date.MustParse("2026-07-01")),
		PlanIDs: []string{"P-1001", "plan"}, TeamIDs: make([]string, 21)}
	for i := range d.Filters.TeamIDs {
		d.Filters.TeamIDs[i] = "T-1001"
	}
	assertProblems(t, d, map[string]string{
		"filters.dateTo": msgDateRange, "filters.planIds": msgUnknownPlan, "filters.teamIds": msgTooManyTeams,
	})
}

func TestValidateWidget_RulesByType_FR015(t *testing.T) {
	t.Parallel()
	two := []Measure{{Aggregation: "sum", Field: ptr("plannedCost")}, {Aggregation: "count", Field: nil}}
	series := &GroupBy{Field: "campaign"}
	tests := []struct {
		name   string
		change func(w *Widget)
		want   map[string]string
	}{
		{name: "bar with a group and one measure", change: func(*Widget) {}, want: map[string]string{}},
		{name: "bar without a group", change: func(w *Widget) { w.Query.Dimension = nil },
			want: map[string]string{"widgets[0].query.dimension": msgChooseGroup}},
		{name: "bar broken down with two measures", change: func(w *Widget) { w.Query.Series, w.Query.Measures = series, two },
			want: map[string]string{"widgets[0].query.measures": "A bar chart with a breakdown shows one measure."}},
		{name: "line with two measures", change: func(w *Widget) { w.Type, w.Query.Measures = typeLine, two },
			want: map[string]string{}},
		{name: "area with five measures", change: func(w *Widget) { w.Type, w.Query.Measures = typeArea, append(two, two[0], two[0], two[0]) },
			want: map[string]string{"widgets[0].query.measures": "An area chart shows 1 to 4 measures."}},
		{name: "combo with one measure and a breakdown", change: func(w *Widget) { w.Type, w.Query.Series = typeCombo, series },
			want: map[string]string{"widgets[0].query.series": "A bar and line chart has no breakdown.",
				"widgets[0].query.measures": "A bar and line chart shows 2 to 4 measures."}},
		{name: "pie with two measures", change: func(w *Widget) { w.Type, w.Query.Measures = typePie, two },
			want: map[string]string{"widgets[0].query.measures": "A pie chart shows one measure."}},
		{name: "donut with a breakdown", change: func(w *Widget) { w.Type, w.Query.Series = typeDonut, series },
			want: map[string]string{"widgets[0].query.series": "A donut chart has no breakdown."}},
		{name: "scatter with one measure", change: func(w *Widget) { w.Type = typeScatter },
			want: map[string]string{"widgets[0].query.measures": "A scatter chart shows 2 to 3 measures."}},
		{name: "funnel", change: func(w *Widget) { w.Type = typeFunnel }, want: map[string]string{}},
		{name: "gauge with a group", change: func(w *Widget) { w.Type = typeGauge },
			want: map[string]string{"widgets[0].query.dimension": "A gauge has no grouping."}},
		{name: "kpi without a group", change: func(w *Widget) { w.Type, w.Query.Dimension = typeKpi, nil },
			want: map[string]string{}},
		{name: "heatmap without a breakdown", change: func(w *Widget) { w.Type = typeHeatmap },
			want: map[string]string{"widgets[0].query.series": msgChooseSeries}},
		{name: "treemap averaging", change: func(w *Widget) { w.Type, w.Query.Measures[0].Aggregation = typeTreemap, "avg" },
			want: map[string]string{"widgets[0].query.measures[0].aggregation": msgTreemapAggregation}},
		{name: "radar broken down", change: func(w *Widget) { w.Type, w.Query.Series = typeRadar, series },
			want: map[string]string{}},
		{name: "table without a group and two measures", change: func(w *Widget) { w.Type, w.Query.Dimension, w.Query.Measures = typeTable, nil, two },
			want: map[string]string{}},
		{name: "text with a query", change: func(w *Widget) { w.Type = typeText },
			want: map[string]string{"widgets[0].query": msgTextHasNoQuery}},
		{name: "text note", change: func(w *Widget) { w.Type, w.Query, w.Note = typeText, nil, "Figures in USD." },
			want: map[string]string{}},
		{name: "a chart without a query and with a note", change: func(w *Widget) { w.Query, w.Note = nil, "Note" },
			want: map[string]string{"widgets[0].query": msgChooseMeasure, "widgets[0].note": msgNoteOnlyText}},
		{name: "an unknown type", change: func(w *Widget) { w.Type = "sankey" },
			want: map[string]string{"widgets[0].type": msgWidgetType}},
		{name: "a bad id, a long title and subtitle", change: func(w *Widget) {
			w.ID, w.Title, w.Subtitle = "w0", strings.Repeat("t", 81), strings.Repeat("s", 121)
		}, want: map[string]string{"widgets[0].id": msgWidgetID, "widgets[0].title": msgTitleTooLong,
			"widgets[0].subtitle": msgSubtitleTooLong}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := validBar("w1", 0, 0)
			tt.change(&w)
			assertProblems(t, validDashboard(w), tt.want)
		})
	}
}

func TestValidateDashboard_WidgetCountAndIDs_FR013(t *testing.T) {
	t.Parallel()
	widgets := make([]Widget, 31)
	for i := range widgets {
		widgets[i] = validBar("w1", 6*(i%2), 4*(i/2))
	}
	got := problemsOf(validDashboard(widgets...))
	if got[widgetsField] != msgTooManyWidgets || got["widgets[1].id"] != msgWidgetID {
		t.Errorf("31 widgets with one id: problems = %v", got)
	}
}

func TestValidate_EveryProblemAtOnce_FR033(t *testing.T) {
	t.Parallel()
	pie := validBar("w2", 3, 0)
	pie.Type, pie.Query.Measures = typePie, append(pie.Query.Measures, Measure{Aggregation: "count"})
	d := validDashboard(validBar("w1", 0, 0), pie)
	d.Name = " "

	assertProblems(t, d, map[string]string{
		nameField:                   msgEnterName,
		"widgets[1].layout":         "This widget overlaps 'Planned cost by type'.",
		"widgets[1].query.measures": "A pie chart shows one measure.",
	})
	if err := validate(d).Err(); codeOf(err) != errs.CodeValidationFailed {
		t.Errorf("validate().Err() = %v, want validation_failed", err)
	}
}

func codeOf(err error) string {
	if apiErr, ok := errors.AsType[*errs.Error](err); ok {
		return apiErr.Code
	}
	return ""
}
