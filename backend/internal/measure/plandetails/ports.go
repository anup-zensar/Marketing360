package plandetails

import (
	"context"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

// Scoped is the frame's step (039 R1, R2): the year in the reports.view scope, matched, and the currency's rate.
type Scoped interface {
	Matched(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (Matching, error)
	Attributes(ctx context.Context) ([]allwork.Attribute, error)
}

// Matching holds the year's nodes in scope; Matched names the records of every level that match the filter.
type Matching struct {
	Year          int
	Nodes         []allwork.Node
	Matched       map[string]bool
	ActivityTypes []allwork.Option
	Currency      allwork.Currency
}
