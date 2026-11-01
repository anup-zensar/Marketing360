package app

import (
	"context"

	"mapp/internal/activity"
	"mapp/internal/history"
	"mapp/internal/performance"
	"mapp/internal/performance/simulatedfeed"
)

const feedSeed = 46

// performanceApprovals answers 046-C1: In progress and Complete are display forms of approved.
type performanceApprovals struct {
	activities interface {
		ListByIDs(ctx context.Context, ids []string) ([]activity.Activity, error)
	}
}

func (p performanceApprovals) ApprovedIDs(ctx context.Context, ids []string) (map[string]bool, error) {
	found, err := p.activities.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	approved := make(map[string]bool, len(found))
	for _, a := range found {
		approved[a.ID] = a.Status == activity.StatusApproved
	}
	return approved, nil
}

// performanceArea builds feature 046 after the planning areas, whose activities its facts match to.
func (w *wiring) performanceArea() error {
	tools := history.Tools{Tx: w.set.Tx, IDs: w.generator, Log: w.history, Now: w.now}
	service := performance.NewService(tools, w.set.Performance)
	service.UseApprovals(performanceApprovals{activities: w.set.Activities})
	service.UseFeed(simulatedfeed.New(feedSeed))
	if err := w.register(service.Kinds()); err != nil {
		return err
	}
	mount(w.signedIn, performance.NewHandler(service, w.now))
	return nil
}
