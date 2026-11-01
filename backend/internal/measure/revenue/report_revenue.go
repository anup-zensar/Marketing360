package revenue

import (
	"context"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// Report is Revenue against goal (US1): the cards and the period series for the activities in view.
func (s *Service) Report(ctx context.Context, a actor.Actor, f measure.Filter) (RevenueReport, error) {
	var out RevenueReport
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		sc, views, err := s.scopeAndActivities(ctx, a, f, categoryOf(f.Own))
		if err != nil {
			return err
		}
		out = s.buildRevenueReport(views, f.Period, sc)
		return nil
	})
	return out, err
}

func (s *Service) buildRevenueReport(views []activityView, period string, sc measure.Scope) RevenueReport {
	w := window{today: date.Today(s.now()), year: sc.Year}
	display := money.DisplayIn(sc.Currency.Code, sc.Currency.Rate.String())
	periods := revenuePeriods(revenueMonths(views, w), period, w, display)
	anyGoal := false
	for _, v := range views {
		if v.projection.Present {
			anyGoal = true
			break
		}
	}
	out := RevenueReport{Cards: revenueCards(periods, anyGoal), Period: periodShown(period), Periods: periods}
	if stamp := stampOf(views); stamp != nil {
		out.Projection = stamp
	}
	if len(views) == 0 {
		out.Reason = ReasonNoData
	} else if !anyGoal {
		out.Reason = ReasonNoGoal
	}
	return out
}

func periodShown(period string) string {
	if period == measure.PeriodQuarter {
		return measure.PeriodQuarter
	}
	return measure.PeriodMonth
}

// stampOf names the first model behind a projected figure, so a reviewer can reproduce any of them
// (047 FR-012). Reports mixing models name one; the ranked list's own entries carry the rest.
func stampOf(views []activityView) *Stamp {
	for _, v := range views {
		if v.projection.Present && v.projection.Stamp.ModelName != "" {
			return &Stamp{ModelName: v.projection.Stamp.ModelName, ModelChangedAt: v.projection.Stamp.ModelChangedAt.Format("2006-01-02T15:04:05Z07:00")}
		}
	}
	return nil
}
