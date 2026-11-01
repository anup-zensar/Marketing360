package widgetdata

import (
	"math/big"
	"strconv"
)

const maxShareTenths = 99999

// SharePercent is a percent with one decimal: "37.5" is 375 tenths.
type SharePercent struct{ tenths int64 }

func (p SharePercent) String() string {
	sign, tenths := "", p.tenths
	if tenths < 0 {
		sign, tenths = "-", -tenths
	}
	return sign + strconv.FormatInt(tenths/10, 10) + "." + strconv.FormatInt(tenths%10, 10)
}

func (p SharePercent) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(p.String())), nil
}

// percentOf is part ÷ whole in percent, half up to one decimal; nil when whole is 0 or the result does not fit.
func percentOf(part int64, whole *big.Int) *SharePercent {
	if whole.Sign() == 0 {
		return nil
	}
	num := new(big.Int).Mul(big.NewInt(part), big.NewInt(1000))
	tenths := roundHalfUp(num, whole)
	if !tenths.IsInt64() || tenths.Int64() > maxShareTenths || tenths.Int64() < -maxShareTenths {
		return nil
	}
	return &SharePercent{tenths: tenths.Int64()}
}

// roundHalfUp divides, rounding halves away from zero as the money package does.
func roundHalfUp(num, den *big.Int) *big.Int {
	quotient, remainder := new(big.Int).QuoRem(num, den, new(big.Int))
	twice := new(big.Int).Mul(new(big.Int).Abs(remainder), big.NewInt(2))
	if twice.Cmp(new(big.Int).Abs(den)) < 0 {
		return quotient
	}
	if num.Sign()*den.Sign() < 0 {
		return quotient.Sub(quotient, big.NewInt(1))
	}
	return quotient.Add(quotient, big.NewInt(1))
}
