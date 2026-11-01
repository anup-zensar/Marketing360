// Package datasettest is the in-memory fake of the datasets store, for tests that build services or the whole
// API without a database.
package datasettest

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"mapp/internal/dataset"
	"mapp/internal/people"
	"mapp/internal/platform/fake"
)

type rowSet struct {
	DatasetID string        `json:"datasetId"`
	Rows      []dataset.Row `json:"rows"`
}

type Store struct {
	datasets    *fake.Table[dataset.Dataset]
	rows        *fake.Table[rowSet]
	memberships *fake.Table[people.Membership]
}

func NewStore(db *fake.DB) Store {
	return Store{
		datasets:    fake.TableOf(db, "datasets", func(d dataset.Dataset) string { return d.ID }),
		rows:        fake.TableOf(db, "dataset_rows", func(r rowSet) string { return r.DatasetID }),
		memberships: fake.TableOf(db, "team_memberships", func(m people.Membership) string { return m.ID }),
	}
}

func (s Store) Get(_ context.Context, id string) (dataset.Dataset, bool, error) {
	return s.datasets.Get(id)
}

func (s Store) GetVisible(ctx context.Context, personID, id string) (dataset.Dataset, bool, error) {
	visible, err := s.ListVisible(ctx, personID)
	if err != nil {
		return dataset.Dataset{}, false, err
	}
	i := slices.IndexFunc(visible, func(d dataset.Dataset) bool { return d.ID == id })
	if i < 0 {
		return dataset.Dataset{}, false, nil
	}
	return visible[i], true, nil
}

// ListVisible orders by name ignoring case, then id, as the SQL does.
func (s Store) ListVisible(_ context.Context, personID string) ([]dataset.Dataset, error) {
	memberships, err := s.memberships.Where(func(m people.Membership) bool { return m.PersonID == personID })
	if err != nil {
		return nil, err
	}
	visible, err := s.datasets.Where(func(d dataset.Dataset) bool {
		return d.OwnerID == personID || slices.ContainsFunc(memberships, func(m people.Membership) bool {
			return slices.Contains(d.SharedWith, m.TeamID)
		})
	})
	slices.SortStableFunc(visible, func(a, b dataset.Dataset) int {
		return cmp.Or(strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), cmp.Compare(len(a.ID), len(b.ID)),
			strings.Compare(a.ID, b.ID))
	})
	return visible, err
}

func (s Store) CountOfOwner(_ context.Context, ownerID string) (int, error) {
	owned, err := s.datasets.Where(func(d dataset.Dataset) bool { return d.OwnerID == ownerID })
	return len(owned), err
}

func (s Store) OwnerOf(_ context.Context, id string) (string, bool, error) {
	d, found, err := s.datasets.Get(id)
	return d.OwnerID, found, err
}

func (s Store) Rows(_ context.Context, id string) ([]dataset.Row, error) {
	set, _, err := s.rows.Get(id)
	return append([]dataset.Row{}, set.Rows...), err
}

func (s Store) RowsPage(ctx context.Context, id string, offset, limit int) ([]dataset.Row, error) {
	rows, err := s.Rows(ctx, id)
	from := min(offset, len(rows))
	return rows[from:min(from+limit, len(rows))], err
}

func (s Store) Insert(ctx context.Context, d dataset.Dataset, rows []dataset.Row) error {
	if err := s.datasets.Put(ctx, d); err != nil {
		return err
	}
	return s.rows.Put(ctx, rowSet{DatasetID: d.ID, Rows: rows})
}

func (s Store) Update(ctx context.Context, d dataset.Dataset) error {
	return s.change(ctx, d.ID, func(stored *dataset.Dataset) {
		stored.Name, stored.Description, stored.Meta = d.Name, d.Description, d.Meta
	})
}

func (s Store) UpdateLabels(ctx context.Context, id string, columns []dataset.Column) error {
	return s.change(ctx, id, func(stored *dataset.Dataset) { stored.Columns = columns })
}

func (s Store) ReplaceShares(ctx context.Context, id string, teamIDs []string) error {
	return s.change(ctx, id, func(stored *dataset.Dataset) { stored.SharedWith = slices.Sorted(slices.Values(teamIDs)) })
}

func (s Store) change(ctx context.Context, id string, apply func(*dataset.Dataset)) error {
	stored, found, err := s.datasets.Get(id)
	if err != nil || !found {
		return err
	}
	apply(&stored)
	return s.datasets.Put(ctx, stored)
}

func (s Store) Delete(ctx context.Context, id string) error {
	if err := s.rows.Delete(ctx, id); err != nil {
		return err
	}
	return s.datasets.Delete(ctx, id)
}
