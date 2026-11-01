package attribution

import (
	"cmp"
	"math/big"
	"slices"

	"mapp/internal/platform/money"
)

// Attribute shares a conversion's revenue among its journey's activities; the credits add up to it exactly.
func Attribute(c Conversion, m Model) []Credit {
	if !c.Journey {
		return []Credit{{ActivityID: c.ActivityID, Amount: c.Revenue}}
	}
	touches := slices.DeleteFunc(slices.Clone(c.Touches), func(t Touch) bool { return t.Date.After(c.Date) })
	if len(touches) == 0 {
		return []Credit{{Amount: c.Revenue}}
	}
	slices.SortStableFunc(touches, func(x, y Touch) int { return x.Date.Compare(y.Date) })
	shares := allocate(c.Revenue.Cents(), weights(m, len(touches)))
	var credits []Credit
	for i, t := range touches {
		at := slices.IndexFunc(credits, func(c Credit) bool { return c.ActivityID == t.ActivityID })
		if at < 0 {
			credits, at = append(credits, Credit{ActivityID: t.ActivityID}), len(credits)
		}
		credits[at].Amount = credits[at].Amount.Add(money.FromCents(shares[i]))
	}
	return slices.DeleteFunc(credits, func(c Credit) bool { return c.Amount.IsZero() && c.ActivityID != "" })
}

// weights are the touches' relative weights, earliest first (FR-003).
func weights(m Model, n int) []int64 {
	w := make([]int64, n)
	switch {
	case m == FirstTouch:
		w[0] = 1
	case m == LastTouch:
		w[n-1] = 1
	case m == LinearDecay:
		for i := range w {
			w[i] = int64(i + 1)
		}
	case m == UShaped && n > 2:
		for i := range w {
			w[i] = 1
		}
		w[0], w[n-1] = int64(2*(n-2)), int64(2*(n-2))
	default:
		for i := range w {
			w[i] = 1
		}
	}
	return w
}

// allocate splits total cents by the weights, rounding down; the residual goes to the largest weight, the newest
// touch on a tie, so nothing is lost or made up (FR-005).
func allocate(total int64, w []int64) []int64 {
	var sum int64
	for _, x := range w {
		sum += x
	}
	shares := make([]int64, len(w))
	var given int64
	for i, x := range w {
		share := new(big.Int).Mul(big.NewInt(total), big.NewInt(x))
		shares[i] = share.Quo(share, big.NewInt(sum)).Int64()
		given += shares[i]
	}
	largest := 0
	for i, x := range w {
		if cmp.Compare(x, w[largest]) >= 0 {
			largest = i
		}
	}
	shares[largest] += total - given
	return shares
}
