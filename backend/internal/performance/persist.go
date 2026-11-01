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
	writer := factWriter{s: s, a: a, run: run, current: current}
	for _, result := range results {
		switch result.Outcome {
		case OutcomeMatched, OutcomeHeld, OutcomeUnattributed:
			if err := writer.row(ctx, result); err != nil {
				return err
			}
		case OutcomeRejected, OutcomeUnmatched:
			if err := s.writeReviewRow(ctx, a, run, result); err != nil {
				return err
			}
		}
	}
	return nil
}

// factWriter carries one load's write state, as 015's runner does, keeping each step within the limits.
type factWriter struct {
	s       *Service
	a       actor.Actor
	run     Import
	current map[identity]Fact
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

func (s *Service) writeReviewRow(ctx context.Context, a actor.Actor, run Import, result RowResult) error {
	id, err := s.ids.Next(ctx, ids.PrefixPerformanceReviewRow)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(rawOf(result.Row))
	if err != nil {
		return err
	}
	return s.repo.PutReviewRow(ctx, ReviewRow{
		ID: id, ImportID: run.ID, SourceID: run.SourceID, Raw: raw,
		Reason: result.Reason, Faults: result.Faults, Candidates: result.Candidates,
		Status: ReviewOpen, Meta: record.NewMeta(a, s.now()),
	})
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
