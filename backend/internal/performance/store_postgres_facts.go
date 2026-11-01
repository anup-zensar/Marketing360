package performance

import (
	"context"
	"fmt"
	"time"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/database/sqlcdb"
	"mapp/internal/platform/money"
)

func factOf(f sqlcdb.PerformanceFact) Fact {
	fact := Fact{
		ID: f.ID, SourceID: f.SourceID, ImportID: f.ImportID, SourceReference: f.SourceReference,
		ActivityID: text(f.ActivityID), MatchedOn: MatchedOn(f.MatchedOn), MatchedKey: f.MatchedKey, FactDate: f.FactDate,
		Stage: f.Stage, WebMetric: WebMetric(f.WebMetric), Count: f.CountValue,
		Held: f.Held, ReplacedByID: text(f.ReplacedByID),
		Meta: metaOf(f.Version, metaRow{f.CreatedAt, f.CreatedByID, f.CreatedByName,
			f.UpdatedAt, f.UpdatedByID, f.UpdatedByName}),
	}
	if f.AmountMinor != nil && f.Currency != nil {
		fact.Amount = &money.Money{Amount: money.FromCents(*f.AmountMinor), Currency: *f.Currency}
	}
	return fact
}

func factParams(f Fact) sqlcdb.PutPerformanceFactParams {
	p := sqlcdb.PutPerformanceFactParams{
		ID: f.ID, SourceID: f.SourceID, ImportID: f.ImportID, SourceReference: f.SourceReference,
		ActivityID: optional(f.ActivityID), MatchedOn: string(f.MatchedOn), MatchedKey: f.MatchedKey, FactDate: f.FactDate,
		Stage: f.Stage, WebMetric: string(f.WebMetric), CountValue: f.Count,
		Held: f.Held, ReplacedByID: optional(f.ReplacedByID),
		Version: f.Version, CreatedAt: f.CreatedAt, CreatedByID: f.CreatedBy.ID,
		CreatedByName: f.CreatedBy.Name, UpdatedAt: f.UpdatedAt, UpdatedByID: f.UpdatedBy.ID,
		UpdatedByName: f.UpdatedBy.Name,
	}
	if f.Amount != nil {
		cents := f.Amount.Amount.Cents()
		currency := f.Amount.Currency
		p.AmountMinor, p.Currency = &cents, &currency
	}
	return p
}

