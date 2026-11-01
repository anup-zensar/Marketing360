package approvalreporttest

import (
	"testing"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/approval/approvaltest"
	"mapp/internal/measure"
	"mapp/internal/measure/approvalreport"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/money"
)

const (
	PlanWithFinance    = "plan-1"
	PlanWithoutFinance = "plan-2"
	Year               = 2026
)

// The activities, one round each; names say the outcome the report must show.
const (
	ActivityWaiting      = "act-waiting"
	ActivityApprovedAny  = "act-approved-any"
	ActivityDeclinedAll  = "act-declined-all"
	ActivityRecalled     = "act-recalled"
	ActivityReplaced     = "act-replaced"
	ActivityCancelled    = "act-cancelled"
	ActivityAutomatic    = "act-automatic"
	ActivitySkippedLevel = "act-skipped-level"
	RequestWaiting       = "pr-waiting"
)

func Alice() actor.Actor { return actor.Actor{ID: "p-alice", Name: "Alice Ng"} }
func Bob() actor.Actor   { return actor.Actor{ID: "p-bob", Name: "Bob Ruiz"} }
func Carol() actor.Actor { return actor.Actor{ID: "p-carol", Name: "Carol Dev"} }
func Dan() actor.Actor   { return actor.Actor{ID: "p-dan", Name: "Dan Ito"} }
func Erin() actor.Actor  { return actor.Actor{ID: "p-erin", Name: "Erin Poe"} }

// Fixture is every port of the report on one in-memory database; Now is the fixed clock.
type Fixture struct {
	Deps  approvalreport.Deps
	DB    *fake.DB
	Scope measure.Scope
	Now   time.Time
}

// New builds the fixture: plan-1 shows its costs, plan-2 has none to show (no finance.view), and every
// round outcome, approver mode, automatic level and skipped level appears once.
func New(tb testing.TB) Fixture {
	tb.Helper()
	db, now := fake.NewDB(), time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC)
	rounds := approvaltest.NewStore(db)
	for _, r := range roundsAt(now) {
		if err := rounds.PutRound(tb.Context(), r); err != nil {
			tb.Fatalf("PutRound(%s) error: %v", r.ID, err)
		}
	}
	for _, a := range actionsAt(now) {
		if err := rounds.PutAction(tb.Context(), a); err != nil {
			tb.Fatalf("PutAction(%s) error: %v", a.ID, err)
		}
	}
	return Fixture{DB: db, Now: now, Scope: scope(), Deps: approvalreport.Deps{
		Scope:     Scoper{Result: scope()},
		Rounds:    rounds,
		Purchases: Purchases{Entries: []Entry{WaitingRequest()}},
	}}
}

func WaitingRequest() Entry {
	return Entry{
		PurchaseRequest: approvalreport.PurchaseRequest{ID: RequestWaiting, Name: "Event stand", PlanID: PlanWithFinance,
			OwnerID: Erin().ID, Amount: money.FromCents(250000), PurchaseOrders: []string{"4500001", "4500002"}},
		LineItemIDs: []string{"li-" + ActivityWaiting},
	}
}

func scope() measure.Scope {
	list := nodes()
	matched := allwork.Matched{Plans: []string{PlanWithFinance, PlanWithoutFinance}, Activities: []string{},
		LineItems: []string{}}
	for _, n := range list {
		switch n.Level {
		case allwork.LevelActivity:
			matched.Activities = append(matched.Activities, n.ID)
		case allwork.LevelLineItem:
			matched.LineItems = append(matched.LineItems, n.ID)
		case allwork.LevelPlan, allwork.LevelCampaign, allwork.LevelInitiative, allwork.LevelUnassigned:
		}
	}
	return measure.Scope{Year: Year, Nodes: list, Matched: matched,
		Choices:  allwork.Choices{SeesFinance: true, FinancePlans: []string{PlanWithFinance}, People: people()},
		Currency: allwork.Currency{Code: money.BaseCurrency, Rate: money.One}}
}

func nodes() []allwork.Node {
	cost := money.FromCents(120000)
	ids := []string{ActivityWaiting, ActivityApprovedAny, ActivityDeclinedAll, ActivityRecalled, ActivityReplaced,
		ActivityCancelled, ActivityAutomatic}
	list := make([]allwork.Node, 0, len(ids)+3)
	for _, id := range ids {
		list = append(list, activityNode(id, PlanWithFinance, &cost))
	}
	list = append(list, activityNode(ActivitySkippedLevel, PlanWithoutFinance, nil))
	for _, id := range []string{ActivityWaiting} {
		list = append(list, allwork.Node{Level: allwork.LevelLineItem, ID: "li-" + id, ParentID: id, PlanID: PlanWithFinance})
	}
	return list
}

func activityNode(id, planID string, cost *money.Amount) allwork.Node {
	return allwork.Node{Level: allwork.LevelActivity, ID: id, PlanID: planID, Name: "Activity " + id,
		Owner: &allwork.Owner{ID: Alice().ID, Name: Alice().Name}, PlannedCost: cost}
}

func people() []allwork.Option {
	list := []allwork.Option{}
	for _, p := range []actor.Actor{Alice(), Bob(), Carol(), Dan(), Erin()} {
		list = append(list, allwork.Option{ID: p.ID, Label: p.Name})
	}
	return list
}
