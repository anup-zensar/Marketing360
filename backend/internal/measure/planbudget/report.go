package planbudget

import (
	"context"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

// report is Plan budget in 039's registry: built, with its own address parameters for the ranked list (R5).
type report struct {
	service *Service
}

func NewReport(s *Service) measure.Report {
	return report{service: s}
}

func (report) Key() string     { return measure.KeyPlanBudget }
func (report) Available() bool { return true }

func (report) Params() []string {
	return []string{paramSort, paramOrder, paramGroup, paramSize, paramSearch, paramDrill}
}

func (r report) Sheets(ctx context.Context, a actor.Actor, f measure.Filter) ([]measure.WidgetSheet, error) {
	return r.service.sheets(ctx, a, f)
}
