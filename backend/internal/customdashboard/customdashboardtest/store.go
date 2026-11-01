// Package customdashboardtest is the in-memory fake of the custom dashboards store, for tests that build
// services or the whole API without a database.
package customdashboardtest

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"mapp/internal/customdashboard"
	"mapp/internal/people"
	"mapp/internal/platform/fake"
)

var errChanged = errors.New("customdashboardtest: the dashboard changed meanwhile")

type Store struct {
	dashboards  *fake.Table[customdashboard.Dashboard]
	memberships *fake.Table[people.Membership]
}

func NewStore(db *fake.DB) Store {
	return Store{
		dashboards:  fake.TableOf(db, "custom_dashboards", func(d customdashboard.Dashboard) string { return d.ID }),
		memberships: fake.TableOf(db, "team_memberships", func(m people.Membership) string { return m.ID }),
	}
}

func (s Store) ListVisible(_ context.Context, personID string) ([]customdashboard.Dashboard, error) {
	teams, err := s.teamsOf(personID)
	if err != nil {
		return nil, err
	}
	return s.dashboards.Where(func(d customdashboard.Dashboard) bool { return isVisible(d, personID, teams) })
}

func (s Store) GetVisible(_ context.Context, personID, id string) (customdashboard.Dashboard, bool, error) {
	d, found, err := s.dashboards.Get(id)
	if err != nil || !found {
		return customdashboard.Dashboard{}, false, err
	}
	teams, err := s.teamsOf(personID)
	if err != nil || !isVisible(d, personID, teams) {
		return customdashboard.Dashboard{}, false, err
	}
	return d, true, nil
}

func (s Store) Get(_ context.Context, id string) (customdashboard.Dashboard, bool, error) {
	return s.dashboards.Get(id)
}

func (s Store) Insert(ctx context.Context, d customdashboard.Dashboard) error {
	return s.dashboards.Put(ctx, d)
}

func (s Store) Update(ctx context.Context, d customdashboard.Dashboard, previousVersion int) error {
	stored, found, err := s.dashboards.Get(d.ID)
	if err != nil {
		return err
	}
	if !found || stored.Version != previousVersion {
		return fmt.Errorf("%w: %s", errChanged, d.ID)
	}
	d.Widgets, d.SharedWith = stored.Widgets, stored.SharedWith
	return s.dashboards.Put(ctx, d)
}

func (s Store) Delete(ctx context.Context, id string) error {
	return s.dashboards.Delete(ctx, id)
}

func (s Store) ReplaceWidgets(ctx context.Context, id string, widgets []customdashboard.Widget) error {
	return s.change(ctx, id, func(d *customdashboard.Dashboard) { d.Widgets = widgets })
}

func (s Store) ReplaceShares(ctx context.Context, id string, teamIDs []string) error {
	return s.change(ctx, id, func(d *customdashboard.Dashboard) { d.SharedWith = teamIDs })
}

func (s Store) UsingDataset(_ context.Context, personID, datasetID string) ([]customdashboard.Ref, error) {
	teams, err := s.teamsOf(personID)
	if err != nil {
		return nil, err
	}
	using, err := s.dashboards.Where(func(d customdashboard.Dashboard) bool {
		return isVisible(d, personID, teams) && slices.Contains(customdashboard.DatasetSources(d.Widgets), datasetID)
	})
	refs := make([]customdashboard.Ref, 0, len(using))
	for _, d := range using {
		refs = append(refs, customdashboard.Ref{ID: d.ID, Name: d.Name})
	}
	slices.SortFunc(refs, func(a, b customdashboard.Ref) int {
		return cmp.Or(strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), strings.Compare(a.ID, b.ID))
	})
	return refs, err
}

func (s Store) NameTaken(_ context.Context, ownerID, name, exceptID string) (bool, error) {
	same, err := s.dashboards.Where(func(d customdashboard.Dashboard) bool {
		return d.OwnerID == ownerID && d.ID != exceptID && strings.EqualFold(d.Name, name)
	})
	return len(same) > 0, err
}

func (s Store) OwnerOf(_ context.Context, id string) (ownerID string, found bool, err error) {
	d, found, err := s.dashboards.Get(id)
	return d.OwnerID, found, err
}

func (s Store) change(ctx context.Context, id string, apply func(d *customdashboard.Dashboard)) error {
	d, found, err := s.dashboards.Get(id)
	if err != nil || !found {
		return err
	}
	apply(&d)
	return s.dashboards.Put(ctx, d)
}

func (s Store) teamsOf(personID string) ([]string, error) {
	of, err := s.memberships.Where(func(m people.Membership) bool { return m.PersonID == personID })
	teams := make([]string, 0, len(of))
	for _, m := range of {
		teams = append(teams, m.TeamID)
	}
	return teams, err
}

func isVisible(d customdashboard.Dashboard, personID string, teams []string) bool {
	return d.OwnerID == personID || slices.ContainsFunc(d.SharedWith, func(t string) bool { return slices.Contains(teams, t) })
}
