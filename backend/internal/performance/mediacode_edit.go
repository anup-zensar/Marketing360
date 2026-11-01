package performance

import (
	"context"
	"fmt"
	"strings"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

// EditMediaCode changes attributes and landing page only (FR-041): the number and the activity are the
// code's identity and never move.
func (s *Service) EditMediaCode(ctx context.Context, a actor.Actor, codeID string, in MediaCodeInput) (MediaCode, error) {
	if err := authz.Require(ctx, PermissionLink); err != nil {
		return MediaCode{}, err
	}
	var saved MediaCode
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		stored, findErr := s.ownCode(ctx, codeID)
		if findErr != nil {
			return findErr
		}
		if stored.Version != in.Version {
			return errs.VersionConflict("media code", stored.UpdatedBy.Name, stored.UpdatedAt)
		}
		saved = stored
		saved.Attributes = attributesOf(in)
		saved.LandingPage = strings.TrimSpace(in.LandingPage)
		saved.Version++
		saved.UpdatedAt, saved.UpdatedBy = s.now(), a
		if putErr := s.repo.PutMediaCode(ctx, saved); putErr != nil {
			return putErr
		}
		_, logErr := s.history.Log(ctx, a, history.LogInput{
			Kind: mediaCodeKind, RecordID: saved.ID, RecordName: codeText(saved.Code),
			ParentID: saved.ActivityID, Action: history.ActionChanged,
		})
		return logErr
	})
	if err != nil {
		return MediaCode{}, fmt.Errorf("edit media code %s: %w", codeID, err)
	}
	return saved, nil
}

// DeleteMediaCode removes an unused code to the recycle bin; one with facts behind it deactivates instead,
// because deleting it would orphan what already matched (FR-041).
func (s *Service) DeleteMediaCode(ctx context.Context, a actor.Actor, codeID string) error {
	if err := authz.Require(ctx, PermissionLink); err != nil {
		return err
	}
	return s.tx.Write(ctx, func(ctx context.Context) error {
		stored, err := s.ownCode(ctx, codeID)
		if err != nil {
			return err
		}
		used, err := s.repo.MediaCodeUsed(ctx, stored.Code)
		if err != nil {
			return err
		}
		if used {
			return errs.NotAllowed(msgCodeUsed)
		}
		_, err = s.history.MoveToBin(ctx, a, mediaCodeKind, codeID)
		return err
	})
}

func (s *Service) DeactivateMediaCode(ctx context.Context, a actor.Actor, codeID string) (MediaCode, error) {
	if err := authz.Require(ctx, PermissionLink); err != nil {
		return MediaCode{}, err
	}
	var saved MediaCode
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		stored, findErr := s.ownCode(ctx, codeID)
		if findErr != nil {
			return findErr
		}
		saved = stored
		saved.Status = StatusInactive
		saved.Version++
		saved.UpdatedAt, saved.UpdatedBy = s.now(), a
		if putErr := s.repo.PutMediaCode(ctx, saved); putErr != nil {
			return putErr
		}
		_, logErr := s.history.Log(ctx, a, history.LogInput{
			Kind: mediaCodeKind, RecordID: saved.ID, RecordName: codeText(saved.Code),
			ParentID: saved.ActivityID, Action: history.ActionStatusChanged, Verb: deactivatedVerb,
		})
		return logErr
	})
	if err != nil {
		return MediaCode{}, fmt.Errorf("deactivate media code %s: %w", codeID, err)
	}
	return saved, nil
}

// ownCode reads a code and gates it by its activity's visibility: a code of an activity outside the
// reader's scope answers as missing.
func (s *Service) ownCode(ctx context.Context, codeID string) (MediaCode, error) {
	stored, found, err := s.repo.GetMediaCode(ctx, codeID)
	if err != nil {
		return MediaCode{}, err
	}
	if !found {
		return MediaCode{}, errs.NotFound(msgCodeNotFound)
	}
	if _, err := s.visibleActivity(ctx, stored.ActivityID); err != nil {
		return MediaCode{}, errs.NotFound(msgCodeNotFound)
	}
	return stored, nil
}
