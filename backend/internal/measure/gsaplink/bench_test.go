package gsaplink

import (
	"context"
	"fmt"
	"testing"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// The volume of 028 and SC-001: 5,000 activities, 20,000 line items and requests' lines, 100,000 transactions.
const (
	benchPlans        = 50
	benchActivities   = 100
	benchItems        = 4
	benchTransactions = 5
)

// volume reads nothing: the year read is 028's and 036's to measure, so each tab's own work is what is timed (R6).
type volume struct {
	scope        measure.Scope
	approved     map[string]time.Time
	requests     []Request
	transactions []Transaction
}

func (v *volume) Scope(context.Context, actor.Actor, measure.Filter, []string) (measure.Scope, error) {
	return v.scope, nil
}

func (v *volume) ActivitiesApprovedAt(context.Context, []string) (map[string]time.Time, error) {
	return v.approved, nil
}

func (v *volume) WaitingForGsap(context.Context, []string) ([]Request, error) { return v.requests, nil }

func (v *volume) OrderTransactions(context.Context, []string) ([]Transaction, error) {
	return v.transactions, nil
}

func (v *volume) Rates(context.Context) (money.Rates, error) {
	return money.Rates{"EUR": money.RateFromMicros(500_000), money.BaseCurrency: money.One}, nil
}

func newVolume() *volume {
	v := &volume{approved: map[string]time.Time{}, scope: measure.Scope{Year: testYear,
		Choices: allwork.Choices{SeesFinance: true}, Currency: allwork.Currency{Code: "EUR", Rate: money.RateFromMicros(500_000)}}}
	for p := range benchPlans {
		plan, campaign, initiative := fmt.Sprintf("P-%d", p), fmt.Sprintf("C-%d", p), fmt.Sprintf("I-%d", p)
		v.node(allwork.LevelPlan, plan, "", plan)
		v.node(allwork.LevelCampaign, campaign, plan, plan)
		v.node(allwork.LevelInitiative, initiative, campaign, plan)
		for a := range benchActivities {
			v.activity(fmt.Sprintf("A-%d-%d", p, a), initiative, plan, p*benchActivities+a)
		}
	}
	return v
}

func (v *volume) node(level allwork.Level, id, parentID, planID string) {
	v.scope.Nodes = append(v.scope.Nodes, allwork.Node{Level: level, ID: id, ParentID: parentID, PlanID: planID, Name: id,
		Swbs: swbsReady})
	if level == allwork.LevelPlan {
		v.scope.Matched.Plans = append(v.scope.Matched.Plans, id)
	}
	if level == allwork.LevelLineItem {
		v.scope.Matched.LineItems = append(v.scope.Matched.LineItems, id)
	}
}

// activity adds an activity approved n hours ago, its line items, one purchase request on them and their transactions.
func (v *volume) activity(id, initiative, plan string, n int) {
	v.node(allwork.LevelActivity, id, initiative, plan)
	v.approved[id] = now.Add(-time.Duration(n) * time.Hour)
	request := Request{ID: fmt.Sprintf("PR-%d", n), Name: id, PlanID: plan,
		Routings: []Routing{{Type: RoutingOriginal, At: now.Add(-time.Duration(n) * time.Hour)}}}
	for l := range benchItems {
		item := fmt.Sprintf("L-%d-%d", n, l)
		v.node(allwork.LevelLineItem, item, id, plan)
		amount := money.Money{Amount: money.FromCents(int64(10_000 + n + l)), Currency: "EUR"}
		request.Lines = append(request.Lines, RequestLine{LineItemID: item, Amount: amount})
		kinds := []string{TypePurchaseOrder, TypePurchaseOrder, TypeInvoice, TypeInvoice, TypePayment}
		for t := range benchTransactions {
			v.transactions = append(v.transactions, Transaction{ID: fmt.Sprintf("FT-%d-%d-%d", n, l, t), Type: kinds[t],
				LineItemID: item, PoNumber: fmt.Sprintf("45%08d", n), LineNumber: fmt.Sprint(10 * (t + 1)),
				PostingDate: date.New(2026, 1, 1+t), Amount: amount})
		}
	}
	v.requests = append(v.requests, request)
}

// BenchmarkGsapLink_SC001 times each tab's first page, sorted, at the 028 volume; SC-001 leaves it under 300 ms.
func BenchmarkGsapLink_SC001(b *testing.B) {
	v := newVolume()
	s := NewService(fakeTx{}, Deps{Scoped: v, Store: v, Rates: v}, func() time.Time { return now })
	for _, tab := range []string{TabSwbs, TabWaiting, TabOrders} {
		b.Run(tab, func(b *testing.B) {
			for b.Loop() {
				if _, err := s.Report(b.Context(), actor.Actor{}, measure.Filter{}, Table{Tab: tab}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
