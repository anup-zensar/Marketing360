package settings

import (
	"context"
	"fmt"
	"slices"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/database/sqlcdb"
	"mapp/internal/platform/record"
)

func (s *PostgresStore) GetActivityCategory(ctx context.Context, id string) (ActivityCategory, bool, error) {
	categories, err := s.activityCategories(ctx)
	row, found := categories.Find(id)
	if err != nil || !found {
		return ActivityCategory{}, false, err
	}
	return activityCategoryFromRow(row), true, nil
}

func (s *PostgresStore) ListActivityCategories(ctx context.Context) ([]ActivityCategory, error) {
	rows, err := s.activityCategories(ctx)
	if err != nil {
		return nil, err
	}
	categories := make([]ActivityCategory, 0, len(rows.Rows))
	for _, row := range rows.Rows {
		categories = append(categories, activityCategoryFromRow(row))
	}
	return categories, nil
}

// PutActivityCategory saves the category and its activity types. The assignment's primary key is one row
// per type, so assigning a type here takes it out of the category it was in (047-C7).
func (s *PostgresStore) PutActivityCategory(ctx context.Context, c ActivityCategory) error {
	q := s.change(ctx)
	err := q.PutActivityCategory(ctx, sqlcdb.PutActivityCategoryParams{
		ID: c.ID, Name: c.Name, IsDefault: c.IsDefault, Active: c.Active, Version: c.Version,
		CreatedAt: c.CreatedAt, CreatedByID: c.CreatedBy.ID, CreatedByName: c.CreatedBy.Name,
		UpdatedAt: c.UpdatedAt, UpdatedByID: c.UpdatedBy.ID, UpdatedByName: c.UpdatedBy.Name,
	})
	if err == nil {
		err = q.ClearActivityCategoryTypes(ctx, c.ID)
	}
	for _, typeID := range c.ActivityTypeIDs {
		if err != nil {
			break
		}
		err = q.AssignActivityCategoryType(ctx, sqlcdb.AssignActivityCategoryTypeParams{ActivityTypeID: typeID, CategoryID: c.ID})
	}
	if err != nil {
		return fmt.Errorf("save activity category %s: %w", c.ID, err)
	}
	return nil
}

func (s *PostgresStore) DeleteActivityCategory(ctx context.Context, id string) error {
	if err := s.change(ctx).DeleteActivityCategory(ctx, id); err != nil {
		return fmt.Errorf("delete activity category %s: %w", id, err)
	}
	return nil
}

func activityCategoryFromRow(row sqlcdb.ListActivityCategoriesRow) ActivityCategory {
	c := row.ActivityCategory
	return ActivityCategory{
		ID: c.ID, Name: c.Name, IsDefault: c.IsDefault, Active: c.Active,
		ActivityTypeIDs: slices.Clone(row.ActivityTypeIds),
		Meta: record.Meta{
			Version:   c.Version,
			CreatedAt: c.CreatedAt,
			CreatedBy: actor.Actor{ID: c.CreatedByID, Name: c.CreatedByName},
			UpdatedAt: c.UpdatedAt,
			UpdatedBy: actor.Actor{ID: c.UpdatedByID, Name: c.UpdatedByName},
		},
	}
}
