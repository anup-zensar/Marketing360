package customdashboard

import "testing"

func TestLayout_InsideAndNoOverlap_FR013(t *testing.T) {
	t.Parallel()
	const at = "widgets[1].layout"
	tests := []struct {
		name   string
		second Layout
		title  string
		want   map[string]string
	}{
		{name: "side by side", second: Layout{X: 6, Y: 0, W: 6, H: 4}, want: map[string]string{}},
		{name: "right below", second: Layout{X: 0, Y: 4, W: 12, H: 2}, want: map[string]string{}},
		{name: "the last cell of the grid", second: Layout{X: 10, Y: 198, W: 2, H: 2}, want: map[string]string{}},
		{name: "one cell shared", second: Layout{X: 5, Y: 3, W: 2, H: 2},
			want: map[string]string{at: "This widget overlaps 'Planned cost by type'."}},
		{name: "overlapping an untitled widget", second: Layout{X: 0, Y: 0, W: 2, H: 2}, title: " ",
			want: map[string]string{at: "This widget overlaps the Bar widget (w1)."}},
		{name: "past the right edge", second: Layout{X: 8, Y: 10, W: 6, H: 2}, want: map[string]string{at: msgPastRightEdge}},
		{name: "past the bottom edge", second: Layout{X: 0, Y: 199, W: 2, H: 2}, want: map[string]string{at: msgPastBottomEdge}},
		{name: "column 13", second: Layout{X: 12, Y: 10, W: 2, H: 2}, want: map[string]string{at: msgColumnRange}},
		{name: "row 0", second: Layout{X: 0, Y: -1, W: 2, H: 2}, want: map[string]string{at: msgRowRange}},
		{name: "one column wide", second: Layout{X: 0, Y: 10, W: 1, H: 2}, want: map[string]string{at: msgWidthRange}},
		{name: "13 rows high", second: Layout{X: 0, Y: 10, W: 2, H: 13}, want: map[string]string{at: msgHeightRange}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			first, second := validBar("w1", 0, 0), validBar("w2", 0, 0)
			if tt.title != "" {
				first.Title = ""
			}
			second.Layout = tt.second
			assertProblems(t, validDashboard(first, second), tt.want)
		})
	}
}
