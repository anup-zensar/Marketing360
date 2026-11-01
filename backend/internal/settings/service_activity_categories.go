package settings

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

const maxActivityCategoryName = 80

const (
	activityCategoryGoneMessage = "This activity category no longer exists — someone may have deleted it."
	activityCategoryNameTaken   = "An activity category called '%s' already exists."
	tickedTypeGoneMessage       = "An activity type you ticked no longer exists — reload the page and try again."
)

// CreateActivityCategory creates an active activity category (FR-018, FR-019, FR-020): a name of 1 to 80
// characters that no other category has ignoring letter case, and the activity types it groups — each type
// leaves the category it was in. Marking it the default unmarks the previous default in the same
// transaction. Every problem is reported at once.
func (s *Service) CreateActivityCategory(ctx context.Context, a actor.Actor, input ActivityCategoryInput) (ActivityCategoryView, error) {
	typeIDs := input.ActivityTypeIDs
	if typeIDs == nil {
		typeIDs = []string{}
	}
	all := ActivityCategoryUpdate{Name: &input.Name, IsDefault: &input.IsDefault, ActivityTypeIDs: &typeIDs}
	var view ActivityCategoryView
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		var problems errs.Problems
		c, err := s.applyActivityCategory(ctx, ActivityCategory{}, all, &problems)
		if err != nil {
			return err
		}
		if err = problems.Err(); err != nil {
			return err
		}
		if c.ID, err = s.ids.Next(ctx, ids.PrefixActivityCategory); err != nil {
			return err
		}
		c.Active, c.Meta = true, record.NewMeta(a, s.now())
		if err = s.saveActivityCategory(ctx, a, c); err != nil {
			return err
		}
		if err = s.logCreated(ctx, a, activityCategoryKind, c.ID, c.Name, ""); err != nil {
			return err
		}
		view, err = s.activityCategoryView(ctx, c)
		return err
	})
	return view, err
}

// UpdateActivityCategory changes the fields of the activity category id that update sends, from the version
// the person started at (FR-007, FR-018). activityTypeIds, when sent, is the complete new list of types:
// a type listed here leaves the category it was in (FR-019), and one left out becomes Uncategorised.
func (s *Service) UpdateActivityCategory(ctx context.Context, a actor.Actor, id string, update ActivityCategoryUpdate) (ActivityCategoryView, error) {
	var view ActivityCategoryView
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		stored, err := s.storedActivityCategory(ctx, id)
		if err != nil {
			return err
		}
		var problems errs.Problems
		checkVersion(&problems, update.Version)
		changed, err := s.applyActivityCategory(ctx, stored, update, &problems)
		if err != nil {
			return err
		}
		if err = problems.Err(); err != nil {
			return err
		}
		if err = stored.CheckVersion(*update.Version, activityCategoryNoun); err != nil {
			return err
		}
		changes, err := history.Diff(stored, changed)
		if err != nil {
			return err
		}
		if len(changes) > 0 {
			changed.Touch(a, s.now())
			if err = s.saveActivityCategory(ctx, a, changed); err != nil {
				return err
			}
			if err = s.logChanged(ctx, a, activityCategoryKind, changed.ID, changed.Name, changes, ""); err != nil {
				return err
			}
			stored = changed
		}
		view, err = s.activityCategoryView(ctx, stored)
		return err
	})
	return view, err
}

// DeleteActivityCategory moves the activity category id to the recycle bin, so that it can be undone
// (FR-021). Its types are released to Uncategorised by the database; the snapshot keeps them, so an Undo or
// a Restore assigns them back.
func (s *Service) DeleteActivityCategory(ctx context.Context, a actor.Actor, id string) (history.BinEntry, error) {
	var entry history.BinEntry
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		c, err := s.storedActivityCategory(ctx, id)
		if err != nil {
			return err
		}
		entry, err = s.history.MoveToBin(ctx, a, activityCategoryKind, c.ID)
		return err
	})
	return entry, err
}

func (s *Service) saveActivityCategory(ctx context.Context, a actor.Actor, c ActivityCategory) error {
	if c.IsDefault {
		if err := s.unmarkDefaultExcept(ctx, a, c.ID); err != nil {
			return err
		}
	}
	return s.repo.PutActivityCategory(ctx, c)
}

func (s *Service) unmarkDefaultExcept(ctx context.Context, a actor.Actor, exceptID string) error {
	categories, err := s.repo.ListActivityCategories(ctx)
	if err != nil {
		return err
	}
	for _, stored := range categories {
		if !stored.IsDefault || stored.ID == exceptID {
			continue
		}
		changed := stored
		changed.IsDefault = false
		changes, err := history.Diff(stored, changed)
		if err != nil {
			return err
		}
		changed.Touch(a, s.now())
		if err = s.repo.PutActivityCategory(ctx, changed); err != nil {
			return err
		}
		if err = s.logChanged(ctx, a, activityCategoryKind, changed.ID, changed.Name, changes, ""); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) applyActivityCategory(ctx context.Context, c ActivityCategory, update ActivityCategoryUpdate,
	problems *errs.Problems) (ActivityCategory, error) {
	if update.Name != nil {
		name := strings.TrimSpace(*update.Name)
		checkText(problems, "name", name, enterNameMessage, maxActivityCategoryName)
		if err := s.checkCategoryName(ctx, c.ID, name, problems); err != nil {
			return ActivityCategory{}, err
		}
		c.Name = name
	}
	if update.IsDefault != nil {
		c.IsDefault = *update.IsDefault
	}
	if update.Active != nil {
		c.Active = *update.Active
	}
	if update.ActivityTypeIDs == nil {
		return c, nil
	}
	typeIDs, err := s.checkCategoryTypes(ctx, *update.ActivityTypeIDs, problems)
	c.ActivityTypeIDs = typeIDs
	return c, err
}

func (s *Service) checkCategoryName(ctx context.Context, id, name string, problems *errs.Problems) error {
	if name == "" || utf8.RuneCountInString(name) > maxActivityCategoryName {
		return nil
	}
	categories, err := s.repo.ListActivityCategories(ctx)
	if err != nil {
		return err
	}
	other, taken := findNamed(categories, name, id, func(c ActivityCategory) (string, string) { return c.ID, c.Name })
	if taken {
		problems.Add("name", fmt.Sprintf(activityCategoryNameTaken, other.Name))
	}
	return nil
}

func (s *Service) checkCategoryTypes(ctx context.Context, typeIDs []string, problems *errs.Problems) ([]string, error) {
	types, err := s.repo.ListActivityTypes(ctx)
	if err != nil {
		return nil, err
	}
	wanted := unique(typeIDs)
	for _, id := range wanted {
		if !slices.ContainsFunc(types, func(t ActivityType) bool { return t.ID == id }) {
			problems.Add("activityTypeIds", tickedTypeGoneMessage)
			break
		}
	}
	return inTypeOrder(types, wanted), nil
}
