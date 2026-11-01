package attribution

import (
	"context"
	"slices"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/clock"
	"mapp/internal/platform/errs"
)

const (
	settingsKind = "attributionSettings"
	settingsID   = "attribution"
	settingsName = "Attribution"
	settingsNoun = "attribution setting"
	msgPickModel = "Choose one of the listed models."
	fieldModel   = "model"
)

type Service struct {
	tx   Transactor
	repo Store
	log  *history.Service
	now  clock.Clock
}

func NewService(tools history.Tools, repo Store) *Service {
	return &Service{tx: tools.Tx, repo: repo, log: tools.Log, now: tools.Now}
}

func (s *Service) Settings(ctx context.Context) (SettingsView, error) {
	var view SettingsView
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		st, err := s.repo.Settings(ctx)
		view = settingsView(ctx, st)
		return err
	})
	return view, err
}

// Designated is the model every attributed figure uses (FR-008, FR-010); it joins the caller's transaction.
func (s *Service) Designated(ctx context.Context) (Model, error) {
	st, err := s.repo.Settings(ctx)
	return st.Model, err
}

// SaveSettings needs Manage settings; a stale version is refused and the change is logged (FR-002).
func (s *Service) SaveSettings(ctx context.Context, a actor.Actor, in SettingsInput) (SettingsView, error) {
	if err := authz.Require(ctx, authz.SettingsManage); err != nil {
		return SettingsView{}, err
	}
	if !slices.ContainsFunc(Models(), func(c ModelChoice) bool { return c.Key == in.Model }) {
		return SettingsView{}, errs.Validation(errs.FieldError{Field: fieldModel, Message: msgPickModel})
	}
	var view SettingsView
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		before, err := s.repo.Settings(ctx)
		if err != nil {
			return err
		}
		if err = before.CheckVersion(in.Version, settingsNoun); err != nil {
			return err
		}
		after := before
		after.Model = in.Model
		if after.Model == before.Model {
			view = settingsView(ctx, before)
			return nil
		}
		after.Touch(a, s.now())
		if err = s.repo.PutSettings(ctx, after); err != nil {
			return err
		}
		changes, err := history.Diff(before, after)
		if err != nil {
			return err
		}
		_, err = s.log.Log(ctx, a, history.LogInput{Kind: settingsKind, RecordID: settingsID, RecordName: settingsName,
			Action: history.ActionChanged, Changes: changes})
		view = settingsView(ctx, after)
		return err
	})
	return view, err
}

func settingsView(ctx context.Context, st Settings) SettingsView {
	view := SettingsView{Model: st.Model, Models: Models(), Version: st.Version,
		CanChange: authz.AvailableAnywhere(ctx, authz.SettingsManage)}
	if !st.UpdatedAt.IsZero() && st.UpdatedBy.ID != actor.System.ID {
		at, by := st.UpdatedAt, st.UpdatedBy
		view.UpdatedAt, view.UpdatedBy = &at, &by
	}
	return view
}
