package performance

import (
	"context"
	"time"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

type Approvals interface {
	ApprovedIDs(ctx context.Context, activityIDs []string) (map[string]bool, error)
}

func (s *Service) UseApprovals(a Approvals) {
	s.approvals = a
}

// Feed stands in for a live connector (ADR-0020): one pull becomes a CSV that takes an upload's path.
type Feed interface {
	Pull(source Source, mappings []Mapping, today time.Time) (string, []byte, error)
}

func (s *Service) UseFeed(f Feed) {
	s.feed = f
}

type GoalFacts struct {
	INQ                  *int64
	MQL                  *int64
	Revenue              *money.Money
	SetLaterStagesMyself bool
	StartDate            *date.Date
	EndDate              *date.Date
	Ongoing              bool
	ModelID              string
	ActivityTypeID       string
	PlanYear             int
}

// Figures answers the goals as shown, the plan's model, type and year; unset, the panel has no goals.
type Figures interface {
	GoalFacts(ctx context.Context, activityID string) (GoalFacts, bool, error)
}

func (s *Service) UseFigures(f Figures) {
	s.figures = f
}

func (s *Service) UseRates(rates func(ctx context.Context) (money.Rates, error)) {
	s.rates = rates
}
