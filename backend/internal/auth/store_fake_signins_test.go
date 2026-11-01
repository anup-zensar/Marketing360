package auth

import (
	"context"
	"time"
)

// signIn is the fake's row of a person's latest sign-in.
type signIn struct {
	PersonID string
	At       time.Time
}

func (s fakeStore) RecordSignIn(ctx context.Context, personID string, at time.Time) error {
	return s.signIns.Put(ctx, signIn{PersonID: personID, At: at})
}

func (s fakeStore) LastSignIns(context.Context) (map[string]time.Time, error) {
	list, err := s.signIns.List()
	last := make(map[string]time.Time, len(list))
	for _, r := range list {
		last[r.PersonID] = r.At
	}
	return last, err
}

func (s fakeStore) SignInsRecordedSince(context.Context) (time.Time, error) {
	return time.Time{}, nil
}
