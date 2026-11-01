package attribution_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"mapp/internal/history"
	"mapp/internal/history/historytest"
	"mapp/internal/measure/attribution"
	"mapp/internal/measure/attribution/attributiontest"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/apitest"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/ids"
)

var alex = actor.Actor{ID: "U-1001", Name: "Alex Morgan"}

func withPermissions(perms ...authz.Permission) context.Context {
	role := authz.GrantedRole{ID: "RL-1010", Name: "Custom", Scope: authz.ScopeAll}
	g := authz.NewGrants("U-1001", []authz.GrantedRole{role}, nil, map[string][]authz.Permission{"RL-1010": perms}, nil, authz.Rules{})
	return authz.With(context.Background(), g)
}

func settingsService(t *testing.T) (*attribution.Service, *history.Service) {
	t.Helper()
	db := fake.NewDB()
	generator := ids.NewGenerator(fake.NewCounters(db))
	now := func() time.Time { return time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC) }
	log := history.NewService(db, historytest.NewStore(db), generator, now)
	s := attribution.NewService(history.Tools{Tx: db, IDs: generator, Log: log, Now: now}, attributiontest.NewStore(db))
	for _, k := range s.Kinds() {
		if err := log.Register(k); err != nil {
			t.Fatal(err)
		}
	}
	return s, log
}

func codeOf(err error) string {
	var e *errs.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

func TestSettings_SeedAndWhoMayChange_FR001_FR002(t *testing.T) {
	t.Parallel()
	s, _ := settingsService(t)
	admin, planner := withPermissions(authz.SettingsManage), withPermissions(authz.PlansView)

	seed, err := s.Settings(planner)
	if err != nil || seed.Model != attribution.Equal || len(seed.Models) != 5 || seed.CanChange.Available {
		t.Fatalf("seed = %+v, %v: equally distributed, five models, a planner may not change it", seed, err)
	}
	_, err = s.SaveSettings(planner, alex, attribution.SettingsInput{Version: 1, Model: attribution.UShaped})
	if codeOf(err) != errs.CodeForbidden {
		t.Errorf("a planner's change = %v, want forbidden", err)
	}
	_, err = s.SaveSettings(admin, alex, attribution.SettingsInput{Version: 1, Model: "halfLife"})
	if codeOf(err) != errs.CodeValidationFailed {
		t.Errorf("an unknown model = %v, want a validation problem", err)
	}
}

func TestSettings_VersionedAndLogged_FR002(t *testing.T) {
	t.Parallel()
	s, log := settingsService(t)
	admin := withPermissions(authz.SettingsManage)

	saved, err := s.SaveSettings(admin, alex, attribution.SettingsInput{Version: 1, Model: attribution.UShaped})
	if err != nil || saved.Model != attribution.UShaped || saved.Version != 2 || saved.UpdatedBy == nil {
		t.Fatalf("change = %+v, %v: U-shaped at version 2, by Alex", saved, err)
	}
	_, err = s.SaveSettings(admin, alex, attribution.SettingsInput{Version: 1, Model: attribution.LastTouch})
	if codeOf(err) != errs.CodeVersionConflict {
		t.Errorf("a stale version = %v, want a version conflict", err)
	}
	entries, err := log.Entries(admin)
	if err != nil || len(entries) != 1 || entries[0].Changes[0].NewText != "U-shaped" {
		t.Errorf("log = %+v, %v: one entry naming the new model", entries, err)
	}
}

func TestAttributionContract(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	admin, planner := server.SignIn(t, "alex.morgan@shell.com"), server.SignIn(t, "priya.raman@shell.com")

	if r := planner.Do(t, http.MethodGet, "/api/v1/settings/attribution", nil); r.Status != http.StatusOK {
		t.Errorf("a planner reads the setting = %d %s", r.Status, r.Body)
	}
	change := map[string]any{"version": 1, "model": "lastTouch"}
	if r := planner.Do(t, http.MethodPut, "/api/v1/settings/attribution", change); r.Status != http.StatusForbidden {
		t.Errorf("a planner's change = %d, want 403", r.Status)
	}
	if r := admin.Do(t, http.MethodPut, "/api/v1/settings/attribution", change); r.Status != http.StatusOK {
		t.Errorf("the Administrator's change = %d %s", r.Status, r.Body)
	}
	if r := admin.Do(t, http.MethodPut, "/api/v1/settings/attribution", change); r.Status != http.StatusConflict {
		t.Errorf("the same version again = %d, want 409", r.Status)
	}
}
