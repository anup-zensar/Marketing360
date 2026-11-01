package app

import (
	"mapp/internal/history"
	"mapp/internal/performance"
)

// performanceArea builds feature 046 after the planning areas, whose activities its facts match to.
func (w *wiring) performanceArea() error {
	tools := history.Tools{Tx: w.set.Tx, IDs: w.generator, Log: w.history, Now: w.now}
	service := performance.NewService(tools, w.set.Performance)
	if err := w.register(service.Kinds()); err != nil {
		return err
	}
	mount(w.signedIn, performance.NewHandler(service, w.now))
	return nil
}
