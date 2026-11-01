package app

import "mapp/internal/measure"

// registerPurchaseRequests registers the report of feature 042; its lane replaces the stub with the built report.
func registerPurchaseRequests(_ *wiring, r *measure.Registry) error {
	return r.Register(measure.Stub(measure.KeyPurchaseRequests))
}
