package measure

import (
	"context"

	"mapp/internal/allwork"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
	Write(ctx context.Context, fn func(ctx context.Context) error) error
}

// Planning is All work's source (010): the year's plans in the person's plans.view scope.
type Planning interface {
	Years(ctx context.Context) ([]int, error)
	Year(ctx context.Context, year int, today date.Date, columns []string) (allwork.YearData, error)
	Choices(ctx context.Context, personID string) (allwork.Choices, error)
	Attributes(ctx context.Context) ([]allwork.Attribute, error)
	Currency(ctx context.Context, personID string) (allwork.Currency, error)
	RatesUsed(ctx context.Context) ([]string, error)
}

// Plans gives the IDs of the plans in the person's reports.view scope, read by the scoped plan store (R2).
type Plans interface {
	InReportsScope(ctx context.Context) ([]string, error)
}

type Currency struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Currencies interface {
	Active(ctx context.Context) ([]Currency, error)
	RateOf(ctx context.Context, code string) (money.Rate, error)
}

type Deps struct {
	Planning   Planning
	Plans      Plans
	Currencies Currencies
	Views      Store
}
