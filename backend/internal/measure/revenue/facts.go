package revenue

import (
	"context"

	"mapp/internal/platform/authz"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
)

const stageRevenue = "revenue"

// activityTotals is one activity's counted facts, laid out in the engine's own shape so they line up month
// by month with its projected series (mirrors 046's activityprojection.go).
type activityTotals struct {
	counts  map[string][]projection.MonthFigure // by stage ("inq", "mql")
	revenue []projection.MoneyFigure
}

// factsByActivity reads the activities' counted facts for the plan year and lays each one out as a twelve-
// month series per stage, one query whatever the number of activities (research R2, SC-003). Revenue is
// converted to the current rate once per month; a currency with no rate is excluded, as 046-C7 has it.
func (s *Service) factsByActivity(ctx context.Context, scope authz.Scope, activityIDs []string, year int) (map[string]activityTotals, error) {
	if len(activityIDs) == 0 {
		return map[string]activityTotals{}, nil
	}
	rows, err := s.deps.Facts.ReportTotals(ctx, scope, activityIDs, year)
	if err != nil {
		return nil, err
	}
	rates, err := s.deps.Rates.Rates(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]activityTotals, len(activityIDs))
	for _, r := range rows {
		s.addRow(out, r, year, rates)
	}
	return out, nil
}

func (s *Service) addRow(out map[string]activityTotals, r FactTotal, year int, rates money.Rates) {
	t, ok := out[r.ActivityID]
	if !ok {
		t = activityTotals{counts: map[string][]projection.MonthFigure{}}
	}
	i := monthIndex(r.Month, year)
	if i < 0 {
		return
	}
	if r.Stage == stageRevenue {
		t.revenue = addRevenue(t.revenue, r, i, rates)
	} else {
		t.counts[r.Stage] = appendAt(t.counts[r.Stage], i, r.Counts, r.Month)
	}
	out[r.ActivityID] = t
}

func addRevenue(months []projection.MoneyFigure, r FactTotal, i int, rates money.Rates) []projection.MoneyFigure {
	if months == nil {
		months = emptyRevenueMonths(r.Month, i)
	}
	converted, err := rates.Convert(money.Money{Amount: money.FromCents(r.AmountMinor), Currency: r.Currency},
		money.BaseCurrency)
	if err != nil {
		return months
	}
	months[i].Value.Amount = months[i].Value.Amount.Add(converted.Amount)
	return months
}

func emptyRevenueMonths(sampleMonth string, atIndex int) []projection.MoneyFigure {
	months := make([]projection.MoneyFigure, projection.Months)
	year := sampleMonth[:4]
	for i := range months {
		months[i].Value = money.USD(money.Amount{})
		months[i].Month = monthKey(year, i)
	}
	_ = atIndex
	return months
}

func appendAt(months []projection.MonthFigure, i int, value int64, sampleMonth string) []projection.MonthFigure {
	if months == nil {
		months = make([]projection.MonthFigure, projection.Months)
		year := sampleMonth[:4]
		for j := range months {
			months[j].Month = monthKey(year, j)
		}
	}
	months[i].Value += value
	return months
}

func monthKey(year string, zeroBasedMonth int) string {
	m := zeroBasedMonth + 1
	digits := "0123456789"
	tens, ones := m/10, m%10
	return year + "-" + string(digits[tens]) + string(digits[ones])
}

// monthIndex places a fact month ("2026-01") in the plan year, or -1 when it belongs to another year.
func monthIndex(month string, year int) int {
	if len(month) < 7 {
		return -1
	}
	y := 0
	for _, r := range month[:4] {
		if r < '0' || r > '9' {
			return -1
		}
		y = y*10 + int(r-'0')
	}
	if y != year {
		return -1
	}
	m := 0
	for _, r := range month[5:7] {
		if r < '0' || r > '9' {
			return -1
		}
		m = m*10 + int(r-'0')
	}
	return m - 1
}
