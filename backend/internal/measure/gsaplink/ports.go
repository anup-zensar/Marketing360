package gsaplink

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

// Scoped is the frame's step (039 R1, R2): the year in the reports.view scope, matched, and the currency's rate.
type Scoped interface {
	Scope(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (measure.Scope, error)
}

type Rates interface {
	Rates(ctx context.Context) (money.Rates, error)
}

// Store reads what the year's nodes do not carry, only for the IDs it is given (research R2).
type Store interface {
	ActivitiesApprovedAt(ctx context.Context, activityIDs []string) (map[string]time.Time, error)
	WaitingForGsap(ctx context.Context, lineItemIDs []string) ([]Request, error)
	OrderTransactions(ctx context.Context, lineItemIDs []string) ([]Transaction, error)
}

type Request struct {
	ID           string
	Name         string
	Vendor       string
	Owner        string
	PlanID       string
	GsapPrNumber string
	Routing      string
	RoutedAt     time.Time
	Lines        []RequestLine
}

type RequestLine struct {
	LineItemID string
	Amount     money.Money
}

// Transaction is a purchase order line, invoice or payment linked to a line item and referring to a PO (015).
type Transaction struct {
	ID                string
	Type              string
	LineItemID        string
	PoNumber          string
	LineNumber        string
	PostingDate       date.Date
	Amount            money.Money
	Vendor            string
	PurchaseRequestID string
	GsapPrNumber      string
}

const (
	TypePurchaseOrder = "purchaseOrder"
	TypeInvoice       = "invoice"
	TypePayment       = "payment"
)
