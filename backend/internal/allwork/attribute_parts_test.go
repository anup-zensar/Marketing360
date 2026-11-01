package allwork

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/money"
)

func TestAttributeParts_016_FR041(t *testing.T) {
	t.Parallel()
	half, err := money.ParsePercent("50")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, amount string
		v            AttrValue
		want         []string
		wantEach     bool
	}{
		{name: "one option gets it all", amount: "100.00", v: AttrValue{OptionIDs: []string{"a"}}, want: []string{"100.00"}},
		{name: "a % split, half up, the rest on the last", amount: "0.05", v: AttrValue{OptionIDs: []string{"a", "b"},
			Shares: map[string]money.Percent{"a": half, "b": half}}, want: []string{"0.03", "0.02"}},
		{name: "several without shares each get it all", amount: "100.00", v: AttrValue{OptionIDs: []string{"a", "b"}},
			want: []string{"100.00", "100.00"}, wantEach: true},
		{name: "no option", amount: "100.00", want: []string{}},
	}
	for _, tt := range tests {
		amount, err := money.ParseAmount(tt.amount)
		if err != nil {
			t.Fatal(err)
		}
		parts, each := AttributeParts(tt.v, amount)
		got := []string{}
		for _, p := range parts {
			got = append(got, p.String())
		}
		if diff := cmp.Diff(tt.want, got); diff != "" || each != tt.wantEach {
			t.Errorf("%s (-want +got):\n%s each = %v", tt.name, diff, each)
		}
	}
}
