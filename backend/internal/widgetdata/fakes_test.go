package widgetdata_test

import (
	"context"
	"testing"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/money"
	"mapp/internal/widgetdata"
)

var viewer = actor.Actor{ID: "U-1009", Name: "Vera Viewer"}

func signedIn(t *testing.T, perms ...authz.Permission) context.Context {
	t.Helper()
	byRole := map[string][]authz.Permission{"RL-9": perms}
	roles := []authz.GrantedRole{{ID: "RL-9", Name: "Test", Scope: authz.ScopeAll}}
	return authz.With(t.Context(), authz.NewGrants(viewer.ID, roles, nil, byRole, nil, authz.Rules{}))
}

type fakeTx struct{}

func (fakeTx) Read(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }

type fakeReader struct {
	rows  widgetdata.Rows
	asked []widgetdata.Prefilter
	who   []string
}

func (f *fakeReader) Rows(_ context.Context, a actor.Actor, p widgetdata.Prefilter) (widgetdata.Rows, error) {
	f.asked = append(f.asked, p)
	f.who = append(f.who, a.ID)
	return f.rows, nil
}

type fakeDatasets struct {
	tables map[string]widgetdata.Table
}

func (f fakeDatasets) Visible(_ context.Context, _ actor.Actor) ([]widgetdata.DatasetInfo, error) {
	out := []widgetdata.DatasetInfo{}
	for _, t := range f.tables {
		out = append(out, t.Info)
	}
	return out, nil
}

func (f fakeDatasets) Table(_ context.Context, _ actor.Actor, id string) (widgetdata.Table, error) {
	t, found := f.tables[id]
	if !found {
		return widgetdata.Table{}, errs.NotFound("This dataset doesn't exist.")
	}
	return t, nil
}

type fakeRates money.Rates

func (f fakeRates) Rates(context.Context) (money.Rates, error) { return money.Rates(f), nil }

func text(s string) widgetdata.Value { return widgetdata.TextValue(s) }

func usd(t *testing.T, s string) widgetdata.Value {
	t.Helper()
	a, err := money.ParseAmount(s)
	if err != nil {
		t.Fatalf("ParseAmount(%q): %v", s, err)
	}
	return widgetdata.MoneyValue(&a)
}

func service(readers map[string]widgetdata.SourceReader, datasets widgetdata.DatasetTables) *widgetdata.Service {
	rates := fakeRates{"EUR": money.RateFromMicros(500_000)}
	return widgetdata.NewService(fakeTx{}, widgetdata.Deps{Sources: readers, Datasets: datasets, Rates: rates})
}
