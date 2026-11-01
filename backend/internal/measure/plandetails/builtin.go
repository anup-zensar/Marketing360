package plandetails

import (
	"fmt"
	"slices"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// builtIn is the level's section of fields every record has (FR-006, FR-007, 043-C1); empty with no record.
func (v view) builtIn() []Widget {
	if len(v.records) == 0 {
		return []Widget{}
	}
	switch v.level {
	case LevelPlan:
		return []Widget{v.wordsWidget("statusCount", "status", planStatusWords(), statusOf)}
	case LevelCampaign, LevelInitiative:
		return []Widget{v.wordsWidget("progressCount", "approval progress", progressWords(), statusOf),
			v.wordsWidget("timingCount", "timing", timingWords(), timingOf)}
	case LevelActivity:
		return v.activityBuiltIn()
	}
	return []Widget{}
}

func (v view) activityBuiltIn() []Widget {
	types, statuses, starts := newTally(), newTally(), newTally()
	entered, goals := newTally(), newTally()
	starts.seed(v.periods()...)
	entered.seed(yesNoPart(true), yesNoPart(false))
	goals.seed(yesNoPart(true), yesNoPart(false))
	for _, a := range v.records {
		amount := plannedCostOf(a)
		typePart := v.typePart(a.ActivityTypeID)
		typePart.amount = amount
		types.add(a.ID, []part{typePart}, amount)
		statuses.add(a.ID, []part{wordPart(a.Status, activityStatusWords())}, money.Amount{})
		starts.add(a.ID, []part{v.startPart(a.StartDate)}, money.Amount{})
		entered.add(a.ID, []part{yesNoPart(amount.Sign() > 0)}, money.Amount{})
		goals.add(a.ID, []part{yesNoPart(hasGoal(a))}, money.Amount{})
	}
	startTitle := "Activities by start month"
	if v.period == measure.PeriodQuarter {
		startTitle = "Activities by start quarter"
	}
	return []Widget{
		types.widget(spec{key: "activityTypeCount", title: "Activities by activity type", measure: MeasureCount,
			filter: FilterActivityType}, v.currency),
		types.widget(spec{key: "activityTypePlannedCost", title: "Planned cost by activity type",
			measure: MeasurePlannedCost, filter: FilterActivityType}, v.currency),
		statuses.widget(spec{key: "statusCount", title: "Activities by status", measure: MeasureCount,
			filter: FilterActivityStatus}, v.currency),
		starts.widget(spec{key: "startPeriodCount", title: startTitle, measure: MeasureCount, timeOrder: true}, v.currency),
		entered.widget(spec{key: "plannedCostEntered", title: "Planned cost entered", measure: MeasureCount,
			timeOrder: true}, v.currency),
		goals.widget(spec{key: "goalEntered", title: "Goal entered", measure: MeasureCount, timeOrder: true}, v.currency),
	}
}

func (v view) typePart(typeID string) part {
	if typeID == "" {
		p := notFilledInPart()
		p.linkable = false
		return p
	}
	p := part{key: typeID, label: typeID, counted: true, linkable: true}
	if i := slices.IndexFunc(v.types, func(o allwork.Option) bool { return o.ID == typeID }); i >= 0 {
		p.label, p.inactive = v.types[i].Label, v.types[i].Inactive
	}
	if p.inactive {
		p.label += labelInactive
	}
	return p
}

// periods are the months or quarters of the plan year, in order (039 FR-016).
func (v view) periods() []part {
	count := 12
	if v.period == measure.PeriodQuarter {
		count = 4
	}
	parts := make([]part, 0, count)
	for n := 1; n <= count; n++ {
		parts = append(parts, v.periodPart(n))
	}
	return parts
}

// periodPart is the n-th month or quarter of the plan year.
func (v view) periodPart(n int) part {
	if v.period == measure.PeriodQuarter {
		return part{key: fmt.Sprintf("%d-Q%d", v.year, n), label: fmt.Sprintf("Q%d %d", n, v.year), order: n, counted: true}
	}
	return part{key: fmt.Sprintf("%d-%02d", v.year, n), label: fmt.Sprintf("%s %d", time.Month(n).String()[:3], v.year),
		order: n, counted: true}
}

// startPart is the period of an activity's start date, Outside <year> or No dates.
func (v view) startPart(start *date.Date) part {
	switch {
	case start == nil:
		return part{key: KeyNoDates, label: labelNoDates, order: orderNoDates, counted: true}
	case start.Year() != v.year:
		return part{key: KeyOutsideYear, label: fmt.Sprintf("Outside %d", v.year), order: orderOutside, counted: true}
	case v.period == measure.PeriodQuarter:
		return v.periodPart((int(start.Month())-1)/3 + 1)
	}
	return v.periodPart(int(start.Month()))
}

func hasGoal(a *allwork.Node) bool {
	positive := func(n *int64) bool { return n != nil && *n > 0 }
	return positive(a.INQ) || positive(a.MQL) || (a.Revenue != nil && a.Revenue.Sign() > 0)
}
