package plandetails

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/platform/money"
)

func usdCurrency() allwork.Currency {
	return allwork.Currency{Code: money.BaseCurrency, Rate: money.One}
}

func valuePart(key string, order int) part {
	return part{key: key, label: key, order: order, counted: true, linkable: true}
}

func TestValues_LargestFirstThenNoneAndDoesNotApply_FR009(t *testing.T) {
	t.Parallel()
	tl := newTally()
	add := func(n int, p part) {
		for i := range n {
			tl.add(fmt.Sprintf("%s-%d", p.key, i), []part{p}, money.Amount{})
		}
	}
	add(1, valuePart("a", 0))
	add(5, notFilledInPart())
	add(3, valuePart("b", 1))
	add(2, doesNotApplyPart())

	w := tl.widget(spec{key: "attr1Count", title: "Activities by X", measure: MeasureCount}, usdCurrency())

	if diff := cmp.Diff([]string{"b", "a", KeyNotFilledIn, KeyDoesNotApply}, keys(w)); diff != "" || w.Total.Count != 11 {
		t.Errorf("values (-want +got):\n%s total %d, want 11", diff, w.Total.Count)
	}
}

func TestOther_AfterTenLargest_FR017(t *testing.T) {
	t.Parallel()
	tl := newTally()
	for v := range 12 {
		for i := range 12 - v {
			tl.add(fmt.Sprintf("R-%d-%d", v, i), []part{valuePart(fmt.Sprintf("V%02d", v), v)}, money.Amount{})
		}
	}

	w := tl.widget(spec{key: "attr1Count", title: "Activities by X", measure: MeasureCount}, usdCurrency())

	if len(w.Values) != 12 || w.Other == nil || w.Other.Key != KeyOther || w.Other.Count != 3 || w.Other.Linkable {
		t.Errorf("values %d, other %+v: want every value listed and the last two (2 + 1) as Other", len(w.Values), w.Other)
	}
}

func TestPlannedCost_InReportCurrency_FR013(t *testing.T) {
	t.Parallel()
	rate, err := money.ParseRate("0.50")
	if err != nil {
		t.Fatal(err)
	}
	tl := newTally()
	for i := range 3 {
		cent := *amount(t, "0.01")
		p := valuePart("a", 0)
		p.amount = cent
		tl.add(fmt.Sprintf("A-%d", i), []part{p}, cent)
	}

	w := tl.widget(spec{key: "attr1PlannedCost", title: "Planned cost by X", measure: MeasurePlannedCost},
		allwork.Currency{Code: "GBP", Rate: rate})

	if got := w.Values[0].PlannedCost; got == nil || got.Amount.String() != "0.02" || got.Currency != "GBP" ||
		w.Total.PlannedCost.Amount.String() != "0.02" {
		t.Errorf("planned cost %+v total %+v, want 0.03 USD summed, then converted once: 0.02 GBP", got, w.Total.PlannedCost)
	}
}

func TestKeys_ReportWidgetPattern(t *testing.T) {
	t.Parallel()
	pattern := regexp.MustCompile(`^[a-z][A-Za-z0-9]{0,39}$`)
	for _, key := range []string{attrKey("SA-1003", suffixCount), attrKey("SA-7", suffixPlannedCost),
		attrKey("SA-12", suffixFilledIn)} {
		if !pattern.MatchString(key) {
			t.Errorf("key %q does not match 039's ReportWidget pattern", key)
		}
	}
	if got := attrKey("SA-12", suffixCount); got != "attr12Count" {
		t.Errorf("attrKey(SA-12) = %q, want attr12Count", got)
	}
}

func TestValues_LinkableAndFilterParam_FR014(t *testing.T) {
	t.Parallel()
	tl := newTally()
	tl.add("A-1", []part{valuePart("GB01", 0)}, money.Amount{})
	tl.add("A-2", []part{notFilledInPart()}, money.Amount{})
	tl.add("A-3", []part{doesNotApplyPart()}, money.Amount{})

	w := tl.widget(spec{key: "attr1Count", title: "Activities by X", measure: MeasureCount, filter: FilterAttr},
		usdCurrency())

	linkable := map[string]bool{}
	for _, v := range w.Values {
		linkable[v.Key] = v.Linkable
	}
	want := map[string]bool{"GB01": true, KeyNotFilledIn: true, KeyDoesNotApply: false}
	if diff := cmp.Diff(want, linkable); diff != "" || w.FilterParam == nil || *w.FilterParam != FilterAttr ||
		w.Total.Linkable {
		t.Errorf("linkable (-want +got):\n%s filterParam %v", diff, w.FilterParam)
	}
}
