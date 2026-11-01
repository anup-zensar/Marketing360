package performance

import (
	"context"
	"encoding/json"
	"strings"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

// persist writes what Classify decided: facts for the rows that carry data, replacing the version a
// repeated reference supersedes (FR-013), and review rows for the rejected and unmatched ones (FR-023).
func (s *Service) persist(ctx context.Context, a actor.Actor, run Import, results []RowResult) error {
	current, err := s.currentByIdentity(ctx, run.SourceID, results)
	if err != nil {
		return err
	}
	if err := s.supersedeAnsweredRows(ctx, a, run.SourceID, results); err != nil {
		return err
	}
	waiting, waitErr := s.openReviewRaws(ctx, run.SourceID)
	if waitErr != nil {
		return waitErr
	}
	writer := factWriter{s: s, a: a, run: run, current: current, waiting: waiting}
	for _, result := range results {
		switch result.Outcome {
		case OutcomeMatched, OutcomeHeld, OutcomeUnattributed:
			if err := writer.row(ctx, result); err != nil {
				return err
			}
		case OutcomeRejected, OutcomeUnmatched:
			if err := writer.review(ctx, result); err != nil {
				return err
			}
		}
	}
	return nil
}

// openReviewRaws keys the waiting rows - open and dismissed alike - by their raw content, so re-loading an
// identical file neither queues it twice nor resurrects a dismissal (SC-004).
func (s *Service) openReviewRaws(ctx context.Context, sourceID string) (map[string]bool, error) {
	rows, err := s.repo.ListReviewRows(ctx, ReviewFilter{StatusNot: string(ReviewResolved), SourceID: sourceID, Limit: maxRows})
	if err != nil {
		return nil, err
	}
	waiting := make(map[string]bool, len(rows))
	for _, row := range rows {
		waiting[string(row.Raw)] = true
	}
	return waiting, nil
}

// factWriter carries one load's write state, as 015's runner does, keeping each step within the limits.
type factWriter struct {
	s       *Service
	a       actor.Actor
	run     Import
	current map[identity]Fact
	waiting map[string]bool
}

type identity struct {
	reference string
	measure   Measure
}

func (s *Service) currentByIdentity(ctx context.Context, sourceID string, results []RowResult) (map[identity]Fact, error) {
	var references []string
	for _, result := range results {
		if len(result.Measured) > 0 {
			references = append(references, result.reference())
		}
	}
	facts, err := s.repo.CurrentFacts(ctx, sourceID, references)
	if err != nil {
		return nil, err
	}
	current := make(map[identity]Fact, len(facts))
	for _, f := range facts {
		current[identity{f.SourceReference, Measure{Stage: f.Stage, WebMetric: f.WebMetric}}] = f
	}
	return current, nil
}

func (w factWriter) row(ctx context.Context, result RowResult) error {
	for _, measured := range result.Measured {
		next := factOfResult(w.run, result, measured, record.NewMeta(w.a, w.s.now()))
		stored, exists := w.current[identity{next.SourceReference, measured.Measure}]
		if exists && unchanged(stored, next) {
			continue
		}
		id, err := w.s.ids.Next(ctx, ids.PrefixPerformanceFact)
		if err != nil {
			return err
		}
		next.ID = id
		if exists {
			stored.ReplacedByID = id
			stored.Version++
			stored.UpdatedAt, stored.UpdatedBy = w.s.now(), w.a
			if putErr := w.s.repo.PutFact(ctx, stored); putErr != nil {
				return putErr
			}
		}
		if putErr := w.s.repo.PutFact(ctx, next); putErr != nil {
			return putErr
		}
	}
	return nil
}

func factOfResult(run Import, result RowResult, measured Measured, meta record.Meta) Fact {
	return Fact{
		SourceID: run.SourceID, ImportID: run.ID, SourceReference: result.reference(),
		ActivityID: result.ActivityID, MatchedOn: result.MatchedOn, FactDate: result.FactDate,
		Stage: measured.Stage, WebMetric: measured.WebMetric,
		Count: measured.Count, Amount: measured.Amount,
		Held: result.Outcome == OutcomeHeld, Meta: meta,
	}
}

// unchanged compares the fields that make a fact, so re-uploading an identical file writes nothing (SC-004).
func unchanged(stored, next Fact) bool {
	sameAmount := (stored.Amount == nil) == (next.Amount == nil) &&
		(stored.Amount == nil || *stored.Amount == *next.Amount)
	sameCount := (stored.Count == nil) == (next.Count == nil) &&
		(stored.Count == nil || *stored.Count == *next.Count)
	return stored.ActivityID == next.ActivityID && stored.Held == next.Held &&
		stored.FactDate.Equal(next.FactDate) && sameCount && sameAmount
}

func (w factWriter) review(ctx context.Context, result RowResult) error {
	raw, err := json.Marshal(rawOf(result.Row))
	if err != nil {
		return err
	}
	if w.waiting[string(raw)] {
		return nil
	}
	id, err := w.s.ids.Next(ctx, ids.PrefixPerformanceReviewRow)
	if err != nil {
		return err
	}
	if putErr := w.s.repo.PutReviewRow(ctx, ReviewRow{
		ID: id, ImportID: w.run.ID, SourceID: w.run.SourceID, Raw: raw, Parsed: interpretedOf(result.Row),
		Reason: result.Reason, Faults: result.Faults, Candidates: result.Candidates,
		Status: ReviewOpen, Meta: record.NewMeta(w.a, w.s.now()),
	}); putErr != nil {
		return putErr
	}
	w.waiting[string(raw)] = true
	return nil
}

// interpretedOf keeps the load-time reading of the row, so a re-check never reinterprets it through the
// mappings of a later day (FR-007).
func interpretedOf(row Row) Interpreted {
	return Interpreted{
		Reference: row.Reference.Value, Date: row.Date.Value, Currency: row.Currency.Value,
		Keys: row.Keys, Readings: row.Readings,
	}
}

func rawOf(row Row) map[string]string {
	if row.Raw != nil {
		return row.Raw
	}
	return map[string]string{}
}

func (r RowResult) reference() string {
	return strings.TrimSpace(r.Row.Reference.Value)
}

// supersedeAnsweredRows closes the open rows a load answers, whatever the load's own verdicts: the source
// wins over a correction in progress (FR-026). An uncorrected row the file re-sent byte-identically stays
// waiting, so an identical re-upload still changes nothing (SC-004).
func (s *Service) supersedeAnsweredRows(ctx context.Context, a actor.Actor, sourceID string, results []RowResult) error {
	incoming := map[string]bool{}
	identical := map[string]bool{}
	for _, result := range results {
		if reference := result.reference(); reference != "" {
			incoming[reference] = true
		}
		if raw, err := json.Marshal(rawOf(result.Row)); err == nil {
			identical[string(raw)] = true
		}
	}
	open, err := s.repo.ListReviewRows(ctx, ReviewFilter{Status: string(ReviewOpen), SourceID: sourceID, Limit: maxRows})
	if err != nil {
		return err
	}
	for _, row := range open {
		if !answeredByLoad(row.ReviewRow, incoming, identical) {
			continue
		}
		if resolveErr := s.resolveRow(ctx, a, row.ReviewRow, supersededVerb); resolveErr != nil {
			return resolveErr
		}
	}
	return nil
}

func answeredByLoad(row ReviewRow, incoming, identical map[string]bool) bool {
	answered := incoming[strings.TrimSpace(row.Parsed.Reference)]
	resent := identical[string(row.Raw)] && !row.Parsed.Corrected
	return answered && !resent
}
