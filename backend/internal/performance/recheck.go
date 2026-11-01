package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
)

// RecheckSource re-checks a source's open rows and held facts on demand (FR-025).
func (s *Service) RecheckSource(ctx context.Context, a actor.Actor, sourceID string) (Counts, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return Counts{}, err
	}
	return s.RecheckAfterLoad(ctx, a, sourceID)
}

func (s *Service) RecheckAfterLoad(ctx context.Context, a actor.Actor, sourceID string) (Counts, error) {
	var resolved Counts
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		if _, loadErr := s.loadableSource(ctx, sourceID); loadErr != nil {
			return loadErr
		}
		var recheckErr error
		resolved, recheckErr = s.recheck(ctx, a, sourceID)
		if recheckErr != nil {
			return recheckErr
		}
		return s.releaseApprovedHolds(ctx, a, sourceID)
	})
	if err != nil {
		return Counts{}, fmt.Errorf("re-check source %s: %w", sourceID, err)
	}
	return resolved, nil
}

// releaseApprovedHolds is FR-025's second half: a release the approval hook never saw - a restore from the
// bin, above all - still happens here.
func (s *Service) releaseApprovedHolds(ctx context.Context, a actor.Actor, sourceID string) error {
	if s.approvals == nil {
		return nil
	}
	held, err := s.repo.HeldActivityIDs(ctx, sourceID)
	if err != nil {
		return err
	}
	approved, err := s.approvals.ApprovedIDs(ctx, held)
	if err != nil {
		return err
	}
	for _, activityID := range held {
		if !approved[activityID] {
			continue
		}
		if _, releaseErr := s.repo.ReleaseHeld(ctx, activityID, a, s.now()); releaseErr != nil {
			return releaseErr
		}
	}
	return nil
}

// recheck classifies each batch from its stored load-time interpretation, never today's mappings (FR-007).
func (s *Service) recheck(ctx context.Context, a actor.Actor, sourceID string) (Counts, error) {
	counts := Counts{}
	seen := map[string]bool{}
	for {
		rows, err := s.openBatch(ctx, sourceID, seen)
		if err != nil || len(rows) == 0 {
			return counts, err
		}
		if err := s.recheckBatch(ctx, a, rows, &counts); err != nil {
			return counts, err
		}
	}
}

