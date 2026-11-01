package settings

import (
	"context"
	"encoding/json"
	"slices"

	"mapp/internal/history"
)

// defaultReleasedNote says why a category comes back without being the default (FR-020).
const defaultReleasedNote = "No longer the default because another category is now the default."

// The fields of an activity category and their labels in the activity log (data model §12).
const (
	fieldName            = "name"
	fieldIsDefault       = "isDefault"
	fieldActivityTypeIDs = "activityTypeIds"
	labelName            = "Name"
	labelStatus          = "Status"
	textYes              = "Yes"
)

func (s *Service) describeActivityCategory(ctx context.Context, field string, value any) (string, string, error) {
	switch field {
	case fieldName:
		return labelName, "", nil
	case fieldIsDefault:
		return "Default category", yesNoText(value), nil
	case fieldActivityTypeIDs:
		text, err := s.namesOf(ctx, value, func(ctx context.Context, id string) (string, bool, error) {
			t, found, err := s.repo.GetActivityType(ctx, id)
			return t.Name, found, err
		})
		return "Activity types", text, err
	case statusActive:
		return labelStatus, statusText(value), nil
	}
	return "", "", nil
}

func yesNoText(value any) string {
	if value == nil {
		return ""
	}
	if yes, ok := value.(bool); ok && yes {
		return textYes
	}
	return "No"
}

// prepareActivityCategoryRestore applies the name rule before a category comes back, and lets the category
// that is the default now keep it, so at most one is ever the default (FR-020).
func (s *Service) prepareActivityCategoryRestore(ctx context.Context, rc history.RestoreContext, snap *history.Snapshot) ([]string, error) {
	var c ActivityCategory
	if err := json.Unmarshal(snap.Data, &c); err != nil {
		return nil, err
	}
	categories, err := s.repo.ListActivityCategories(ctx)
	if err != nil {
		return nil, err
	}
	name, notes, err := restoreName(rc, categoryNoun, c.Name, func(name string) (string, bool) {
		other, taken := findNamed(categories, name, c.ID, func(x ActivityCategory) (string, string) { return x.ID, x.Name })
		return other.Name, taken
	})
	if err != nil {
		return nil, err
	}
	changed := name != c.Name
	c.Name = name
	if c.IsDefault && slices.ContainsFunc(categories, func(x ActivityCategory) bool { return x.IsDefault && x.ID != c.ID }) {
		notes = append(notes, defaultReleasedNote)
		c.IsDefault, changed = false, true
	}
	if !changed {
		return nil, nil
	}
	c.Touch(rc.Actor, rc.Now)
	return notes, resnapshot(snap, c, c.Name)
}
