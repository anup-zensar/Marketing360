package app

import "mapp/internal/measure"

// registerPlanDetails registers the report of feature 043; its lane replaces the stub with the built report.
func registerPlanDetails(_ *wiring, r *measure.Registry) error {
	return r.Register(measure.Stub(measure.KeyPlanDetails))
}
