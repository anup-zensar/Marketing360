package useraccess

import (
	"context"
	"time"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
)

// Service works out the User access report: read-only, nothing on it changes a record (FR-004).
type Service struct {
	tx      Transactor
	people  People
	signIns SignIns
	now     func() time.Time
}

func NewService(tx Transactor, people People, signIns SignIns, now func() time.Time) *Service {
	return &Service{tx: tx, people: people, signIns: signIns, now: now}
}

// Report lists the people matching q with their cards and bands, sorted, then one page of them, in one Read.
func (s *Service) Report(ctx context.Context, a actor.Actor, q Query) (Report, error) {
	q = withDefaults(q)
	var r Report
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		list, since, err := s.matching(ctx, a, q)
		if err != nil {
			return err
		}
		bands, cards := bandsAndCards(list, s.now())
		r = Report{Total: len(list), Page: q.Page, PageSize: q.PageSize, Sort: q.Sort, Cards: cards, Bands: bands,
			RecordedSince: since, Rows: pageOf(list, q.Page, q.PageSize)}
		return nil
	})
	if err != nil {
		return Report{}, err
	}
	return r, nil
}

// matching gives every matching row in q's sort, and the day recording started (today when unknown).
func (s *Service) matching(ctx context.Context, a actor.Actor, q Query) ([]Row, date.Date, error) {
	people, err := s.people.PeopleAccess(ctx, a)
	if err != nil {
		return nil, date.Date{}, err
	}
	last, err := s.signIns.LastSignIns(ctx)
	if err != nil {
		return nil, date.Date{}, err
	}
	since, err := s.signIns.SignInsRecordedSince(ctx)
	if err != nil {
		return nil, date.Date{}, err
	}
	if since.IsZero() {
		since = s.now()
	}
	list := matching(rows(people, last), q)
	sortRows(list, q.Sort)
	return list, date.Today(since.UTC()), nil
}
