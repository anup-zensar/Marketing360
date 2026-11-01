package performance

import (
	"context"
	"time"

	"mapp/internal/platform/money"
)

// Approvals answers which activities count as approved (046-C1), a batch at a time; unset, it fails closed.
type Approvals interface {
	ApprovedIDs(ctx context.Context, activityIDs []string) (map[string]bool, error)
}

func (s *Service) UseApprovals(a Approvals) {
	s.approvals = a
}

// Feed is the simulated feed behind its port (ADR-0020): one pull becomes the CSV the source's mappings
// describe, so its rows take the identical path as an uploaded file; a live connector later implements this.
type Feed interface {
	Pull(source Source, mappings []Mapping, today time.Time) (string, []byte, error)
}

func (s *Service) UseFeed(f Feed) {
	s.feed = f
}

// GoalFacts is what the panel needs of the activity's own side: the goals as shown (typed or calculated by
// the model, domain model 4.5), the plan's model and its year.
type GoalFacts struct {
	INQ            *int64
	MQL            *int64
	Revenue        *money.Money
	ModelID        string
	ActivityTypeID string
	PlanYear       int
}

// Figures answers GoalFacts from the activity area; unset, the panel reads as having no goals.
type Figures interface {
	GoalFacts(ctx context.Context, activityID string) (GoalFacts, bool, error)
}

func (s *Service) UseFigures(f Figures) {
	s.figures = f
}

// UseRates gives the panel the current exchange rates, so a converted revenue total is computed here and
// never by the web app (constitution non-negotiable 4).
func (s *Service) UseRates(rates func(ctx context.Context) (money.Rates, error)) {
	s.rates = rates
}
