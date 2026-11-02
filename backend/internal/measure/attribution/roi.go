package attribution

import (
	"math/big"

	"mapp/internal/platform/money"
)

type Basis string

const (
	BasisActual  Basis = "actual"
	BasisPlanned Basis = "planned"
)

// ROI is (revenue − cost) ÷ cost as a whole per cent, halves away from zero; nil without a cost above zero, which
// shows as "—" (FR-007).
func ROI(revenue money.Amount, cost *money.Amount) *int {
	if cost == nil || cost.Sign() <= 0 {
		return nil
	}
	gain := new(big.Rat).SetFrac(big.NewInt((revenue.Cents()-cost.Cents())*100), big.NewInt(cost.Cents()))
	half := big.NewRat(1, 2)
	if gain.Sign() < 0 {
		half.Neg(half)
	}
	rounded := new(big.Rat).Add(gain, half)
	whole := int(new(big.Int).Quo(rounded.Num(), rounded.Denom()).Int64())
	return &whole
}
