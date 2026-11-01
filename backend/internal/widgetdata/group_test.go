package widgetdata

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func labelsAndRows(groups []group) map[string][]int {
	out := map[string][]int{}
	for _, g := range groups {
		out[g.key.label()] = g.rows
	}
	return out
}

func TestGroup_MultiValuedCountsInEach_D373(t *testing.T) {
	t.Parallel()
	team := grouping{index: 0, field: Field{Key: "team", Kind: KindText, Multi: true}}
	rows := []Row{{TextsValue([]string{"Europe", "Asia"})}, {TextsValue([]string{"Europe", "Europe"})}, {TextsValue(nil)}}

	groups, multi := groupRows(rows, []int{0, 1, 2}, &team)

	want := map[string][]int{"Europe": {0, 1}, "Asia": {0}, labelNotSet: {2}}
	if diff := cmp.Diff(want, labelsAndRows(groups)); diff != "" || !multi {
		t.Errorf("groups (-want +got):\n%s multi = %v, want true", diff, multi)
	}
}

func TestGroup_EmptyIsNotSet(t *testing.T) {
	t.Parallel()
	status := grouping{index: 0, field: Field{Key: "status", Kind: KindText}}
	rows := []Row{{TextValue("Draft")}, {TextValue("")}, {noValue()}}

	groups, multi := groupRows(rows, []int{0, 1, 2}, &status)

	want := map[string][]int{"Draft": {0}, labelNotSet: {1, 2}}
	if diff := cmp.Diff(want, labelsAndRows(groups)); diff != "" || multi {
		t.Errorf("groups (-want +got):\n%s multi = %v, want false", diff, multi)
	}
	if k := groups[1].key.answer(); k.Value != nil || k.Kind != KeyEmpty {
		t.Errorf("empty key answer = %+v, want a null value of kind empty", k)
	}
}

func TestGroup_NumbersAndDateBuckets_FR017(t *testing.T) {
	t.Parallel()
	year := grouping{index: 0, field: Field{Key: "year", Kind: KindNumber}}
	month := grouping{index: 1, field: Field{Key: "start", Kind: KindDate}, bucket: bucketMonth}
	rows := []Row{{WholeValue(2026), dateValue(t, "2026-03-09")}, {WholeValue(2027), dateValue(t, "2026-03-31")}}

	years, _ := groupRows(rows, []int{0, 1}, &year)
	months, _ := groupRows(rows, []int{0, 1}, &month)

	if diff := cmp.Diff(map[string][]int{"2026": {0}, "2027": {1}}, labelsAndRows(years)); diff != "" {
		t.Errorf("years (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(map[string][]int{"2026-03-01": {0, 1}}, labelsAndRows(months)); diff != "" {
		t.Errorf("months (-want +got):\n%s", diff)
	}
	if compareKeys(*years[0].key, *years[1].key, KindNumber) >= 0 {
		t.Error("2026 sorts before 2027")
	}
}
