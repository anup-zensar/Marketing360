package widgetdata

import "testing"

func TestFilters_EveryOperator_FR019(t *testing.T) {
	t.Parallel()
	status := Field{Key: "status", Kind: KindText}
	team := Field{Key: "team", Kind: KindText, Multi: true}
	cost := Field{Key: "plannedCost", Kind: KindMoney}
	days := Field{Key: "daysToDecide", Kind: KindNumber}
	start := Field{Key: "startDate", Kind: KindDate}
	tests := []struct {
		name   string
		field  Field
		op     string
		values []Value
		cell   Value
		want   bool
	}{
		{"is matches the same text", status, opEq, []Value{TextValue("Approved")}, TextValue("Approved"), true},
		{"is is exact", status, opEq, []Value{TextValue("Approved")}, TextValue("approved"), false},
		{"is never matches an empty value", status, opEq, []Value{TextValue("Approved")}, noValue(), false},
		{"is not holds for another text", status, opNe, []Value{TextValue("Declined")}, TextValue("Approved"), true},
		{"is not holds for an empty value", status, opNe, []Value{TextValue("Declined")}, noValue(), true},
		{"is not fails on the same text", status, opNe, []Value{TextValue("Declined")}, TextValue("Declined"), false},
		{"is one of matches any", status, opIn, []Value{TextValue("Created"), TextValue("Approved")}, TextValue("Approved"), true},
		{"is one of fails outside", status, opIn, []Value{TextValue("Created")}, TextValue("Approved"), false},
		{"contains ignores case", status, opContains, []Value{TextValue("PROV")}, TextValue("Approved"), true},
		{"contains fails without the text", status, opContains, []Value{TextValue("lin")}, TextValue("Approved"), false},
		{"is empty holds for an empty text", status, opIsEmpty, nil, TextValue(""), true},
		{"is not empty fails for an empty value", status, opNotEmpty, nil, noValue(), false},
		{"is not empty holds for a text", status, opNotEmpty, nil, TextValue("Draft"), true},
		{"a multi field is one of its values", team, opEq, []Value{TextValue("Asia")}, TextsValue([]string{"Europe", "Asia"}), true},
		{"a multi field is not a value it holds", team, opNe, []Value{TextValue("Asia")}, TextsValue([]string{"Europe", "Asia"}), false},
		{"a multi field without teams is empty", team, opIsEmpty, nil, TextsValue([]string{}), true},
		{"a multi field contains in any value", team, opContains, []Value{TextValue("eur")}, TextsValue([]string{"Asia", "Europe"}), true},
		{"greater than on money", cost, opGt, []Value{moneyValue(t, "100")}, moneyValue(t, "100.01"), true},
		{"greater than is strict", cost, opGt, []Value{moneyValue(t, "100")}, moneyValue(t, "100"), false},
		{"at least includes the value", cost, opGte, []Value{moneyValue(t, "100")}, moneyValue(t, "100"), true},
		{"less than on numbers", days, opLt, []Value{numberValue(t, "2.5")}, numberValue(t, "2.4999"), true},
		{"at most includes the value", days, opLte, []Value{numberValue(t, "2.5")}, numberValue(t, "2.5"), true},
		{"at most fails above", days, opLte, []Value{numberValue(t, "2.5")}, numberValue(t, "2.5001"), false},
		{"a number equals the same decimal", days, opEq, []Value{numberValue(t, "3")}, numberValue(t, "3.0"), true},
		{"between includes both ends", start, opBetween, []Value{dateValue(t, "2026-07-01"), dateValue(t, "2026-07-31")},
			dateValue(t, "2026-07-31"), true},
		{"between fails after the end", start, opBetween, []Value{dateValue(t, "2026-07-01"), dateValue(t, "2026-07-31")},
			dateValue(t, "2026-08-01"), false},
		{"a comparison never holds for an empty date", start, opLt, []Value{dateValue(t, "2026-07-01")}, noValue(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := condition{index: 0, field: tt.field, op: tt.op, values: tt.values}
			if got := c.holds(Row{tt.cell}); got != tt.want {
				t.Errorf("%s %s %+v on %+v = %v, want %v", tt.field.Key, tt.op, tt.values, tt.cell, got, tt.want)
			}
		})
	}
}
