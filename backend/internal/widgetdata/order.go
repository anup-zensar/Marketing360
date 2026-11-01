package widgetdata

import (
	"cmp"
	"slices"
	"strings"
)

const (
	sortByDimension = "dimension"
	sortByMeasure   = "measure"
	directionAsc    = "asc"
	directionDesc   = "desc"
	defaultLimit    = 50
	maxSeries       = 20
	maxRadarSpokes  = 12
	maxRadarShapes  = 8
)

// ranked is a group with its measures' values, sorted and limited as the query says.
type ranked struct {
	group
	values []Value
}

func (p *compiled) rank(rows []Row, groups []group) ([]ranked, error) {
	out := make([]ranked, 0, len(groups))
	for _, g := range groups {
		values, err := p.valuesOf(rows, g.rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ranked{group: g, values: values})
	}
	return out, nil
}

func (p *compiled) valuesOf(rows []Row, idx []int) ([]Value, error) {
	values := make([]Value, len(p.measures))
	for i, m := range p.measures {
		v, err := aggregate(m, rows, idx)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}
	return values, nil
}

// sortGroups orders by the sort, "Not set" and empty values last whatever the direction, ties by label.
func (p *compiled) sortGroups(rs []ranked) {
	slices.SortStableFunc(rs, func(a, b ranked) int {
		return cmp.Or(p.compareRanked(a, b), compareLabels(a.key, b.key))
	})
}

func (p *compiled) compareRanked(a, b ranked) int {
	sign := 1
	if p.sort.Direction == directionDesc {
		sign = -1
	}
	if p.sort.By == sortByDimension && p.dimension != nil {
		if a.key.kind != b.key.kind {
			return cmp.Compare(keyRank(a.key.kind), keyRank(b.key.kind))
		}
		return sign * compareKeys(*a.key, *b.key, p.dimension.field.Kind)
	}
	m := p.sort.MeasureIndex
	va, vb := a.values[m], b.values[m]
	if va.Empty || vb.Empty {
		return cmp.Compare(boolRank(va.Empty), boolRank(vb.Empty))
	}
	return sign * compareValues(va, vb, p.measures[m].kind())
}

// limited keeps the first limit groups and, with groupRest, makes the others one "Other" group.
func (p *compiled) limited(rows []Row, rs []ranked) ([]ranked, error) {
	limit := p.limit
	if p.typ == typeRadar {
		limit = min(limit, maxRadarSpokes)
	}
	if len(rs) <= limit {
		return rs, nil
	}
	kept := slices.Clone(rs[:limit])
	if !p.groupRest {
		return kept, nil
	}
	rest := make([]group, 0, len(rs)-limit)
	for _, r := range rs[limit:] {
		rest = append(rest, r.group)
	}
	other, err := p.rank(rows, []group{merged(rest, otherKey)})
	return append(kept, other...), err
}

// topSeries are the series kept: the largest by the first measure; the others join "Other".
func (p *compiled) topSeries(rows []Row, idx []int) (map[groupKey]bool, error) {
	all, _ := groupRows(rows, idx, p.series)
	rs, err := p.rank(rows, all)
	if err != nil {
		return nil, err
	}
	byFirst := &compiled{measures: p.measures, sort: Sort{By: sortByMeasure, Direction: directionDesc}}
	byFirst.sortGroups(rs)
	keep := map[groupKey]bool{}
	for i, r := range rs {
		if i < p.seriesCap {
			keep[*r.key] = true
		}
	}
	return keep, nil
}

// seriesOf groups one dimension group's rows by series, folding series not kept into "Other", by key.
func (p *compiled) seriesOf(rows []Row, idx []int, keep map[groupKey]bool) []group {
	groups, _ := groupRows(rows, idx, p.series)
	var out, rest []group
	for _, g := range groups {
		if keep[*g.key] {
			out = append(out, g)
		} else {
			rest = append(rest, g)
		}
	}
	if len(rest) > 0 {
		out = append(out, merged(rest, otherKey))
	}
	slices.SortStableFunc(out, func(a, b group) int { return compareKeys(*a.key, *b.key, p.series.field.Kind) })
	return out
}

// merged is one group of the groups' rows, each row once.
func merged(groups []group, key groupKey) group {
	seen := map[int]bool{}
	var rows []int
	for _, g := range groups {
		for _, i := range g.rows {
			if !seen[i] {
				seen[i] = true
				rows = append(rows, i)
			}
		}
	}
	slices.Sort(rows)
	return group{key: &key, rows: rows}
}

func compareLabels(a, b *groupKey) int {
	if a == nil || b == nil {
		return 0
	}
	return strings.Compare(strings.ToLower(a.label()), strings.ToLower(b.label()))
}

func boolRank(b bool) int {
	if b {
		return 1
	}
	return 0
}
