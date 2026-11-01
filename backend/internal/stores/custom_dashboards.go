package stores

import (
	"context"

	"mapp/internal/customdashboard"
)

// CustomDashboardStore is the custom dashboards store as the wiring and the seed use it.
type CustomDashboardStore interface {
	customdashboard.Store
	OwnerOf(ctx context.Context, id string) (ownerID string, found bool, err error)
	RemoveTeamShares(ctx context.Context, teamID string) error
}
