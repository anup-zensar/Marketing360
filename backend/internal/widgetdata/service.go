package widgetdata

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

const datasetPrefix = "DS-"

type Service struct {
	tx   Transactor
	deps Deps
}

func NewService(tx Transactor, deps Deps) *Service {
	return &Service{tx: tx, deps: deps}
}

// FinanceNote is the note a reader gives when the person sees financial data in only some of the plans.
func FinanceNote(seen, of int) string {
	return fmt.Sprintf(financeNote, seen, of)
}

// Catalogue lists the built-in sources a may read, then the datasets a may see by name.
func (s *Service) Catalogue(ctx context.Context, a actor.Actor) ([]Source, error) {
	if err := mayUse(ctx); err != nil {
		return nil, err
	}
	var out []Source
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		for _, src := range builtIns() {
			if holds(ctx, src.need) {
				out = append(out, src)
			}
		}
		datasets, err := s.datasets(ctx, a)
		out = append(out, datasets...)
		return err
	})
	return out, err
}

func (s *Service) datasets(ctx context.Context, a actor.Actor) ([]Source, error) {
	if s.deps.Datasets == nil {
		return nil, nil
	}
	infos, err := s.deps.Datasets.Visible(ctx, a)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(infos, func(x, y DatasetInfo) int {
		return cmp.Or(strings.Compare(strings.ToLower(x.Name), strings.ToLower(y.Name)), strings.Compare(x.ID, y.ID))
	})
	out := make([]Source, 0, len(infos))
	for _, info := range infos {
		out = append(out, datasetSource(info))
	}
	return out, nil
}

// Query works one widget's figures out for a: 403 for a built-in source a may not read, 404 for a dataset a
// may not see, 400 with every problem of the query.
func (s *Service) Query(ctx context.Context, a actor.Actor, r Request) (Data, error) {
	if err := mayUse(ctx); err != nil {
		return Data{}, err
	}
	var data Data
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		in, err := s.input(ctx, a, r)
		if err != nil {
			return err
		}
		data, err = in.answer(ctx)
		return err
	})
	if errors.Is(err, errTooLarge) {
		return Data{}, errs.Validation(errs.FieldError{Field: "query.measures", Message: msgTooLarge})
	}
	return data, err
}

func (s *Service) input(ctx context.Context, a actor.Actor, r Request) (*input, error) {
	in := &input{service: s, actor: a, request: r}
	src, err := s.source(ctx, a, r.Query.Source, in)
	if err != nil {
		return nil, err
	}
	in.plan, err = compile(r, src)
	return in, err
}

func (s *Service) source(ctx context.Context, a actor.Actor, key string, in *input) (Source, error) {
	if strings.HasPrefix(key, datasetPrefix) {
		return s.dataset(ctx, a, key, in)
	}
	src, found := builtIn(key)
	if !found {
		return Source{}, errs.Validation(errs.FieldError{Field: "query.source", Message: msgNoSource})
	}
	if err := authz.Require(ctx, src.need); err != nil {
		return Source{}, err
	}
	reader, wired := s.deps.Sources[key]
	if !wired {
		return Source{}, fmt.Errorf("widgetdata: no reader for the source %s", key)
	}
	in.reader = reader
	return src, nil
}

func (s *Service) dataset(ctx context.Context, a actor.Actor, id string, in *input) (Source, error) {
	if s.deps.Datasets == nil {
		return Source{}, errs.NotFound(msgNoDatasets)
	}
	table, err := s.deps.Datasets.Table(ctx, a, id)
	if err != nil {
		return Source{}, err
	}
	rows, notes, err := s.inUSD(ctx, table)
	in.rows, in.notes = rows, notes
	return datasetSource(table.Info), err
}

func mayUse(ctx context.Context) error {
	if holds(ctx, authz.DashboardsView) || holds(ctx, authz.DashboardsBuild) {
		return nil
	}
	return errs.Forbidden(msgNoAccess)
}

func holds(ctx context.Context, p authz.Permission) bool {
	g, restricted := authz.From(ctx)
	return !restricted || g.Has(p)
}
