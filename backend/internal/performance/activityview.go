package performance

import (
	"context"
	"fmt"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/money"
)

const (
	PanelOK          = "ok"
	PanelNotApproved = "notApproved"
	reasonNoSource   = "noSource"
	reasonNoGoal     = "noGoal"
	reasonFailed     = "failed"
	reasonStale      = "stale"
	reasonNoData     = "noData"
)

type StageFigure struct {
	Stage      string   `json:"stage"`
	Goal       *int64   `json:"goal"`
	Actual     *int64   `json:"actual"`
	Conversion *float64 `json:"conversion"`
	Reason     string   `json:"reason"`
}

type RevenueFigure struct {
	Goal          *money.Money `json:"goal"`
	Actual        *money.Money `json:"actual"`
	ExcludedCount int          `json:"excludedCount"`
	Reason        string       `json:"reason"`
}

type MonthCount struct {
	Month string `json:"month"`
	Value int64  `json:"value"`
}

type MetricFigure struct {
	Metric WebMetric    `json:"metric"`
	Months []MonthCount `json:"months"`
}

type PanelSource struct {
	SourceID    string    `json:"sourceId"`
	SourceName  string    `json:"sourceName"`
	Freshness   Freshness `json:"freshness"`
	DataThrough *string   `json:"dataThrough"`
}

// Panel is the activity's Performance area (FR-029). Projected is 047's and absent until it lands; the
// revenue model's name is 050's and arrives with it (Assumptions).
type Panel struct {
	State      string         `json:"state"`
	Stages     []StageFigure  `json:"stages"`
	Revenue    RevenueFigure  `json:"revenue"`
	WebMetrics []MetricFigure `json:"webMetrics"`
	Sources    []PanelSource  `json:"sources"`
}

func (s *Service) ActivityPanel(ctx context.Context, _ actor.Actor, activityID string) (Panel, error) {
	var panel Panel
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		if _, readErr := s.visibleActivity(ctx, activityID); readErr != nil {
			return readErr
		}
		approved, gateErr := s.countsAsApproved(ctx, activityID)
		if gateErr != nil {
			return gateErr
		}
		if !approved {
			panel = Panel{State: PanelNotApproved, Stages: []StageFigure{},
				Revenue: RevenueFigure{Reason: PanelNotApproved}, WebMetrics: []MetricFigure{}, Sources: []PanelSource{}}
			return nil
		}
		var buildErr error
		panel, buildErr = s.buildPanel(ctx, activityID)
		return buildErr
	})
	if err != nil {
		return Panel{}, fmt.Errorf("read the performance of %s: %w", activityID, err)
	}
	return panel, nil
}

// countsAsApproved applies 046-C1 through the same port the matcher holds to: unset, nothing is approved.
func (s *Service) countsAsApproved(ctx context.Context, activityID string) (bool, error) {
	if s.approvals == nil {
		return false, nil
	}
	approved, err := s.approvals.ApprovedIDs(ctx, []string{activityID})
	if err != nil {
		return false, err
	}
	return approved[activityID], nil
}

func (s *Service) buildPanel(ctx context.Context, activityID string) (Panel, error) {
	goals, err := s.goalFacts(ctx, activityID)
	if err != nil {
		return Panel{}, err
	}
	totals, err := s.repo.ActivityTotals(ctx, authz.ScopeOf(ctx), activityID, goals.PlanYear)
	if err != nil {
		return Panel{}, err
	}
	suppliers, err := s.repo.ActiveSuppliers(ctx, goals.ModelID)
	if err != nil {
		return Panel{}, err
	}
	sources, err := s.panelSources(ctx, activityID)
	if err != nil {
		return Panel{}, err
	}
	in := panelInput{goals: goals, totals: totals, suppliers: suppliers, words: sourceWords(sources)}
	panel := Panel{
		State:      PanelOK,
		Stages:     stageFigures(goals, totals, suppliers, in.words),
		Revenue:    s.revenueFigure(ctx, in),
		WebMetrics: metricFigures(totals),
		Sources:    sources,
	}
	return panel, nil
}

func (s *Service) goalFacts(ctx context.Context, activityID string) (GoalFacts, error) {
	if s.figures == nil {
		return GoalFacts{PlanYear: s.now().Year()}, nil
	}
	goals, found, err := s.figures.GoalFacts(ctx, activityID)
	if err != nil || !found {
		return GoalFacts{PlanYear: s.now().Year()}, err
	}
	return goals, nil
}

// sourceWords ranks the contributing sources' troubles: failed beats stale, and both only matter to a
// figure that has nothing to show (FR-032); otherwise they stay source annotations.
type troubleWords struct {
	failed bool
	stale  bool
}

func sourceWords(sources []PanelSource) troubleWords {
	words := troubleWords{}
	for _, src := range sources {
		words.failed = words.failed || src.Freshness == FreshnessFailed
		words.stale = words.stale || src.Freshness == FreshnessStale
	}
	return words
}

func (w troubleWords) emptyReason() string {
	switch {
	case w.failed:
		return reasonFailed
	case w.stale:
		return reasonStale
	}
	return reasonNoData
}
