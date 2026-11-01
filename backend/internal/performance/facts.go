package performance

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
)

// ActivityFacts lists an approved-or-not activity's facts (FR-030): the list itself is not gated on
// approval - only figures are (FR-027) - but the activity must be within the reader's scope.
func (s *Service) ActivityFacts(ctx context.Context, _ actor.Actor, q FactQuery) ([]FactView, error) {
	var facts []FactView
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		if _, readErr := s.visibleActivity(ctx, q.ActivityID); readErr != nil {
			return readErr
		}
		var readErr error
		facts, readErr = s.repo.ListFactsForActivity(ctx, authz.ScopeOf(ctx), q)
		return readErr
	})
	if err != nil {
		return nil, fmt.Errorf("list the facts of %s: %w", q.ActivityID, err)
	}
	return facts, nil
}

type ExportFile struct {
	Name        string
	ContentType string
	Body        []byte
}

// ExportFacts writes the list as CSV; an export leaves the building, so it needs data.export (D-367).
func (s *Service) ExportFacts(ctx context.Context, a actor.Actor, q FactQuery) (ExportFile, error) {
	if err := authz.Require(ctx, authz.DataExport); err != nil {
		return ExportFile{}, err
	}
	facts, err := s.ActivityFacts(ctx, a, q)
	if err != nil {
		return ExportFile{}, err
	}
	return ExportFile{
		Name:        fmt.Sprintf("performance-facts-%s.csv", strings.ToLower(q.ActivityID)),
		ContentType: "text/csv; charset=utf-8",
		Body:        factsCSV(facts),
	}, nil
}

func factsCSV(facts []FactView) []byte {
	var out bytes.Buffer
	writer := csv.NewWriter(&out)
	_ = writer.Write([]string{"Date", "Source", "Stage or metric", "Value", "Currency", "Key matched on", "Upload"})
	for _, fact := range facts {
		_ = writer.Write([]string{
			fact.FactDate.Format(time.DateOnly), fact.SourceName, measureWord(fact.Fact),
			valueWord(fact.Fact), currencyWord(fact.Fact), string(fact.MatchedOn), fact.ImportID,
		})
	}
	writer.Flush()
	return out.Bytes()
}

func measureWord(f Fact) string {
	if f.WebMetric != "" {
		return string(f.WebMetric)
	}
	return f.Stage
}

func valueWord(f Fact) string {
	if f.Amount != nil {
		return f.Amount.Amount.String()
	}
	if f.Count != nil {
		return strconv.FormatInt(*f.Count, 10)
	}
	return ""
}

func currencyWord(f Fact) string {
	if f.Amount != nil {
		return f.Amount.Currency
	}
	return ""
}
