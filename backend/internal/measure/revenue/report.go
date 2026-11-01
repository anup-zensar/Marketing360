package revenue

import (
	"context"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

// RankedRevenue is the ranked list of entries sorted by actual or projected revenue (FR-006, FR-012).
func (s *Service) RankedRevenue(ctx context.Context, a actor.Actor, f measure.Filter, q RankedQuery) (RevenueRanked, error) {
	var out RevenueRanked
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		sc, views, err := s.scopeAndActivities(ctx, a, f, categoryOf(f.Own))
		if err != nil {
			return err
		}
		show := showLevel(f.Show)
		out = revenueRanked(views, sc.Nodes, show, q)
		return nil
	})
	return out, err
}

func showLevel(show string) string {
	if show == "" {
		return measure.ShowCampaign
	}
	return show
}

// revenueReport is Revenue in 039's registry.
type revenueReport struct {
	service *Service
}

func NewRevenueReport(s *Service) measure.Report {
	return revenueReport{service: s}
}

func (revenueReport) Key() string     { return measure.KeyRevenue }
func (revenueReport) Available() bool { return true }

func (revenueReport) Params() []string {
	return []string{paramCategory, paramSort, paramOrder, paramGroup, paramSize, paramSearch, paramDrill}
}

func (r revenueReport) Sheets(_ context.Context, _ actor.Actor, _ measure.Filter) ([]measure.WidgetSheet, error) {
	return nil, nil
}

// funnelReport is Funnel stages in 039's registry, sharing Revenue's service (research R1).
type funnelReport struct {
	service *Service
}

func NewFunnelReport(s *Service) measure.Report {
	return funnelReport{service: s}
}

func (funnelReport) Key() string     { return measure.KeyFunnel }
func (funnelReport) Available() bool { return true }

func (funnelReport) Params() []string {
	return []string{paramCategory, paramStage, paramSort, paramOrder, paramGroup, paramSize, paramSearch, paramDrill}
}

func (r funnelReport) Sheets(_ context.Context, _ actor.Actor, _ measure.Filter) ([]measure.WidgetSheet, error) {
	return nil, nil
}
