package approvalreport

import (
	"context"

	"mapp/internal/approval"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/money"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

type Scoper interface {
	Scope(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (measure.Scope, error)
}

type Rounds interface {
	RoundsOfSubjects(ctx context.Context, subjectIDs []string) ([]approval.Round, error)
	ActionsOfSubjects(ctx context.Context, subjectIDs []string) ([]approval.Action, error)
}

type Purchases interface {
	OnLineItems(ctx context.Context, lineItemIDs []string) ([]PurchaseRequest, error)
}

// PurchaseRequest is what the grid shows of one: Amount in USD as feature 032 counts it, one purchase order
// number per line that has one, in line order.
type PurchaseRequest struct {
	ID             string
	Name           string
	PlanID         string
	OwnerID        string
	CanOpen        bool
	Amount         money.Amount
	PurchaseOrders []string
}

type Deps struct {
	Scope     Scoper
	Rounds    Rounds
	Purchases Purchases
}
