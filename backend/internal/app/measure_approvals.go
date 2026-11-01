package app

import "mapp/internal/measure"

// registerApprovals registers the report of feature 041; its lane replaces the stub with the built report.
func registerApprovals(_ *wiring, r *measure.Registry) error {
	return r.Register(measure.Stub(measure.KeyApprovals))
}
