package widgetdata

import (
	"errors"
	"math"
	"testing"
)

func TestAggregate_EveryKind_FR018(t *testing.T) {
	t.Parallel()
	cost := Field{Key: "plannedCost", Kind: KindMoney}
	days := Field{Key: "days", Kind: KindNumber}
	start := Field{Key: "start", Kind: KindDate}
	team := Field{Key: "team", Kind: KindText, Multi: true}
	rows := []Row{
		{moneyValue(t, "10.00"), numberValue(t, "1"), dateValue(t, "2026-03-01"), TextsValue([]string{"A", "B"})},
		{moneyValue(t, "0.01"), numberValue(t, "2"), dateValue(t, "2026-01-15"), TextsValue([]string{"A"})},
		{noValue(), noValue(), noValue(), TextsValue(nil)},
	}
	all := []int{0, 1, 2}
	tests := []struct {
		name string
		m    measure
		idx  []int
		want string
	}{
		{"count of records counts every row", measure{agg: aggCount, index: -1}, all, "3"},
		{"count of a field skips empty values", measure{agg: aggCount, index: 0, field: cost}, all, "2"},
		{"count distinct counts each value of a multi field once", measure{agg: aggCountDistinct, index: 3, field: team}, all, "2"},
		{"sum is exact and ignores empty values", measure{agg: aggSum, index: 0, field: cost}, all, "10.01"},
		{"sum of nothing is zero", measure{agg: aggSum, index: 0, field: cost}, []int{2}, "0.00"},
		{"average of money rounds half up to the cent", measure{agg: aggAvg, index: 0, field: cost}, all, "5.01"},
		{"average of numbers keeps four decimals", measure{agg: aggAvg, index: 1, field: days}, all, "1.5"},
		{"average of nothing is empty", measure{agg: aggAvg, index: 1, field: days}, []int{2}, "—"},
		{"minimum of dates", measure{agg: aggMin, index: 2, field: start}, all, "2026-01-15"},
		{"maximum of money", measure{agg: aggMax, index: 0, field: cost}, all, "10.00"},
		{"maximum of nothing is empty", measure{agg: aggMax, index: 2, field: start}, []int{2}, "—"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := aggregate(tt.m, rows, tt.idx)
			if err != nil {
				t.Fatalf("aggregate: %v", err)
			}
			if text := valueText(got, tt.m.kind()); text != tt.want {
				t.Errorf("%s = %s, want %s", tt.m.agg, text, tt.want)
			}
		})
	}
}

func TestAggregate_AverageHalfUp_FR018(t *testing.T) {
	t.Parallel()
	days := Field{Key: "days", Kind: KindNumber}
	rows := []Row{{numberValue(t, "0.0001")}, {numberValue(t, "0")}, {numberValue(t, "-0.0001")}, {numberValue(t, "-0.0002")}}
	tests := []struct {
		name string
		idx  []int
		want string
	}{
		{"half a step rounds up", []int{0, 1}, "0.0001"},
		{"half a negative step rounds away from zero", []int{2, 1}, "-0.0001"},
		{"below half rounds down", []int{0, 1, 1}, "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _ := aggregate(measure{agg: aggAvg, index: 0, field: days}, rows, tt.idx)
			if got.Number.String() != tt.want {
				t.Errorf("average = %s, want %s", got.Number, tt.want)
			}
		})
	}
}

func TestAggregate_OverflowIsRefused_FR018(t *testing.T) {
	t.Parallel()
	big := NumberOfUnits(math.MaxInt64 - 1)
	rows := []Row{{NumberValue(&big)}, {numberValue(t, "1")}}
	_, err := aggregate(measure{agg: aggSum, index: 0, field: Field{Kind: KindNumber}}, rows, []int{0, 1})
	if !errors.Is(err, errTooLarge) {
		t.Errorf("sum beyond the range error = %v, want errTooLarge", err)
	}
	cents := Field{Kind: KindMoney}
	most := moneyValue(t, "999999999999999.99")
	_, err = aggregate(measure{agg: aggAvg, index: 0, field: cents}, []Row{{most}, {most}}, []int{0, 1})
	if !errors.Is(err, errTooLarge) {
		t.Errorf("money beyond fifteen digits error = %v, want errTooLarge", err)
	}
}

func valueText(v Value, kind FieldKind) string {
	if v.Empty {
		return "—"
	}
	switch kind {
	case KindMoney:
		return v.Money.Amount.String()
	case KindDate:
		return v.Date.String()
	case KindNumber, KindText:
	}
	return v.Number.String()
}
