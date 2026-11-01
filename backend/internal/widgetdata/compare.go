package widgetdata

import (
	"math/big"

	"mapp/internal/platform/date"
	"mapp/internal/platform/errs"
)

const (
	compareTarget   = "target"
	comparePrevious = "previousPeriod"
	maxFilterIDs    = 20
	hoursPerDay     = 24
)

func checkFilters(f *Filters, problems *errs.Problems) {
	if f == nil {
		return
	}
	if f.DateFrom != nil && f.DateTo != nil && f.DateFrom.After(*f.DateTo) {
		problems.Add("filters.dateTo", msgDateOrder)
	}
	if len(f.PlanIDs) > maxFilterIDs {
		problems.Add("filters.planIds", msgTooManyPlans)
	}
	if len(f.TeamIDs) > maxFilterIDs {
		problems.Add("filters.teamIds", msgTooManyTeams)
	}
}

func checkComparison(r Request, src Source, problems *errs.Problems) {
	c := r.Comparison
	switch {
	case c == nil:
		return
	case r.Type != typeKpi:
		problems.Add("comparison", msgComparisonKpi)
	case c.Kind == compareTarget:
		if _, ok := targetOf(c); !ok {
			problems.Add("comparison.target", msgEnterTarget)
		}
	case c.Kind != comparePrevious:
		problems.Add("comparison.kind", msgComparisonKind)
	case src.Filters.Date == "":
		problems.Add("comparison.kind", msgPeriodNeedsDate)
	case !hasRange(r.Filters):
		problems.Add("filters", msgPeriodNeedsRange)
	}
}

func hasRange(f *Filters) bool {
	return f != nil && f.DateFrom != nil && f.DateTo != nil
}

func targetOf(c *ComparisonRequest) (Number, bool) {
	if c.Target == nil {
		return Number{}, false
	}
	n, err := ParseNumber(*c.Target)
	return n, err == nil
}

// againstTarget is the target in the measure's kind, with the share of it reached.
func againstTarget(c *ComparisonRequest, m measure, current Value) *Comparison {
	target, _ := targetOf(c)
	value := Value{Number: target}
	units := target.Units()
	if m.kind() == KindMoney {
		units = target.Units() / (numberScale / 100)
		value = unitsValue(KindMoney, units)
	}
	return &Comparison{Kind: compareTarget, Value: cellOf(value, m.kind()), Change: changeOf(current, units, m.kind(), false)}
}

// previousPeriod is the range of the same length just before from–to.
func previousPeriod(from, to date.Date) Period {
	days := int(to.Time().Sub(from.Time()).Hours()/hoursPerDay) + 1
	return Period{From: from.AddDays(-days), To: from.AddDays(-1)}
}

func againstPrevious(m measure, current, previous Value, period Period) *Comparison {
	c := &Comparison{Kind: comparePrevious, Value: cellOf(previous, m.kind()), Period: &period}
	if !previous.Empty {
		c.Change = changeOf(current, unitsIn(previous, m.kind()), m.kind(), true)
	}
	return c
}

// changeOf is current as a percent of base, or, as a change, (current − base) as a percent of |base|.
func changeOf(current Value, base int64, kind FieldKind, asChange bool) *SharePercent {
	if current.Empty || base == 0 {
		return nil
	}
	part := unitsIn(current, kind)
	whole := big.NewInt(base)
	if asChange {
		part -= base
		whole.Abs(whole)
	}
	return percentOf(part, whole)
}
