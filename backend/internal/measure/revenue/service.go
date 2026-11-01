package revenue

import (
	"context"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/clock"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
)

type Deps struct {
	Scoper     Scoper
	Facts      Facts
	Goals      Goals
	Categories Categories
	Rates      Rates
}

type Service struct {
	tx   Transactor
	deps Deps
	now  clock.Clock
}

func NewService(tx Transactor, deps Deps, now clock.Clock) *Service {
	return &Service{tx: tx, deps: deps, now: now}
}

// activityView is one activity with its projected series (research R3, R4) and its counted facts.
type activityView struct {
	node       allwork.Node
	projection projection.Projection
	facts      activityTotals
}

// scopeAndActivities resolves the scope, applies the category parameter (049-C10), and projects each activity.
func (s *Service) scopeAndActivities(ctx context.Context, a actor.Actor, f measure.Filter, category string) (
	measure.Scope, []activityView, error,
) {
	sc, err := s.deps.Scoper.Scope(ctx, a, f, nil)
	if err != nil {
		return measure.Scope{}, nil, err
	}
	categories, err := s.deps.Categories.ListCategories(ctx)
	if err != nil {
		return measure.Scope{}, nil, err
	}
	nodes := s.activitiesInView(sc, category, categories)
	activityIDs := idsOf(nodes)

	scope := authz.ScopeOf(ctx)
	goals, err := s.deps.Goals.ActivityGoals(ctx, scope, nodes)
	if err != nil {
		return measure.Scope{}, nil, err
	}
	facts, err := s.factsByActivity(ctx, scope, activityIDs, sc.Year)
	if err != nil {
		return measure.Scope{}, nil, err
	}
	today := date.Today(s.now())
	views := make([]activityView, 0, len(nodes))
	for _, n := range nodes {
		g := goals[n.ID]
		var revenueGoal *money.Money
		if n.Revenue != nil {
			usd := money.USD(*n.Revenue)
			revenueGoal = &usd
		}
		proj := projection.Project(projection.Inputs{
			Goals:     projection.Goals{INQ: n.INQ, MQL: n.MQL, Revenue: revenueGoal, SetLaterStagesMyself: g.SetLaterStagesMyself},
			Model:     g.Model,
			PlanYear:  sc.Year,
			StartDate: n.StartDate, EndDate: n.EndDate, Ongoing: n.Ongoing,
			Today: today,
		})
		views = append(views, activityView{node: n, projection: proj, facts: facts[n.ID]})
	}
	return sc, views, nil
}

// activitiesInView is the matched activities, narrowed to one category or "uncategorised" (049-C10).
func (s *Service) activitiesInView(sc measure.Scope, category string, categories []Category) []allwork.Node {
	typeCategory := map[string]string{}
	for _, c := range categories {
		for _, t := range c.ActivityTypeIDs {
			typeCategory[t] = c.ID
		}
	}
	byID := make(map[string]allwork.Node, len(sc.Nodes))
	for _, n := range sc.Nodes {
		byID[n.ID] = n
	}
	nodes := make([]allwork.Node, 0, len(sc.Matched.Activities))
	for _, id := range sc.Matched.Activities {
		n, found := byID[id]
		if !found || n.Level != allwork.LevelActivity {
			continue
		}
		if category != "" && !matchesCategory(n.ActivityTypeID, category, typeCategory) {
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes
}

func matchesCategory(activityTypeID, want string, typeCategory map[string]string) bool {
	id, has := typeCategory[activityTypeID]
	if want == categoryUncategorised {
		return !has
	}
	return has && id == want
}

const categoryUncategorised = "uncategorised"

func idsOf(nodes []allwork.Node) []string {
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n.ID)
	}
	return out
}

// nodesForRanking keeps the whole hierarchy but narrows the activity leaves to the ones in view (049-C10).
func nodesForRanking(sc measure.Scope, views []activityView) []allwork.Node {
	kept := make([]allwork.Node, 0, len(sc.Nodes))
	for _, n := range sc.Nodes {
		if n.Level != allwork.LevelActivity {
			kept = append(kept, n)
		}
	}
	for _, v := range views {
		kept = append(kept, v.node)
	}
	return kept
}
