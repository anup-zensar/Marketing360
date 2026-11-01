package widgetdata

import (
	"errors"
	"math"
	"math/big"

	"mapp/internal/platform/money"
)

const maxCents int64 = 99_999_999_999_999_999

var errTooLarge = errors.New("widgetdata: values too large to add up")

// aggregate works out one measure over the rows idx; sums of nothing are 0, averages, minimums and maximums empty.
func aggregate(m measure, rows []Row, idx []int) (Value, error) {
	switch m.agg {
	case aggCount:
		return WholeValue(int64(countOf(m, rows, idx))), nil
	case aggCountDistinct:
		return WholeValue(int64(distinctOf(m, rows, idx))), nil
	case aggMin, aggMax:
		return extremeOf(m, rows, idx), nil
	}
	total, n, err := sumOf(m, rows, idx)
	if err != nil || m.agg == aggSum {
		return unitsValue(m.field.Kind, total), err
	}
	if n == 0 {
		return noValue(), nil
	}
	return unitsValue(m.field.Kind, roundHalfUp(big.NewInt(total), big.NewInt(int64(n))).Int64()), nil
}

func countOf(m measure, rows []Row, idx []int) int {
	if m.index < 0 {
		return len(idx)
	}
	n := 0
	for _, i := range idx {
		if !isEmpty(rows[i][m.index], m.field) {
			n++
		}
	}
	return n
}

func distinctOf(m measure, rows []Row, idx []int) int {
	seen := map[string]bool{}
	for _, i := range idx {
		v := rows[i][m.index]
		if isEmpty(v, m.field) {
			continue
		}
		for _, k := range distinctKeys(v, m.field) {
			seen[k] = true
		}
	}
	return len(seen)
}

func distinctKeys(v Value, f Field) []string {
	switch f.Kind {
	case KindNumber:
		return []string{v.Number.String()}
	case KindMoney:
		return []string{v.Money.Amount.String()}
	case KindDate:
		return []string{v.Date.String()}
	case KindText:
	}
	return textsOf(v, f)
}

// sumOf adds number units or cents exactly, refusing a total beyond what the answer can hold.
func sumOf(m measure, rows []Row, idx []int) (total int64, n int, err error) {
	limit := int64(math.MaxInt64)
	if m.field.Kind == KindMoney {
		limit = maxCents
	}
	for _, i := range idx {
		v := rows[i][m.index]
		if isEmpty(v, m.field) {
			continue
		}
		units := unitsIn(v, m.field.Kind)
		if (units > 0 && total > limit-units) || (units < 0 && total < -limit-units) {
			return 0, 0, errTooLarge
		}
		total, n = total+units, n+1
	}
	return total, n, nil
}

func extremeOf(m measure, rows []Row, idx []int) Value {
	best := noValue()
	for _, i := range idx {
		v := rows[i][m.index]
		if isEmpty(v, m.field) {
			continue
		}
		c := compareValues(v, best, m.field.Kind)
		if best.Empty || (m.agg == aggMin && c < 0) || (m.agg == aggMax && c > 0) {
			best = v
		}
	}
	return best
}

func unitsIn(v Value, kind FieldKind) int64 {
	if kind == KindMoney {
		return v.Money.Amount.Cents()
	}
	return v.Number.Units()
}

func unitsValue(kind FieldKind, units int64) Value {
	if kind == KindMoney {
		return Value{Money: money.USD(money.FromCents(units))}
	}
	return Value{Number: NumberOfUnits(units)}
}
