package customdashboard

import (
	"context"
	"fmt"
	"strings"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/clock"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

type Service struct {
	tx        Transactor
	repo      Store
	ids       *ids.Generator
	history   *history.Service
	directory Directory
	now       clock.Clock
}

func NewService(tools history.Tools, repo Store, directory Directory) *Service {
	return &Service{tx: tools.Tx, repo: repo, ids: tools.IDs, history: tools.Log, directory: directory, now: tools.Now}
}

// List answers the person's own dashboards and those shared with their teams, each list by name.
func (s *Service) List(ctx context.Context, a actor.Actor) (List, error) {
	var list List
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		visible, err := s.repo.ListVisible(ctx, a.ID)
		if err != nil {
			return err
		}
		list, err = s.summaries(ctx, a, visible)
		return err
	})
	return list, err
}

// Get answers a dashboard the person may not see exactly like a missing one.
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
	return view, err
}

func (s *Service) Create(ctx context.Context, a actor.Actor, in CreateInput) (View, error) {
	if err := authz.Require(ctx, authz.DashboardsBuild); err != nil {
		return View{}, err
	}
	d := normalised(Dashboard{Name: strings.TrimSpace(in.Name), Description: strings.TrimSpace(in.Description),
		OwnerID: a.ID, Filters: in.Filters, Widgets: in.Widgets})
	if err := validate(d).Err(); err != nil {
		return View{}, err
	}
	var view View
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		var err error
		if d, err = s.insert(ctx, a, d); err != nil {
			return err
		}
		if _, err = s.history.Log(ctx, a, history.LogInput{
			Kind: kindDashboard, RecordID: d.ID, RecordName: d.Name, Action: history.ActionCreated,
		}); err != nil {
			return err
		}
		view, err = s.view(ctx, a, d)
		return err
	})
	return view, err
}

// Update saves only the fields sent; a save that changes nothing creates no version and no log entry.
func (s *Service) Update(ctx context.Context, a actor.Actor, id string, in UpdateInput) (View, error) {
	var view View
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		stored, err := s.owned(ctx, a, id)
		if err != nil {
			return err
		}
		changed := applied(stored, in)
		problems := validate(changed)
		if in.Version == nil {
			problems.Add(versionField, msgEnterVersion)
		}
		if err = problems.Err(); err != nil {
			return err
		}
		if err = stored.CheckVersion(*in.Version, dashboardNoun); err != nil {
			return err
		}
		changed, err = s.save(ctx, a, stored, changed)
		if err != nil {
			return err
		}
		view, err = s.view(ctx, a, changed)
		return err
	})
	return view, err
}

func (s *Service) Delete(ctx context.Context, a actor.Actor, id string) (history.BinEntry, error) {
	var entry history.BinEntry
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		if _, err := s.owned(ctx, a, id); err != nil {
			return err
		}
		var err error
		entry, err = s.history.MoveToBin(ctx, a, kindDashboard, id)
		return err
	})
	return entry, err
}

func applied(stored Dashboard, in UpdateInput) Dashboard {
	changed := stored
	if in.Name != nil {
		changed.Name = strings.TrimSpace(*in.Name)
	}
	if in.Description != nil {
		changed.Description = strings.TrimSpace(*in.Description)
	}
	if in.Filters != nil {
		changed.Filters = *in.Filters
	}
	if in.Widgets != nil {
		changed.Widgets = *in.Widgets
	}
	return normalised(changed)
}

// save stores what changed between stored and changed and logs it in the same Write.
func (s *Service) save(ctx context.Context, a actor.Actor, stored, changed Dashboard) (Dashboard, error) {
	if err := s.checkNameFree(ctx, changed, stored.ID); err != nil {
		return Dashboard{}, err
	}
	changes, err := logChanges(stored, changed)
	if err != nil || len(changes) == 0 {
		return stored, err
	}
	changed.Touch(a, s.now())
	if err = s.repo.Update(ctx, changed, stored.Version); err != nil {
		return Dashboard{}, err
	}
	if err = s.repo.ReplaceWidgets(ctx, changed.ID, changed.Widgets); err != nil {
		return Dashboard{}, err
	}
	_, err = s.history.Log(ctx, a, history.LogInput{
		Kind: kindDashboard, RecordID: changed.ID, RecordName: changed.Name, Action: history.ActionChanged, Changes: changes,
	})
	return changed, err
}

// insert gives d its id and bookkeeping and stores it, refusing a name the owner already uses.
func (s *Service) insert(ctx context.Context, a actor.Actor, d Dashboard) (Dashboard, error) {
	if err := s.checkNameFree(ctx, d, ""); err != nil {
		return Dashboard{}, err
	}
	id, err := s.ids.Next(ctx, ids.PrefixDashboard)
	if err != nil {
		return Dashboard{}, err
	}
	d.ID, d.Meta = id, record.NewMeta(a, s.now())
	return d, s.repo.Insert(ctx, d)
}

func (s *Service) checkNameFree(ctx context.Context, d Dashboard, exceptID string) error {
	taken, err := s.repo.NameTaken(ctx, d.OwnerID, d.Name, exceptID)
	if err != nil || !taken {
		return err
	}
	return errs.NameTaken(fmt.Sprintf(msgNameTaken, d.Name))
}

func (s *Service) visible(ctx context.Context, a actor.Actor, id string) (Dashboard, error) {
	d, found, err := s.repo.GetVisible(ctx, a.ID, id)
	if err != nil {
		return Dashboard{}, err
	}
	if !found {
		return Dashboard{}, errs.NotFound(authz.NotFoundMessage(dashboardNoun))
	}
	return d, nil
}

// owned answers 404 for a dashboard the person cannot see and 403 for one they see but may not change.
func (s *Service) owned(ctx context.Context, a actor.Actor, id string) (Dashboard, error) {
	d, err := s.visible(ctx, a, id)
	if err != nil {
		return Dashboard{}, err
	}
	if d.OwnerID != a.ID {
		owner, err := s.ownerName(ctx, d.OwnerID)
		if err != nil {
			return Dashboard{}, err
		}
		return Dashboard{}, errs.Forbidden(fmt.Sprintf(msgNotOwner, owner))
	}
	return d, authz.Require(ctx, authz.DashboardsBuild)
}
