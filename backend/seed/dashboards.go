package seed

import (
	"context"

	"mapp/internal/customdashboard"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

// The demo dashboard of feature 051: DB-1001 by Priya Raman, shared with her team; its last widget reads
// her dataset DS-1001, which she has not shared, so her teammates see "You don't have access to this data".
const (
	demoDashboard      = "Lubricants Europe overview"
	demoDashboardOwner = "Priya Raman"
	demoDashboardTeam  = "Lubricants Europe"
	demoDashboardText  = "Figures are in US dollars at today's rates. Each person sees only the plans they may see."
	demoDataset        = "DS-1001"
	plannedCostField   = "plannedCost"
	barChart           = "bar"
)

func (d *demo) addDashboards(ctx context.Context) error {
	var r resolver
	ownerID := r.id(d.people, "person", demoDashboardOwner)
	teamID := r.id(d.teams, "team", demoDashboardTeam)
	if r.err != nil {
		return r.err
	}
	id, err := d.next(ctx, ids.PrefixDashboard)
	if err != nil {
		return err
	}
	return d.stores.CustomDashboards.Insert(ctx, customdashboard.Dashboard{
		ID: id, Name: demoDashboard, Description: "Planned cost and budgets of the Lubricants Europe plans.",
		OwnerID: ownerID, SharedWith: []string{teamID},
		Filters: customdashboard.Filters{PlanIDs: []string{}, TeamIDs: []string{}},
		Widgets: demoWidgets(),
		Meta:    record.NewMeta(actor.Actor{ID: ownerID, Name: demoDashboardOwner}, d.at),
	})
}

func demoWidgets() []customdashboard.Widget {
	byType := grouped("activities", "activityType", nil, sumOf(plannedCostField))
	byMonth := grouped("plannedByMonth", "month", ptrTo("month"), sumOf(plannedCostField))
	byMonth.Series = &customdashboard.GroupBy{Field: "campaign"}
	byMonth.Sort = inTimeOrder()
	budgets := grouped("campaigns", "campaign", nil, sumOf("budget"), sumOf(plannedCostField))
	budgets.Sort = inTimeOrder()
	total := grouped("plans", "", nil, sumOf("budget"))
	total.Dimension = nil
	leads := grouped(demoDataset, "c2", ptrTo("month"), sumOf("c3"))
	leads.Sort = inTimeOrder()
	return []customdashboard.Widget{
		{ID: "w1", Type: barChart, Title: "Planned cost by activity type", Layout: layout(0, 0, 6, 4), Query: byType},
		{ID: "w2", Type: "line", Title: "Planned cost by month", Subtitle: "By campaign", Layout: layout(6, 0, 6, 4), Query: byMonth},
		{ID: "w3", Type: "kpi", Title: "Plan budgets", Layout: layout(0, 4, 3, 2), Query: total},
		{ID: "w4", Type: "table", Title: "Campaigns", Layout: layout(3, 4, 6, 5), Query: budgets},
		{ID: "w5", Type: "text", Title: "About these figures", Layout: layout(9, 4, 3, 2), Note: demoDashboardText},
		{ID: "w6", Type: barChart, Title: "Event leads by month", Layout: layout(0, 9, 6, 4), Query: leads},
	}
}

func grouped(source, field string, dateGroup *string, measures ...customdashboard.Measure) *customdashboard.Query {
	return &customdashboard.Query{
		Source: source, Dimension: &customdashboard.GroupBy{Field: field, DateGroup: dateGroup}, Measures: measures,
		Filters: []customdashboard.Filter{}, Sort: customdashboard.Sort{By: "measure", Direction: "desc"}, Limit: 50,
	}
}

func sumOf(field string) customdashboard.Measure {
	return customdashboard.Measure{Aggregation: "sum", Field: &field}
}

// inTimeOrder sorts by the group ascending: months in order, campaigns by name.
func inTimeOrder() customdashboard.Sort {
	return customdashboard.Sort{By: "dimension", Direction: "asc"}
}

func layout(x, y, w, h int) customdashboard.Layout {
	return customdashboard.Layout{X: x, Y: y, W: w, H: h}
}

func ptrTo(text string) *string {
	return &text
}
