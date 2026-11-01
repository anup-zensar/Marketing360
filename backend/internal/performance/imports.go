package performance

import (
	"context"
	"fmt"
	"time"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

type ImportView struct {
	Import
	// DataThrough shadows the domain's timestamp as the contract's plain date, null until something loads.
	DataThrough *string     `json:"dataThrough"`
	Breakdown   []Breakdown `json:"breakdown"`
}

func viewOfImport(run Import, breakdown []Breakdown) ImportView {
	view := ImportView{Import: run, Breakdown: breakdown}
	if run.DataThrough != nil {
		through := run.DataThrough.Format(time.DateOnly)
		view.DataThrough = &through
	}
	return view
}

// Imports and ImportByID are permission-gated rather than scoped (046-C15): a load is operational work,
// not a plan-owned record, so performance.import sees every upload.
func (s *Service) Imports(ctx context.Context, _ actor.Actor, f ImportFilter) ([]ImportView, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return nil, err
	}
	f.ID = ""
	var views []ImportView
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		found, readErr := s.repo.ListImports(ctx, f)
		if readErr != nil {
			return readErr
		}
		views, readErr = s.withBreakdowns(ctx, found)
		return readErr
	})
	if err != nil {
		return nil, fmt.Errorf("list performance imports: %w", err)
	}
	return views, nil
}

func (s *Service) ImportByID(ctx context.Context, _ actor.Actor, id string) (ImportView, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return ImportView{}, err
	}
	var views []ImportView
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		found, readErr := s.repo.ListImports(ctx, ImportFilter{ID: id, Limit: 1})
		if readErr != nil {
			return readErr
		}
		views, readErr = s.withBreakdowns(ctx, found)
		return readErr
	})
	if err != nil {
		return ImportView{}, fmt.Errorf("get performance import %s: %w", id, err)
	}
	if len(views) == 0 {
		return ImportView{}, errs.NotFound(msgImportNotFound)
	}
	return views[0], nil
}

func (s *Service) withBreakdowns(ctx context.Context, imports []Import) ([]ImportView, error) {
	ids := make([]string, 0, len(imports))
	for _, run := range imports {
		ids = append(ids, run.ID)
	}
	groups, err := s.repo.Breakdowns(ctx, ids)
	if err != nil {
		return nil, err
	}
	views := make([]ImportView, 0, len(imports))
	for _, run := range imports {
		breakdown := groups[run.ID]
		if breakdown == nil {
			breakdown = []Breakdown{}
		}
		views = append(views, viewOfImport(run, breakdown))
	}
	return views, nil
}
