package settings_test

import (
	"net/http"
	"testing"

	"mapp/internal/platform/apitest"
)

const (
	categoryGoneBody = `{"error":{"code":"not_found","message":"This activity category no longer exists — someone may have deleted it.","fields":[]}}`
	groupsOneType    = `{"inUse":true,"summary":"Groups 1 activity type","parts":` +
		`[{"kind":"activityType","text":"Groups 1 activity type","count":1,"inRecycleBin":0}],"records":[]}`
	eventRef = `{"id":"AT-1002","name":"Event","active":true}`
)

// digitalBody is a new activity category grouping the activity types.
func digitalBody(isDefault bool, typeIDs ...string) map[string]any {
	if typeIDs == nil {
		typeIDs = []string{}
	}
	return map[string]any{"name": "Digital", "isDefault": isDefault, "activityTypeIds": typeIDs}
}

// digitalAnswer is the answer for the category digitalBody (CAT-1001) at version 1, with its activity
// types (JSON objects, comma separated), its default flag and its usage, created by Alex.
func digitalAnswer(types, isDefault, usage string) string {
	return `{"id":"CAT-1001","name":"Digital","isDefault":` + isDefault + `,"activityTypes":[` + types +
		`],"active":true,"usage":` + usage + `,"version":1,` + alexAtStart + `}`
}

// The categories are listed and read with the activity types they group; a new one answers 201 and takes
// the types out of the category they were in; every problem of a new one comes at once (FR-018, FR-019).
func TestActivityCategoriesAPI_ListCreateGet_FR018(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	alex := server.SignIn(t, alexsEmail)

	empty := alex.Do(t, http.MethodGet, "/api/v1/settings/activity-categories", nil)
	created := mustCreate(t, alex, "/api/v1/settings/activity-categories", digitalBody(true, "AT-1002"))
	got := alex.Do(t, http.MethodGet, "/api/v1/settings/activity-categories/CAT-1001", nil)
	gone := alex.Do(t, http.MethodGet, "/api/v1/settings/activity-categories/CAT-9999", nil)
	invalid := alex.Do(t, http.MethodPost, "/api/v1/settings/activity-categories", map[string]any{
		"name": " digital ", "isDefault": false, "activityTypeIds": []string{"AT-9999"},
	})
	moved := mustCreate(t, alex, "/api/v1/settings/activity-categories", map[string]any{
		"name": "Offline", "isDefault": false, "activityTypeIds": []string{"AT-1002"},
	})
	after := alex.Do(t, http.MethodGet, "/api/v1/settings/activity-categories/CAT-1001", nil)

	assertAnswer(t, empty, http.StatusOK, `{"items":[],"total":0}`)
	assertAnswer(t, created, http.StatusCreated, digitalAnswer(eventRef, "true", groupsOneType))
	assertAnswer(t, got, http.StatusOK, digitalAnswer(eventRef, "true", groupsOneType))
	assertAnswer(t, gone, http.StatusNotFound, categoryGoneBody)
	assertAnswer(t, invalid, http.StatusBadRequest, validationBody(
		`{"field":"name","message":"An activity category called 'Digital' already exists."},`+
			`{"field":"activityTypeIds","message":"An activity type you ticked no longer exists — reload the page and try again."}`))
	if moved.Status != http.StatusCreated {
		t.Fatalf("creating Offline = %d %s", moved.Status, moved.Body)
	}
	assertAnswer(t, after, http.StatusOK, digitalAnswer("", "true", unusedBody))
}

// A category is deleted to the recycle bin, which releases its activity types, and Undo brings it back with
// them (FR-021).
func TestActivityCategoriesAPI_DeleteAndUndo_FR021(t *testing.T) {
	t.Parallel()
	server := apitest.New(t, apitest.WithoutDemoHistory())
	alex := server.SignIn(t, alexsEmail)
	mustCreate(t, alex, "/api/v1/settings/activity-categories", digitalBody(false, "AT-1002"))

	deleted := alex.Do(t, http.MethodDelete, "/api/v1/settings/activity-categories/CAT-1001", nil)
	afterDelete := alex.Do(t, http.MethodGet, "/api/v1/settings/activity-categories/CAT-1001", nil)
	undone := alex.Do(t, http.MethodPost, "/api/v1/recycle-bin/BIN-1001/undo", nil)

	assertAnswer(t, deleted, http.StatusOK, `{"recycleBinEntry":{"id":"BIN-1001","recordType":"activityCategory",`+
		`"recordTypeLabel":"activity category","recordId":"CAT-1001","recordName":"Digital","path":[],"parent":null,`+
		`"deletedAt":"2026-03-12T09:00:00Z","deletedBy":{"id":"U-1001","name":"Alex Morgan"},"contents":[]}}`)
	assertAnswer(t, afterDelete, http.StatusNotFound, categoryGoneBody)
	assertAnswer(t, undone, http.StatusOK, `{"recordType":"activityCategory","recordTypeLabel":"activity category",`+
		`"recordId":"CAT-1001","recordName":"Digital","details":""}`)
	assertAnswer(t, alex.Do(t, http.MethodGet, "/api/v1/settings/activity-categories/CAT-1001", nil),
		http.StatusOK, digitalAnswer(eventRef, "false", groupsOneType))
}

// Reading the categories needs only a sign-in; changing them needs settings.manage (FR-023).
func TestActivityCategoriesAPI_ChangedWithSettingsManage_FR023(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	mustCreate(t, server.SignIn(t, alexsEmail), "/api/v1/settings/activity-categories", digitalBody(false))
	daniel := server.SignIn(t, danielsEmail)

	read := daniel.Do(t, http.MethodGet, "/api/v1/settings/activity-categories/CAT-1001", nil)
	refused := daniel.Do(t, http.MethodPatch, "/api/v1/settings/activity-categories/CAT-1001",
		map[string]any{"name": "Online", "version": 1})
	refusedDelete := daniel.Do(t, http.MethodDelete, "/api/v1/settings/activity-categories/CAT-1001", nil)
	server.GrantRole(t, danielsEmail, "administrator", "all")
	allowed := daniel.Do(t, http.MethodPatch, "/api/v1/settings/activity-categories/CAT-1001",
		map[string]any{"name": "Online", "version": 1})

	if read.Status != http.StatusOK {
		t.Errorf("Daniel reading the category = %d %s, want 200", read.Status, read.Body)
	}
	assertAnswer(t, refused, http.StatusForbidden, `{"error":{"code":"forbidden","message":"Changing Settings isn't part of your roles `+
		`(Approver, Planner). Ask an administrator if you need it.","fields":[]}}`)
	if refusedDelete.Status != http.StatusForbidden {
		t.Errorf("Daniel deleting the category = %d %s, want 403", refusedDelete.Status, refusedDelete.Body)
	}
	if allowed.Status != http.StatusOK {
		t.Errorf("Daniel with settings.manage = %d %s, want 200", allowed.Status, allowed.Body)
	}
}
