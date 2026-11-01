package performance

import (
	"context"
	"fmt"
	"strings"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

type IntegrationIDInput struct {
	Value string `json:"value"`
}

func (s *Service) IntegrationIDs(ctx context.Context, _ actor.Actor, activityID string) ([]IntegrationID, error) {
	var keys []IntegrationID
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		if _, readErr := s.visibleActivity(ctx, activityID); readErr != nil {
			return readErr
		}
		var readErr error
		keys, readErr = s.repo.ListIntegrationIDs(ctx, activityID)
		return readErr
	})
	if err != nil {
		return nil, fmt.Errorf("list the integration ids of %s: %w", activityID, err)
	}
	return keys, nil
}

// SetIntegrationID registers the outside system's own id for this activity: at most one per source, and a
// second is refused rather than replaced, so a swap is a deliberate delete and set (FR-037).
func (s *Service) SetIntegrationID(ctx context.Context, a actor.Actor, ref KeyRef, in IntegrationIDInput) (IntegrationID, error) {
	if err := authz.Require(ctx, PermissionLink); err != nil {
		return IntegrationID{}, err
	}
	value := strings.TrimSpace(in.Value)
	if value == "" || len(value) > maxNameLength {
		return IntegrationID{}, errs.Validation(errs.FieldError{Field: "value", Message: msgIntegrationValue})
	}
	var saved IntegrationID
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		var writeErr error
		saved, writeErr = s.writeIntegrationID(ctx, a, ref, value)
		return writeErr
	})
	if err != nil {
		return IntegrationID{}, fmt.Errorf("set the integration id of %s for %s: %w", ref.ActivityID, ref.SourceID, err)
	}
	return saved, nil
}

type KeyRef struct {
	ActivityID string
	SourceID   string
}

func (s *Service) writeIntegrationID(ctx context.Context, a actor.Actor, ref KeyRef, value string) (IntegrationID, error) {
	if _, err := s.visibleActivity(ctx, ref.ActivityID); err != nil {
		return IntegrationID{}, err
	}
	sources, err := s.repo.ListSources(ctx, ref.SourceID)
	if err != nil {
		return IntegrationID{}, err
	}
	if len(sources) == 0 {
		return IntegrationID{}, errs.NotFound(msgSourceNotFound)
	}
	existing, err := s.repo.ListIntegrationIDs(ctx, ref.ActivityID)
	if err != nil {
		return IntegrationID{}, err
	}
	for _, key := range existing {
		if key.SourceID == ref.SourceID {
			return IntegrationID{}, errs.NotAllowed(fmt.Sprintf(msgIntegrationExists, sources[0].Name, key.Value))
		}
	}
	id, err := s.ids.Next(ctx, ids.PrefixIntegrationID)
	if err != nil {
		return IntegrationID{}, err
	}
	saved := IntegrationID{ID: id, ActivityID: ref.ActivityID, SourceID: ref.SourceID, Value: value,
		Meta: record.NewMeta(a, s.now())}
	if putErr := s.repo.PutIntegrationID(ctx, saved); putErr != nil {
		return IntegrationID{}, putErr
	}
	_, err = s.history.Log(ctx, a, history.LogInput{
		Kind: integrationKind, RecordID: saved.ID, RecordName: saved.Value,
		ParentID: ref.ActivityID, Action: history.ActionCreated,
	})
	return saved, err
}

// DeleteIntegrationID stops the source's rows matching this activity through that key (FR-037).
func (s *Service) DeleteIntegrationID(ctx context.Context, a actor.Actor, ref KeyRef) error {
	if err := authz.Require(ctx, PermissionLink); err != nil {
		return err
	}
	return s.tx.Write(ctx, func(ctx context.Context) error {
		if _, err := s.visibleActivity(ctx, ref.ActivityID); err != nil {
			return err
		}
		keys, err := s.repo.ListIntegrationIDs(ctx, ref.ActivityID)
		if err != nil {
			return err
		}
		for _, key := range keys {
			if key.SourceID != ref.SourceID {
				continue
			}
			if _, removeErr := s.repo.RemoveIntegrationID(ctx, ref.ActivityID, ref.SourceID); removeErr != nil {
				return removeErr
			}
			_, err = s.history.Log(ctx, a, history.LogInput{
				Kind: integrationKind, RecordID: key.ID, RecordName: key.Value,
				ParentID: ref.ActivityID, Action: history.ActionDeleted,
			})
			return err
		}
		return errs.NotFound(msgIntegrationNotFound)
	})
}
