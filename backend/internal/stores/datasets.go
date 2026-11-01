package stores

import (
	"context"

	"mapp/internal/dataset"
)

// DatasetStore is the datasets store as the wiring and the seed use it (feature 051).
type DatasetStore interface {
	dataset.Store
	RemoveTeamShares(ctx context.Context, teamID string) error
}
