package plandetails

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
)

// amounts returns each part's key with its amount and whether it counts the record, for comparison.
func amounts(parts []part) map[string]string {
	got := map[string]string{}
	for _, p := range parts {
		counted := ""
		if !p.counted {
			counted = " (not counted)"
		}
		got[p.key] = p.amount.String() + counted
	}
	return got
}

func TestAttribution_Split_FR011(t *testing.T) {
	t.Parallel()
	country := attribute("SA-2", "Country / OU", allwork.TypeMultipleChoiceSplit, "DE", "FR")
	var n allwork.Node
	split(t, "SA-2", map[string]string{"DE": "60.00", "FR": "40.00"})(&n)

	parts, each := partsOf(country, &n, nil, *amount(t, "10000.00"))

	if diff := cmp.Diff(map[string]string{"DE": "6000.00", "FR": "4000.00"}, amounts(parts)); diff != "" || each {
		t.Errorf("parts (-want +got):\n%s countsInEach = %v", diff, each)
	}

	thirds := attribute("SA-3", "Region", allwork.TypeMultipleChoiceSplit, "A", "B", "C")
	split(t, "SA-3", map[string]string{"A": "33.33", "B": "33.33", "C": "33.34"})(&n)
	parts, _ = partsOf(thirds, &n, nil, *amount(t, "100.01"))

	if diff := cmp.Diff(map[string]string{"A": "33.33", "B": "33.33", "C": "33.35"}, amounts(parts)); diff != "" {
		t.Errorf("half up with the rest on the last value (-want +got):\n%s", diff)
	}
}

func TestAttribution_CountsInEach_FR011(t *testing.T) {
	t.Parallel()
	channel := attribute("SA-4", "Channel / audience", allwork.TypeMultipleChoice, "EXT", "INT")
	var n allwork.Node
	choice("SA-4", "EXT", "INT")(&n)

	parts, each := partsOf(channel, &n, nil, *amount(t, "500.00"))

	if diff := cmp.Diff(map[string]string{"EXT": "500.00", "INT": "500.00"}, amounts(parts)); diff != "" || !each {
		t.Errorf("parts (-want +got):\n%s countsInEach = %v, want true", diff, each)
	}
}

func TestAttribution_ZeroShare_FR011(t *testing.T) {
	t.Parallel()
	country := attribute("SA-2", "Country / OU", allwork.TypeMultipleChoiceSplit, "DE", "FR")
	var n allwork.Node
	split(t, "SA-2", map[string]string{"DE": "100.00", "FR": "0.00"})(&n)

	parts, _ := partsOf(country, &n, nil, *amount(t, "80.00"))

	want := map[string]string{"DE": "80.00", "FR": "0.00 (not counted)"}
	if diff := cmp.Diff(want, amounts(parts)); diff != "" {
		t.Errorf("a 0 %% share neither counts the record nor gets an amount (-want +got):\n%s", diff)
	}
}

func TestAttribution_DoesNotApply_FR012(t *testing.T) {
	t.Parallel()
	launch := attribute("SA-5", "Product launch", allwork.TypeSingleChoice, "NEW")
	launch.Condition = []string{"BU-LUB"}
	var empty, filled allwork.Node
	choice("SA-5", "NEW")(&filled)

	tests := []struct {
		name  string
		n     *allwork.Node
		units []string
		want  string
	}{
		{"hidden by the Business unit condition", &filled, []string{"BU-MOB"}, KeyDoesNotApply},
		{"shown and empty", &empty, []string{"BU-LUB"}, KeyNotFilledIn},
		{"shown and filled", &filled, []string{"BU-LUB"}, "NEW"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parts, _ := partsOf(launch, tt.n, tt.units, *amount(t, "10.00"))
			if len(parts) != 1 || parts[0].key != tt.want || parts[0].amount.String() != "10.00" {
				t.Errorf("parts = %+v, want one %s with 10.00", parts, tt.want)
			}
		})
	}
}

func TestAttribution_YesNo_FR009(t *testing.T) {
	t.Parallel()
	agency := attribute("SA-6", "Agency involved", allwork.TypeYesNo)
	var n allwork.Node
	yesNo("SA-6", false)(&n)

	parts, _ := partsOf(agency, &n, nil, *amount(t, "1.00"))

	if len(parts) != 1 || parts[0].key != KeyNo || parts[0].label != "No" {
		t.Errorf("parts = %+v, want one No", parts)
	}
}
