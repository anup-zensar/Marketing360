// Package storestest builds the API's store set in memory from every area's fake, for tests that run the
// whole API or the seed without a database.
package storestest

import (
	"context"
	"strings"

	"mapp/internal/performance"
	"mapp/internal/performance/performancetest"

	"mapp/internal/access/accesstest"
	"mapp/internal/activity/activitytest"
	"mapp/internal/allwork/allworktest"
	"mapp/internal/approval/approvaltest"
	"mapp/internal/auth/authtest"
	"mapp/internal/budget/budgettest"
	"mapp/internal/bulk/bulktest"
	"mapp/internal/campaign/campaigntest"
	"mapp/internal/financials/financialstest"
	"mapp/internal/history/historytest"
	"mapp/internal/lineitem/lineitemtest"
	"mapp/internal/measure/attribution/attributiontest"
	"mapp/internal/measure/measuretest"
	"mapp/internal/overspend/overspendtest"
	"mapp/internal/people/peopletest"
	"mapp/internal/plan/plantest"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/fake"
	"mapp/internal/purchaserequest/purchaserequesttest"
	"mapp/internal/sample/sampletest"
	"mapp/internal/settings/settingstest"
	"mapp/internal/stores"
	"mapp/internal/swbs/swbstest"
	"mapp/internal/workbook/workbooktest"
)

func New(db *fake.DB) stores.Set {
	counters := fake.NewCounters(db)
	return stores.Set{
		Tx: db, Counters: counters, Samples: sampletest.NewStore(db),
		Organisation: organisation(db), Planning: planning(db), Finance: finance(db),
		Ready:   func(context.Context) error { return nil },
		IsEmpty: func(context.Context) (bool, error) { return !counters.Used(), nil },
	}
}

func organisation(db *fake.DB) stores.Organisation {
	return stores.Organisation{
		People: peopletest.NewStore(db), History: historytest.NewStore(db), Auth: authtest.NewStore(db),
		Access: accesstest.NewStore(db), Settings: settingstest.NewStore(db),
	}
}

func planning(db *fake.DB) stores.Planning {
	return stores.Planning{
		Plans: plantest.NewStore(db), Campaigns: campaigntest.NewStore(db), Activities: activitytest.NewStore(db),
		LineItems: lineitemtest.NewStore(db), Approvals: approvaltest.NewStore(db), AllWork: allworktest.NewStore(db),
		Overspend: overspendtest.NewStore(db),
		Bulk:      bulktest.NewStore(db), Workbook: workbooktest.NewStore(db, plantest.NewStore(db)),
		Measure: measuretest.NewStore(db), Attribution: attributiontest.NewStore(db),
	}
}

func finance(db *fake.DB) stores.Finance {
	return stores.Finance{
		Budgets: budgettest.NewStore(db), Financials: financialstest.NewStore(db), SwbsRuns: swbstest.NewStore(db),
		Performance:      performanceStore(db),
		PurchaseRequests: purchaserequesttest.NewStore(db),
	}
}

// performanceStore answers 046's scoped activity read from the shared activity fake, as the SQL does
// through activity_in_scope().
func performanceStore(db *fake.DB) *performancetest.Store {
	store := performancetest.NewStore(db)
	activities := activitytest.NewStore(db)
	store.ActivityLookup = func(ctx context.Context, scope authz.Scope, id string) (performance.KeyFacts, bool, error) {
		found, err := activities.ListInScopeOf(ctx, scope, []string{id})
		if err != nil || len(found) == 0 {
			return performance.KeyFacts{}, false, err
		}
		a := found[0]
		return performance.KeyFacts{ExternalName: a.ExternalName,
			GoalEntered: a.MQL != nil || a.Revenue != nil}, true, nil
	}
	store.ResolveActivityIDs = func(ctx context.Context, scope authz.Scope, keys []string) (map[string][]string, error) {
		all, err := activities.ListInScope(ctx, scope)
		if err != nil {
			return nil, err
		}
		wanted := map[string]bool{}
		for _, key := range keys {
			wanted[key] = true
		}
		found := map[string][]string{}
		for _, a := range all {
			lowered := strings.ToLower(a.ID)
			if wanted[lowered] {
				found[lowered] = append(found[lowered], a.ID)
			}
		}
		return found, nil
	}
	return store
}
