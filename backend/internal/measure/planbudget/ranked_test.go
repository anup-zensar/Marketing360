package planbudget_test

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure"
	"mapp/internal/measure/planbudget"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/sheet"
)

func ranked(t *testing.T, f measure.Filter, query string) planbudget.Ranked {
	t.Helper()
	values, _ := url.ParseQuery(query)
	q, err := planbudget.ParseRanked(values)
	if err != nil {
		t.Fatalf("ParseRanked(%q): %v", query, err)
	}
	r, err := service(scope(everything, true)).Ranked(t.Context(), priya, f, q)
	if err != nil {
		t.Fatalf("Ranked: %v", err)
	}
	return r
}

// lines are the page's groups and entries as text: "group: entry planned [children]".
func lines(r planbudget.Ranked) []string {
	var out []string
	for _, g := range r.Groups {
		for _, e := range g.Entries {
			var children []string
			for _, c := range e.Children {
				children = append(children, c.Name)
			}
			out = append(out, fmt.Sprintf("%s: %s %s %v", g.Label, e.Name, shown(e.Figures.Planned), children))
		}
	}
	return out
}

func TestRanked_LevelsAndSort_FR008_FR009(t *testing.T) {
	t.Parallel()
	byPlanned := ranked(t, measure.Filter{}, "")
	if diff := cmp.Diff([]string{": Helix launch 300.00 [Launch events]", ": Rimula fleet 200.00 [Fleet days]"},
		lines(byPlanned)); diff != "" || byPlanned.Level != "campaign" || byPlanned.Sort != planbudget.SortPlanned {
		t.Errorf("campaigns by planned cost, largest first (-want +got):\n%s", diff)
	}
	byLeft := ranked(t, measure.Filter{Show: measure.ShowInitiative}, "sort=leftToPlan&order=asc")
	if diff := cmp.Diff([]string{": Fleet days -100.00 [Fleet day]", ": Launch events 100.00 [Trade show]"},
		leftToPlanLines(byLeft)); diff != "" {
		t.Errorf("initiatives by left to plan, lowest first (-want +got):\n%s", diff)
	}
	fleet := byLeft.Groups[0].Entries[0]
	if !fleet.Figures.OverAllocated || byLeft.Groups[0].Entries[1].Figures.OverAllocated {
		t.Errorf("Fleet days plans 200 against a budget of 100: over-allocated = %t", fleet.Figures.OverAllocated)
	}
}

func leftToPlanLines(r planbudget.Ranked) []string {
	var out []string
	for _, e := range r.Groups[0].Entries {
		out = append(out, fmt.Sprintf(": %s %s [%s]", e.Name, shown(e.Figures.LeftToPlan), e.Children[0].Name))
	}
	return out
}

func TestRanked_Drill_FR010(t *testing.T) {
	t.Parallel()
	toActivity := ranked(t, measure.Filter{}, "drill=A-1")

	var trail []string
	for _, step := range toActivity.Trail {
		trail = append(trail, step.Name)
	}
	if diff := cmp.Diff([]string{"FY2026 Lubricants", "Helix launch", "Launch events", "Trade show"}, trail); diff != "" {
		t.Errorf("trail (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{": Trade show 300.00 [Stand]"}, lines(toActivity)); diff != "" {
		t.Errorf("a drilled activity lists its line items (040-C1) (-want +got):\n%s", diff)
	}
	lost := ranked(t, measure.Filter{}, "drill=A-99")
	if !lost.DrillLost || len(lines(lost)) != 2 {
		t.Errorf("a drill to a record not in view: lost %t, %d entries; want the level's list back", lost.DrillLost, len(lines(lost)))
	}
}

func TestRanked_TotalsEqualCards_FR015(t *testing.T) {
	t.Parallel()
	cards := report(t, scope(everything, true), measure.Filter{}).Cards
	for _, show := range []string{measure.ShowPlan, measure.ShowCampaign, measure.ShowInitiative, measure.ShowActivity} {
		totals := ranked(t, measure.Filter{Show: show}, "").Groups[0].Totals
		got := []string{shown(totals.Planned), shown(totals.Committed), shown(totals.Actual)}
		if diff := cmp.Diff([]string{shown(cards.Planned), shown(cards.Committed), shown(cards.Actual)}, got); diff != "" {
			t.Errorf("%s totals against the cards (-want +got):\n%s", show, diff)
		}
	}
}

func TestRanked_Groups_FR013(t *testing.T) {
	t.Parallel()
	byOwner := lines(ranked(t, measure.Filter{}, "group=owner"))
	if diff := cmp.Diff([]string{"Priya Raman: Helix launch 300.00 [Launch events]",
		"Daniel Okoro: Rimula fleet 200.00 [Fleet days]"}, byOwner); diff != "" {
		t.Errorf("by owner (-want +got):\n%s", diff)
	}
	byBrand := lines(ranked(t, measure.Filter{}, "group=attr:SA-1022"))
	if diff := cmp.Diff([]string{"Not filled in: Rimula fleet 200.00 [Fleet days]", "Helix: Helix launch 180.00 [Launch events]",
		"Pennzoil: Helix launch 120.00 [Launch events]"}, byBrand); diff != "" {
		t.Errorf("by Brand, the 60/40 split counted by share (-want +got):\n%s", diff)
	}
}

func TestRanked_PagesAndSearch_FR014(t *testing.T) {
	t.Parallel()
	second := ranked(t, measure.Filter{}, "page=2")
	if second.Total != 2 || len(second.Groups) != 0 {
		t.Errorf("page 2 of 2 campaigns at 10 a page: total %d, %d groups", second.Total, len(second.Groups))
	}
	found := ranked(t, measure.Filter{}, "q=RIM")
	if diff := cmp.Diff([]string{": Rimula fleet 200.00 [Fleet days]"}, lines(found)); diff != "" || found.Total != 1 {
		t.Errorf("search ignores case (-want +got):\n%s", diff)
	}
}

func TestParseRanked_FR014(t *testing.T) {
	t.Parallel()
	values, _ := url.ParseQuery("sort=name&order=up&group=team&size=7&page=0&drill=bad&q=" + strings.Repeat("x", 101))
	_, err := planbudget.ParseRanked(values)

	var e *errs.Error
	if !errors.As(err, &e) || len(e.Fields) != 7 {
		t.Fatalf("ParseRanked = %v, want the seven problems at once", err)
	}
}

func TestSheets_FR016(t *testing.T) {
	t.Parallel()
	f := measure.Filter{Own: url.Values{"sort": {"actual"}, "order": {"asc"}, "page": {"2"}}}
	sheets, err := planbudget.NewReport(service(scope(everything, true))).Sheets(t.Context(), priya, f)
	if err != nil {
		t.Fatalf("Sheets: %v", err)
	}
	var got []string
	for _, s := range sheets {
		got = append(got, fmt.Sprintf("%s %s %d", s.Key, s.Sheet.Name, len(s.Sheet.Rows)))
	}
	if diff := cmp.Diff([]string{"cards Cards 11", "periods Periods 14", "ranked Ranked list 5"}, got); diff != "" {
		t.Errorf("sheets: cards, months with a total, every campaign with its child whatever the page (-want +got):\n%s", diff)
	}
	if sheets[2].Sheet.Rows[1][2] != sheet.Text("Rimula fleet") {
		t.Errorf("first ranked row = %+v, want Rimula fleet: lowest actual first, as on screen", sheets[2].Sheet.Rows[1])
	}
}
