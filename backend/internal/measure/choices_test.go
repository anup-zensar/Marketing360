package measure_test

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure"
)

func TestChoices_FR008_FR011(t *testing.T) {
	t.Parallel()
	choices := frameOf(t, frameService(fakePlanning{years: []int{2025}}), url.Values{}).Choices

	if diff := cmp.Diff([]int{2025, 2026}, choices.Years); diff != "" {
		t.Errorf("years with plans plus the current (-want +got):\n%s", diff)
	}
	want := []measure.PlanChoice{{ID: "P-1", Name: "FY2026 Lubricants", PlanType: "operational", TeamIDs: []string{"T-1"}}}
	if diff := cmp.Diff(want, choices.Plans); diff != "" {
		t.Errorf("plans of the year in scope (-want +got):\n%s", diff)
	}
	owners := []measure.NamedChoice{{ID: "U-3", Name: "Daniel Okoro"}, {ID: "U-2", Name: "Priya Raman"}}
	types := []measure.NamedChoice{{ID: "AT-2", Name: "Digital"}, {ID: "AT-1", Name: "Event"}}
	if diff := cmp.Diff(owners, choices.Owners); diff != "" {
		t.Errorf("owners of the selected plans, by name (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(types, choices.ActivityTypes); diff != "" {
		t.Errorf("activity types of the selected plans, by name (-want +got):\n%s", diff)
	}
	type sizes struct{ Teams, Currencies, Statuses, PlanTypes int }
	got := sizes{len(choices.Teams), len(choices.Currencies), len(choices.ActivityStatuses), len(choices.PlanTypes)}
	if diff := cmp.Diff(sizes{2, 2, 6, 3}, got); diff != "" || !choices.Teams[0].Mine {
		t.Errorf("teams (the person's marked), currencies, statuses and plan types (-want +got):\n%s", diff)
	}
}

// One group per active choice or yes/no attribute of each level, never merged by label; an inactive option only
// while a record has it; no text, long-text or number attribute; Not filled in as none, last (FR-008, FR-010).
func TestChoices_AttributeGroups_FR008(t *testing.T) {
	t.Parallel()
	groups := frameOf(t, frameService(fakePlanning{}), url.Values{}).Choices.AttributeGroups

	ids := func(values []measure.AttributeValue) []string {
		var got []string
		for _, v := range values {
			got = append(got, v.ID)
		}
		return got
	}
	got := map[string][]string{}
	for _, g := range groups {
		got[g.Level+" "+g.Label] = ids(g.Values)
	}
	want := map[string][]string{
		"campaign Cluster": {"SO-EU", "SO-AS", "none"},
		"activity Brand":   {"SO-HX", "SO-OLD", "none"},
		"activity FIT4":    {"yes", "no", "none"},
		"activity Cluster": {"SO-X", "none"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("attribute groups (-want +got):\n%s", diff)
	}
}

// Plans of another year in the address are dropped and counted (FR-011).
func TestChoices_PlansOfAnotherYearDropped_FR011(t *testing.T) {
	t.Parallel()
	frame := frameOf(t, frameService(fakePlanning{}), url.Values{"plan": {"P-1", "P-2025"}})

	if frame.LeftOut != 1 || len(frame.Filter.Plans) != 1 {
		t.Errorf("left out %d, plans %v: want P-2025 dropped and counted", frame.LeftOut, frame.Filter.Plans)
	}
}
