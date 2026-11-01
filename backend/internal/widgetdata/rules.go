package widgetdata

import (
	"fmt"
	"slices"

	"mapp/internal/platform/errs"
)

const (
	typeBar     = "bar"
	typeLine    = "line"
	typeArea    = "area"
	typeCombo   = "combo"
	typePie     = "pie"
	typeDonut   = "donut"
	typeScatter = "scatter"
	typeFunnel  = "funnel"
	typeGauge   = "gauge"
	typeHeatmap = "heatmap"
	typeTreemap = "treemap"
	typeRadar   = "radar"
	typeKpi     = "kpi"
	typeTable   = "table"
	typeText    = "text"

	needNone     = 0
	needOptional = 1
	needRequired = 2
	maxLimit     = 500
)

// rule is a widget type's needs (data model §2.3).
type rule struct {
	dimension, series int
	minMeasures       int
	maxMeasures       int
}

var rules = map[string]rule{ //nolint:gochecknoglobals // the widget rules of data model §2.3
	typeBar: {needRequired, needOptional, 1, 4}, typeLine: {needRequired, needOptional, 1, 4},
	typeArea: {needRequired, needOptional, 1, 4}, typeCombo: {needRequired, needNone, 2, 4},
	typePie: {needRequired, needNone, 1, 1}, typeDonut: {needRequired, needNone, 1, 1},
	typeFunnel: {needRequired, needNone, 1, 1}, typeScatter: {needRequired, needOptional, 2, 3},
	typeGauge: {needNone, needNone, 1, 1}, typeKpi: {needNone, needNone, 1, 1},
	typeHeatmap: {needRequired, needRequired, 1, 1}, typeTreemap: {needRequired, needOptional, 1, 1},
	typeRadar: {needRequired, needOptional, 1, 1}, typeTable: {needOptional, needOptional, 1, 8},
}

// checkShape applies the widget rules, the sort and the limit; it never reads a field.
func checkShape(r Request, problems *errs.Problems) {
	q := r.Query
	if r.Type == typeText {
		problems.Add("type", msgTextNote)
		return
	}
	need, found := rules[r.Type]
	if !found {
		problems.Add("type", msgChooseType)
		return
	}
	name := typeNames[r.Type]
	checkNeed(groupNeed{need.dimension, q.Dimension, "query.dimension", msgChooseGroup, fmt.Sprintf(msgNoGroup, name)}, problems)
	checkNeed(groupNeed{need.series, q.Series, "query.series", msgChooseBreakdown, fmt.Sprintf(msgNoBreakdown, name)}, problems)
	if sameGrouping(q.Dimension, q.Series) {
		problems.Add("query.series", msgSameBreakdown)
	}
	checkMeasureCount(r, need, problems)
	checkSort(q, problems)
	if q.Limit < 1 || q.Limit > maxLimit {
		problems.Add("query.limit", msgLimit)
	}
}

type groupNeed struct {
	need                  int
	given                 *GroupBy
	path, missing, refuse string
}

func checkNeed(n groupNeed, problems *errs.Problems) {
	switch {
	case n.need == needRequired && n.given == nil:
		problems.Add(n.path, n.missing)
	case n.need == needNone && n.given != nil:
		problems.Add(n.path, n.refuse)
	}
}

func checkMeasureCount(r Request, need rule, problems *errs.Problems) {
	n, name := len(r.Query.Measures), typeNames[r.Type]
	oneWithSeries := r.Query.Series != nil && slices.Contains([]string{typeBar, typeLine, typeArea}, r.Type)
	switch {
	case oneWithSeries && n != 1:
		problems.Add("query.measures", fmt.Sprintf(msgOneWithBreakdown, lower(name)))
	case need.maxMeasures == 1 && n != 1:
		problems.Add("query.measures", fmt.Sprintf(msgOneMeasure, name))
	case n < need.minMeasures || n > need.maxMeasures:
		problems.Add("query.measures", fmt.Sprintf(msgMeasureRange, name, need.minMeasures, need.maxMeasures))
	}
}

func checkSort(q Query, problems *errs.Problems) {
	if q.Sort.By != sortByDimension && q.Sort.By != sortByMeasure {
		problems.Add("query.sort.by", msgSortBy)
	}
	if q.Sort.MeasureIndex < 0 || q.Sort.MeasureIndex >= max(len(q.Measures), 1) {
		problems.Add("query.sort.measureIndex", msgSortMeasure)
	}
	if q.Sort.Direction != directionAsc && q.Sort.Direction != directionDesc {
		problems.Add("query.sort.direction", msgSortDirection)
	}
}

// checkMeasureKinds applies the type rules that depend on what a measure works out.
func checkMeasureKinds(typ string, measures []measure, problems *errs.Problems) {
	for i, m := range measures {
		switch {
		case typ == typeScatter && m.kind() != KindNumber && m.kind() != KindMoney:
			problems.Add(fmt.Sprintf("query.measures[%d].field", i), msgScatterNeeds)
		case typ == typeTreemap && m.agg != aggSum && m.agg != aggCount:
			problems.Add(fmt.Sprintf("query.measures[%d].aggregation", i), msgTreemapNeeds)
		}
	}
}

func sameGrouping(a, b *GroupBy) bool {
	if a == nil || b == nil || a.Field != b.Field {
		return false
	}
	return groupBucket(a) == groupBucket(b)
}

func groupBucket(g *GroupBy) string {
	if g.DateGroup == nil {
		return ""
	}
	return *g.DateGroup
}

func lower(name string) string {
	if name == "" {
		return name
	}
	return string(name[0]+('a'-'A')) + name[1:]
}
