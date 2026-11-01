package projection_test

import (
	"testing"
	"time"

	"mapp/internal/platform/date"
	"mapp/internal/projection"
)

func counts(values ...int64) []projection.MonthFigure {
	months := make([]projection.MonthFigure, 0, len(values))
	for i, v := range values {
		months = append(months, projection.MonthFigure{Month: monthOf(i), Value: v})
	}
	return months
}

func monthOf(i int) string {
	return date.New(2026, timeMonth(i+1), 1).Time().Format("2006-01")
}

func TestSplitCounts_ActualBeforeTodayProjectedFromIt_FR010(t *testing.T) {
	t.Parallel()
	actual := counts(10, 10, 10, 10, 10, 10, 4, 0, 0, 0, 0, 0) // July is half gone
	projected := counts(9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9)

	got := projection.SplitCounts(actual, projected, date.MustParse("2026-07-15"), 2026)

	assertMonthsSplit(t, got)
	// Actual to the end of June (60) plus projected from July (6 months × 9 = 54).
	if got.Actual != 60 || got.Projected != 54 || got.Total != 114 {
		t.Errorf("split = %d actual + %d projected = %d, want 60 + 54 = 114 with no month counted twice",
			got.Actual, got.Projected, got.Total)
	}
}

func assertMonthsSplit(t *testing.T, got projection.Split) {
	t.Helper()
	july := got.Months[6]
	if !july.InProgress || !carries(july.Actual, 4) || !carries(july.Projected, 9) {
		t.Errorf("July = %+v, want the month in progress carrying both its actual and its projected value", july)
	}
	if june := got.Months[5]; june.Actual == nil || june.Projected != nil {
		t.Errorf("June = %+v, want a whole month behind today to be actual only", june)
	}
	if august := got.Months[7]; august.Actual != nil || august.Projected == nil {
		t.Errorf("August = %+v, want a month ahead of today to be projected only", august)
	}
}

func carries(value *int64, want int64) bool { return value != nil && *value == want }

func TestSplitCounts_BeforeAndAfterThePlanYear_FR010(t *testing.T) {
	t.Parallel()
	actual := counts(1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1)
	projected := counts(2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2)

	before := projection.SplitCounts(actual, projected, date.MustParse("2025-12-31"), 2026)
	if before.Actual != 0 || before.Projected != 24 {
		t.Errorf("before the plan year = %d actual + %d projected, want everything still expected", before.Actual, before.Projected)
	}

	after := projection.SplitCounts(actual, projected, date.MustParse("2027-01-01"), 2026)
	if after.Actual != 12 || after.Projected != 0 {
		t.Errorf("after the plan year = %d actual + %d projected, want everything already happened", after.Actual, after.Projected)
	}
}

// FuzzProject_SeriesStayInsideTheYearAndReconcile proves what every consumer relies on, whatever the goal,
// the dates and the velocities: no negative month, twelve months exactly, and the goal stage adding up to
// the goal itself whenever the activity covers any of the plan year (SC-003, SC-005).
func FuzzProject_SeriesStayInsideTheYearAndReconcile_SC003(f *testing.F) {
	f.Add(int64(365_000), 1, 12, 1, 2)
	f.Add(int64(1), 6, 6, 0, 0)
	f.Add(int64(0), 1, 12, 24, 24)
	f.Add(int64(1_000_000_000), 12, 1, 3, 4)

	f.Fuzz(func(t *testing.T, goal int64, startMonth, endMonth, inqVelocity, mqlVelocity int) {
		in := inputs(goal)
		in.Model.InqToMqlVelocityMonths = boundedVelocity(inqVelocity)
		in.Model.MqlToDealVelocityMonths = boundedVelocity(mqlVelocity)
		start := date.New(2026, timeMonth(boundedMonth(startMonth)), 1)
		end := date.New(2026, timeMonth(boundedMonth(endMonth)), 28)
		in.StartDate, in.EndDate = &start, &end

		got := projection.Project(in)

		for _, series := range got.Stages {
			assertSeriesSane(t, series, goal)
		}
		inq, found := got.Find(projection.StageINQ)
		if found && goal > 0 && !start.After(end) && inq.Total != goal {
			t.Fatalf("INQ totals %d, want the whole goal %d phased with nothing lost", inq.Total, goal)
		}
	})
}

func assertSeriesSane(t *testing.T, series projection.Series, goal int64) {
	t.Helper()
	if len(series.Months) != projection.Months {
		t.Fatalf("%s has %d months, want the plan year's %d", series.Stage, len(series.Months), projection.Months)
	}
	for _, m := range series.Months {
		if m.Value < 0 {
			t.Fatalf("%s %s = %d, want no negative month", series.Stage, m.Month, m.Value)
		}
	}
	if series.Total > max(goal, 0) {
		t.Fatalf("%s totals %d, want no more than the goal %d: a later stage only ever loses volume",
			series.Stage, series.Total, goal)
	}
}

func boundedVelocity(n int) int {
	if n < 0 {
		n = -n
	}
	return n % (projection.Months*2 + 1)
}

func boundedMonth(n int) int {
	if n < 0 {
		n = -n
	}
	return n%projection.Months + 1
}

func timeMonth(n int) time.Month { return time.Month(n) }
