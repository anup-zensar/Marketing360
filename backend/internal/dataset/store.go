package dataset

import (
	"context"

	"mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/settings"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
	Write(ctx context.Context, fn func(ctx context.Context) error) error
}

// Store reads a dataset with its columns and shares. The visible reads answer only what personID owns or what is
// shared with a team they are a member of now; Get and OwnerOf read whoever asks (history, the owner check).
type Store interface {
	Get(ctx context.Context, id string) (Dataset, bool, error)
	GetVisible(ctx context.Context, personID, id string) (Dataset, bool, error)
	ListVisible(ctx context.Context, personID string) ([]Dataset, error)
	CountOfOwner(ctx context.Context, ownerID string) (int, error)
	OwnerOf(ctx context.Context, id string) (string, bool, error)
	Rows(ctx context.Context, id string) ([]Row, error)
	RowsPage(ctx context.Context, id string, offset, limit int) ([]Row, error)
	Insert(ctx context.Context, d Dataset, rows []Row) error
	Update(ctx context.Context, d Dataset) error
	UpdateLabels(ctx context.Context, id string, columns []Column) error
	ReplaceShares(ctx context.Context, id string, teamIDs []string) error
	Delete(ctx context.Context, id string) error
}

type Directory interface {
	PersonRefs(ctx context.Context, ids []string) (map[string]people.PersonRef, error)
	TeamRefs(ctx context.Context, ids []string) (map[string]people.TeamRef, error)
	TeamIDsOf(ctx context.Context, personID string) ([]string, error)
}

type Currencies interface {
	Currency(ctx context.Context, code string) (settings.Currency, bool, error)
}

// Usage names the dashboards the actor may see whose widgets read the dataset.
type Usage interface {
	UsingDataset(ctx context.Context, a actor.Actor, datasetID string) ([]DashboardRef, error)
}

type Deps struct {
	Directory  Directory
	Currencies Currencies
}
