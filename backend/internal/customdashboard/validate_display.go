package customdashboard

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"mapp/internal/platform/errs"
)

// checkDisplay refuses an option the type does not draw, then checks the values of the options it does.
func checkDisplay(at string, w Widget, rule typeRule, p *errs.Problems) {
	drawn := displayOptions()
	refused := false
	for _, name := range setOptions(w.Display) {
		if !slices.Contains(drawn[name], w.Type) {
			p.Add(at+"."+name, fmt.Sprintf(msgOptionNotForType, withArticle(rule.noun)))
			refused = true
		}
	}
	if refused {
		return
	}
	checkChoices(at, w.Display, p)
	checkNumbers(at, w.Display, p)
	checkTypeOptions(at, w, p)
}

func setOptions(d Display) []string {
	encoded, err := json.Marshal(d)
	if err != nil {
		return nil
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(encoded, &fields) != nil {
		return nil
	}
	return slices.Sorted(maps.Keys(fields))
}

func checkChoices(at string, d Display, p *errs.Problems) {
	oneOf(at+".palette", d.Palette, p, "categorical", "planning", "levels", "sequential")
	if d.Legend != nil {
		oneOf(at+".legend.position", &d.Legend.Position, p, "top", "bottom", "left", "right")
	}
	oneOf(at+".stack", d.Stack, p, "none", "stacked", "percent")
	oneOf(at+".orientation", d.Orientation, p, "vertical", "horizontal")
	oneOf(at+".numberFormat", d.NumberFormat, p, "number", "money", "percent", "compact")
}

func oneOf(field string, value *string, p *errs.Problems, allowed ...string) {
	if value != nil && !slices.Contains(allowed, *value) {
		p.Add(field, msgChooseOption)
	}
}

func checkNumbers(at string, d Display, p *errs.Problems) {
	if d.Decimals != nil && (*d.Decimals < 0 || *d.Decimals > maxDecimals) {
		p.Add(at+".decimals", msgDecimals)
	}
	if t := d.AxisTitles; t != nil {
		for name, title := range map[string]string{"x": t.X, "y": t.Y, "y2": t.Y2} {
			if tooLong(title, maxLabelLength) {
				p.Add(at+".axisTitles."+name, msgLabelTooLong)
			}
		}
	}
	if line := d.ReferenceLine; line != nil {
		if !decimalPattern.MatchString(line.Value) {
			p.Add(at+".referenceLine.value", msgDecimal)
		}
		if tooLong(line.Label, maxLabelLength) {
			p.Add(at+".referenceLine.label", msgLabelTooLong)
		}
	}
}

func checkTypeOptions(at string, w Widget, p *errs.Problems) {
	d, q := w.Display, w.Query
	if q == nil {
		q = &Query{}
	}
	if d.Stack != nil && *d.Stack != "none" && q.Series == nil {
		p.Add(at+".stack", msgStackNeedsSeries)
	}
	if d.MeasureStyles != nil && !areMeasureStyles(d.MeasureStyles, len(q.Measures)) {
		p.Add(at+".measureStyles", msgMeasureStyles)
	}
	if d.Gauge != nil {
		checkGauge(at+".gauge", *d.Gauge, p)
	}
	if d.Kpi != nil {
		checkKpi(at+".kpi", *d.Kpi, p)
	}
	if d.Table != nil && !areTableColumns(d.Table.Columns, *q) {
		p.Add(at+".table.columns", msgTableColumns)
	}
}

func areMeasureStyles(styles []MeasureStyle, measures int) bool {
	wrong := func(s MeasureStyle) bool {
		return (s.As != "bar" && s.As != "line") || (s.Axis != "left" && s.Axis != "right")
	}
	return len(styles) == measures && !slices.ContainsFunc(styles, wrong)
}

func checkGauge(at string, g GaugeOptions, p *errs.Problems) {
	least, minIsNumber := decimalOf(g.Min)
	if !minIsNumber {
		p.Add(at+".min", msgDecimal)
	}
	if g.Max == nil {
		return
	}
	most, maxIsNumber := decimalOf(*g.Max)
	switch {
	case !maxIsNumber:
		p.Add(at+".max", msgDecimal)
	case minIsNumber && least.Cmp(most) >= 0:
		p.Add(at, msgGaugeRange)
	}
}

func checkKpi(at string, k KpiOptions, p *errs.Problems) {
	oneOf(at+".comparison", &k.Comparison, p, "none", "target", "previousPeriod")
	switch {
	case k.Target != nil && !decimalPattern.MatchString(*k.Target):
		p.Add(at+".target", msgDecimal)
	case k.Target == nil && k.Comparison == "target":
		p.Add(at+".target", msgEnterTarget)
	}
}

func areTableColumns(columns []string, q Query) bool {
	allowed := []string{}
	if q.Dimension != nil {
		allowed = append(allowed, "dimension")
	}
	if q.Series != nil {
		allowed = append(allowed, "series")
	}
	for i := range q.Measures {
		allowed = append(allowed, fmt.Sprintf("m%d", i))
	}
	seen := map[string]bool{}
	for _, c := range columns {
		if seen[c] || !slices.Contains(allowed, c) {
			return false
		}
		seen[c] = true
	}
	return len(columns) <= maxTableColumns
}
