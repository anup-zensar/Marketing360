// Package widgetdata answers a dashboard widget's data from built-in sources and datasets, read within the person's scope.
package widgetdata

import (
	"context"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

type Prefilter struct {
	From, To         *date.Date
	PlanIDs, TeamIDs []string
}

type Rows struct {
	Items []Row
	Notes []string
}

type SourceReader interface {
	Rows(ctx context.Context, a actor.Actor, f Prefilter) (Rows, error)
}

// DatasetTables' Table answers errs.NotFound for a dataset the actor may not see or that is deleted.
type DatasetTables interface {
	Visible(ctx context.Context, a actor.Actor) ([]DatasetInfo, error)
	Table(ctx context.Context, a actor.Actor, id string) (Table, error)
}

type Rates interface {
	Rates(ctx context.Context) (money.Rates, error)
}

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

type Deps struct {
	Sources  map[string]SourceReader
	Datasets DatasetTables
	Rates    Rates
}
