package app

import (
	"mapp/internal/measure"
	"mapp/internal/measure/gsaplink"
)

// registerGsapLink registers and mounts the GSAP link report (feature 044) on 039's shared Scope step.
func registerGsapLink(w *wiring, r *measure.Registry) error {
	service := gsaplink.NewService(w.set.Tx, gsaplink.Deps{Scoped: w.reports, Store: w.set.GsapLink, Rates: w.settings},
		w.now)
	mount(w.signedIn, gsaplink.NewHandler(service))
	return r.Register(service)
}
