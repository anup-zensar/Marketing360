package customdashboard

import (
	"regexp"
	"slices"
	"strings"
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
)

type need int

const (
	needNone need = iota
	needOptional
	needRequired
)

// typeRule is what a widget type needs (data model §2.3); noun names it mid-sentence.
type typeRule struct {
	name, label, noun string
	dimension, series need
	minMeasures       int
	maxMeasures       int
}

func typeRules() map[string]typeRule {
	chart := func(label, noun string, series need, measures ...int) typeRule {
		return typeRule{label: label, noun: noun, dimension: needRequired, series: series,
			minMeasures: measures[0], maxMeasures: measures[1]}
	}
	rules := map[string]typeRule{
		typeBar:     chart("Bar", "bar chart", needOptional, 1, 4),
		typeLine:    chart("Line", "line chart", needOptional, 1, 4),
		typeArea:    chart("Area", "area chart", needOptional, 1, 4),
		typeCombo:   chart("Bar and line", "bar and line chart", needNone, 2, 4),
		typePie:     chart("Pie", "pie chart", needNone, 1, 1),
		typeDonut:   chart("Donut", "donut chart", needNone, 1, 1),
		typeScatter: chart("Scatter", "scatter chart", needOptional, 2, 3),
		typeFunnel:  chart("Funnel", "funnel chart", needNone, 1, 1),
		typeGauge:   {label: "Gauge", noun: "gauge", minMeasures: 1, maxMeasures: 1},
		typeHeatmap: chart("Heatmap", "heatmap", needRequired, 1, 1),
		typeTreemap: chart("Treemap", "treemap", needOptional, 1, 1),
		typeRadar:   chart("Radar", "radar chart", needOptional, 1, 1),
		typeKpi:     {label: "Key figure", noun: "key figure", minMeasures: 1, maxMeasures: 1},
		typeTable: {label: "Table", noun: "table", dimension: needOptional, series: needOptional,
			minMeasures: 1, maxMeasures: 8},
		typeText: {label: "Text note", noun: "text note"},
	}
	for name, rule := range rules {
		rule.name = name
		rules[name] = rule
	}
	return rules
}

// oneMeasureWithSeries lists the types that show a single measure once they are broken down.
func oneMeasureWithSeries(widgetType string) bool {
	return slices.Contains([]string{typeBar, typeLine, typeArea}, widgetType)
}

// displayOptions says which widget types draw each display option (data model §3).
func displayOptions() map[string][]string {
	charts := []string{typeBar, typeLine, typeArea, typeCombo, typePie, typeDonut, typeScatter, typeFunnel, typeGauge,
		typeHeatmap, typeTreemap, typeRadar}
	axis := []string{typeBar, typeLine, typeArea, typeCombo, typeScatter, typeHeatmap}
	withValues := append(slices.Clone(charts), typeKpi, typeTable)
	return map[string][]string{
		"palette":       charts,
		"legend":        slices.DeleteFunc(slices.Clone(charts), func(t string) bool { return t == typeGauge }),
		"dataLabels":    charts,
		"axisTitles":    axis,
		"gridlines":     axis,
		"smooth":        {typeLine, typeArea, typeCombo},
		"stack":         {typeBar, typeLine, typeArea},
		"orientation":   {typeBar, typeFunnel},
		"numberFormat":  withValues,
		"decimals":      withValues,
		"zoom":          {typeBar, typeLine, typeArea, typeCombo, typeScatter},
		"referenceLine": {typeBar, typeLine, typeArea, typeCombo},
		"measureStyles": {typeCombo},
		"gauge":         {typeGauge},
		"kpi":           {typeKpi},
		"table":         {typeTable},
	}
}

func builtInSources() []string {
	return []string{"plans", "campaigns", "initiatives", "activities", "lineItems", "plannedByMonth",
		"financialTransactions", "approvals", "purchaseRequests", "globalBudgets"}
}

var (
	datasetSourcePattern = regexp.MustCompile(`^DS-[0-9]+$`)
	widgetIDPattern      = regexp.MustCompile(`^w[1-9][0-9]{0,2}$`)
	decimalPattern       = regexp.MustCompile(`^-?[0-9]{1,15}(\.[0-9]{1,4})?$`)
	planIDPattern        = regexp.MustCompile(`^P-[0-9]+$`)
	teamIDPattern        = regexp.MustCompile(`^T-[0-9]+$`)
)

func isDatasetSource(source string) bool {
	return datasetSourcePattern.MatchString(source)
}

// DatasetSources lists, once each and sorted, the datasets the widgets read.
func DatasetSources(widgets []Widget) []string {
	sources := []string{}
	for _, w := range widgets {
		if w.Query != nil && isDatasetSource(w.Query.Source) && !slices.Contains(sources, w.Query.Source) {
			sources = append(sources, w.Query.Source)
		}
	}
	slices.SortFunc(sources, func(a, b string) int {
		if len(a) != len(b) {
			return len(a) - len(b)
		}
		return strings.Compare(a, b)
	})
	return sources
}
