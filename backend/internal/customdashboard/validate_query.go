package customdashboard

import (
	"fmt"
	"math/big"
	"slices"
	"strings"

	"mapp/internal/platform/date"
	"mapp/internal/platform/errs"
)

const (
	operatorContains = "contains"
	operatorBetween  = "between"
	aggregationCount = "count"
	aggregationSum   = "sum"
)

type groupCheck struct {
	at               string
	need             need
	missing, refused string
}

// valueRule is how many values an operator takes (data model §2.2) and what to say otherwise.
type valueRule struct {
	min, max int
	message  string
}

func valueRules() map[string]valueRule {
	one := valueRule{min: 1, max: 1, message: msgOneValue}
	none := valueRule{message: msgNoValues}
	return map[string]valueRule{
		"eq": one, "ne": one, operatorContains: one, "gt": one, "gte": one, "lt": one, "lte": one,
		"in":            {min: 1, max: maxFilterValues, message: msgInValues},
		operatorBetween: {min: 2, max: 2, message: msgBetweenValues},
		"isEmpty":       none, "notEmpty": none,
	}
}

func checkQuery(at string, rule typeRule, q Query, p *errs.Problems) {
	if !slices.Contains(builtInSources(), q.Source) && !isDatasetSource(q.Source) {
		p.Add(at+".source", msgUnknownSource)
	}
	checkGroups(at, rule, q, p)
	checkMeasureCount(at, rule, q, p)
	for i, m := range q.Measures {
		checkMeasure(indexed(at, "measures", i), rule, m, p)
	}
	checkWidgetFilters(at, q.Filters, p)
	checkSort(at, q, p)
	if q.Limit < 1 || q.Limit > maxLimit {
		p.Add(at+".limit", msgLimit)
	}
}

func checkGroups(at string, rule typeRule, q Query, p *errs.Problems) {
	noun := withArticle(rule.noun)
	checkGroup(groupCheck{at: at + ".dimension", need: rule.dimension, missing: msgChooseGroup,
		refused: fmt.Sprintf(msgNoGroup, noun)}, q.Dimension, p)
	checkGroup(groupCheck{at: at + ".series", need: rule.series, missing: msgChooseSeries,
		refused: fmt.Sprintf(msgNoSeries, noun)}, q.Series, p)
	if q.Dimension != nil && q.Series != nil && sameGroup(*q.Dimension, *q.Series) {
		p.Add(at+".series", msgSameAsGroup)
	}
}

func checkGroup(c groupCheck, g *GroupBy, p *errs.Problems) {
	switch {
	case g == nil && c.need == needRequired:
		p.Add(c.at, c.missing)
	case g != nil && c.need == needNone:
		p.Add(c.at, c.refused)
	case g != nil:
		checkGroupBy(c.at, *g, p)
	}
}

func checkGroupBy(at string, g GroupBy, p *errs.Problems) {
	if isBlankField(g.Field) {
		p.Add(at+".field", msgFieldName)
	}
	if g.DateGroup != nil && !slices.Contains([]string{"day", "week", "month", "quarter", "year"}, *g.DateGroup) {
		p.Add(at+".dateGroup", msgDateGroup)
	}
}

func sameGroup(a, b GroupBy) bool {
	return a.Field == b.Field && (a.DateGroup == nil) == (b.DateGroup == nil) &&
		(a.DateGroup == nil || *a.DateGroup == *b.DateGroup)
}

func isBlankField(field string) bool {
	return strings.TrimSpace(field) == "" || tooLong(field, maxFieldLength)
}

func checkMeasureCount(at string, rule typeRule, q Query, p *errs.Problems) {
	noun, least, most := rule.noun, rule.minMeasures, rule.maxMeasures
	if q.Series != nil && oneMeasureWithSeries(rule.name) {
		noun, least, most = rule.noun+" with a breakdown", 1, 1
	}
	if n := len(q.Measures); n < least || n > most {
		p.Add(at+".measures", fmt.Sprintf(msgMeasureCount, withArticle(noun), measureCountText(least, most)))
	}
}

func measureCountText(least, most int) string {
	if least == 1 && most == 1 {
		return "one measure"
	}
	return fmt.Sprintf("%d to %d measures", least, most)
}

func checkMeasure(at string, rule typeRule, m Measure, p *errs.Problems) {
	if message := aggregationProblem(rule, m.Aggregation); message != "" {
		p.Add(at+".aggregation", message)
	}
	countsRecords := m.Aggregation == aggregationCount && m.Field == nil
	if !countsRecords && (m.Field == nil || isBlankField(*m.Field)) {
		p.Add(at+".field", msgMeasureField)
	}
	if tooLong(m.Label, maxLabelLength) {
		p.Add(at+".label", msgLabelTooLong)
	}
}

func aggregationProblem(rule typeRule, aggregation string) string {
	switch {
	case !slices.Contains([]string{aggregationCount, "countDistinct", aggregationSum, "avg", "min", "max"}, aggregation):
		return msgAggregation
	case rule.name == typeTreemap && aggregation != aggregationSum && aggregation != aggregationCount:
		return msgTreemapAggregation
	}
	return ""
}

func checkWidgetFilters(at string, filters []Filter, p *errs.Problems) {
	if len(filters) > maxWidgetFilters {
		p.Add(at+".filters", msgTooManyFilters)
		return
	}
	for i, f := range filters {
		checkWidgetFilter(indexed(at, "filters", i), f, p)
	}
}

func checkWidgetFilter(at string, f Filter, p *errs.Problems) {
	if isBlankField(f.Field) {
		p.Add(at+".field", msgFieldName)
	}
	rule, known := valueRules()[f.Operator]
	if !known {
		p.Add(at+".operator", msgOperator)
		return
	}
	if message := valuesProblem(f, rule); message != "" {
		p.Add(at+".values", message)
	}
}

func valuesProblem(f Filter, rule valueRule) string {
	if n := len(f.Values); n < rule.min || n > rule.max {
		return rule.message
	}
	switch {
	case slices.ContainsFunc(f.Values, isBadValue):
		return msgValueLength
	case f.Operator == operatorContains && tooLong(f.Values[0], maxContainsLength):
		return msgContainsLength
	case f.Operator == operatorBetween && isDescending(f.Values[0], f.Values[1]):
		return msgBetweenOrder
	}
	return ""
}

func isBadValue(v string) bool {
	return v == "" || tooLong(v, maxValueLength)
}

// isDescending compares two dates or two decimals; values of another kind are left to the data catalogue.
func isDescending(from, to string) bool {
	if a, err := date.Parse(from); err == nil {
		b, err := date.Parse(to)
		return err == nil && a.After(b)
	}
	a, aIsNumber := decimalOf(from)
	b, bIsNumber := decimalOf(to)
	return aIsNumber && bIsNumber && a.Cmp(b) > 0
}

func decimalOf(text string) (*big.Rat, bool) {
	if !decimalPattern.MatchString(text) {
		return nil, false
	}
	return new(big.Rat).SetString(text)
}

func checkSort(at string, q Query, p *errs.Problems) {
	if q.Sort.By != "dimension" && q.Sort.By != "measure" {
		p.Add(at+".sort.by", msgSortBy)
	}
	if q.Sort.Direction != "asc" && q.Sort.Direction != "desc" {
		p.Add(at+".sort.direction", msgSortDirection)
	}
	if q.Sort.MeasureIndex < 0 || q.Sort.MeasureIndex >= max(1, len(q.Measures)) {
		p.Add(at+".sort.measureIndex", msgSortMeasure)
	}
}
