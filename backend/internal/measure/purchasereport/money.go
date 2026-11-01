package purchasereport

import "mapp/internal/platform/money"

// shown turns USD amounts into the display currency at the current rate.
type shown struct {
	rate money.Rate
	code string
}

func (s shown) of(usd money.Amount) *money.Money {
	return &money.Money{Amount: money.FromUSD(usd, s.rate), Currency: s.code}
}

// tally adds amounts and remembers whether any came from a plan whose financial data the person may not see.
type tally struct {
	usd          money.Amount
	seen, hidden bool
}

func (t *tally) add(usd money.Amount, visible bool) {
	if !visible {
		t.hidden = true
		return
	}
	t.usd, t.seen = t.usd.Add(usd), true
}

func (t tally) money(s shown) *money.Money {
	if t.hidden && !t.seen {
		return nil
	}
	return s.of(t.usd)
}

func (t *tally) merge(o tally) {
	t.usd = t.usd.Add(o.usd)
	t.seen, t.hidden = t.seen || o.seen, t.hidden || o.hidden
}
