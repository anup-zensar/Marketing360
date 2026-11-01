package app

import (
	"context"

	"mapp/internal/access"
	"mapp/internal/measure"
	"mapp/internal/measure/useraccess"
	"mapp/internal/platform/actor"
	"mapp/internal/stores"
)

// registerUserAccess registers and mounts the User access report (feature 045); auth's store gives the sign-ins.
func registerUserAccess(w *wiring, r *measure.Registry) error {
	service := useraccess.NewService(w.set.Tx, userAccessPeople{access: w.access, people: w.set.People}, w.set.Auth,
		w.now)
	mount(w.signedIn, useraccess.NewHandler(service))
	return r.Register(service)
}

// userAccessPeople is the People tab's own read (014 FR-017), with each person's e-mail (045 research R1).
type userAccessPeople struct {
	access *access.Service
	people stores.PeopleStore
}

func (u userAccessPeople) PeopleAccess(ctx context.Context, a actor.Actor) ([]useraccess.Person, error) {
	list, err := u.access.ListPeopleAccess(ctx, a)
	if err != nil {
		return nil, err
	}
	everyone, err := u.people.List(ctx)
	if err != nil {
		return nil, err
	}
	emails := make(map[string]string, len(everyone))
	for _, p := range everyone {
		emails[p.ID] = p.Email
	}
	out := make([]useraccess.Person, 0, len(list))
	for _, p := range list {
		out = append(out, useraccess.Person{Ref: p.Person, Email: emails[p.Person.ID], Roles: p.Roles, Teams: p.Teams})
	}
	return out, nil
}
