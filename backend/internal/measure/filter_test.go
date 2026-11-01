package measure_test

import (
	"errors"
	"net/url"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/errs"
)

// reportQuery is what a report reads of All work's query: the filters of data-model §2.
type reportQuery struct {
	Year                                                                     int
	Plans, PlanTypes, Teams, Owners, ActivityTypes, ActivityStatuses, Others []string
	Attributes                                                               []allwork.AttrFilter
}

func reportQueryOf(q allwork.Query) reportQuery {
	others := []string{q.View, q.Search}
	return reportQuery{q.Year, q.Plans, q.PlanTypes, q.Teams, q.Owners, q.ActivityTypes, q.ActivityStatuses, others,
		q.Attributes}
}

func TestParseFilter_FR008_FR030(t *testing.T) {
	t.Parallel()
	f, err := measure.ParseFilter(url.Values{"year": {"2026"}, "plan": {"P-1", "P-2"}, "planType": {"operational"},
		"team": {"T-1"}, "owner": {"U-1"}, "activityType": {"AT-1"}, "activityStatus": {"approved"},
		"attr": {"SA-1:GB01,none"}, "period": {"quarter"}, "show": {"campaign"}, "currency": {"EUR"},
		"table": {"attr12Count", "statusCount"}, "view": {"budget"}, "search": {"launch"}})
	if err != nil {
		t.Fatalf("ParseFilter() error = %v", err)
	}

	want := measure.Filter{Period: measure.PeriodQuarter, Show: measure.ShowCampaign, Currency: "EUR",
		Tables: []string{"attr12Count", "statusCount"}, Own: url.Values{"view": {"budget"}, "search": {"launch"}}}
	if diff := cmp.Diff(want, f, cmpopts.IgnoreFields(measure.Filter{}, "Query")); diff != "" {
		t.Errorf("controls (-want +got):\n%s", diff)
	}
	wantQuery := reportQuery{Year: 2026, Plans: []string{"P-1", "P-2"}, PlanTypes: []string{"operational"},
		Teams: []string{"T-1"}, Owners: []string{"U-1"}, ActivityTypes: []string{"AT-1"},
		ActivityStatuses: []string{"approved"}, Others: []string{allwork.ViewGrid, ""},
		Attributes: []allwork.AttrFilter{{AttributeID: "SA-1", Values: []string{"GB01", "none"}}}}
	if diff := cmp.Diff(wantQuery, reportQueryOf(f.Query)); diff != "" {
		t.Errorf("All work's report parameters only, view and search ignored (-want +got):\n%s", diff)
	}
}

func TestParseFilter_NoParameterRestrictsNothing_FR030(t *testing.T) {
	t.Parallel()
	f, err := measure.ParseFilter(url.Values{})
	if err != nil {
		t.Fatalf("ParseFilter() error = %v", err)
	}

	want := measure.Filter{Period: measure.PeriodMonth, Tables: []string{}, Own: url.Values{}}
	if diff := cmp.Diff(want, f, cmpopts.IgnoreFields(measure.Filter{}, "Query")); diff != "" ||
		reportQueryOf(f.Query).Year != 0 || len(f.Query.Plans)+len(f.Query.Attributes) != 0 {
		t.Errorf("every value, months, the person's currency, charts (-want +got):\n%s", diff)
	}
}

func TestParseFilter_EveryProblemAtOnce(t *testing.T) {
	t.Parallel()
	_, err := measure.ParseFilter(url.Values{"period": {"year"}, "show": {"lineItem"}, "currency": {"eur"},
		"table": {"Attr-1"}, "activityStatus": {"lost"}})

	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("ParseFilter() error = %v, want field problems", err)
	}
	var fields []string
	for _, f := range e.Fields {
		fields = append(fields, f.Field)
	}
	want := []string{"activityStatus", "currency", "period", "show", "table"}
	if diff := cmp.Diff(want, slices.Sorted(slices.Values(fields))); diff != "" {
		t.Errorf("problems (-want +got):\n%s", diff)
	}
}

func TestCanonicalQuery_FR024(t *testing.T) {
	t.Parallel()
	f, err := measure.ParseFilter(url.Values{"table": {"statusCount", "attr12Count"}, "show": {"campaign"},
		"attr": {"SA-2:SO-2,SO-1", "SA-1:none"}, "plan": {"P-2", "P-1"}, "period": {"month"}, "year": {"2026"}})
	if err != nil {
		t.Fatalf("ParseFilter() error = %v", err)
	}

	got := measure.Canonical(f)

	want := "year=2026&plan=P-1&plan=P-2&attr=SA-1%3Anone&attr=SA-2%3ASO-1%2CSO-2&show=campaign" +
		"&table=attr12Count&table=statusCount"
	if got != want {
		t.Errorf("Canonical() = %q, want %q (data-model order, values sorted, the default month left out)", got, want)
	}
	values, err := url.ParseQuery(got)
	if err != nil {
		t.Fatal(err)
	}
	again, err := measure.ParseFilter(values)
	if err != nil || measure.Canonical(again) != got {
		t.Errorf("round trip = %q, %v; want the same query", measure.Canonical(again), err)
	}
}

func TestParseFilter_KeepsTheReportsOwnParameters_039_B4(t *testing.T) {
	t.Parallel()
	values, _ := url.ParseQuery("year=2026&report=gsapLink&tab=waiting&sort=-age&owner=U-1002")
	f, err := measure.ParseFilter(values)
	if err != nil {
		t.Fatalf("ParseFilter() error = %v", err)
	}
	if diff := cmp.Diff(url.Values{"tab": {"waiting"}, "sort": {"-age"}}, f.Own); diff != "" {
		t.Errorf("the report's own parameters (-want +got):\n%s", diff)
	}
}
