package performance

import (
	"context"
	"encoding/json"
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
	f.KeyText = likeSafe(f.KeyText)
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
	// All or none (046-C15): a partly hidden list would leak, through its length, how many activities
	// exist beyond the reader's sight.
	for i, row := range rows {
		rows[i].Candidates = allVisibleOrNone(row.Candidates, visible)
	}
	return rows, nil
}

func allVisibleOrNone(candidates []string, visible map[string][]string) []string {
	for _, id := range candidates {
		if len(visible[normalKey(MatchedActivityID, id)]) == 0 {
			return []string{}
		}
	}
	return nonNilStrings(candidates)
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

// CorrectInput edits the row's load-time interpretation (FR-007): cells may be fixed, the reference may
// not - a new reference is a new identity, which only a fresh upload may create (FR-013).
type CorrectInput struct {
	Date     *string       `json:"date"`
	Currency *string       `json:"currency"`
	Keys     *RowKeys      `json:"keys"`
	Values   []CorrectCell `json:"values"`
}

type CorrectCell struct {
	Column string `json:"column"`
	Value  string `json:"value"`
}

func (s *Service) Correct(ctx context.Context, a actor.Actor, rowID string, in CorrectInput) (ReviewRow, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return ReviewRow{}, err
	}
	var saved ReviewRow
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		var correctErr error
		saved, correctErr = s.correct(ctx, a, rowID, in)
		return correctErr
	})
	if err != nil {
		return ReviewRow{}, fmt.Errorf("correct review row %s: %w", rowID, err)
	}
	return saved, nil
}

func (s *Service) correct(ctx context.Context, a actor.Actor, rowID string, in CorrectInput) (ReviewRow, error) {
	row, err := s.correctableRow(ctx, rowID, in)
	if err != nil {
		return ReviewRow{}, err
	}
	row.Parsed = corrected(row.Parsed, in)
	row.Parsed.Corrected = true
	row.Version++
	row.UpdatedAt, row.UpdatedBy = s.now(), a
	if putErr := s.repo.PutReviewRow(ctx, row); putErr != nil {
		return ReviewRow{}, putErr
	}
	if _, logErr := s.history.Log(ctx, a, history.LogInput{
		Kind: reviewKind, RecordID: row.ID, RecordName: row.Reason, Action: history.ActionChanged,
	}); logErr != nil {
		return ReviewRow{}, logErr
	}
	if recheckErr := s.recheckOne(ctx, a, row); recheckErr != nil {
		return ReviewRow{}, recheckErr
	}
	latest, err := s.repo.ListReviewRows(ctx, ReviewFilter{ID: rowID, Limit: 1})
	if err != nil || len(latest) == 0 {
		return row, err
	}
	return latest[0].ReviewRow, nil
}

func (s *Service) recheckOne(ctx context.Context, a actor.Actor, row ReviewRow) error {
	raw := map[string]string{}
	_ = json.Unmarshal(row.Raw, &raw)
	counts := Counts{}
	rows := []Row{row.Parsed.Row(raw)}
	env, err := s.recheckEnv(ctx, row.SourceID, rows)
	if err != nil {
		return err
	}
	results := Classify(rows, env)
	current, err := s.currentByIdentity(ctx, row.SourceID, results)
	if err != nil {
		return err
	}
	return s.applyRecheck(ctx, a, recheckOf{row: row, current: current, counts: &counts}, results[0])
}

func corrected(parsed Interpreted, in CorrectInput) Interpreted {
	if in.Date != nil {
		parsed.Date = *in.Date
	}
	if in.Currency != nil {
		parsed.Currency = *in.Currency
	}
	if in.Keys != nil {
		parsed.Keys = *in.Keys
	}
	for _, value := range in.Values {
		for i, reading := range parsed.Readings {
			if reading.Column == value.Column {
				parsed.Readings[i].Value = value.Value
			}
		}
	}
	return parsed
}

// likeSafe caps the key filter and escapes ILIKE's wildcards, so a typed % cannot scan everything.
func likeSafe(text string) string {
	if len(text) > maxKeyFilterLength {
		text = text[:maxKeyFilterLength]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(text)
}

const maxKeyFilterLength = 100

func (s *Service) correctableRow(ctx context.Context, rowID string, in CorrectInput) (ReviewRow, error) {
	rows, err := s.repo.ListReviewRows(ctx, ReviewFilter{ID: rowID, Limit: 1})
	if err != nil {
		return ReviewRow{}, err
	}
	if len(rows) == 0 {
		return ReviewRow{}, errs.NotFound(msgReviewRowNotFound)
	}
	row := rows[0].ReviewRow
	if row.Status != ReviewOpen {
		return ReviewRow{}, errs.NotAllowed(fmt.Sprintf(msgReviewNotChangeable, row.Status))
	}
	if unknown := unknownColumns(row.Parsed, in.Values); len(unknown) > 0 {
		return ReviewRow{}, errs.Validation(errs.FieldError{
			Field: "values", Message: fmt.Sprintf(msgUnknownColumn, strings.Join(unknown, ", "))})
	}
	return row, nil
}

// unknownColumns refuses a correction aimed at a column the row never carried: it would vanish silently.
func unknownColumns(parsed Interpreted, values []CorrectCell) []string {
	known := map[string]bool{}
	for _, reading := range parsed.Readings {
		known[reading.Column] = true
	}
	var unknown []string
	for _, value := range values {
		if !known[value.Column] {
			unknown = append(unknown, value.Column)
		}
	}
	return unknown
}
