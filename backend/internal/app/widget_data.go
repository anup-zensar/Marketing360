package app

import (
	"context"

	"mapp/internal/allwork"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/widgetdata"
)

// widgetData wires feature 051's data catalogue and widget query over the built-in sources' scoped reads.
func (w *wiring) widgetData() {
	s := w.set
	planning := widgetPlanning{years: dashboardPlanning{allWorkSource: w.allWork}, plans: s.Plans, people: s.People,
		spend: w.financials, finance: financialAccess{tx: s.Tx, plans: s.Plans}, now: w.now}
	workflow := workflowReader{planning: planning, approvals: s.Approvals, purchases: s.PurchaseRequests, settings: w.settings}
	rounds := workflow
	rounds.rounds = true
	readers := map[string]widgetdata.SourceReader{
		"plannedByMonth": monthReader{planning: planning}, "approvals": rounds, "purchaseRequests": workflow,
		"financialTransactions": transactionsReader{planning: planning, financials: w.financials},
		"globalBudgets":         globalBudgetsReader{budgets: w.budgets},
	}
	for _, r := range levelReaders(planning) {
		readers[r.source] = r
	}
	service := widgetdata.NewService(s.Tx, widgetdata.Deps{Sources: readers, Datasets: noDatasets{}, Rates: w.settings})
	mount(w.signedIn, widgetdata.NewHandler(service))
}

func levelReaders(planning widgetPlanning) []levelReader {
	return []levelReader{
		{planning: planning, source: "plans", level: allwork.LevelPlan, row: planRow},
		{planning: planning, source: sourceCampaigns, level: allwork.LevelCampaign, dated: true, row: campaignRow},
		{planning: planning, source: sourceInitiatives, level: allwork.LevelInitiative, dated: true, row: initiativeRow},
		{planning: planning, source: sourceActivities, level: allwork.LevelActivity, dated: true, row: activityRow},
		{planning: planning, source: "lineItems", level: allwork.LevelLineItem, row: lineItemRow},
	}
}

// noDatasets stands in for the datasets until they are wired: none is visible, so each is not found.
type noDatasets struct{}

func (noDatasets) Visible(context.Context, actor.Actor) ([]widgetdata.DatasetInfo, error) {
	return []widgetdata.DatasetInfo{}, nil
}

func (noDatasets) Table(context.Context, actor.Actor, string) (widgetdata.Table, error) {
	return widgetdata.Table{}, errs.NotFound("This dataset doesn't exist.")
}
