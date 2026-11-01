package widgetdata_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
	"mapp/internal/widgetdata"
)

var planningKeys = []string{"plans", "campaigns", "initiatives", "activities", "lineItems", "plannedByMonth", "approvals", "purchaseRequests"}

func leadsDataset() fakeDatasets {
	return fakeDatasets{tables: map[string]widgetdata.Table{"DS-1001": {Info: widgetdata.DatasetInfo{
		ID: "DS-1001", Name: "Event leads", Owner: authz.PersonRef{ID: "U-1001", Name: "Priya Raman"},
		Columns: []widgetdata.DatasetColumn{{Key: "c1", Label: "Region", Kind: widgetdata.KindText},
			{Key: "c2", Label: "Spend", Kind: widgetdata.KindMoney, Currency: "EUR"}},
	}}}}
}

func keysOf(sources []widgetdata.Source) []string {
	keys := []string{}
	for _, s := range sources {
		keys = append(keys, s.Key)
	}
	return keys
}

func TestCatalogue_SourcesByPermission_FR016(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		perms []authz.Permission
		want  []string
	}{
		{"a Viewer reads the planning sources and shared datasets", []authz.Permission{authz.DashboardsView, authz.PlansView},
			append(append([]string{}, planningKeys...), "DS-1001")},
		{"a Planner who sees financial data also reads transactions",
			[]authz.Permission{authz.DashboardsBuild, authz.PlansView, authz.FinanceView},
			[]string{"plans", "campaigns", "initiatives", "activities", "lineItems", "plannedByMonth", "financialTransactions",
				"approvals", "purchaseRequests", "DS-1001"}},
		{"global budgets need See global budgets", []authz.Permission{authz.DashboardsView, authz.GlobalBudgetsView},
			[]string{"globalBudgets", "DS-1001"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := service(nil, leadsDataset()).Catalogue(signedIn(t, tt.perms...), viewer)
			if err != nil {
				t.Fatalf("Catalogue: %v", err)
			}
			if diff := cmp.Diff(tt.want, keysOf(got)); diff != "" {
				t.Errorf("sources (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("without See or Build dashboards it is refused", func(t *testing.T) {
		t.Parallel()
		_, err := service(nil, leadsDataset()).Catalogue(signedIn(t, authz.PlansView), viewer)
		if apiErr, ok := errors.AsType[*errs.Error](err); !ok || apiErr.Status != http.StatusForbidden {
			t.Errorf("Catalogue error = %v, want 403", err)
		}
	})
}

func catalogueByKey(t *testing.T) map[string]widgetdata.Source {
	t.Helper()
	got, err := service(nil, leadsDataset()).Catalogue(signedIn(t, authz.DashboardsView, authz.PlansView, authz.FinanceView), viewer)
	if err != nil {
		t.Fatalf("Catalogue: %v", err)
	}
	byKey := map[string]widgetdata.Source{}
	for _, s := range got {
		byKey[s.Key] = s
	}
	return byKey
}

func fieldOf(t *testing.T, s widgetdata.Source, key string) widgetdata.Field {
	t.Helper()
	for _, f := range s.Fields {
		if f.Key == key {
			return f
		}
	}
	t.Fatalf("%s has no field %s", s.Key, key)
	return widgetdata.Field{}
}

func TestCatalogue_FieldsAndFilters_FR016(t *testing.T) {
	t.Parallel()
	byKey := catalogueByKey(t)
	activities := byKey["activities"]
	if activities.Filters != (widgetdata.SourceFilters{Date: "startDate", Plan: true, Team: true}) {
		t.Errorf("activities filters = %+v", activities.Filters)
	}
	field := func(s widgetdata.Source, key string) widgetdata.Field { return fieldOf(t, s, key) }
	if f := field(activities, "team"); !f.Multi || f.Kind != widgetdata.KindText {
		t.Errorf("activities team = %+v, want a multi text field", f)
	}
	if f := field(activities, "committed"); !f.Finance || f.Kind != widgetdata.KindMoney {
		t.Errorf("activities committed = %+v, want a finance money field", f)
	}
	if got := widgetdata.AggregationsOf(field(activities, "startDate")); !cmp.Equal(got, []string{"count", "countDistinct", "min", "max"}) {
		t.Errorf("date aggregations = %v", got)
	}
	if widgetdata.IsGroupable(field(activities, "plannedCost")) || !widgetdata.IsGroupable(field(activities, "year")) {
		t.Error("money is not groupable; numbers are")
	}
}

func TestCatalogue_DatasetColumnsAsFields_FR016(t *testing.T) {
	t.Parallel()
	leads := catalogueByKey(t)["DS-1001"]
	if leads.Kind != widgetdata.SourceDataset || leads.Owner == nil || leads.Filters != (widgetdata.SourceFilters{}) {
		t.Errorf("dataset source = %+v, want a dataset with its owner and no filters", leads)
	}
	if f := fieldOf(t, leads, "c2"); f.Label != "Spend" || f.Kind != widgetdata.KindMoney {
		t.Errorf("dataset column = %+v", f)
	}
}
