package performance

import (
	"context"
	"math"
	"slices"
	"strings"
	"time"

	"mapp/internal/platform/money"
)

func stageFigures(goals GoalFacts, totals []StageTotals, suppliers []Supplier, words troubleWords) []StageFigure {
	actuals := map[string]*int64{}
	for _, total := range totals {
		if total.Stage != "" && total.Stage != stageRevenue {
			counts := total.Counts
			actuals[total.Stage] = sumInto(actuals[total.Stage], counts)
		}
	}
	inq := StageFigure{Stage: "inq", Goal: goals.INQ, Actual: actuals["inq"]}
	mql := StageFigure{Stage: "mql", Goal: goals.MQL, Actual: actuals["mql"]}
	mql.Conversion = conversionOf(mql.Actual, inq.Actual)
	inq.Reason = stageReason(inq, "inq", suppliers, words)
	mql.Reason = stageReason(mql, "mql", suppliers, words)
	return []StageFigure{inq, mql}
}

func sumInto(sum *int64, counts int64) *int64 {
	if sum == nil {
		return &counts
	}
	total := *sum + counts
	return &total
}

// conversionOf is measure-metrics' one definition: actual of a stage over actual of the one before, one
// decimal; absent - never 0.0 - when the earlier stage has nothing to divide by.
func conversionOf(actual, previous *int64) *float64 {
	if actual == nil || previous == nil || *previous == 0 {
		return nil
	}
	rate := math.Round(float64(*actual)/float64(*previous)*1000) / 10
	return &rate
}

func stageReason(figure StageFigure, stage string, suppliers []Supplier, words troubleWords) string {
	if figure.Actual != nil {
		return ""
	}
	if !supplied(suppliers, MeaningStage, stage) {
		return reasonNoSource
	}
	if figure.Goal == nil {
		return reasonNoGoal
	}
	return words.emptyReason()
}

func supplied(suppliers []Supplier, meaning Meaning, stage string) bool {
	for _, supplier := range suppliers {
		if supplier.Meaning == meaning && (meaning != MeaningStage || supplier.Stage == stage) {
			return true
		}
	}
	return false
}

// revenueFigure answers one converted USD total at the current rate, with the count of facts a missing
// rate excluded (046-C7, SC-010).
type panelInput struct {
	goals     GoalFacts
	totals    []StageTotals
	suppliers []Supplier
	words     troubleWords
}

func (s *Service) revenueFigure(ctx context.Context, in panelInput) RevenueFigure {
	goals, totals, suppliers, words := in.goals, in.totals, in.suppliers, in.words
	figure := RevenueFigure{Goal: goals.Revenue}
	rates := s.currentRates(ctx)
	sum := money.Amount{}
	counted := false
	for _, total := range totals {
		if total.Stage != stageRevenue {
			continue
		}
		converted, err := rates.Convert(money.Money{Amount: money.FromCents(total.AmountMinor), Currency: total.Currency}, money.BaseCurrency)
		if err != nil {
			figure.ExcludedCount += total.FactCount
			continue
		}
		sum = sum.Add(converted.Amount)
		counted = true
	}
	if counted {
		usd := money.USD(sum)
		figure.Actual = &usd
	}
	figure.Reason = revenueReason(figure, suppliers, words)
	return figure
}

func (s *Service) currentRates(ctx context.Context) money.Rates {
	if s.rates == nil {
		return money.Rates{}
	}
	rates, err := s.rates(ctx)
	if err != nil {
		return money.Rates{}
	}
	return rates
}

func revenueReason(figure RevenueFigure, suppliers []Supplier, words troubleWords) string {
	if figure.Actual != nil {
		return ""
	}
	if !supplied(suppliers, MeaningRevenue, "") {
		return reasonNoSource
	}
	if figure.Goal == nil {
		return reasonNoGoal
	}
	return words.emptyReason()
}

func metricFigures(totals []StageTotals) []MetricFigure {
	byMetric := map[WebMetric][]MonthCount{}
	for _, total := range totals {
		if total.WebMetric == "" {
			continue
		}
		metric := WebMetric(total.WebMetric)
		byMetric[metric] = append(byMetric[metric], MonthCount{Month: total.Month, Value: total.Counts})
	}
	figures := make([]MetricFigure, 0, len(byMetric))
	for metric, months := range byMetric {
		slices.SortFunc(months, func(a, b MonthCount) int { return strings.Compare(a.Month, b.Month) })
		figures = append(figures, MetricFigure{Metric: metric, Months: months})
	}
	slices.SortFunc(figures, func(a, b MetricFigure) int { return strings.Compare(string(a.Metric), string(b.Metric)) })
	return figures
}

// panelSources lists the contributing sources with their freshness, failed overlaid from the latest load
// (FR-029); data through comes with each.
func (s *Service) panelSources(ctx context.Context, activityID string) ([]PanelSource, error) {
	totals, err := s.repo.FactTotalsBySource(ctx, activityID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(totals))
	for _, total := range totals {
		ids = append(ids, total.SourceID)
	}
	failed, err := s.repo.LatestImportStatuses(ctx, ids)
	if err != nil {
		return nil, err
	}
	sources := make([]PanelSource, 0, len(totals))
	for _, total := range totals {
		view, viewErr := s.sourceView(ctx, total.SourceID)
		if viewErr != nil {
			return nil, viewErr
		}
		view.LastLoadFailed = failed[total.SourceID] == string(ImportFailed)
		src := PanelSource{SourceID: total.SourceID, SourceName: total.SourceName, Freshness: view.FreshnessAt(s.now())}
		if !view.DataThrough.IsZero() {
			through := view.DataThrough.Format(time.DateOnly)
			src.DataThrough = &through
		}
		sources = append(sources, src)
	}
	return sources, nil
}

func (s *Service) sourceView(ctx context.Context, sourceID string) (SourceView, error) {
	views, err := s.repo.ListSources(ctx, sourceID)
	if err != nil {
		return SourceView{}, err
	}
	if len(views) == 0 {
		return SourceView{}, nil
	}
	return views[0], nil
}
