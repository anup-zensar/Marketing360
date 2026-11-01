package app

import (
	"context"
	"maps"
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/widgetdata"
)

// The widest years a spread is worked out over, so a month of another year keeps its month.
const (
	firstSpreadYear = 1
	lastSpreadYear  = 9999
)

// monthReader answers Planned cost by month: each activity's spread as All work's Budget & cost works it out.
type monthReader struct {
	planning widgetPlanning
}

func (r monthReader) Rows(ctx context.Context, a actor.Actor, f widgetdata.Prefilter) (widgetdata.Rows, error) {
	d, err := r.planning.read(ctx, a, f, true)
	if err != nil {
		return widgetdata.Rows{}, err
	}
	fields := widgetdata.BuiltInFields("plannedByMonth")
	out := widgetdata.Rows{Items: []widgetdata.Row{}, Notes: d.notes}
	for _, year := range slices.Sorted(maps.Keys(d.years)) {
		spread := allwork.PlannedByActivityInYears(d.years[year], firstSpreadYear, lastSpreadYear)
		for _, id := range slices.Sorted(maps.Keys(spread)) {
			if activity := d.byID[id]; activity != nil {
				out.Items = append(out.Items, d.monthRows(fields, activity, spread[id], f)...)
			}
		}
	}
	return out, nil
}

func (d *planningData) monthRows(fields []widgetdata.Field, n *allwork.Node, m allwork.Months, f widgetdata.Prefilter) []widgetdata.Row {
	var rows []widgetdata.Row
	initiative := d.byID[n.ParentID]
	base := values{fieldPlan: d.nameOf(n.PlanID), fieldCampaign: d.parentName(initiative), fieldInitiative: d.nameOf(n.ParentID),
		fieldActivity: text(n.Name), fieldActivityType: text(n.Fields["activity.type"].Text)}
	if p := d.byID[n.PlanID]; p != nil {
		base[fieldTeam] = widgetdata.TextsValue(d.teamNamesOf(p.TeamIDs))
	}
	for _, key := range slices.Sorted(maps.Keys(m.ByMonth)) {
		first, err := date.Parse(key + "-01")
		if err != nil || !monthWithin(first, f) {
			continue
		}
		amount := m.ByMonth[key]
		v := maps.Clone(base)
		v[fieldMonth], v[fieldPlannedCost] = widgetdata.DateValue(&first), widgetdata.MoneyValue(&amount)
		rows = append(rows, widgetdata.RowOf(fields, v))
	}
	if !m.NoDates.IsZero() && f.From == nil && f.To == nil {
		base[fieldPlannedCost] = widgetdata.MoneyValue(&m.NoDates)
		rows = append(rows, widgetdata.RowOf(fields, base))
	}
	return rows
}

// monthWithin says whether the month starting on first overlaps the dashboard's range.
func monthWithin(first date.Date, f widgetdata.Prefilter) bool {
	last := date.New(first.Year(), first.Month()+1, 1).AddDays(-1)
	return (f.From == nil || !last.Before(*f.From)) && (f.To == nil || !first.After(*f.To))
}
