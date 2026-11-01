package dataset

import (
	"context"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/clock"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/ids"
)

const (
	defaultPageSize = 50
	maxPageSize     = 200
	offsetField     = "offset"
	limitField      = "limit"
)

type Service struct {
	tx         Transactor
	repo       Store
	ids        *ids.Generator
	history    *history.Service
	now        clock.Clock
	directory  Directory
	currencies Currencies
	usage      Usage
}

func NewService(tools history.Tools, repo Store, deps Deps) *Service {
	return &Service{
		tx: tools.Tx, repo: repo, ids: tools.IDs, history: tools.Log, now: tools.Now,
		directory: deps.Directory, currencies: deps.Currencies, usage: noUsage{},
	}
}

func (s *Service) UseUsage(u Usage) {
	s.usage = u
}

type noUsage struct{}

func (noUsage) UsingDataset(context.Context, actor.Actor, string) ([]DashboardRef, error) {
	return []DashboardRef{}, nil
}

func (s *Service) List(ctx context.Context, a actor.Actor) (List, error) {
	list := List{Mine: []Summary{}, SharedWithMe: []Summary{}}
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		datasets, err := s.repo.ListVisible(ctx, a.ID)
		if err != nil {
			return err
		}
		summaries, err := s.summaries(ctx, a, datasets)
		if err != nil {
			return err
		}
		for i, d := range datasets {
			if d.OwnerID == a.ID {
				list.Mine = append(list.Mine, summaries[i])
			} else {
				list.SharedWithMe = append(list.SharedWithMe, summaries[i])
			}
		}
		return nil
	})
	if err != nil {
		return List{}, err
	}
	return list, nil
}

func (s *Service) Get(ctx context.Context, a actor.Actor, id string) (View, error) {
	var view View
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		d, err := s.visible(ctx, a, id)
		if err != nil {
			return err
		}
		view, err = s.view(ctx, a, d)
		return err
	})
	if err != nil {
		return View{}, err
	}
	return view, nil
}

// Rows answers a page of the rows as stored: text as written, numbers and money as decimals, dates YYYY-MM-DD.
func (s *Service) Rows(ctx context.Context, a actor.Actor, id string, q RowsQuery) (RowPage, error) {
	page, err := pageOf(q)
	if err != nil {
		return RowPage{}, err
	}
	err = s.tx.Read(ctx, func(ctx context.Context) error {
		d, err := s.visible(ctx, a, id)
		if err != nil {
			return err
		}
		rows, err := s.repo.RowsPage(ctx, id, page.Offset, page.Limit)
		if err != nil {
			return err
		}
		page.Columns, page.Total, page.Rows = d.Columns, d.RowCount, make([][]*string, 0, len(rows))
		for _, r := range rows {
			page.Rows = append(page.Rows, rowTexts(d.Columns, r))
		}
		return nil
	})
	if err != nil {
		return RowPage{}, err
	}
	return page, nil
}

func pageOf(q RowsQuery) (RowPage, error) {
	page := RowPage{Limit: defaultPageSize}
	var problems errs.Problems
	if q.Offset != nil {
		page.Offset = *q.Offset
		if page.Offset < 0 || page.Offset > maxRows {
			problems.Add(offsetField, msgOffset)
		}
	}
	if q.Limit != nil {
		page.Limit = *q.Limit
		if page.Limit < 1 || page.Limit > maxPageSize {
			problems.Add(limitField, msgLimit)
		}
	}
	return page, problems.Err()
}

// visible answers not_found for a dataset the person may not see, exactly as for a missing one.
func (s *Service) visible(ctx context.Context, a actor.Actor, id string) (Dataset, error) {
	d, found, err := s.repo.GetVisible(ctx, a.ID, id)
	if err != nil {
		return Dataset{}, err
	}
	if !found {
		return Dataset{}, errs.NotFound(authz.NotFoundMessage(kindSingular))
	}
	return d, nil
}

// owned is visible, and refuses with forbidden a dataset the person sees but does not own.
func (s *Service) owned(ctx context.Context, a actor.Actor, id string) (Dataset, error) {
	d, err := s.visible(ctx, a, id)
	if err != nil {
		return Dataset{}, err
	}
	if d.OwnerID != a.ID {
		return Dataset{}, errs.Forbidden(msgOwnerOnly)
	}
	return d, nil
}

func accessOf(isOwner bool) Access {
	if isOwner {
		yes := Availability{Available: true}
		return Access{Edit: yes, Share: yes, Delete: yes}
	}
	no := Availability{Reason: msgOwnerOnly}
	return Access{Edit: no, Share: no, Delete: no}
}
