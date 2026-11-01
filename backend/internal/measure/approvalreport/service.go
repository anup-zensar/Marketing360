package approvalreport

import (
	"cmp"
	"context"
	"slices"

	"mapp/internal/approval"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

const defaultPageSize = 50

type Service struct {
	tx   Transactor
	deps Deps
}

func NewService(tx Transactor, deps Deps) *Service {
	return &Service{tx: tx, deps: deps}
}

// snapshot is what one Read of the plans in view finds: every round with its history, and the plans' finance cover.
type snapshot struct {
	entries  []entry
	actions  []approval.Action
	covers   FinanceCovers
	rows     []Row
	currency string
}

// Report works the report out in one Read: the hub's scope, then one read of the rounds and one of the history.
func (s *Service) Report(ctx context.Context, a actor.Actor, f measure.Filter, q Query) (View, error) {
	sn, err := s.snapshot(ctx, a, f)
	if err != nil {
		return View{}, err
	}
	return sn.view(q), nil
}

func (s *Service) snapshot(ctx context.Context, a actor.Actor, f measure.Filter) (snapshot, error) {
	var sn snapshot
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		scope, err := s.deps.Scope.Scope(ctx, a, f, nil)
		if err != nil {
			return err
		}
		requests, err := s.deps.Purchases.OnLineItems(ctx, scope.Matched.LineItems)
		if err != nil {
			return err
		}
		subjects := slices.Clone(scope.Matched.Activities)
		planOf := planOfActivities(scope)
		for _, r := range requests {
			subjects = append(subjects, r.ID)
			planOf[r.ID] = r.PlanID
		}
		rounds, err := s.deps.Rounds.RoundsOfSubjects(ctx, subjects)
		if err != nil {
			return err
		}
		if sn.actions, err = s.deps.Rounds.ActionsOfSubjects(ctx, subjects); err != nil {
			return err
		}
		sn.entries = entriesOf(rounds, sn.actions)
		sn.covers = financeCovers(scope, sn.entries, planOf)
		sn.currency = scope.Currency.Code
		sees := func(plan string) bool { return seesFinance(scope, plan) }
		sn.rows = rowsOf(sn.entries, sn.actions, rowSource{records: recordsOf(scope, requests), currency: scope.Currency}, sees)
		return nil
	})
	return sn, err
}

func (sn snapshot) view(q Query) View {
	q.StatusType = cmp.Or(q.StatusType, RecordActivity)
	return View{StatusMix: statusMix(sn.entries, q.StatusType), Average: averages(sn.entries, sn.actions),
		Grid: sn.grid(q), FinanceCovers: sn.covers}
}

func planOfActivities(scope measure.Scope) map[string]string {
	planOf := map[string]string{}
	for _, n := range scope.Nodes {
		planOf[n.ID] = n.PlanID
	}
	return planOf
}

// financeCovers counts the plans that have approval rounds, and those whose amounts the person may see.
func financeCovers(scope measure.Scope, entries []entry, planOf map[string]string) FinanceCovers {
	plans := map[string]bool{}
	for _, e := range entries {
		plans[planOf[e.round.SubjectID]] = true
	}
	covers := FinanceCovers{Of: len(plans)}
	for plan := range plans {
		if seesFinance(scope, plan) {
			covers.Plans++
		}
	}
	return covers
}

func seesFinance(scope measure.Scope, plan string) bool {
	return scope.Choices.SeesFinance && (scope.Choices.FinancePlans == nil || slices.Contains(scope.Choices.FinancePlans, plan))
}

func (sn snapshot) grid(q Query) Grid {
	grid := pageOf(matching(sn.rows, q), q)
	grid.Choices = choicesOf(sn.rows)
	return grid
}
