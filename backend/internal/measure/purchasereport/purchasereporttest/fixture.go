package purchasereporttest

import (
	"fmt"
	"testing"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/purchasereport"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/money"
)

const (
	PlanWithFinance    = "plan-1"
	PlanWithoutFinance = "plan-2"
	Year               = 2026
)

func Alice() actor.Actor { return actor.Actor{ID: "p-alice", Name: "Alice Ng"} }
func Bob() actor.Actor   { return actor.Actor{ID: "p-bob", Name: "Bob Ruiz"} }
func Carol() actor.Actor { return actor.Actor{ID: "p-carol", Name: "Carol Dev"} }

// Fixture is every port of the report; Scope is the hub's scope the fake answers with.
type Fixture struct {
	Deps  purchasereport.Deps
	DB    *fake.DB
	Scope measure.Scope
}

// New builds nine purchase requests of four vendors (one request has none): every status, a line without a
// delivery date, lines with and without a purchase order and invoices, and a request in a plan without finance.view.
func New(tb testing.TB) Fixture {
	tb.Helper()
	scope := scopeOf()
	return Fixture{DB: fake.NewDB(), Scope: scope, Deps: purchasereport.Deps{
		Scope: Scoper{Result: scope}, Purchases: Purchases{Requests: requests()},
	}}
}

func cents(n int64) money.Amount { return money.FromCents(n * 100) }

func day(s string) date.Date { return date.MustParse(s) }

func invoiced(n int64) *money.Amount {
	a := cents(n)
	return &a
}

type lineSpec struct {
	id, item string
	amount   int64
	delivery string
	po, line string
	invoiced *money.Amount
}

func line(s lineSpec) purchasereport.Line {
	l := purchasereport.Line{ID: s.id, LineItemID: s.item, Amount: cents(s.amount), PONumber: s.po, POLine: s.line,
		Invoiced: s.invoiced}
	if s.delivery != "" {
		l.DeliveryDate = day(s.delivery)
	}
	if s.po != "" {
		l.POLineStatus = "Open"
	}
	return l
}

type requestSpec struct {
	id, name, vendor string
	status           purchasereport.Status
	plan, owner      string
}

func request(s requestSpec, lines ...purchasereport.Line) purchasereport.Request {
	return purchasereport.Request{ID: s.id, Name: s.name, Vendor: s.vendor, Status: s.status, PlanID: s.plan, OwnerID: s.owner,
		Workflow: "Standard PR", CreatedAt: time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC), Lines: lines}
}

func requests() []purchasereport.Request {
	pr := func(id, name, vendor string, st purchasereport.Status, owner actor.Actor) requestSpec {
		return requestSpec{id, name, vendor, st, PlanWithFinance, owner.ID}
	}
	approvedPlan2 := requestSpec{"pr-9", "Plan two PR", "Acme", purchasereport.StatusApproved, PlanWithoutFinance, Alice().ID}
	waiting := request(pr("pr-3", "Stage", "Globex", purchasereport.StatusApproved, Alice()),
		line(lineSpec{id: "l4", item: "li-4", amount: 2000, delivery: "2026-04-15"}))
	waiting.WaitingForGSAP, waiting.GSAPNumber = true, "GSAP-3"
	return []purchasereport.Request{
		request(pr("pr-1", "Event stand", "Acme", purchasereport.StatusApproved, Alice()),
			line(lineSpec{id: "l1", item: "li-1", amount: 1000, delivery: "2026-03-10", po: "4500001", line: "10", invoiced: invoiced(400)}),
			line(lineSpec{id: "l2", item: "li-2", amount: 500, delivery: "2026-05-20", po: "4500001", line: "20", invoiced: invoiced(0)})),
		request(pr("pr-2", "Brochures", "Acme", purchasereport.StatusClosed, Bob()),
			line(lineSpec{id: "l3", item: "li-3", amount: 300, delivery: "2026-03-25", po: "4500002", line: "10", invoiced: invoiced(300)})),
		waiting,
		request(pr("pr-4", "Banners", "", purchasereport.StatusApproved, Carol()),
			line(lineSpec{id: "l5", item: "li-5", amount: 100})),
		request(pr("pr-5", "Draft thing", "Globex", purchasereport.StatusCreated, Bob()),
			line(lineSpec{id: "l6", item: "li-1", amount: 700, delivery: "2026-06-01"})),
		request(pr("pr-6", "Submitted one", "Initech", purchasereport.StatusSubmitted, Carol()),
			line(lineSpec{id: "l7", item: "li-2", amount: 50, delivery: "2026-03-02"})),
		request(pr("pr-7", "In review one", "Initech", purchasereport.StatusInReview, Carol()),
			line(lineSpec{id: "l8", item: "li-3", amount: 60, delivery: "2026-03-03"})),
		request(pr("pr-8", "Declined one", "Initech", purchasereport.StatusDeclined, Carol()),
			line(lineSpec{id: "l9", item: "li-4", amount: 70, delivery: "2026-03-04"})),
		request(approvedPlan2, line(lineSpec{id: "l10", item: "li-6", amount: 900, delivery: "2026-03-12"})),
	}
}

func scopeOf() measure.Scope {
	var nodes []allwork.Node
	matched := allwork.Matched{Plans: []string{PlanWithFinance, PlanWithoutFinance}, LineItems: []string{}}
	for i := 1; i <= 6; i++ {
		plan := PlanWithFinance
		if i == 6 {
			plan = PlanWithoutFinance
		}
		activity, item := fmt.Sprintf("act-%d", i), fmt.Sprintf("li-%d", i)
		nodes = append(nodes,
			allwork.Node{Level: allwork.LevelActivity, ID: activity, PlanID: plan, Name: fmt.Sprintf("Activity %d", i)},
			allwork.Node{Level: allwork.LevelLineItem, ID: item, ParentID: activity, PlanID: plan, Name: fmt.Sprintf("Line item %d", i)})
		matched.LineItems = append(matched.LineItems, item)
	}
	return measure.Scope{Year: Year, Nodes: nodes, Matched: matched,
		Choices:  allwork.Choices{SeesFinance: true, FinancePlans: []string{PlanWithFinance}, People: people()},
		Currency: allwork.Currency{Code: money.BaseCurrency, Rate: money.One}}
}

func people() []allwork.Option {
	list := []allwork.Option{}
	for _, p := range []actor.Actor{Alice(), Bob(), Carol()} {
		list = append(list, allwork.Option{ID: p.ID, Label: p.Name})
	}
	return list
}
