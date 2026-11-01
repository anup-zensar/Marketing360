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

// ReviewRows is permission-gated, not scoped (046-C15): the queue is operational work. One field is
// plan-owned all the same - candidates name activities - so each row's list is filtered to the reader's
// scope, and a row whose candidates all lie outside it still appears naming none of them.
func (s *Service) ReviewRows(ctx context.Context, _ actor.Actor, f ReviewFilter) ([]ReviewRowView, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return nil, err
	}
	var rows []ReviewRowView
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		found, readErr := s.repo.ListReviewRows(ctx, f)
		if readErr != nil {
			return readErr
		}
		rows, readErr = s.scopedCandidates(ctx, found)
		return readErr
	})
	if err != nil {
		return nil, fmt.Errorf("list the rows to review: %w", err)
	}
	return rows, nil
}

func (s *Service) ReviewRowByID(ctx context.Context, a actor.Actor, id string) (ReviewRowView, error) {
	rows, err := s.ReviewRows(ctx, a, ReviewFilter{ID: id, Limit: 1})
	if err != nil {
		return ReviewRowView{}, err
	}
	if len(rows) == 0 {
		return ReviewRowView{}, errs.NotFound(msgReviewRowNotFound)
	}
	return rows[0], nil
}

func (s *Service) scopedCandidates(ctx context.Context, rows []ReviewRowView) ([]ReviewRowView, error) {
	var ids []string
	for _, row := range rows {
		ids = append(ids, row.Candidates...)
	}
	if len(ids) == 0 {
		return rows, nil
	}
	visible, err := s.repo.MatchByActivityID(ctx, authz.ScopeOf(ctx), normalKeys(MatchedActivityID, distinct(ids)))
	if err != nil {
		return nil, err
	}
	for i, row := range rows {
		kept := []string{}
		for _, id := range row.Candidates {
			if len(visible[normalKey(MatchedActivityID, id)]) > 0 {
				kept = append(kept, id)
			}
		}
		rows[i].Candidates = kept
	}
	return rows, nil
}

type DismissInput struct {
	Comment string `json:"comment"`
}

// Dismiss parks a row with the required comment (FR-024): the queue must say why work was set aside.
func (s *Service) Dismiss(ctx context.Context, a actor.Actor, rowID string, in DismissInput) (ReviewRow, error) {
	comment := strings.TrimSpace(in.Comment)
	if comment == "" {
		return ReviewRow{}, errs.Validation(errs.FieldError{Field: "comment", Message: msgCommentNeeded})
	}
	return s.setReviewStatus(ctx, a, reviewChange{
		rowID: rowID, from: ReviewOpen, to: ReviewDismissed, comment: comment, verb: dismissedVerb,
	})
}

// Reopen returns a dismissed row to the queue; the dismissal and the reopening both stay in the history.
func (s *Service) Reopen(ctx context.Context, a actor.Actor, rowID string) (ReviewRow, error) {
	return s.setReviewStatus(ctx, a, reviewChange{
		rowID: rowID, from: ReviewDismissed, to: ReviewOpen, verb: reopenedVerb,
	})
}

type reviewChange struct {
	rowID   string
	from    ReviewStatus
	to      ReviewStatus
	comment string
	verb    string
}

func (s *Service) setReviewStatus(ctx context.Context, a actor.Actor, change reviewChange) (ReviewRow, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return ReviewRow{}, err
	}
	var saved ReviewRow
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		rows, readErr := s.repo.ListReviewRows(ctx, ReviewFilter{ID: change.rowID, Limit: 1})
		if readErr != nil {
			return readErr
		}
		if len(rows) == 0 {
			return errs.NotFound(msgReviewRowNotFound)
		}
		saved = rows[0].ReviewRow
		if saved.Status != change.from {
			return errs.NotAllowed(fmt.Sprintf(msgReviewNotChangeable, saved.Status))
		}
		saved.Status, saved.Comment = change.to, change.comment
		saved.Version++
		saved.UpdatedAt, saved.UpdatedBy = s.now(), a
		if putErr := s.repo.PutReviewRow(ctx, saved); putErr != nil {
			return putErr
		}
		_, logErr := s.history.Log(ctx, a, history.LogInput{
			Kind: reviewKind, RecordID: saved.ID, RecordName: saved.Reason,
			Action: history.ActionStatusChanged, Verb: change.verb,
		})
		return logErr
	})
	if err != nil {
		return ReviewRow{}, fmt.Errorf("change review row %s: %w", change.rowID, err)
	}
	return saved, nil
}
