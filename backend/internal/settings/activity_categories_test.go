package settings

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"mapp/internal/history"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/record"
)

// categoryInput is a new activity category with the activity types it groups.
func categoryInput(name string, isDefault bool, typeIDs ...string) ActivityCategoryInput {
	return ActivityCategoryInput{Name: name, IsDefault: isDefault, ActivityTypeIDs: typeIDs}
}

func (f *fixture) createCategory(t *testing.T, input ActivityCategoryInput) ActivityCategoryView {
	t.Helper()
	view, err := f.service.CreateActivityCategory(t.Context(), alex, input)
	if err != nil {
		t.Fatalf("CreateActivityCategory(%q): %v", input.Name, err)
	}
	return view
}

func (f *fixture) categoryNamed(t *testing.T, name string) ActivityCategoryView {
	t.Helper()
	views, err := f.service.ListActivityCategories(t.Context(), alex)
	if err != nil {
		t.Fatalf("ListActivityCategories: %v", err)
	}
	for _, view := range views {
		if view.Name == name {
			return view
		}
	}
	t.Fatalf("no activity category called %q in %+v", name, views)
	return ActivityCategoryView{}
}

// typeNames lists the activity types of a category, by name.
func typeNames(view ActivityCategoryView) []string {
	names := []string{}
	for _, ref := range view.ActivityTypes {
		names = append(names, ref.Name)
	}
	return names
}

// A new activity category is active, keeps its audit trail and its activity types, and is written to the
// activity log (FR-018).
func TestCreateActivityCategory_AuditAndTypes_FR018(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	event := f.addType(t, "Event", true)
	f.clock.now = start.Add(time.Hour)

	created := f.createCategory(t, categoryInput(" Digital ", false, event.ID, event.ID))

	want := ActivityCategoryView{
		ID: "CAT-1001", Name: "Digital", ActivityTypes: []SettingRef{{ID: event.ID, Name: "Event", Active: true}},
		Active: true,
		Usage: Usage{InUse: true, Summary: "Groups 1 activity type", Records: []history.RecordRef{},
			Parts: []UsagePart{{Kind: "activityType", Text: "Groups 1 activity type", Count: 1}}},
		Meta: record.NewMeta(alex, f.clock.now),
	}
	if !reflect.DeepEqual(created, want) {
		t.Errorf("created = %+v\nwant      %+v", created, want)
	}
	entry := f.lastEntry(t)
	if entry.Action != history.ActionCreated || entry.RecordType != "activityCategory" || entry.RecordName != "Digital" {
		t.Errorf("log entry = %+v", entry)
	}
	empty := f.createCategory(t, categoryInput("Offline", false))
	if empty.Usage.InUse || empty.Usage.Summary != notUsedYet || len(empty.ActivityTypes) != 0 {
		t.Errorf("a category without types = %+v", empty)
	}
}

// Every problem of an activity category is reported at once: the name (1 to 80 characters, unique ignoring
// letter case and spaces) and the activity types ticked. Nothing is saved (FR-018).
func TestActivityCategory_Validation_FR018(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	f.createCategory(t, categoryInput("Digital", false))

	_, blank := f.service.CreateActivityCategory(t.Context(), alex, categoryInput(" ", false, "AT-9999"))
	_, taken := f.service.CreateActivityCategory(t.Context(), alex, categoryInput("  DIGITAL ", false))
	_, long := f.service.CreateActivityCategory(t.Context(), alex, categoryInput(strings.Repeat("é", 81), false))
	limit := f.createCategory(t, categoryInput(strings.Repeat("é", 80), false))

	assertError(t, blank, problems("name", "Enter a name.",
		"activityTypeIds", "An activity type you ticked no longer exists — reload the page and try again."))
	assertError(t, taken, problems("name", "An activity category called 'Digital' already exists."))
	assertError(t, long, problems("name", "Use 80 characters or fewer."))
	if limit.ID != "CAT-1002" {
		t.Errorf("a name of 80 characters = %+v", limit)
	}
	if listed, err := f.service.ListActivityCategories(t.Context(), alex); err != nil || len(listed) != 2 {
		t.Errorf("ListActivityCategories = %+v, %v, want only the two saved categories", listed, err)
	}
}

