package app

import "mapp/internal/measure"

// registerPlanBudget registers the report of feature 040; its lane replaces the stub with the built report.
func registerPlanBudget(_ *wiring, r *measure.Registry) error {
	return r.Register(measure.Stub(measure.KeyPlanBudget))
}
