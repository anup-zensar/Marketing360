package app

import "mapp/internal/measure"

// registerGsapLink registers the report of feature 044; its lane replaces the stub with the built report.
func registerGsapLink(_ *wiring, r *measure.Registry) error {
	return r.Register(measure.Stub(measure.KeyGsapLink))
}
