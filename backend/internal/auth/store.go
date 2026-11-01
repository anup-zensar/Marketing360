package auth

import (
	"context"
	"time"

	"mapp/internal/people"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
	Write(ctx context.Context, fn func(ctx context.Context) error) error
}

// Store keeps sessions (listed by id; deleting a missing one changes nothing) and each person's last sign-in (045).
type Store interface {
	Get(ctx context.Context, id string) (Session, bool, error)
	Put(ctx context.Context, s Session) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Session, error)
	RecordSignIn(ctx context.Context, personID string, at time.Time) error
	LastSignIns(ctx context.Context) (map[string]time.Time, error)
	SignInsRecordedSince(ctx context.Context) (time.Time, error)
}

// People is what sign-in reads of people and teams.
type People interface {
	Get(ctx context.Context, id string) (people.Person, bool, error)
	FindByEmail(ctx context.Context, email string) (people.Person, bool, error)
}
