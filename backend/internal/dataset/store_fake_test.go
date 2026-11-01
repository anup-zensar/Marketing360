package dataset

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"mapp/internal/people"
	"mapp/internal/platform/fake"
)

type rowSet struct {
	DatasetID string `json:"datasetId"`
	Rows      []Row  `json:"rows"`
}

// fakeStore is datasettest.Store for this package's own tests, which cannot import datasettest.
type fakeStore struct {
	datasets    *fake.Table[Dataset]
	rows        *fake.Table[rowSet]
	memberships *fake.Table[people.Membership]
}

func newFakeStore(db *fake.DB) fakeStore {
	return fakeStore{
		datasets:    fake.TableOf(db, "datasets", func(d Dataset) string { return d.ID }),
		rows:        fake.TableOf(db, "dataset_rows", func(r rowSet) string { return r.DatasetID }),
		memberships: fake.TableOf(db, "team_memberships", func(m people.Membership) string { return m.ID }),
	}
}

func (s fakeStore) Get(_ context.Context, id string) (Dataset, bool, error) {
	return s.datasets.Get(id)
}

func (s fakeStore) GetVisible(ctx context.Context, personID, id string) (Dataset, bool, error) {
	visible, err := s.ListVisible(ctx, personID)
	if err != nil {
		return Dataset{}, false, err
	}
	i := slices.IndexFunc(visible, func(d Dataset) bool { return d.ID == id })
	if i < 0 {
		return Dataset{}, false, nil
	}
	return visible[i], true, nil
}

func (s fakeStore) ListVisible(_ context.Context, personID string) ([]Dataset, error) {
	memberships, err := s.memberships.Where(func(m people.Membership) bool { return m.PersonID == personID })
	if err != nil {
		return nil, err
	}
	visible, err := s.datasets.Where(func(d Dataset) bool {
		return d.OwnerID == personID || slices.ContainsFunc(memberships, func(m people.Membership) bool {
			return slices.Contains(d.SharedWith, m.TeamID)
		})
	})
	slices.SortStableFunc(visible, func(a, b Dataset) int {
		return cmp.Or(strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), cmp.Compare(len(a.ID), len(b.ID)),
			strings.Compare(a.ID, b.ID))
	})
	return visible, err
}

func (s fakeStore) CountOfOwner(_ context.Context, ownerID string) (int, error) {
	owned, err := s.datasets.Where(func(d Dataset) bool { return d.OwnerID == ownerID })
	return len(owned), err
}

func (s fakeStore) OwnerOf(_ context.Context, id string) (string, bool, error) {
	d, found, err := s.datasets.Get(id)
	return d.OwnerID, found, err
}

func (s fakeStore) Rows(_ context.Context, id string) ([]Row, error) {
	set, _, err := s.rows.Get(id)
	return append([]Row{}, set.Rows...), err
}

func (s fakeStore) RowsPage(ctx context.Context, id string, offset, limit int) ([]Row, error) {
	rows, err := s.Rows(ctx, id)
	from := min(offset, len(rows))
	return rows[from:min(from+limit, len(rows))], err
}

func (s fakeStore) Insert(ctx context.Context, d Dataset, rows []Row) error {
	if err := s.datasets.Put(ctx, d); err != nil {
		return err
	}
	return s.rows.Put(ctx, rowSet{DatasetID: d.ID, Rows: rows})
}

func (s fakeStore) Update(ctx context.Context, d Dataset) error {
	return s.change(ctx, d.ID, func(stored *Dataset) {
		stored.Name, stored.Description, stored.Meta = d.Name, d.Description, d.Meta
	})
}

func (s fakeStore) UpdateLabels(ctx context.Context, id string, columns []Column) error {
	return s.change(ctx, id, func(stored *Dataset) { stored.Columns = columns })
}

func (s fakeStore) ReplaceShares(ctx context.Context, id string, teamIDs []string) error {
	return s.change(ctx, id, func(stored *Dataset) { stored.SharedWith = slices.Sorted(slices.Values(teamIDs)) })
}

func (s fakeStore) change(ctx context.Context, id string, apply func(*Dataset)) error {
	stored, found, err := s.datasets.Get(id)
	if err != nil || !found {
		return err
	}
	apply(&stored)
	return s.datasets.Put(ctx, stored)
}

func (s fakeStore) Delete(ctx context.Context, id string) error {
	if err := s.rows.Delete(ctx, id); err != nil {
		return err
	}
	return s.datasets.Delete(ctx, id)
}
