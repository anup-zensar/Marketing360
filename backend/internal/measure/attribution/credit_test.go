package attribution_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure/attribution"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

func day(d int) date.Date { return date.New(2026, 3, d) }

func journey(revenue int64, touches ...attribution.Touch) attribution.Conversion {
	return attribution.Conversion{ActivityID: "A-9", Date: day(25), Revenue: money.FromCents(revenue), Journey: true,
		Touches: touches}
}

func credits(cs []attribution.Credit) []string {
	var out []string
	for _, c := range cs {
		out = append(out, fmt.Sprintf("%s %s", c.ActivityID, c.Amount))
	}
	return out
}

func threeTouches() attribution.Conversion {
	return journey(100000, attribution.Touch{ActivityID: "A-3", Date: day(20)},
		attribution.Touch{ActivityID: "A-1", Date: day(1)}, attribution.Touch{ActivityID: "A-2", Date: day(10)})
}

func TestCredit_FiveModels_050_FR003(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		model attribution.Model
		want  []string
	}{
		{attribution.FirstTouch, []string{"A-1 1000.00"}},
		{attribution.LastTouch, []string{"A-3 1000.00"}},
		{attribution.Equal, []string{"A-1 333.33", "A-2 333.33", "A-3 333.34"}},
		{attribution.LinearDecay, []string{"A-1 166.66", "A-2 333.33", "A-3 500.01"}},
		{attribution.UShaped, []string{"A-1 400.00", "A-2 200.00", "A-3 400.00"}},
	} {
		if diff := cmp.Diff(c.want, credits(attribution.Attribute(threeTouches(), c.model))); diff != "" {
			t.Errorf("%s (-want +got):\n%s", c.model, diff)
		}
	}
	two := journey(100001, attribution.Touch{ActivityID: "A-1", Date: day(1)}, attribution.Touch{ActivityID: "A-2", Date: day(2)})
	if diff := cmp.Diff([]string{"A-1 500.00", "A-2 500.01"}, credits(attribution.Attribute(two, attribution.UShaped))); diff != "" {
		t.Errorf("U-shaped with two touches: half each, the cent to the newest (-want +got):\n%s", diff)
	}
}

func TestCredit_ResidualExact_FR005(t *testing.T) {
	t.Parallel()
	models := []attribution.Model{attribution.FirstTouch, attribution.LastTouch, attribution.Equal,
		attribution.LinearDecay, attribution.UShaped}
	for _, revenue := range []int64{1, 7, 100, 99999, 123457} {
		for n := 1; n <= 7; n++ {
			c := journey(revenue)
			for i := range n {
				c.Touches = append(c.Touches, attribution.Touch{ActivityID: fmt.Sprintf("A-%d", i), Date: day(i + 1)})
			}
			for _, m := range models {
				var sum money.Amount
				for _, credit := range attribution.Attribute(c, m) {
					sum = sum.Add(credit.Amount)
				}
				if sum.Cents() != revenue {
					t.Errorf("%s, %d touches, %d cents: credited %d", m, n, revenue, sum.Cents())
				}
			}
		}
	}
}

func TestCredit_SingleTouchLateAndUnmatched_FR004_FR006(t *testing.T) {
	t.Parallel()
	single := attribution.Conversion{ActivityID: "A-9", Date: day(25), Revenue: money.FromCents(5000)}
	if diff := cmp.Diff([]string{"A-9 50.00"}, credits(attribution.Attribute(single, attribution.FirstTouch))); diff != "" {
		t.Errorf("no journey: all to its own activity (-want +got):\n%s", diff)
	}
	late := journey(9000, attribution.Touch{ActivityID: "A-1", Date: day(1)}, attribution.Touch{ActivityID: "A-2", Date: day(28)})
	if diff := cmp.Diff([]string{"A-1 90.00"}, credits(attribution.Attribute(late, attribution.Equal))); diff != "" {
		t.Errorf("a touch after the conversion gets nothing (-want +got):\n%s", diff)
	}
	unmatched := journey(9000, attribution.Touch{ActivityID: "A-1", Date: day(1)}, attribution.Touch{Date: day(2)},
		attribution.Touch{ActivityID: "A-1", Date: day(3)})
	if diff := cmp.Diff([]string{"A-1 60.00", " 30.00"}, credits(attribution.Attribute(unmatched, attribution.Equal))); diff != "" {
		t.Errorf("an unmatched touch's share is Unattributed; one activity's touches add up (-want +got):\n%s", diff)
	}
	allLate := journey(9000, attribution.Touch{ActivityID: "A-1", Date: day(30)})
	if diff := cmp.Diff([]string{" 90.00"}, credits(attribution.Attribute(allLate, attribution.LastTouch))); diff != "" {
		t.Errorf("no touch on or before the conversion: all Unattributed (-want +got):\n%s", diff)
	}
}

func TestROI_FR007(t *testing.T) {
	t.Parallel()
	cost := func(cents int64) *money.Amount { a := money.FromCents(cents); return &a }
	for _, c := range []struct {
		revenue int64
		cost    *money.Amount
		want    string
	}{
		{150000, cost(100000), "50"},
		{50000, cost(100000), "-50"},
		{100500, cost(100000), "1"},
		{99500, cost(100000), "-1"},
		{150000, nil, "—"},
		{150000, cost(0), "—"},
	} {
		got := "—"
		if roi := attribution.ROI(money.FromCents(c.revenue), c.cost); roi != nil {
			got = fmt.Sprint(*roi)
		}
		if got != c.want {
			t.Errorf("ROI(%d, %v) = %s, want %s", c.revenue, c.cost, got, c.want)
		}
	}
}
