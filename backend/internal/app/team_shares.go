package app

import (
	"context"

	"mapp/internal/platform/actor"
)

type teamShares interface {
	RemoveTeamShares(ctx context.Context, teamID string) error
}

// sharesLeaveWithTeam runs after a team is deleted, in the same Write: the dashboards and datasets shared with it
// are unshared from it, as the shares' foreign keys cascade in PostgreSQL (D-379).
func sharesLeaveWithTeam(next func(ctx context.Context, a actor.Actor, teamID string) error,
	shares ...teamShares) func(ctx context.Context, a actor.Actor, teamID string) error {
	return func(ctx context.Context, a actor.Actor, teamID string) error {
		if err := next(ctx, a, teamID); err != nil {
			return err
		}
		for _, s := range shares {
			if err := s.RemoveTeamShares(ctx, teamID); err != nil {
				return err
			}
		}
		return nil
	}
}