// A save from an out-of-date copy of an activity category is refused with version_conflict (FR-018).
func TestUpdateActivityCategory_VersionConflict_FR018(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	digital := f.createCategory(t, categoryInput("Digital", false))
	f.clock.now = start.Add(time.Hour)
	if _, err := f.service.UpdateActivityCategory(t.Context(), daniel, digital.ID,
		ActivityCategoryUpdate{Name: text("Digital media"), Version: version(1)}); err != nil {
		t.Fatalf("UpdateActivityCategory: %v", err)
	}

	_, stale := f.service.UpdateActivityCategory(t.Context(), alex, digital.ID,
		ActivityCategoryUpdate{Name: text("Online"), Version: version(1)})
	_, gone := f.service.UpdateActivityCategory(t.Context(), alex, "CAT-9999",
		ActivityCategoryUpdate{Name: text("Online"), Version: version(1)})

	var conflict *errs.Error
	if !errors.As(stale, &conflict) || conflict.Code != errs.CodeVersionConflict ||
		conflict.Message != "Not saved — Daniel Okoro changed this activity category" {
		t.Errorf("stale save = %#v", stale)
	}
	assertError(t, gone, errs.NotFound(activityCategoryGoneMessage))
	if after := f.categoryNamed(t, "Digital media"); after.Version != 2 || after.UpdatedBy != daniel {
		t.Errorf("the category after the change = %+v", after)
	}
}

// An activity type belongs to one category at a time: putting it in another moves it out of the first
// (FR-019); a type in no category stays Uncategorised.
func TestActivityCategory_TypeMovesBetweenCategories_FR019(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	event := f.addType(t, "Event", true)
	webinar := f.addType(t, "Webinar", true)
	digital := f.createCategory(t, categoryInput("Digital", false, event.ID, webinar.ID))

	offline := f.createCategory(t, categoryInput("Offline", false, event.ID))

	if !reflect.DeepEqual(typeNames(offline), []string{"Event"}) {
		t.Errorf("Offline groups %q, want Event", typeNames(offline))
	}
	if left := f.categoryNamed(t, "Digital"); !reflect.DeepEqual(typeNames(left), []string{"Webinar"}) {
		t.Errorf("Digital groups %q after Event moved, want only Webinar", typeNames(left))
	}
	released, err := f.service.UpdateActivityCategory(t.Context(), alex, digital.ID,
		ActivityCategoryUpdate{ActivityTypeIDs: &[]string{}, Version: version(1)})
	if err != nil || len(released.ActivityTypes) != 0 {
		t.Errorf("Digital emptied = %+v, %v", released, err)
	}
	if still := f.categoryNamed(t, "Offline"); !reflect.DeepEqual(typeNames(still), []string{"Event"}) {
		t.Errorf("Offline groups %q, want it untouched", typeNames(still))
	}
}

// At most one activity category is the default: marking a new one unmarks the one that had it, in the same
// transaction (FR-020).
func TestActivityCategory_OneDefault_FR020(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	digital := f.createCategory(t, categoryInput("Digital", true))
	offline := f.createCategory(t, categoryInput("Offline", true))

	if !offline.IsDefault {
		t.Errorf("Offline = %+v, want it to be the default", offline)
	}
	unmarked := f.categoryNamed(t, "Digital")
	if unmarked.IsDefault || unmarked.Version != 2 {
		t.Errorf("Digital = %+v, want it unmarked at version 2", unmarked)
	}
	back, err := f.service.UpdateActivityCategory(t.Context(), alex, digital.ID,
		ActivityCategoryUpdate{IsDefault: yes(true), Version: version(2)})
	if err != nil || !back.IsDefault {
		t.Errorf("Digital marked default again = %+v, %v", back, err)
	}
	if other := f.categoryNamed(t, "Offline"); other.IsDefault {
		t.Errorf("Offline = %+v, want it unmarked", other)
	}
	if entry := f.lastEntry(t); entry.RecordName != "Digital" {
		t.Errorf("last log entry = %+v", entry)
	}
}

// Deleting an activity category moves it to the recycle bin and releases its activity types; Undo brings it
// back with the same types (FR-021).
func TestDeleteActivityCategory_UndoRestoresTypes_FR021(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	event := f.addType(t, "Event", true)
	digital := f.createCategory(t, categoryInput("Digital", true, event.ID))
	if !digital.Usage.InUse || digital.Usage.Summary != "Groups 1 activity type" {
		t.Errorf("Digital usage = %+v, want the warning before it is deleted", digital.Usage)
	}

	entry, err := f.service.DeleteActivityCategory(t.Context(), alex, digital.ID)
	if err != nil {
		t.Fatalf("DeleteActivityCategory: %v", err)
	}
	_, missing := f.service.GetActivityCategory(t.Context(), alex, digital.ID)
	assertError(t, missing, errs.NotFound(activityCategoryGoneMessage))

	if _, err = f.history.Undo(t.Context(), alex, entry.ID); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	back, err := f.service.GetActivityCategory(t.Context(), alex, digital.ID)
	if err != nil || !reflect.DeepEqual(typeNames(back), []string{"Event"}) || !back.IsDefault {
		t.Errorf("Digital after Undo = %+v, %v, want Event back and still the default", back, err)
	}
	_, deletedTwice := f.service.DeleteActivityCategory(t.Context(), alex, "CAT-9999")
	assertError(t, deletedTwice, errs.NotFound(activityCategoryGoneMessage))
}
