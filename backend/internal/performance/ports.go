package performance

import "context"

// Approvals answers which activities count as approved (046-C1), a batch at a time; unset, it fails closed.
type Approvals interface {
	ApprovedIDs(ctx context.Context, activityIDs []string) (map[string]bool, error)
}

func (s *Service) UseApprovals(a Approvals) {
	s.approvals = a
}
