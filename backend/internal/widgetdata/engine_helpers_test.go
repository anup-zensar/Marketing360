package widgetdata

import (
	"testing"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

func dateValue(t *testing.T, s string) Value {
	t.Helper()
	d := date.MustParse(s)
	return DateValue(&d)
}

func moneyValue(t *testing.T, s string) Value {
	t.Helper()
	a, err := money.ParseAmount(s)
	if err != nil {
		t.Fatalf("ParseAmount(%q): %v", s, err)
	}
	return MoneyValue(&a)
}

func numberValue(t *testing.T, s string) Value {
	t.Helper()
	n, err := ParseNumber(s)
	if err != nil {
		t.Fatalf("ParseNumber(%q): %v", s, err)
	}
	return NumberValue(&n)
}
