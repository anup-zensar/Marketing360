package widgetdata

import (
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

func noValue() Value { return Value{Empty: true} }

func TextValue(s string) Value {
	if s == "" {
		return noValue()
	}
	return Value{Text: s}
}

func TextsValue(texts []string) Value {
	kept := make([]string, 0, len(texts))
	for _, t := range texts {
		if t != "" {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		return noValue()
	}
	return Value{Texts: kept}
}

func YesNo(yes bool) Value {
	if yes {
		return Value{Text: "Yes"}
	}
	return Value{Text: "No"}
}

func MoneyValue(a *money.Amount) Value {
	if a == nil {
		return noValue()
	}
	return Value{Money: money.USD(*a)}
}

func DateValue(d *date.Date) Value {
	if d == nil || d.IsZero() {
		return noValue()
	}
	return Value{Date: *d}
}

func WholeValue(n int64) Value {
	return Value{Number: NumberOfUnits(n * numberScale)}
}

func NumberValue(n *Number) Value {
	if n == nil {
		return noValue()
	}
	return Value{Number: *n}
}

// RowOf lays values out in the order of fields; a field without a value is empty.
func RowOf(fields []Field, values map[string]Value) Row {
	row := make(Row, len(fields))
	for i, f := range fields {
		v, found := values[f.Key]
		if !found {
			v = noValue()
		}
		row[i] = v
	}
	return row
}

func isEmpty(v Value, f Field) bool {
	if v.Empty {
		return true
	}
	if f.Multi {
		return len(v.Texts) == 0
	}
	return f.Kind == KindText && v.Text == ""
}

// textsOf are a text cell's values: one, or several for a multi field.
func textsOf(v Value, f Field) []string {
	if f.Multi {
		return v.Texts
	}
	return []string{v.Text}
}
