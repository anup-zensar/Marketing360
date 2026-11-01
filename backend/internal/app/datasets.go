package app

import (
	"context"

	"mapp/internal/customdashboard"
	"mapp/internal/dataset"
	"mapp/internal/history"
	"mapp/internal/platform/actor"
)

// datasetsArea comes after the custom dashboards, which tell a dataset's "Used by".
func (w *wiring) datasetsArea() error {
	tools := history.Tools{Tx: w.set.Tx, IDs: w.generator, Log: w.history, Now: w.now}
	w.datasets = dataset.NewService(tools, w.set.Datasets, dataset.Deps{Directory: w.people, Currencies: w.settings})
	if err := w.register(w.datasets.Kinds()); err != nil {
		return err
	}
	w.datasets.UseUsage(datasetUsage{dashboards: w.customDashboards})
	mount(w.signedIn, dataset.NewHandler(w.datasets))
	return nil
}

// datasetUsage implements dataset.Usage over the custom dashboards the actor may see.
type datasetUsage struct {
	dashboards *customdashboard.Service
}

func (u datasetUsage) UsingDataset(ctx context.Context, a actor.Actor, datasetID string) ([]dataset.DashboardRef, error) {
	refs, err := u.dashboards.UsingDataset(ctx, a, datasetID)
	if err != nil {
		return nil, err
	}
	out := make([]dataset.DashboardRef, 0, len(refs))
	for _, r := range refs {
		out = append(out, dataset.DashboardRef{ID: r.ID, Name: r.Name})
	}
	return out, nil
}
