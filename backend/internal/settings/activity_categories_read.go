package settings

import (
	"context"
	"slices"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
)

// usedByActivityType is the usage kind of the activity types a category groups (contract UsagePart.kind).
const usedByActivityType = "activityType"

// ListActivityCategories returns every activity category, active and inactive, by name ignoring letter
// case, each with the activity types it groups (FR-018, FR-022). Every signed-in person may see them.
func (s *Service) ListActivityCategories(ctx context.Context, _ actor.Actor) ([]ActivityCategoryView, error) {
	var views []ActivityCategoryView
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		categories, err := s.repo.ListActivityCategories(ctx)
		if err != nil {
			return err
		}
		slices.SortStableFunc(categories, func(x, y ActivityCategory) int { return compareNames(x.Name, y.Name) })
		views, err = s.activityCategoryViews(ctx, categories)
		return err
	})
	return views, err
}

// GetActivityCategory returns the activity category id with the types it groups (FR-022).
func (s *Service) GetActivityCategory(ctx context.Context, _ actor.Actor, id string) (ActivityCategoryView, error) {
	var view ActivityCategoryView
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		c, err := s.storedActivityCategory(ctx, id)
		if err != nil {
			return err
		}
		view, err = s.activityCategoryView(ctx, c)
		return err
	})
	return view, err
}

func (s *Service) storedActivityCategory(ctx context.Context, id string) (ActivityCategory, error) {
	c, found, err := s.repo.GetActivityCategory(ctx, id)
	if err != nil {
		return ActivityCategory{}, err
	}
	if !found {
		return ActivityCategory{}, errs.NotFound(activityCategoryGoneMessage)
	}
	return c, nil
}

func (s *Service) activityCategoryView(ctx context.Context, c ActivityCategory) (ActivityCategoryView, error) {
	views, err := s.activityCategoryViews(ctx, []ActivityCategory{c})
	if err != nil {
		return ActivityCategoryView{}, err
	}
	return views[0], nil
}

func (s *Service) activityCategoryViews(ctx context.Context, categories []ActivityCategory) ([]ActivityCategoryView, error) {
	types, err := s.repo.ListActivityTypes(ctx)
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(types, func(x, y ActivityType) int { return compareNames(x.Name, y.Name) })
	views := make([]ActivityCategoryView, 0, len(categories))
	for _, c := range categories {
		grouped := []SettingRef{}
		for _, t := range types {
			if slices.Contains(c.ActivityTypeIDs, t.ID) {
				grouped = append(grouped, SettingRef{ID: t.ID, Name: t.Name, Active: t.Active})
			}
		}
		views = append(views, ActivityCategoryView{
			ID: c.ID, Name: c.Name, IsDefault: c.IsDefault, ActivityTypes: grouped, Active: c.Active,
			Usage: categoryUsage(len(grouped)), Meta: c.Meta,
		})
	}
	return views, nil
}

// categoryUsage counts the activity types a category groups, so the page warns before it is deleted
// (FR-021). Nothing else refers to a category, and deleting one is never refused.
func categoryUsage(types int) Usage {
	if types == 0 {
		return Usage{Summary: notUsedYet, Parts: []UsagePart{}, Records: []history.RecordRef{}}
	}
	text := "Groups " + plural(types, "activity type", "activity types")
	return Usage{
		InUse: true, Summary: text, Records: []history.RecordRef{},
		Parts: []UsagePart{{Kind: usedByActivityType, Text: text, Count: types}},
	}
}
