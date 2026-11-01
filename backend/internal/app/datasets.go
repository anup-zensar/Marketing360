package app

import (
	"mapp/internal/dataset"
	"mapp/internal/history"
)

func (w *wiring) datasetsArea() error {
	tools := history.Tools{Tx: w.set.Tx, IDs: w.generator, Log: w.history, Now: w.now}
	service := dataset.NewService(tools, w.set.Datasets, dataset.Deps{Directory: w.people, Currencies: w.settings})
	if err := w.register(service.Kinds()); err != nil {
		return err
	}
	mount(w.signedIn, dataset.NewHandler(service))
	return nil
}
