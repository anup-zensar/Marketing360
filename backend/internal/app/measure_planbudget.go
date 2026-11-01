package app

import (
	"mapp/internal/measure"
	"mapp/internal/measure/planbudget"
)

// registerPlanBudget registers the Plan budget report (feature 040) and serves its widgets.
func registerPlanBudget(w *wiring, r *measure.Registry) error {
	service := planbudget.NewService(w.set.Tx, planbudget.Deps{Scoper: w.reports, Spending: w.financials}, w.now)
	mount(w.signedIn, planbudget.NewHandler(service))
	return r.Register(planbudget.NewReport(service))
}
