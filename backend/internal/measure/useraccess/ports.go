package useraccess

import (
	"context"
	"time"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
)

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

// People is everyone as Settings › Roles and access › People shows them (014 FR-017), with their e-mail.
type People interface {
	PeopleAccess(ctx context.Context, a actor.Actor) ([]Person, error)
}

type Person struct {
	Ref   authz.PersonRef
	Email string
	Roles []authz.RoleAssignment
	Teams []authz.TeamRef
}

// SignIns are auth's latest sign-in per person and when recording started (zero when unknown).
type SignIns interface {
	LastSignIns(ctx context.Context) (map[string]time.Time, error)
	SignInsRecordedSince(ctx context.Context) (time.Time, error)
}
