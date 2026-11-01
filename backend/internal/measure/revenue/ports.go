// Package revenue serves feature 049's two reports: Revenue and Funnel stages (research R1).
package revenue

import (
	"context"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

type Scoper interface {
	Scope(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (measure.Scope, error)
}

type Facts interface {
	ReportTotals(ctx context.Context, scope authz.Scope, activityIDs []string, year int) ([]FactTotal, error)
}

type FactTotal struct {
	ActivityID  string
	Stage       string
	WebMetric   string
	Month       string // "2006-01"
	Currency    string
	Counts      int64
	AmountMinor int64
	FactCount   int
}

type ActivityGoals struct {
	SetLaterStagesMyself bool
	Model                projection.Model
}

// Goals answers each activity's plan model and "set later stages myself" flag (research R3).
type Goals interface {
	ActivityGoals(ctx context.Context, scope authz.Scope, nodes []allwork.Node) (map[string]ActivityGoals, error)
}

type Category struct {
	ID              string
	Name            string
	IsDefault       bool
	ActivityTypeIDs []string
}

type Categories interface {
	ListCategories(ctx context.Context) ([]Category, error)
}

// Rates answers the current exchange rates, as 046's revenue figure uses them (046-C7).
type Rates interface {
	Rates(ctx context.Context) (money.Rates, error)
}
