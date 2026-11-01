package widgetdata

import (
	"cmp"
	"slices"
	"strings"
)

const (
	opEq       = "eq"
	opNe       = "ne"
	opIn       = "in"
	opContains = "contains"
	opGt       = "gt"
	opGte      = "gte"
	opLt       = "lt"
	opLte      = "lte"
	opBetween  = "between"
	opIsEmpty  = "isEmpty"
	opNotEmpty = "notEmpty"
)

// holds says whether the row passes; an empty value is "not" anything, so only ne, isEmpty and notEmpty take it.
func (c condition) holds(r Row) bool {
	v := r[c.index]
	empty := isEmpty(v, c.field)
	switch c.op {
	case opIsEmpty:
		return empty
	case opNotEmpty:
		return !empty
	case opNe:
		return empty || !c.matchesAny(v)
	}
	return !empty && c.compares(v)
}

func (c condition) compares(v Value) bool {
	switch c.op {
	case opEq, opIn:
		return c.matchesAny(v)
	case opContains:
		return c.contains(v)
	case opGt:
		return compareValues(v, c.values[0], c.field.Kind) > 0
	case opGte:
		return compareValues(v, c.values[0], c.field.Kind) >= 0
	case opLt:
		return compareValues(v, c.values[0], c.field.Kind) < 0
	case opLte:
		return compareValues(v, c.values[0], c.field.Kind) <= 0
	case opBetween:
		return compareValues(v, c.values[0], c.field.Kind) >= 0 && compareValues(v, c.values[1], c.field.Kind) <= 0
	}
	return false
}

func (c condition) matchesAny(v Value) bool {
	return slices.ContainsFunc(c.values, func(want Value) bool {
		if c.field.Kind == KindText {
			return slices.Contains(textsOf(v, c.field), want.Text)
		}
		return compareValues(v, want, c.field.Kind) == 0
	})
}

func (c condition) contains(v Value) bool {
	part := strings.ToLower(c.values[0].Text)
	return slices.ContainsFunc(textsOf(v, c.field), func(t string) bool {
		return strings.Contains(strings.ToLower(t), part)
	})
}

func compareValues(a, b Value, kind FieldKind) int {
	switch kind {
	case KindNumber:
		return cmp.Compare(a.Number.Units(), b.Number.Units())
	case KindMoney:
		return a.Money.Amount.Cmp(b.Money.Amount)
	case KindDate:
		return a.Date.Compare(b.Date)
	case KindText:
	}
	return strings.Compare(a.Text, b.Text)
}

func (p *compiled) keeps(rows []Row) []int {
	idx := make([]int, 0, len(rows))
	for i, r := range rows {
		if p.passes(r) {
			idx = append(idx, i)
		}
	}
	return idx
}

func (p *compiled) passes(r Row) bool {
	for _, c := range p.conditions {
		if !c.holds(r) {
			return false
		}
	}
	return true
}
