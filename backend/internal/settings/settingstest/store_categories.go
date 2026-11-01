package settingstest

import (
	"context"
	"slices"

	"mapp/internal/settings"
)

func (s Store) GetActivityCategory(_ context.Context, id string) (settings.ActivityCategory, bool, error) {
	return s.categories.Get(id)
}

func (s Store) ListActivityCategories(context.Context) ([]settings.ActivityCategory, error) {
	return s.categories.List()
}

func (s Store) PutActivityCategory(ctx context.Context, c settings.ActivityCategory) error {
	c.ActivityTypeIDs = orEmpty(c.ActivityTypeIDs)
	if err := s.releaseCategoryTypes(ctx, c); err != nil {
		return err
	}
	return s.categories.Put(ctx, c)
}

func (s Store) DeleteActivityCategory(ctx context.Context, id string) error {
	return s.categories.Delete(ctx, id)
}

// releaseCategoryTypes mirrors the assignment table's primary key: an activity type assigned to c leaves
// every other category (047-C7).
func (s Store) releaseCategoryTypes(ctx context.Context, c settings.ActivityCategory) error {
	others, err := s.categories.Where(func(o settings.ActivityCategory) bool { return o.ID != c.ID })
	if err != nil {
		return err
	}
	for _, other := range others {
		kept := slices.DeleteFunc(slices.Clone(other.ActivityTypeIDs),
			func(id string) bool { return slices.Contains(c.ActivityTypeIDs, id) })
		if len(kept) == len(other.ActivityTypeIDs) {
			continue
		}
		other.ActivityTypeIDs = kept
		if err := s.categories.Put(ctx, other); err != nil {
			return err
		}
	}
	return nil
}
