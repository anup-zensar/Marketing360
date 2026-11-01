package measure_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
)

func withPermissions(t *testing.T, perms ...authz.Permission) context.Context {
	t.Helper()
	role := authz.GrantedRole{ID: "RL-9001", Name: "Reader", Scope: authz.ScopeAll}
	g := authz.NewGrants("U-1001", []authz.GrantedRole{role}, nil, map[string][]authz.Permission{"RL-9001": perms},
		[]actor.Actor{}, authz.Rules{})
	return authz.With(t.Context(), g)
}

type availableReport struct{ key string }

func (r availableReport) Key() string     { return r.key }
func (r availableReport) Available() bool { return true }
func (r availableReport) Sheets(context.Context, actor.Actor, measure.Filter) ([]measure.WidgetSheet, error) {
	return nil, nil
}

func stubRegistry(t *testing.T) *measure.Registry {
	t.Helper()
	registry := measure.NewRegistry()
	for _, key := range measure.Keys() {
		if err := registry.Register(measure.Stub(key)); err != nil {
			t.Fatalf("Register(%s) error = %v", key, err)
		}
	}
	return registry
}

// The nine reports in order, every stub coming soon, User access only with access.manage.
func TestCatalogue_NineInOrder_FR001(t *testing.T) {
	t.Parallel()
	got := stubRegistry(t).Catalogue(withPermissions(t, authz.PlansView, authz.ReportsView))
	var keys []string
	for _, entry := range got.Items {
		keys = append(keys, entry.Key)
		if entry.Available || entry.Name == "" || entry.Purpose == "" || entry.Path == "" {
			t.Errorf("entry %+v: want a stub with a name, a purpose and a path", entry)
		}
	}
	want := []string{"planBudget", "approvals", "purchaseRequests", "planDetails", "gsapLink", "summary", "revenue", "funnel"}
	if diff := cmp.Diff(want, keys); diff != "" {
		t.Errorf("catalogue without access.manage (-want +got):\n%s", diff)
	}
}

func TestCatalogue_UserAccessWithAccessManage_FR003(t *testing.T) {
	t.Parallel()
	got := stubRegistry(t).Catalogue(withPermissions(t, authz.ReportsView, authz.AccessManage))
	if len(got.Items) != 9 || got.Items[5].Key != "userAccess" || got.Items[5].Path != "user-access" {
		t.Errorf("catalogue = %+v, want nine with userAccess sixth", got.Items)
	}
}

func TestCatalogue_RegisteredReportAvailable_FR001(t *testing.T) {
	t.Parallel()
	registry := measure.NewRegistry()
	if err := registry.Register(availableReport{key: "planBudget"}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	got := registry.Catalogue(withPermissions(t, authz.ReportsView))
	if len(got.Items) != 8 || !got.Items[0].Available || got.Items[1].Available {
		t.Errorf("catalogue = %+v, want planBudget available and the unregistered ones not", got.Items)
	}
}

func TestRegistry_RefusesUnknownKeyAndSecondRegistration(t *testing.T) {
	t.Parallel()
	registry := measure.NewRegistry()
	if err := registry.Register(measure.Stub("someReport")); err == nil {
		t.Error("Register(someReport) error = nil, want an error")
	}
	if err := registry.Register(measure.Stub("approvals")); err != nil {
		t.Fatalf("Register(approvals) error = %v", err)
	}
	if err := registry.Register(measure.Stub("approvals")); err == nil {
		t.Error("Register(approvals) twice error = nil, want an error")
	}
}

func TestKeyOfPath(t *testing.T) {
	t.Parallel()
	if key, found := measure.KeyOfPath("purchase-requests"); !found || key != "purchaseRequests" {
		t.Errorf("KeyOfPath(purchase-requests) = %q, %v", key, found)
	}
	if _, found := measure.KeyOfPath("finance-budget"); found {
		t.Error("KeyOfPath(finance-budget) found a report, want none")
	}
}
