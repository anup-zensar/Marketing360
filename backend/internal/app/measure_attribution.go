package app

import (
	"mapp/internal/history"
	"mapp/internal/measure/attribution"
)

// attributionArea serves Settings › Attribution, the designated model of feature 050, logged like other settings.
func (w *wiring) attributionArea() error {
	tools := history.Tools{Tx: w.set.Tx, IDs: w.generator, Log: w.history, Now: w.now}
	service := attribution.NewService(tools, w.set.Attribution)
	if err := w.register(service.Kinds()); err != nil {
		return err
	}
	mount(w.signedIn, attribution.NewHandler(service))
	return nil
}
