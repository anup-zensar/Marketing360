package customdashboard

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"mapp/internal/people"
	"mapp/internal/platform/fake"
)

var errChanged = errors.New("customdashboard fake: the dashboard changed meanwhile")

// fakeStore is customdashboardtest.Store for this package's own tests, which cannot import it.
type fakeStore struct {
	dashboards  *fake.Table[Dashboard]
	memberships *fake.Table[people.Membership]
}

func newFakeStore(db *fake.DB) fakeStore {
	return fakeStore{
		dashboards:  fake.TableOf(db, "custom_dashboards", func(d Dashboard) string { return d.ID }),
		memberships: fake.TableOf(db, "team_memberships", func(m people.Membership) string { return m.ID }),
	}
}

func (s fakeStore) ListVisible(_ context.Context, personID string) ([]Dashboard, error) {
	teams, err := s.teamsOf(personID)
	if err != nil {
		return nil, err
	}
	return s.dashboards.Where(func(d Dashboard) bool { return isVisibleTo(d, personID, teams) })
}

func (s fakeStore) GetVisible(_ context.Context, personID, id string) (Dashboard, bool, error) {
	d, found, err := s.dashboards.Get(id)
	if err != nil || !found {
		return Dashboard{}, false, err
	}
	teams, err := s.teamsOf(personID)
	if err != nil || !isVisibleTo(d, personID, teams) {
		return Dashboard{}, false, err
	}
	return d, true, nil
}

func (s fakeStore) Get(_ context.Context, id string) (Dashboard, bool, error) {
	return s.dashboards.Get(id)
}

func (s fakeStore) Insert(ctx context.Context, d Dashboard) error {
	return s.dashboards.Put(ctx, d)
}

func (s fakeStore) Update(ctx context.Context, d Dashboard, previousVersion int) error {
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

func (s fakeStore) Delete(ctx context.Context, id string) error {
	return s.dashboards.Delete(ctx, id)
}

func (s fakeStore) ReplaceWidgets(ctx context.Context, id string, widgets []Widget) error {
	return s.change(ctx, id, func(d *Dashboard) { d.Widgets = widgets })
}

func (s fakeStore) ReplaceShares(ctx context.Context, id string, teamIDs []string) error {
	return s.change(ctx, id, func(d *Dashboard) { d.SharedWith = teamIDs })
}

func (s fakeStore) UsingDataset(_ context.Context, personID, datasetID string) ([]Ref, error) {
	teams, err := s.teamsOf(personID)
	if err != nil {
		return nil, err
	}
	using, err := s.dashboards.Where(func(d Dashboard) bool {
		return isVisibleTo(d, personID, teams) && slices.Contains(DatasetSources(d.Widgets), datasetID)
	})
	refs := make([]Ref, 0, len(using))
	for _, d := range using {
		refs = append(refs, Ref{ID: d.ID, Name: d.Name})
	}
	slices.SortFunc(refs, func(a, b Ref) int {
		return cmp.Or(strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), strings.Compare(a.ID, b.ID))
	})
	return refs, err
}

func (s fakeStore) NameTaken(_ context.Context, ownerID, name, exceptID string) (bool, error) {
	same, err := s.dashboards.Where(func(d Dashboard) bool {
		return d.OwnerID == ownerID && d.ID != exceptID && strings.EqualFold(d.Name, name)
	})
	return len(same) > 0, err
}

func (s fakeStore) OwnerOf(_ context.Context, id string) (ownerID string, found bool, err error) {
	d, found, err := s.dashboards.Get(id)
	return d.OwnerID, found, err
}

func (s fakeStore) change(ctx context.Context, id string, apply func(d *Dashboard)) error {
	d, found, err := s.dashboards.Get(id)
	if err != nil || !found {
		return err
	}
	apply(&d)
	return s.dashboards.Put(ctx, d)
}

func (s fakeStore) teamsOf(personID string) ([]string, error) {
	of, err := s.memberships.Where(func(m people.Membership) bool { return m.PersonID == personID })
	teams := make([]string, 0, len(of))
	for _, m := range of {
		teams = append(teams, m.TeamID)
	}
	return teams, err
}

func isVisibleTo(d Dashboard, personID string, teams []string) bool {
	return d.OwnerID == personID || slices.ContainsFunc(d.SharedWith, func(t string) bool { return slices.Contains(teams, t) })
}
