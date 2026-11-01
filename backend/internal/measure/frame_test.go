package measure_test

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

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

func TestFrame_UnknownReport_404(t *testing.T) {
	t.Parallel()
	_, err := frameService(fakePlanning{}).Frame(t.Context(), priyaActs(), "nope", measure.Filter{})

	var e *errs.Error
	if !errors.As(err, &e) || e.Status != 404 {
		t.Errorf("Frame(nope) error = %v, want 404", err)
	}
}
