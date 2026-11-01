package customdashboard

import (
	"context"

	"mapp/internal/people"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
	Write(ctx context.Context, fn func(ctx context.Context) error) error
}

// Store reads dashboards with their widgets and shares. A visible dashboard is one personID owns or that is
// shared with a team they belong to now; Get reads any, for history only.
type Store interface {
	ListVisible(ctx context.Context, personID string) ([]Dashboard, error)
	GetVisible(ctx context.Context, personID, id string) (Dashboard, bool, error)
	Get(ctx context.Context, id string) (Dashboard, bool, error)
	Insert(ctx context.Context, d Dashboard) error
	Update(ctx context.Context, d Dashboard, previousVersion int) error
	Delete(ctx context.Context, id string) error
	ReplaceWidgets(ctx context.Context, id string, widgets []Widget) error
	ReplaceShares(ctx context.Context, id string, teamIDs []string) error
	UsingDataset(ctx context.Context, personID, datasetID string) ([]Ref, error)
	NameTaken(ctx context.Context, ownerID, name, exceptID string) (bool, error)
}

// Directory names people and teams, and says which teams a person belongs to now.
type Directory interface {
	PersonRefs(ctx context.Context, ids []string) (map[string]people.PersonRef, error)
	TeamRefs(ctx context.Context, ids []string) (map[string]people.TeamRef, error)
	TeamsOf(ctx context.Context, personID string) ([]string, error)
}
