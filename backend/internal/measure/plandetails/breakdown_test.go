package plandetails

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/allwork/allworktest"
	"mapp/internal/history"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/money"
)

// yearSource is an allwork.Source over one in-memory year, so All work's own Breakdown can be compared.
type yearSource struct {
	nodes      []allwork.Node
	attributes []allwork.Attribute
}

func (s yearSource) Years(context.Context) ([]int, error) { return []int{testYearNumber}, nil }

func (s yearSource) Year(context.Context, int, date.Date, []string) (allwork.YearData, error) {
	return allwork.YearData{Nodes: s.nodes}, nil
}

func (s yearSource) PlanYear(context.Context, string, int, date.Date) (allwork.YearData, error) {
	return allwork.YearData{}, nil
}

func (s yearSource) Choices(context.Context, string) (allwork.Choices, error) {
	return allwork.Choices{}, nil
}

func (s yearSource) Attributes(context.Context) ([]allwork.Attribute, error) {
	return s.attributes, nil
}

func (s yearSource) Currency(context.Context, string) (allwork.Currency, error) {
	return usdCurrency(), nil
}

func (s yearSource) Locate(context.Context, string) (allwork.Location, bool, error) {
	return allwork.Location{}, false, nil
}

func (s yearSource) InRecycleBin(context.Context, string) (bool, error) { return false, nil }

func (s yearSource) RatesUsed(context.Context) ([]string, error) { return nil, nil }

// breakdownYear mixes every choice and yes/no kind, empty values, uneven splits and a Business unit condition.
func breakdownYear(t *testing.T) (*year, []allwork.Attribute) {
	t.Helper()
	launch := attribute("SA-5", "Product launch", allwork.TypeSingleChoice, "NEW", "OLD")
	launch.Condition = []string{"BU-LUB"}
	attributes := []allwork.Attribute{
		companyCode(),
		attribute("SA-2", "Country / OU", allwork.TypeMultipleChoiceSplit, "DE", "FR", "IT"),
		attribute("SA-4", "Channel / audience", allwork.TypeMultipleChoice, "EXT", "INT"),
		launch,
		attribute("SA-6", "Agency involved", allwork.TypeYesNo),
		atLevel(attribute("SA-10", "Plan focus", allwork.TypeSingleChoice, "B2B"), allwork.LevelPlan),
		cluster(),
		atLevel(attribute("SA-12", "Theme", allwork.TypeMultipleChoice, "T1", "T2"), allwork.LevelInitiative),
	}
	y := newYear(t).plan("P-1", "BU-LUB").on("P-1", choice("SA-10", "B2B")).campaign("C-1", "P-1", choice("SA-11", "EU")).
		initiative("I-1", "C-1", choice("SA-12", "T1", "T2")).plan("P-2", "BU-MOB").campaign("C-2", "P-2")
	splits := []map[string]string{{"DE": "33.33", "FR": "33.33", "IT": "33.34"}, {"DE": "70.00", "FR": "30.00"},
		{"FR": "100.00", "IT": "0.00"}}
	for i := range 30 {
		changes := []func(*allwork.Node){split(t, "SA-2", splits[i%3])}
		if i%4 != 0 {
			changes = append(changes, choice("SA-1", []string{"GB01", "CN04"}[i%2]))
		}
		if i%5 != 0 {
			changes = append(changes, choice("SA-4", []string{"EXT", "INT"}[:1+i%2]...), choice("SA-5", "NEW"))
		}
		if i%3 != 0 {
			changes = append(changes, yesNo("SA-6", i%2 == 0))
		}
		y.activity(fmt.Sprintf("A-%d", i+1), []string{"C-1", "C-2", "I-1"}[i%3], fmt.Sprintf("%d.%02d", 100+i*37, i*7%100),
			changes...)
	}
	return y, attributes
}

func TestAttribution_EqualsAllWorkBreakdown_SC002(t *testing.T) {
	t.Parallel()
	y, attributes := breakdownYear(t)
	allWork := allWorkOf(yearSource{nodes: y.nodes, attributes: attributes})

	for _, a := range attributes {
		t.Run(a.Label, func(t *testing.T) {
			t.Parallel()
			br, err := allWork.Breakdown(t.Context(), actor.Actor{ID: "U-1002"},
				allwork.Query{Year: testYearNumber, By: "attr:" + a.ID})
			if err != nil {
				t.Fatalf("Breakdown() error = %v", err)
			}
			want := map[string]string{}
			for _, g := range br.Groups {
				want[g.Key] = g.PlannedCost.Amount.String()
			}
			rows := newView(y.matching(), Level(a.Level), measure.PeriodMonth).rows([]allwork.Attribute{a})
			got := plannedCosts(t, rows[0].Widgets[1])
			// An activity without a record of the level (an activity right under its campaign has no initiative) is
			// on no tab of that level; the Breakdown counts it as Not filled in.
			if outside := outsideLevel(y.nodes, a.Level); !outside.IsZero() {
				if text, found := got["notFilledIn"]; found {
					outside = outside.Add(*amount(t, text))
				}
				got["notFilledIn"] = outside.String()
			}
			if got["notFilledIn"] == "0.00" && want["notFilledIn"] == "" { // the Breakdown leaves out an empty group
				delete(got, "notFilledIn")
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("planned cost by value against All work's Breakdown (-want +got):\n%s", diff)
			}
		})
	}
}

func outsideLevel(nodes []allwork.Node, level allwork.Level) money.Amount {
	byID := map[string]*allwork.Node{}
	for i := range nodes {
		byID[nodes[i].ID] = &nodes[i]
	}
	var outside money.Amount
	for i := range nodes {
		if nodes[i].Level == allwork.LevelActivity && ancestorAt(byID, &nodes[i], level) == nil {
			outside = outside.Add(plannedCostOf(&nodes[i]))
		}
	}
	return outside
}

// plannedCosts returns the widget's amount per value with Not filled in and Does not apply joined under
// Breakdown's key for Not filled in.
func plannedCosts(t *testing.T, w Widget) map[string]string {
	t.Helper()
	got := map[string]string{}
	var notFilledIn *string
	for _, v := range w.Values {
		if v.PlannedCost == nil {
			t.Fatalf("value %s has no planned cost", v.Key)
		}
		if v.Key != KeyNotFilledIn && v.Key != KeyDoesNotApply {
			got[v.Key] = v.PlannedCost.Amount.String()
			continue
		}
		sum := v.PlannedCost.Amount
		if notFilledIn != nil {
			previous := *amount(t, *notFilledIn)
			sum = sum.Add(previous)
		}
		text := sum.String()
		notFilledIn = &text
	}
	if notFilledIn != nil {
		got["notFilledIn"] = *notFilledIn
	}
	return got
}

// allWorkOf is All work's own service over an in-memory year, for the tests that compare this report with it.
func allWorkOf(source yearSource) *allwork.Service {
	db := fake.NewDB()
	tools := history.Tools{Tx: db, IDs: ids.NewGenerator(fake.NewCounters(db)), Now: func() time.Time {
		return time.Date(testYearNumber, time.October, 18, 14, 5, 0, 0, time.UTC)
	}}
	return allwork.NewService(tools, allworktest.NewStore(db), source)
}
