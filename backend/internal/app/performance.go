package app

import (
	"context"

	"mapp/internal/activity"
	"mapp/internal/history"
	"mapp/internal/performance"
	"mapp/internal/performance/simulatedfeed"
	"mapp/internal/platform/actor"
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

// heldRelease late-binds the release hook: approvals are wired before the performance area exists.
type heldRelease struct {
	service *performance.Service
}

func (r *heldRelease) activityApproved(ctx context.Context, a actor.Actor, activityID string) error {
	if r.service == nil {
		return nil
	}
	_, err := r.service.Release(ctx, a, activityID)
	return err
}

// performanceArea builds feature 046 after the planning areas, whose activities its facts match to.
func (w *wiring) performanceArea(release *heldRelease) error {
	tools := history.Tools{Tx: w.set.Tx, IDs: w.generator, Log: w.history, Now: w.now}
	service := performance.NewService(tools, w.set.Performance)
	service.UseApprovals(performanceApprovals{activities: w.set.Activities})
	service.UseFeed(simulatedfeed.New(feedSeed))
	release.service = service
	if err := w.register(service.Kinds()); err != nil {
		return err
	}
	mount(w.signedIn, performance.NewHandler(service, w.now))
	return nil
}
