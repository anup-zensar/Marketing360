package authtest

import (
	"context"
	"time"
)

// SignIn is a person's latest sign-in as the fake keeps it (feature 045).
type SignIn struct {
	PersonID string
	At       time.Time
}

func (s Store) RecordSignIn(ctx context.Context, personID string, at time.Time) error {
	return s.signIns.Put(ctx, SignIn{PersonID: personID, At: at})
}

func (s Store) LastSignIns(context.Context) (map[string]time.Time, error) {
	list, err := s.signIns.List()
	last := make(map[string]time.Time, len(list))
	for _, r := range list {
		last[r.PersonID] = r.At
	}
	return last, err
}

// SignInsRecordedSince is unknown in memory: recording starts with the process.
func (s Store) SignInsRecordedSince(context.Context) (time.Time, error) {
	return time.Time{}, nil
}
