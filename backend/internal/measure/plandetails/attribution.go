package plandetails

import (
	"math/big"
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/platform/money"
)

const (
	labelNotFilledIn  = "Not filled in"
	labelDoesNotApply = "Does not apply"
	labelYes          = "Yes"
	labelNo           = "No"
	labelInactive     = " (inactive)"
)

const (
	orderSpecial = 1 << 28
	orderOutside = orderSpecial + iota
	orderNoDates
	orderNotFilledIn
	orderDoesNotApply
)

// part is one value a record has in a widget, with the share of the amount it brings (016 FR-041).
type part struct {
	key, label string
	order      int
	inactive   bool
	counted    bool // false for a % split option with a 0 % share: it gets its amount but not the record
	linkable   bool
	amount     money.Amount
}

func notFilledInPart() part {
	return part{key: KeyNotFilledIn, label: labelNotFilledIn, order: orderNotFilledIn, counted: true, linkable: true}
}

func doesNotApplyPart() part {
	return part{key: KeyDoesNotApply, label: labelDoesNotApply, order: orderDoesNotApply, counted: true}
}

func yesNoPart(yes bool) part {
	if yes {
		return part{key: KeyYes, label: labelYes, order: 0, counted: true, linkable: true}
	}
	return part{key: KeyNo, label: labelNo, order: 1, counted: true, linkable: true}
}

// partsOf returns the values of a on the record n with amount divided as All work's Breakdown divides it, and true
// when several values without shares each get it all; hidden by a's Business unit condition, it does not apply (043-C2).
func partsOf(a allwork.Attribute, n *allwork.Node, units []string, amount money.Amount) ([]part, bool) {
	var parts []part
	switch v, found := n.Attrs[a.ID]; {
	case !shownFor(a, units):
		parts = []part{doesNotApplyPart()}
	case !found || (len(v.OptionIDs) == 0 && v.Yes == nil):
		parts = []part{notFilledInPart()}
	case v.Yes != nil:
		parts = []part{yesNoPart(*v.Yes)}
	default:
		return divide(optionParts(a, v), v.Shares, amount)
	}
	parts[0].amount = amount
	return parts, false
}

func shownFor(a allwork.Attribute, units []string) bool {
	return len(a.Condition) == 0 || slices.ContainsFunc(a.Condition, func(u string) bool { return slices.Contains(units, u) })
}

func optionParts(a allwork.Attribute, v allwork.AttrValue) []part {
	parts := make([]part, 0, len(v.OptionIDs))
	for _, id := range v.OptionIDs {
		p := part{key: id, label: id, order: len(a.Options), counted: true, linkable: true}
		if i := slices.IndexFunc(a.Options, func(o allwork.Option) bool { return o.ID == id }); i >= 0 {
			p.label, p.order, p.inactive = a.Options[i].Label, i, a.Options[i].Inactive
		}
		if p.inactive {
			p.label += labelInactive
		}
		if v.Shares != nil {
			p.counted = v.Shares[id].Hundredths() > 0
		}
		parts = append(parts, p)
	}
	return parts
}

// divide gives a % split's values amount × share ÷ 100, half up, and the last the rest; any other value gets it all.
func divide(parts []part, shares map[string]money.Percent, amount money.Amount) ([]part, bool) {
	if len(parts) == 1 || shares == nil {
		for i := range parts {
			parts[i].amount = amount
		}
		return parts, len(parts) > 1
	}
	var given money.Amount
	for i := range parts {
		parts[i].amount = amount.Sub(given)
		if i < len(parts)-1 {
			parts[i].amount = shareOfAmount(amount, shares[parts[i].key])
			given = given.Add(parts[i].amount)
		}
	}
	return parts, false
}

// shareOfAmount is a × p ÷ 100, rounded half away from zero to the cent, as All work's Breakdown rounds it.
func shareOfAmount(a money.Amount, p money.Percent) money.Amount {
	num := new(big.Int).Mul(big.NewInt(a.Cents()), big.NewInt(p.Hundredths()))
	return money.FromCents(roundHalfAway(num, big.NewInt(10000)))
}

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
