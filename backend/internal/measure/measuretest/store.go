// Package measuretest is the in-memory fake of the measure store: the saved report views.
package measuretest

import (
	"context"

	"mapp/internal/measure"
	"mapp/internal/platform/fake"
)

type Store struct {
	views *fake.Table[measure.StoredView]
}

func NewStore(db *fake.DB) Store {
	return Store{views: fake.TableOf(db, "report_views", func(v measure.StoredView) string { return v.ID })}
}

func (s Store) ListViews(_ context.Context, personID, report string) ([]measure.StoredView, error) {
	return s.views.Where(func(v measure.StoredView) bool { return v.PersonID == personID && v.Report == report })
}

func (s Store) GetView(_ context.Context, id string) (measure.StoredView, bool, error) {
	return s.views.Get(id)
}

func (s Store) PutView(ctx context.Context, v measure.StoredView) error {
	return s.views.Put(ctx, v)
}

func (s Store) DeleteView(ctx context.Context, id string) error {
	return s.views.Delete(ctx, id)
}
