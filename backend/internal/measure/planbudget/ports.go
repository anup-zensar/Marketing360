package planbudget

import (
	"context"

	"mapp/internal/financials"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

type Scoper interface {
	Scope(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (measure.Scope, error)
}

// Spending gives payments and purchase-order lines by posting month, as 036's dashboard reads them (R2).
type Spending interface {
	SpendByMonth(ctx context.Context, a actor.Actor, lineItemIDs []string) (map[string]map[string]financials.Facts, error)
}
