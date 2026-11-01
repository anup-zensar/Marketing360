package measure

import (
	"cmp"
	"context"
	"slices"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/clock"
	"mapp/internal/platform/date"
	"mapp/internal/platform/ids"
)

// Service is what every report shares: the catalogue, the year a report covers and the report frame.
type Service struct {
	tx       Transactor
	ids      *ids.Generator
	log      *history.Service
	deps     Deps
	registry *Registry
	now      clock.Clock
}

func NewService(tools history.Tools, deps Deps, registry *Registry) *Service {
	return &Service{tx: tools.Tx, ids: tools.IDs, log: tools.Log, deps: deps, registry: registry, now: tools.Now}
}

func (s *Service) Catalogue(ctx context.Context) ReportCatalogue {
	return s.registry.Catalogue(ctx)
}

// Scope is the year a report covers: the plans in the reports.view scope, what matches, and the currency (R1, R2).
type Scope struct {
	Year        int
	Nodes       []allwork.Node
	Matched     allwork.Matched
	Choices     allwork.Choices
	Attributes  []allwork.Attribute
	Currency    allwork.Currency
	FinanceAsOf *time.Time
}

// Scope reads the scope of f inside the caller's Read, so a report and the frame cover the same records (FR-007).
func (s *Service) Scope(ctx context.Context, a actor.Actor, f Filter, columns []string) (Scope, error) {
	today := date.Today(s.now())
	year := cmp.Or(f.Query.Year, today.Year())
	data, err := s.deps.Planning.Year(ctx, year, today, columns)
	if err != nil {
		return Scope{}, err
	}
	inScope, err := s.deps.Plans.InReportsScope(ctx)
	if err != nil {
		return Scope{}, err
	}
	choices, err := s.deps.Planning.Choices(ctx, a.ID)
	if err != nil {
		return Scope{}, err
	}
	attributes, err := s.deps.Planning.Attributes(ctx)
	if err != nil {
		return Scope{}, err
	}
	currency, err := s.currency(ctx, a, f.Currency)
	nodes := slices.DeleteFunc(slices.Clone(data.Nodes), func(n allwork.Node) bool {
		return !slices.Contains(inScope, n.PlanID)
	})
	return Scope{Year: year, Nodes: nodes, Matched: allwork.Match(nodes, f.Query, attributes, choices), Choices: choices,
		Attributes: attributes, Currency: currency, FinanceAsOf: data.FinanceAsOf}, err
}

// currency is the chosen active currency at its current rate, or else the person's display currency (FR-018).
func (s *Service) currency(ctx context.Context, a actor.Actor, code string) (allwork.Currency, error) {
	if code != "" {
		active, err := s.deps.Currencies.Active(ctx)
		if err != nil {
			return allwork.Currency{}, err
		}
		if slices.ContainsFunc(active, func(c Currency) bool { return c.Code == code }) {
			rate, err := s.deps.Currencies.RateOf(ctx, code)
			return allwork.Currency{Code: code, Rate: rate}, err
		}
	}
	return s.deps.Planning.Currency(ctx, a.ID)
}
