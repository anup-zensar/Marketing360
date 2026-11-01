package widgetdata

import (
	"cmp"
	"fmt"
	"slices"
	"unicode/utf8"

	"mapp/internal/platform/date"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/money"
)

const (
	maxMeasures     = 8
	maxFilters      = 10
	maxLabel        = 40
	maxFilterValues = 50
	maxText         = 200
	maxContains     = 100
)

// compile checks a request against its source's fields and answers every problem at once.
func compile(r Request, src Source) (*compiled, error) {
	var problems errs.Problems
	checkShape(r, &problems)
	q := r.Query
	c := &compiled{typ: r.Type, source: src, sort: q.Sort, limit: q.Limit, groupRest: q.GroupRest, seriesCap: maxSeries}
	c.dimension = groupingOf(src, q.Dimension, "query.dimension", &problems)
	c.series = groupingOf(src, q.Series, "query.series", &problems)
	c.measures = measuresOf(src, q.Measures, &problems)
	checkMeasureKinds(r.Type, c.measures, &problems)
	c.conditions = conditionsOf(src, q.Filters, &problems)
	checkFilters(r.Filters, &problems)
	checkComparison(r, src, &problems)
	if err := problems.Err(); err != nil {
		return nil, err
	}
	switch r.Type {
	case typeFunnel:
		c.sort = Sort{By: sortByMeasure, Direction: directionDesc}
	case typeRadar:
		c.seriesCap = maxRadarShapes
	}
	return c, nil
}

func (s Source) field(key string) (int, Field, bool) {
	i := slices.IndexFunc(s.Fields, func(f Field) bool { return f.Key == key })
	if i < 0 {
		return -1, Field{}, false
	}
	return i, s.Fields[i], true
}

func groupingOf(src Source, g *GroupBy, path string, problems *errs.Problems) *grouping {
	if g == nil {
		return nil
	}
	i, f, found := src.field(g.Field)
	bucket := groupBucket(g)
	switch {
	case !found:
		problems.Add(path+".field", msgNoField)
	case !IsGroupable(f):
		problems.Add(path+".field", msgNotGroupable)
	case f.Kind == KindDate && !slices.Contains(buckets, bucket):
		problems.Add(path+".dateGroup", msgChooseBucket)
	case f.Kind != KindDate && bucket != "":
		problems.Add(path+".dateGroup", msgBucketOnlyDates)
	default:
		return &grouping{index: i, field: f, bucket: bucket}
	}
	return nil
}

func measuresOf(src Source, in []Measure, problems *errs.Problems) []measure {
	out := make([]measure, 0, len(in))
	for i, m := range in[:min(len(in), maxMeasures)] {
		path := fmt.Sprintf("query.measures[%d]", i)
		if utf8.RuneCountInString(m.Label) > maxLabel {
			problems.Add(path+".label", msgLabelLength)
		}
		if c, ok := measureOf(src, m, path, problems); ok {
			out = append(out, c)
		}
	}
	return out
}

func measureOf(src Source, m Measure, path string, problems *errs.Problems) (measure, bool) {
	if _, known := aggregationWords[m.Aggregation]; !known {
		problems.Add(path+".aggregation", msgChooseAggregation)
		return measure{}, false
	}
	c := measure{agg: m.Aggregation, index: -1, label: m.Label}
	if m.Field == nil {
		if m.Aggregation != aggCount {
			problems.Add(path+".field", msgChooseField)
			return measure{}, false
		}
		c.label = cmp.Or(c.label, "Count of records")
		return c, true
	}
	i, f, found := src.field(*m.Field)
	if !found {
		problems.Add(path+".field", msgNoField)
		return measure{}, false
	}
	if !slices.Contains(AggregationsOf(f), m.Aggregation) {
		problems.Add(path+".field", aggregationNeeds(m.Aggregation))
		return measure{}, false
	}
	c.index, c.field = i, f
	c.label = cmp.Or(c.label, aggregationWords[m.Aggregation]+" "+f.Label)
	return c, true
}

func aggregationNeeds(agg string) string {
	switch agg {
	case aggSum:
		return msgSumNeeds
	case aggAvg:
		return msgAvgNeeds
	case aggMin:
		return msgMinNeeds
	}
	return msgMaxNeeds
}

func conditionsOf(src Source, in []Filter, problems *errs.Problems) []condition {
	if len(in) > maxFilters {
		problems.Add("query.filters", msgTooManyFilters)
		return nil
	}
	out := make([]condition, 0, len(in))
	for i, f := range in {
		path := fmt.Sprintf("query.filters[%d]", i)
		index, field, found := src.field(f.Field)
		if !found {
			problems.Add(path+".field", msgNoField)
			continue
		}
		if !slices.Contains(operatorsOf(field.Kind), f.Operator) {
			problems.Add(path+".operator", operatorProblem(f.Operator))
			continue
		}
		values, ok := valuesOf(field, f, path+".values", problems)
		if ok {
			out = append(out, condition{index: index, field: field, op: f.Operator, values: values})
		}
	}
	return out
}

func operatorsOf(kind FieldKind) []string {
	every := []string{opEq, opNe, opIn, opIsEmpty, opNotEmpty}
	if kind == KindText {
		return append(every, opContains)
	}
	return append(every, opGt, opGte, opLt, opLte, opBetween)
}

func operatorProblem(op string) string {
	if slices.Contains(operatorsOf(KindText), op) || slices.Contains(operatorsOf(KindDate), op) {
		return msgConditionKind
	}
	return msgChooseCondition
}

func valuesOf(field Field, f Filter, path string, problems *errs.Problems) ([]Value, bool) {
	if msg := countProblem(f.Operator, len(f.Values)); msg != "" {
		problems.Add(path, msg)
		return nil, false
	}
	values := make([]Value, 0, len(f.Values))
	for _, text := range f.Values {
		v, msg := parseValue(field.Kind, f.Operator, text)
		if msg != "" {
			problems.Add(path, msg)
			return nil, false
		}
		values = append(values, v)
	}
	if f.Operator == opBetween && compareValues(values[0], values[1], field.Kind) > 0 {
		problems.Add(path, msgBetweenOrder)
		return nil, false
	}
	return values, true
}

func countProblem(op string, n int) string {
	switch op {
	case opIsEmpty, opNotEmpty:
		return pick(n == 0, msgNoValues)
	case opIn:
		return pick(n >= 1 && n <= maxFilterValues, msgSomeValues)
	case opBetween:
		return pick(n == 2, msgTwoValues)
	}
	return pick(n == 1, msgOneValue)
}

func parseValue(kind FieldKind, op, text string) (Value, string) {
	switch kind {
	case KindNumber:
		n, err := ParseNumber(text)
		return NumberValue(&n), pick(err == nil, msgEnterNumber)
	case KindMoney:
		a, err := money.ParseAmount(text)
		return MoneyValue(&a), pick(err == nil, msgEnterMoney)
	case KindDate:
		d, err := date.Parse(text)
		return DateValue(&d), pick(err == nil, msgEnterDate)
	case KindText:
	}
	length := utf8.RuneCountInString(text)
	if op == opContains {
		return TextValue(text), pick(length >= 1 && length <= maxContains, msgContainsLength)
	}
	return TextValue(text), pick(length >= 1 && length <= maxText, msgTextLength)
}

// pick is "" when ok, else the message.
func pick(ok bool, message string) string {
	if ok {
		return ""
	}
	return message
}
