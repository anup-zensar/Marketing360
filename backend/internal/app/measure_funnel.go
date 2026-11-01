package app

import "mapp/internal/measure"

// registerFunnel registers the report of feature 049; its lane replaces the stub with the built report.
func registerFunnel(_ *wiring, r *measure.Registry) error {
	return r.Register(measure.Stub(measure.KeyFunnel))
}