func (s *PostgresStore) ListFactsForActivity(ctx context.Context, scope authz.Scope, q FactQuery) ([]FactView, error) {
	a := argsOf(scope)
	rows, err := s.db.Queries(ctx).ListFactsForActivity(ctx, sqlcdb.ListFactsForActivityParams{
		ActivityID: optional(q.ActivityID), CountedOnly: q.CountedOnly, Search: optional(q.Search),
		MaxRows: rowLimit(q.Limit), SeesEverything: a.seesEverything,
		ViewerID: a.viewerID, ViewsPlans: a.viewsPlans, ViewerTeamIds: a.teamIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("select the facts of activity %s: %w", q.ActivityID, err)
	}
	facts := make([]FactView, 0, len(rows))
	for _, r := range rows {
		facts = append(facts, FactView{Fact: factOf(r.PerformanceFact), SourceName: r.SourceName})
	}
	return facts, nil
}

func (s *PostgresStore) PutFact(ctx context.Context, f Fact) error {
	if err := s.db.Queries(ctx).PutPerformanceFact(ctx, factParams(f)); err != nil {
		return fmt.Errorf("save performance fact %s: %w", f.ID, err)
	}
	return nil
}

func (s *PostgresStore) CurrentFacts(ctx context.Context, sourceID string, references []string) ([]Fact, error) {
	rows, err := s.db.Queries(ctx).CurrentFactsByReferences(ctx, sqlcdb.CurrentFactsByReferencesParams{
		SourceID: sourceID, SourceReferences: references,
	})
	if err != nil {
		return nil, fmt.Errorf("select the current facts of %s: %w", sourceID, err)
	}
	facts := make([]Fact, 0, len(rows))
	for _, r := range rows {
		facts = append(facts, factOf(r.PerformanceFact))
	}
	return facts, nil
}

// ReleaseHeld clears the hold on an activity's facts in one statement, inside the caller's transaction, so
// approval and release commit together (research R4).
func (s *PostgresStore) ReleaseHeld(ctx context.Context, activityID string, by actor.Actor, at time.Time) (int, error) {
	released, err := s.db.Queries(ctx).ReleaseHeldFactsForActivity(ctx, sqlcdb.ReleaseHeldFactsForActivityParams{
		ActivityID: optional(activityID), UpdatedAt: at, UpdatedByID: by.ID, UpdatedByName: by.Name,
	})
	if err != nil {
		return 0, fmt.Errorf("release the held facts of activity %s: %w", activityID, err)
	}
	return int(released), nil
}

func (s *PostgresStore) ListFactsOfImport(ctx context.Context, importID string) ([]Fact, error) {
	rows, err := s.db.Queries(ctx).ListFactsOfImport(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("select the facts of import %s: %w", importID, err)
	}
	facts := make([]Fact, 0, len(rows))
	for _, r := range rows {
		facts = append(facts, factOf(r.PerformanceFact))
	}
	return facts, nil
}

func (s *PostgresStore) FactsReplacedBy(ctx context.Context, replacerIDs []string) ([]Fact, error) {
	rows, err := s.db.Queries(ctx).FactsReplacedBy(ctx, replacerIDs)
	if err != nil {
		return nil, fmt.Errorf("select the facts the import replaced: %w", err)
	}
	facts := make([]Fact, 0, len(rows))
	for _, r := range rows {
		facts = append(facts, factOf(r.PerformanceFact))
	}
	return facts, nil
}

func (s *PostgresStore) HeldActivityIDs(ctx context.Context, sourceID string) ([]string, error) {
	ids, err := s.db.Queries(ctx).HeldActivityIDsOfSource(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("select the held activities of %s: %w", sourceID, err)
	}
	return ids, nil
}

// StageTotals is the panel's one aggregate: a row per (measure, month, currency) of counted facts.
type StageTotals struct {
	Stage       string
	WebMetric   string
	Month       string
	Currency    string
	Counts      int64
	AmountMinor int64
	FactCount   int
}

func (s *PostgresStore) ActivityTotals(ctx context.Context, scope authz.Scope, activityID string, year int) ([]StageTotals, error) {
	a := argsOf(scope)
	rows, err := s.db.Queries(ctx).ActivityFactTotals(ctx, sqlcdb.ActivityFactTotalsParams{
		ActivityID: activityID, Year: year,
		SeesEverything: a.seesEverything, ViewerID: a.viewerID,
		ViewsPlans: a.viewsPlans, ViewerTeamIds: a.teamIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("sum the facts of activity %s: %w", activityID, err)
	}
	totals := make([]StageTotals, 0, len(rows))
	for _, r := range rows {
		totals = append(totals, StageTotals{
			Stage: r.Stage, WebMetric: r.WebMetric, Month: r.Month.Format(monthOnly), Currency: r.Currency,
			Counts: r.Counts, AmountMinor: r.AmountMinor, FactCount: int(r.FactCount),
		})
	}
	return totals, nil
}

const monthOnly = "2006-01"

func (s *PostgresStore) LatestImportStatuses(ctx context.Context, sourceIDs []string) (map[string]string, error) {
	rows, err := s.db.Queries(ctx).LatestImportStatuses(ctx, sourceIDs)
	if err != nil {
		return nil, fmt.Errorf("select the latest load of %d sources: %w", len(sourceIDs), err)
	}
	latest := make(map[string]string, len(rows))
	for _, r := range rows {
		latest[r.SourceID] = r.Status
	}
	return latest, nil
}
