package approvalreport_test

import (
	"testing"

	"mapp/internal/allwork"
	"mapp/internal/measure/approvalreport"
	"mapp/internal/measure/approvalreport/approvalreporttest"
	"mapp/internal/platform/money"
)

func rowOf(t *testing.T, f approvalreporttest.Fixture, itemID string) approvalreport.Row {
	t.Helper()
	for _, r := range allRows(t, f, approvalreport.Query{}) {
		if r.ItemID == itemID {
			return r
		}
	}
	t.Fatalf("no row for %s", itemID)
	return approvalreport.Row{}
}

func amountText(m *money.Money) string {
	if m == nil {
		return "—"
	}
	return m.Amount.String() + " " + m.Currency
}

func TestAmounts_041_FR018(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)

	for itemID, want := range map[string]string{
		approvalreporttest.ActivityWaiting:      "1200.00 USD",
		approvalreporttest.ActivitySkippedLevel: "—",
		approvalreporttest.RequestWaiting:       "2500.00 USD",
	} {
		if got := amountText(rowOf(t, f, itemID).Amount); got != want {
			t.Errorf("amount of %s = %s, want %s: planned cost now, a dash for a plan without finance.view", itemID, got, want)
		}
	}
}

func TestAmounts_DisplayCurrency_041_FR018(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	rate, err := money.ParseRate("0.92")
	if err != nil {
		t.Fatalf("ParseRate() error: %v", err)
	}
	f.Scope.Currency = allwork.Currency{Code: "EUR", Rate: rate}
	f.Deps.Scope = approvalreporttest.Scoper{Result: f.Scope}

	if got := amountText(rowOf(t, f, approvalreporttest.ActivityWaiting).Amount); got != "1104.00 EUR" {
		t.Errorf("activity amount = %s, want 1104.00 EUR", got)
	}
	if got := amountText(rowOf(t, f, approvalreporttest.RequestWaiting).Amount); got != "2300.00 EUR" {
		t.Errorf("request amount = %s, want 2300.00 EUR", got)
	}
}

func TestAmounts_NoFinanceView_041_FR018(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)
	f.Scope.Choices = allwork.Choices{SeesFinance: false, People: f.Scope.Choices.People}
	f.Deps.Scope = approvalreporttest.Scoper{Result: f.Scope}

	for _, r := range allRows(t, f, approvalreport.Query{}) {
		if r.Amount != nil {
			t.Errorf("row %s shows %s, want a dash without finance.view", keyOf(r), amountText(r.Amount))
		}
	}
}

func TestCanOpen_041_FR017(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)

	for _, itemID := range []string{approvalreporttest.ActivityWaiting, approvalreporttest.RequestWaiting} {
		if !rowOf(t, f, itemID).CanOpen {
			t.Errorf("%s is in view through the person's own scope, so its page opens for them", itemID)
		}
	}
}
