package widgetdata

import (
	"math/big"
	"slices"
)

// outcome is the engine's answer before the service adds the source, the notes and the comparison.
type outcome struct {
	rows       []DataRow
	totals     []Cell
	values     []Value
	groupCount int
	multi      bool
	limited    bool
}

// run takes rows already limited to the scope and the dashboard filters: widget filters, groups, measures,
// shares, sort, limit and "Other", in that order.
func (p *compiled) run(rows []Row) (outcome, error) {
	idx := p.keeps(rows)
	dims, multi := groupRows(rows, idx, p.dimension)
	rs, err := p.rank(rows, dims)
	if err != nil {
		return outcome{}, err
	}
	p.sortGroups(rs)
	denominators := p.denominators(rs)
	shown, err := p.limited(rows, rs)
	if err != nil {
		return outcome{}, err
	}
	out := outcome{groupCount: len(rs), multi: multi, limited: len(shown) < len(rs)}
	if out.values, err = p.valuesOf(rows, idx); err != nil {
		return outcome{}, err
	}
	out.totals = p.cells(out.values, make([]*big.Int, len(out.values)))
	if p.series == nil {
		out.rows = p.dimensionRows(shown, denominators)
		return out, nil
	}
	return p.withSeries(rows, idx, shown, out)
}

func (p *compiled) withSeries(rows []Row, idx []int, shown []ranked, out outcome) (outcome, error) {
	keep, err := p.topSeries(rows, idx)
	if err != nil {
		return outcome{}, err
	}
	for _, dim := range shown {
		series, err := p.rank(rows, p.seriesOf(rows, dim.rows, keep))
		if err != nil {
			return outcome{}, err
		}
		denominators := p.denominators(series)
		for _, s := range series {
			out.rows = append(out.rows, DataRow{Dimension: p.dimensionKey(dim), Series: s.key.answer(),
				Values: p.cells(s.values, denominators)})
		}
		out.multi = out.multi || p.series.field.Multi && hasSeveral(rows, dim.rows, p.series)
	}
	return out, nil
}

func (p *compiled) dimensionRows(shown []ranked, denominators []*big.Int) []DataRow {
	out := make([]DataRow, 0, len(shown))
	for _, r := range shown {
		out = append(out, DataRow{Dimension: p.dimensionKey(r), Values: p.cells(r.values, denominators)})
	}
	return out
}

func (p *compiled) dimensionKey(r ranked) *Key {
	if p.dimension == nil {
		return nil
	}
	return r.key.answer()
}

// denominators are, per measure, the sum of the siblings' values that shares are of; nil when a share means
// nothing: an aggregation other than a count or sum, or a negative or empty value among them.
func (p *compiled) denominators(siblings []ranked) []*big.Int {
	out := make([]*big.Int, len(p.measures))
	for m, measure := range p.measures {
		if !slices.Contains([]string{aggCount, aggCountDistinct, aggSum}, measure.agg) {
			continue
		}
		sum := new(big.Int)
		for _, s := range siblings {
			v := s.values[m]
			units := unitsIn(v, measure.kind())
			if v.Empty || units < 0 {
				sum = nil
				break
			}
			sum.Add(sum, big.NewInt(units))
		}
		out[m] = sum
	}
	return out
}

func (p *compiled) cells(values []Value, denominators []*big.Int) []Cell {
	out := make([]Cell, len(values))
	for m, v := range values {
		kind := p.measures[m].kind()
		out[m] = cellOf(v, kind)
		if denominators[m] != nil {
			out[m].Share = percentOf(unitsIn(v, kind), denominators[m])
		}
	}
	return out
}

func cellOf(v Value, kind FieldKind) Cell {
	if v.Empty {
		return Cell{}
	}
	switch kind {
	case KindMoney:
		m := v.Money
		return Cell{Money: &m}
	case KindDate:
		d := v.Date
		return Cell{Date: &d}
	case KindNumber, KindText:
	}
	n := v.Number
	return Cell{Number: &n}
}

func hasSeveral(rows []Row, idx []int, g *grouping) bool {
	for _, i := range idx {
		if len(g.keysOf(rows[i])) > 1 {
			return true
		}
	}
	return false
}
