package app

import (
	"context"

	"mapp/internal/customdashboard"
	"mapp/internal/history"
	"mapp/internal/people"
	"mapp/internal/stores"
)

func (w *wiring) customDashboardsArea() error {
	tools := history.Tools{Tx: w.set.Tx, IDs: w.generator, Log: w.history, Now: w.now}
	dashboards := customdashboard.NewService(tools, w.set.CustomDashboards,
		customDashboardDirectory{Service: w.people, repo: w.set.People})
	if err := w.register(dashboards.Kinds()); err != nil {
		return err
	}
	mount(w.signedIn, customdashboard.NewHandler(dashboards))
	return nil
}

// customDashboardDirectory implements customdashboard.Directory over the people service and its store.
type customDashboardDirectory struct {
	*people.Service
	repo stores.PeopleStore
}

func (d customDashboardDirectory) TeamsOf(ctx context.Context, personID string) ([]string, error) {
	memberships, err := d.repo.MembershipsOfPerson(ctx, personID)
	if err != nil {
		return nil, err
	}
	teams := make([]string, 0, len(memberships))
	for _, m := range memberships {
		teams = append(teams, m.TeamID)
	}
	return teams, nil
}
