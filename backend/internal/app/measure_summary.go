package app

import "mapp/internal/measure"

// registerSummary registers the report of feature 048; its lane replaces the stub with the built report.
func registerSummary(_ *wiring, r *measure.Registry) error {
	return r.Register(measure.Stub(measure.KeySummary))
}
