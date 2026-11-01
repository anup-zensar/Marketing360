package performance

import (
	"context"
	"time"
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
