package app

import (
	"context"

	"mapp/internal/allwork"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/widgetdata"
)

type values = map[string]widgetdata.Value

// levelReader answers one level of the plan tree as a built-in source, a row per record.
type levelReader struct {
	planning widgetPlanning
	source   string
	level    allwork.Level
	dated    bool
	row      func(d *planningData, n *allwork.Node) values
}

func (r levelReader) Rows(ctx context.Context, a actor.Actor, f widgetdata.Prefilter) (widgetdata.Rows, error) {
	d, err := r.planning.read(ctx, a, f, r.dated)
	if err != nil {
		return widgetdata.Rows{}, err
	}
	fields := widgetdata.BuiltInFields(r.source)
	out := widgetdata.Rows{Items: []widgetdata.Row{}, Notes: d.notes}
	for _, n := range d.nodes {
		if n.Level == r.level && (!r.dated || within(n.StartDate, f)) {
			out.Items = append(out.Items, widgetdata.RowOf(fields, r.row(d, n)))
		}
	}
	return out, nil
}

func planRow(d *planningData, n *allwork.Node) values {
	v := d.common(n)
	labels := allwork.LabelsOf(n)
	v[fieldPlan], v[fieldStatus], v[fieldPlanType] = text(n.Name), text(labels.Status), text(labels.PlanType)
	v[fieldCurrency] = text(n.PlanCurrency)
	return d.withBudget(n, v)
}

func campaignRow(d *planningData, n *allwork.Node) values {
	v := d.common(n)
	v[fieldCampaign], v[fieldStatus], v[fieldTiming] = text(n.Name), text(allwork.LabelsOf(n).Status), text(allwork.LabelsOf(n).Timing)
	v[fieldStartDate], v[fieldEndDate] = widgetdata.DateValue(n.StartDate), widgetdata.DateValue(n.EndDate)
	return d.withBudget(n, v)
}

func initiativeRow(d *planningData, n *allwork.Node) values {
	v := d.common(n)
	v[fieldInitiative], v[fieldCampaign], v[fieldStatus] = text(n.Name), d.nameOf(n.ParentID), text(allwork.LabelsOf(n).Status)
	v[fieldStartDate], v[fieldEndDate] = widgetdata.DateValue(n.StartDate), widgetdata.DateValue(n.EndDate)
	return d.withBudget(n, v)
}

func activityRow(d *planningData, n *allwork.Node) values {
	v := d.common(n)
	initiative := d.byID[n.ParentID]
	v[fieldActivity], v[fieldInitiative], v[fieldCampaign] = text(n.Name), d.nameOf(n.ParentID), d.parentName(initiative)
	v[fieldActivityType], v[fieldStatus] = text(n.Fields["activity.type"].Text), text(allwork.LabelsOf(n).Status)
	v[fieldComplete], v[fieldPlannedCost] = widgetdata.YesNo(n.Missing == 0), widgetdata.MoneyValue(n.PlannedCost)
	v[fieldStartDate], v[fieldEndDate] = widgetdata.DateValue(n.StartDate), widgetdata.DateValue(n.EndDate)
	return v
}

func lineItemRow(d *planningData, n *allwork.Node) values {
	v := d.common(n)
	activity := d.byID[n.ParentID]
	v[fieldLineItem], v[fieldActivity], v[fieldCampaign] = text(n.Name), d.nameOf(n.ParentID), d.parentName(d.byID[parentOf(activity)])
	v[fieldFinanceBudgetItem], v[fieldSwbsStatus] = text(n.Fields["lineItem.financeBudgetItem"].Text), text(allwork.LabelsOf(n).Swbs)
	v[fieldComplete], v[fieldPlannedCost] = widgetdata.YesNo(n.Missing == 0), widgetdata.MoneyValue(n.PlannedCost)
	if n.LocalPlannedCost != nil {
		v[fieldCurrency] = text(n.LocalPlannedCost.Currency)
	}
	return v
}

// common are the fields every level shares: its plan, year, teams, owner and, where seen, GSAP figures.
func (d *planningData) common(n *allwork.Node) values {
	p := d.byID[n.PlanID]
	v := values{fieldPlan: d.nameOf(n.PlanID), fieldYear: widgetdata.WholeValue(int64(n.PlanYear))}
	if p != nil {
		v[fieldTeam] = widgetdata.TextsValue(d.teamNamesOf(p.TeamIDs))
	}
	if n.Owner != nil {
		v[fieldOwner] = text(n.Owner.Name)
	}
	if d.finance[n.PlanID] {
		s := d.sums[n.ID]
		v[fieldCommitted], v[fieldActual] = widgetdata.MoneyValue(&s.committed), widgetdata.MoneyValue(&s.actual)
	}
	return v
}

func (d *planningData) withBudget(n *allwork.Node, v values) values {
	planned := d.sums[n.ID].planned
	v[fieldBudget], v[fieldPlannedCost] = widgetdata.MoneyValue(n.Budget), widgetdata.MoneyValue(&planned)
	if n.Budget != nil {
		left := n.Budget.Sub(planned)
		v[fieldLeftToPlan] = widgetdata.MoneyValue(&left)
	}
	return v
}

func (d *planningData) nameOf(id string) widgetdata.Value {
	if n := d.byID[id]; n != nil {
		return text(n.Name)
	}
	return widgetdata.TextValue("")
}

func (d *planningData) parentName(n *allwork.Node) widgetdata.Value {
	return d.nameOf(parentOf(n))
}

func (d *planningData) teamNamesOf(ids []string) []string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		names = append(names, d.teams[id])
	}
	return names
}

func parentOf(n *allwork.Node) string {
	if n == nil {
		return ""
	}
	return n.ParentID
}

func text(s string) widgetdata.Value { return widgetdata.TextValue(s) }

// within says whether a record's date falls in the dashboard's range, both ends included; no date never does.
func within(d *date.Date, f widgetdata.Prefilter) bool {
	if f.From == nil && f.To == nil {
		return true
	}
	if d == nil {
		return false
	}
	return (f.From == nil || !d.Before(*f.From)) && (f.To == nil || !d.After(*f.To))
}
