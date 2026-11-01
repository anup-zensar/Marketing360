package allwork

import (
	"fmt"
	"math/big"

	"mapp/internal/platform/money"
)

// amountOf is a record's planned cost before conversion: USD, or the plan currency; nil when it has none.
func (w *breakdownWork) amountOf(n *Node) *money.Amount {
	if w.planCurrency != "" {
		if n.PlannedCostPlan == nil {
			return nil
		}
		return &n.PlannedCostPlan.Amount
	}
	return n.PlannedCost
}

// pieces split activity a's planned cost (data-model §3.5): the activity, or its line items and "Not on a line item".
func (w *breakdownWork) pieces(a *tnode, byLineItems bool) []piece {
	planned := w.amountOf(a.n)
	if !byLineItems {
		var amount money.Amount
		if planned != nil {
			amount = *planned
		}
		return []piece{{activity: a, amount: amount}}
	}
	var list []piece
	var assigned money.Amount
	for _, li := range a.children {
		var amount money.Amount
		if cost := w.amountOf(li.n); cost != nil {
			amount = *cost
		}
		assigned = assigned.Add(amount)
		list = append(list, piece{activity: a, lineItem: li, amount: amount})
	}
	switch {
	case planned == nil:
		w.withoutPC = w.withoutPC.Add(assigned)
		if len(list) == 0 {
			list = append(list, piece{activity: a, unassigned: true})
		}
	case !planned.Sub(assigned).IsZero() || len(list) == 0:
		list = append(list, piece{activity: a, unassigned: true, amount: planned.Sub(assigned)})
	}
	return list
}

type part struct {
	value
	amount money.Amount
}

// split divides amount among the values by AttributeParts' rule; the Breakdown says when each gets it all.
func (w *breakdownWork) split(amount money.Amount, values []value) []part {
	shares := make([]*money.Percent, len(values))
	for i, v := range values {
		shares[i] = v.share
	}
	amounts, each := splitAmount(amount, shares)
	w.countsInEach = w.countsInEach || each
	parts := make([]part, 0, len(values))
	for i, v := range values {
		parts = append(parts, part{value: v, amount: amounts[i]})
	}
	return parts
}

// AttributeParts divides amount over a record's attribute value by option (data-model §3.5): one gets it all; a %
// split gives each amount × share ÷ 100, half up, the last the rest; several without shares each get it all (true).
func AttributeParts(v AttrValue, amount money.Amount) ([]money.Amount, bool) {
	shares := make([]*money.Percent, len(v.OptionIDs))
	for i, id := range v.OptionIDs {
		if v.Shares != nil {
			share := v.Shares[id]
			shares[i] = &share
		}
	}
	return splitAmount(amount, shares)
}

func splitAmount(amount money.Amount, shares []*money.Percent) ([]money.Amount, bool) {
	amounts := make([]money.Amount, len(shares))
	if len(shares) <= 1 || shares[0] == nil {
		for i := range amounts {
			amounts[i] = amount
		}
		return amounts, len(shares) > 1
	}
	var given money.Amount
	for i, share := range shares {
		amounts[i] = amount.Sub(given)
		if i < len(shares)-1 {
			amounts[i] = shareOfAmount(amount, *share)
			given = given.Add(amounts[i])
		}
	}
	return amounts, false
}

func shareOfAmount(a money.Amount, p money.Percent) money.Amount {
	num := new(big.Int).Mul(big.NewInt(a.Cents()), big.NewInt(p.Hundredths()))
	return money.FromCents(roundHalfAway(num, big.NewInt(10000)))
}

// roundHalfAway divides num by den, rounding half away from zero.
func roundHalfAway(num, den *big.Int) int64 {
	negative := num.Sign()*den.Sign() < 0
	n, d := new(big.Int).Abs(num), new(big.Int).Abs(den)
	q, r := new(big.Int).QuoRem(n, d, new(big.Int))
	if new(big.Int).Mul(r, big.NewInt(2)).Cmp(d) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	if negative {
		q.Neg(q)
	}
	return q.Int64()
}

// shareText2 is part ÷ whole × 100, half up to two decimals ("37.50"); "0.00" when whole is 0.00.
func shareText2(part, whole money.Amount) string {
	if whole.IsZero() {
		return "0.00"
	}
	num := new(big.Int).Mul(big.NewInt(part.Cents()), big.NewInt(10000))
	hundredths := roundHalfAway(num, big.NewInt(whole.Cents()))
	sign := ""
	if hundredths < 0 {
		sign, hundredths = "-", -hundredths
	}
	return fmt.Sprintf("%s%d.%02d", sign, hundredths/100, hundredths%100)
}
