package measure_test

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"mapp/internal/history"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
)

func priyaActs() actor.Actor { return actor.Actor{ID: "U-2", Name: "Priya Raman"} }

func frameOf(t *testing.T, s *measure.Service, values url.Values) measure.Frame {
	t.Helper()
	f, err := measure.ParseFilter(values)
	if err != nil {
		t.Fatalf("ParseFilter(%v) error = %v", values, err)
	}
	frame, err := s.Frame(t.Context(), priyaActs(), measure.KeyPlanBudget, f)
	if err != nil {
		t.Fatalf("Frame() error = %v", err)
	}
	return frame
}

// In view counts the matching records and the line items of the matching activities only (FR-007).
func TestFrame_InView_FR007(t *testing.T) {
	t.Parallel()
	frame := frameOf(t, frameService(fakePlanning{}), url.Values{"activityType": {"AT-1"}})

	want := measure.InView{Plans: 1, Campaigns: 1, Activities: 1, LineItems: 2}
	if diff := cmp.Diff(want, frame.InView); diff != "" {
		t.Errorf("in view (-want +got):\n%s", diff)
	}
}

func TestFrame_OnlyReportsScope_FR003(t *testing.T) {
	t.Parallel()
	frame := frameOf(t, frameService(fakePlanning{}), url.Values{})

	body, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	if frame.InView != (measure.InView{Plans: 1, Campaigns: 1, Activities: 2, LineItems: 3}) ||
		strings.Contains(string(body), "P-2") || strings.Contains(string(body), "Wei Zhang") {
		t.Errorf("frame = %s, want P-1 only and nothing of P-2", body)
	}
}

func TestFrame_FinanceCoverage_FR004(t *testing.T) {
	t.Parallel()
	nodes := frameNodes()
	both := frameService(fakePlanning{nodes: nodes, financePlan: []string{"P-1"}})
	none := frameService(fakePlanning{nodes: nodes, financePlan: []string{"P-9"}})

	if got := frameOf(t, both, url.Values{}).Finance; got != (measure.FinanceCoverage{Plans: 1, Of: 1}) {
		t.Errorf("finance = %+v, want 1 of 1", got)
	}
	if got := frameOf(t, none, url.Values{}).Finance; got != (measure.FinanceCoverage{Plans: 0, Of: 1}) {
		t.Errorf("finance = %+v, want 0 of 1", got)
	}
}

// Values that are unknown or out of scope are left out and counted, never named (FR-015).
func TestFrame_LeftOutCountedNotNamed_FR015(t *testing.T) {
	t.Parallel()
	frame := frameOf(t, frameService(fakePlanning{}), url.Values{"plan": {"P-1", "P-2"}, "owner": {"U-404"}})

	body, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	if frame.LeftOut != 2 || strings.Contains(string(body), "P-2") || strings.Contains(string(body), "U-404") {
		t.Errorf("left out %d in %s, want 2 counted and neither named", frame.LeftOut, body)
	}
	if diff := cmp.Diff([]string{"P-1"}, frame.Filter.Plans); diff != "" {
		t.Errorf("applied plans (-want +got):\n%s", diff)
	}
}

func TestFrame_OutOfScopeAndUnknownValuesLeftOut_FR015(t *testing.T) {
	t.Parallel()
	frame := frameOf(t, frameService(fakePlanning{}), url.Values{"owner": {"U-3", "U-9"}, "team": {"T-2"},
		"currency": {"XYZ"}, "attr": {"SA-B:SO-HX,SO-404", "SA-404:SO-X", "SA-OFF:SO-X"}})

	type applied struct {
		LeftOut       int
		Owners, Teams []string
		Attributes    []allwork.AttrFilter
		Currency      string
	}
	got := applied{frame.LeftOut, frame.Filter.Owners, frame.Filter.Teams, frame.Filter.Attributes, frame.Filter.Currency}
	want := applied{6, []string{"U-3"}, []string{}, []allwork.AttrFilter{{AttributeID: "SA-B", Values: []string{"SO-HX"}}}, "GBP"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Wei (another market), T-2, SO-404, SA-404, inactive SA-OFF and XYZ left out (-want +got):\n%s", diff)
	}
	body, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"U-9", "T-2", "SO-404", "SA-404", "SA-OFF", "XYZ"} {
		if strings.Contains(string(body), `"`+id+`"`) {
			t.Errorf("the frame names %s, which was left out", id)
		}
	}
}

