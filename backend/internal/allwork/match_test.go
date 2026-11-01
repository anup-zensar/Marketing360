package allwork

import (
	"maps"
	"net/url"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// matchFilters are one filter of every kind on twoYears (039 T015): plan, plan type, team, owner, activity type and
// status, an attribute value, Not filled in, a % split share above 0, a value hidden by a Business unit condition, a
// line-item attribute and a value that is not known.
func matchFilters() map[string]url.Values {
	return map[string]url.Values{
		"plan":           {"plan": {"P-1001"}},
		"plan type":      {"planType": {"crossMarket"}},
		"team":           {"team": {"T-1001"}},
		"owner":          {"owner": {daniel.ID}},
		"activity type":  {"activityType": {"AT-1001"}},
		"activity state": {"activityStatus": {"created", "approved"}},
		"attribute":      {"attr": {"SA-B:SO-HX"}},
		"not filled in":  {"attr": {"SA-B:none"}},
		"split share":    {"attr": {"SA-C:SO-FR"}},
		"hidden value":   {"attr": {"SA-M:SO-X,none"}},
		"line item":      {"attr": {"SA-S:SO-ME"}},
		"unknown value":  {"owner": {"U-9999"}, "plan": {"P-1001", "P-9999"}},
	}
}

func matchNodes() []Node {
	nodes := withAttr(withAttr(twoYears(), "A-1", "SA-M", options("SO-X")), "A-3", "SA-M", options("SO-X"))
	for i := range nodes {
		switch nodes[i].ID {
		case "A-1":
			nodes[i].Status = statusApproved
		case "A-2", "A-4":
			nodes[i].Status = statusCreated
		}
	}
	return nodes
}

// TestMatch_EqualsAllWorkCounts_039_FR012 checks every filter alone and every pair of them (the mix) against the
// summary and rows All work's own grid shows for the same filters.
func TestMatch_EqualsAllWorkCounts_039_FR012(t *testing.T) {
	t.Parallel()
	filters := matchFilters()
	names := slices.Sorted(maps.Keys(filters))
	for i, first := range names {
		for _, second := range names[i:] {
			values := url.Values{}
			maps.Copy(values, filters[first])
			for key, v := range filters[second] {
				values[key] = append(slices.Clone(values[key]), v...)
			}
			t.Run(first+" and "+second, func(t *testing.T) {
				t.Parallel()
				assertMatchEqualsGrid(t, values)
			})
		}
	}
}

func TestMatch_KeepsOnlyTheMatches_039_FR012(t *testing.T) {
	t.Parallel()
	q, err := ParseQuery(url.Values{"attr": {"SA-C:SO-FR"}})
	if err != nil {
		t.Fatal(err)
	}

	got := Match(matchNodes(), q, rangeAttributes, fixtureChoices(priya.ID))

	if !slices.Equal(got.Activities, []string{"A-1"}) || !slices.Equal(got.LineItems, []string{"L-1", "L-2"}) {
		t.Errorf("France above 0 %% = %v with %v, want A-1 with L-1 and L-2", got.Activities, got.LineItems)
	}
}

func assertMatchEqualsGrid(t *testing.T, values url.Values) {
	t.Helper()
	q, err := ParseQuery(values)
	if err != nil {
		t.Fatalf("ParseQuery(%v) error = %v", values, err)
	}
	nodes := matchNodes()

	got := Match(nodes, q, rangeAttributes, fixtureChoices(priya.ID))
	grid := rangeView(nodes, q, 2026, 2027).result

	want := Matched{
		Plans: matchingRows(grid, LevelPlan), Campaigns: matchingRows(grid, LevelCampaign),
		Initiatives: matchingRows(grid, LevelInitiative), Activities: matchingRows(grid, LevelActivity),
		LineItems: lineItemsOf(grid, got.Activities), LeftOut: grid.LeftOut,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Match(%v) (-grid +match):\n%s", values, diff)
	}
	if s := grid.Summary; s.Campaigns != len(got.Campaigns) || s.Initiatives != len(got.Initiatives) ||
		s.Activities != len(got.Activities) {
		t.Errorf("summary %d · %d · %d, match %d · %d · %d", s.Campaigns, s.Initiatives, s.Activities,
			len(got.Campaigns), len(got.Initiatives), len(got.Activities))
	}
}

// matchingRows returns the grid's rows of level whose record matches the filters rather than only holding a match.
func matchingRows(r Result, level Level) []string {
	keys := []string{}
	for _, row := range r.Rows {
		if row.Level == level && row.Role == RoleMatch {
			keys = append(keys, row.RecordID)
		}
	}
	return keys
}

func lineItemsOf(r Result, activities []string) []string {
	keys := []string{}
	for _, row := range r.Rows {
		if row.Level == LevelLineItem && slices.Contains(activities, row.ActivityID) {
			keys = append(keys, row.RecordID)
		}
	}
	return keys
}
