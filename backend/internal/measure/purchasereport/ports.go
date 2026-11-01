package purchasereport

import (
	"context"
	"time"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

type Scoper interface {
	Scope(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (measure.Scope, error)
}

type Purchases interface {
	OnLineItems(ctx context.Context, lineItemIDs []string) (Requests, error)
}

// Requests are the purchase requests on the line items, and whether the organisation has switched purchase support on.
type Requests struct {
	Items      []Request
	SwitchedOn bool
}

// Request is a purchase request with every amount already in USD at today's rates, as feature 032 counts it.
type Request struct {
	ID             string
	Name           string
	Vendor         string
	Status         Status
	PlanID         string
	OwnerID        string
	Workflow       string
	GSAPNumber     string
	WaitingForGSAP bool
	CreatedAt      time.Time
	Lines          []Line
}

// Line is one line of a request; Invoiced is nil while the line has no purchase order line.
type Line struct {
	ID           string
	LineItemID   string
	Amount       money.Amount
	DeliveryDate date.Date
	SentOn       *date.Date
	PONumber     string
	POLine       string
	POLineStatus string
	Invoiced     *money.Amount
}

type Deps struct {
	Scope     Scoper
	Purchases Purchases
}