func TestFrame_AppliedFilterAndCurrency_FR018_FR030(t *testing.T) {
	t.Parallel()
	s := frameService(fakePlanning{})

	defaults := frameOf(t, s, url.Values{})
	euro := frameOf(t, s, url.Values{"currency": {"EUR"}, "show": {"campaign"}, "period": {"quarter"}})
	unknown := frameOf(t, s, url.Values{"currency": {"XYZ"}})

	type outline struct {
		Year                     int
		Period, Show, Currency   string
		Display, UnknownFallback string
	}
	show := func(f measure.AppliedFilter) string {
		if f.Show == nil {
			return ""
		}
		return *f.Show
	}
	got := []outline{
		{defaults.Filter.Year, defaults.Filter.Period, show(defaults.Filter), defaults.Filter.Currency,
			defaults.DisplayCurrency, unknown.Filter.Currency},
		{euro.Filter.Year, euro.Filter.Period, show(euro.Filter), euro.Filter.Currency, euro.DisplayCurrency, ""},
	}
	want := []outline{{2026, "month", "", "GBP", "GBP", "GBP"}, {2026, "quarter", "campaign", "EUR", "GBP", ""}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("defaults, then EUR and campaign; an unknown code falls back to GBP (-want +got):\n%s", diff)
	}
}

func TestFrame_NoPlansInYear_FR031(t *testing.T) {
	t.Parallel()
	frame := frameOf(t, frameService(fakePlanning{years: []int{2025, 2026}}), url.Values{"year": {"2027"}})

	if frame.InView != (measure.InView{}) || len(frame.Choices.Plans) != 0 || frame.Filter.Year != 2027 {
		t.Errorf("2027: in view %+v, %d plans to choose, year %d: want nothing in view", frame.InView,
			len(frame.Choices.Plans), frame.Filter.Year)
	}
	if diff := cmp.Diff([]int{2025, 2026, 2027}, frame.Choices.Years); diff != "" {
		t.Errorf("years to choose: the ones that have plans and the year shown (-want +got):\n%s", diff)
	}
}

func TestFrame_YearsOfferTheCurrentYear_FR008(t *testing.T) {
	t.Parallel()
	frame := frameOf(t, frameService(fakePlanning{years: []int{2024}}), url.Values{"year": {"2024"}})

	if diff := cmp.Diff([]int{2024, 2026}, frame.Choices.Years); diff != "" {
		t.Errorf("years to choose on 2024 with plans only in 2024, today in 2026 (-want +got):\n%s", diff)
	}
}

func TestFrame_YearsOfTheReportsScope_FR008(t *testing.T) {
	t.Parallel()
	s := measure.NewService(history.Tools{Tx: fakeTx{}, Now: func() time.Time { return time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC) }},
		measure.Deps{Planning: fakePlanning{nodes: frameNodes(), years: []int{2024, 2025}},
			Plans: fakePlans{ids: []string{"P-1"}, years: []int{2025}}, Currencies: fakeCurrencies{}}, measure.NewRegistry())

	frame := frameOf(t, s, url.Values{"year": {"2025"}})

	if diff := cmp.Diff([]int{2025, 2026}, frame.Choices.Years); diff != "" {
		t.Errorf("2024 has plans Priya may plan but not report on (-want +got):\n%s", diff)
	}
}

func TestFrame_UnknownReport_404(t *testing.T) {
	t.Parallel()
	_, err := frameService(fakePlanning{}).Frame(t.Context(), priyaActs(), "nope", measure.Filter{})

	var e *errs.Error
	if !errors.As(err, &e) || e.Status != 404 {
		t.Errorf("Frame(nope) error = %v, want 404", err)
	}
}