// openBatch pages by what the last pass left open, skipping undone imports; a queue over maxRows that will
// not shrink re-checks its newest page only (noted in the PR).
func (s *Service) openBatch(ctx context.Context, sourceID string, seen map[string]bool) ([]ReviewRowView, error) {
	page, err := s.repo.ListReviewRows(ctx, ReviewFilter{Status: string(ReviewOpen), SourceID: sourceID, Limit: maxRows})
	if err != nil {
		return nil, err
	}
	undone, err := s.undoneImports(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	fresh := page[:0:0]
	for _, row := range page {
		if !seen[row.ID] && !undone[row.ImportID] {
			seen[row.ID] = true
			fresh = append(fresh, row)
		}
	}
	return fresh, nil
}

func (s *Service) undoneImports(ctx context.Context, sourceID string) (map[string]bool, error) {
	imports, err := s.repo.ListImports(ctx, "", sourceID, maxRows)
	if err != nil {
		return nil, err
	}
	undone := map[string]bool{}
	for _, run := range imports {
		if run.Status == ImportUndone {
			undone[run.ID] = true
		}
	}
	return undone, nil
}

func (s *Service) recheckBatch(ctx context.Context, a actor.Actor, views []ReviewRowView, counts *Counts) error {
	views, err := s.newestPerReference(ctx, a, views)
	if err != nil {
		return err
	}
	rows := make([]Row, 0, len(views))
	for _, view := range views {
		raw := map[string]string{}
		_ = json.Unmarshal(view.Raw, &raw)
		rows = append(rows, view.Parsed.Row(raw))
	}
	env, err := s.recheckEnv(ctx, views[0].SourceID, rows)
	if err != nil {
		return err
	}
	results := Classify(rows, env)
	current, err := s.currentByIdentity(ctx, views[0].SourceID, results)
	if err != nil {
		return err
	}
	for i, result := range results {
		at := recheckOf{row: views[i].ReviewRow, current: current, counts: counts}
		if err := s.applyRecheck(ctx, a, at, result); err != nil {
			return err
		}
	}
	return nil
}

type recheckOf struct {
	row     ReviewRow
	current map[identity]Fact
	counts  *Counts
}

// recheckEnv matches with the organisation-wide scope: a re-check restates matching for everyone, so its
// outcome must not depend on who pressed the button (deviation from load-time matching, noted in the PR).
func (s *Service) recheckEnv(ctx context.Context, sourceID string, rows []Row) (Env, error) {
	currencies, err := s.repo.CurrencyCodes(ctx)
	if err != nil {
		return Env{}, err
	}
	resolved, err := s.resolveWith(ctx, authz.Scope{Everything: true}, sourceID, rows)
	if err != nil {
		return Env{}, err
	}
	approved, err := s.approvedOf(ctx, resolved)
	if err != nil {
		return Env{}, err
	}
	return Env{Today: s.now(), Currencies: currencies, Resolved: resolved, Approved: approved}, nil
}

// newestPerReference supersedes all but the newest row per reference, so the in-file rule stays in-file.
func (s *Service) newestPerReference(ctx context.Context, a actor.Actor, views []ReviewRowView) ([]ReviewRowView, error) {
	kept := make([]ReviewRowView, 0, len(views))
	byReference := map[string]int{}
	for _, view := range views {
		reference := strings.TrimSpace(view.Parsed.Reference)
		if reference == "" {
			kept = append(kept, view)
			continue
		}
		at, found := byReference[reference]
		if !found {
			byReference[reference] = len(kept)
			kept = append(kept, view)
			continue
		}
		older := kept[at]
		if view.CreatedAt.After(older.CreatedAt) {
			kept[at] = view
			view = older
		}
		if err := s.resolveRow(ctx, a, view.ReviewRow, supersededVerb); err != nil {
			return nil, err
		}
	}
	return kept, nil
}

// applyRecheck writes insert-only: a current fact is never replaced from here, because a later load owns
// the identity and the source wins (FR-026); the row resolves as superseded instead.
func (s *Service) applyRecheck(ctx context.Context, a actor.Actor, at recheckOf, result RowResult) error {
	row := at.row
	switch result.Outcome {
	case OutcomeMatched, OutcomeHeld, OutcomeUnattributed:
		verb := resolvedVerb
		if superseded(at.current, result) {
			verb = supersededVerb
		} else {
			writer := factWriter{s: s, a: a, run: Import{ID: row.ImportID, SourceID: row.SourceID},
				current: at.current, waiting: map[string]bool{}}
			if err := writer.row(ctx, result); err != nil {
				return err
			}
			at.counts.add(result.Outcome)
		}
		return s.resolveRow(ctx, a, row, verb)
	case OutcomeRejected, OutcomeUnmatched:
		return s.restateRow(ctx, a, row, result)
	}
	return nil
}

// superseded reports a current fact that differs from what the row would write: a later load owns it.
func superseded(current map[identity]Fact, result RowResult) bool {
	for _, measured := range result.Measured {
		stored, exists := current[identity{result.reference(), measured.Measure}]
		if !exists {
			continue
		}
		next := factOfResult(Import{}, result, measured, stored.Meta)
		next.SourceReference = result.reference()
		if !unchanged(stored, next) {
			return true
		}
	}
	return false
}

// restateRow records a changed verdict with its history entry (FR-045); an unchanged one stays quiet.
func (s *Service) restateRow(ctx context.Context, a actor.Actor, row ReviewRow, result RowResult) error {
	sameFaults := fmt.Sprint(row.Faults) == fmt.Sprint(result.Faults)
	sameCandidates := fmt.Sprint(row.Candidates) == fmt.Sprint(result.Candidates)
	if row.Reason == result.Reason && sameFaults && sameCandidates {
		return nil
	}
	row.Reason, row.Faults, row.Candidates = result.Reason, result.Faults, result.Candidates
	row.Version++
	row.UpdatedAt, row.UpdatedBy = s.now(), a
	if err := s.repo.PutReviewRow(ctx, row); err != nil {
		return err
	}
	_, err := s.history.Log(ctx, a, history.LogInput{
		Kind: reviewKind, RecordID: row.ID, RecordName: row.Reason, Action: history.ActionChanged,
	})
	return err
}

func (s *Service) resolveRow(ctx context.Context, a actor.Actor, row ReviewRow, verb string) error {
	row.Status, row.Comment = ReviewResolved, ""
	row.Version++
	row.UpdatedAt, row.UpdatedBy = s.now(), a
	if err := s.repo.PutReviewRow(ctx, row); err != nil {
		return err
	}
	_, err := s.history.Log(ctx, a, history.LogInput{
		Kind: reviewKind, RecordID: row.ID, RecordName: row.Reason,
		Action: history.ActionStatusChanged, Verb: verb,
	})
	return err
}
