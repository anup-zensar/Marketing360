package approvalreport

import (
	"slices"
	"strings"

	"mapp/internal/allwork"
	"mapp/internal/approval"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// record is what a row shows of the activity or purchase request a round is about.
type record struct {
	name   string
	owner  PersonRef
	planID string
	orders *string
	amount *money.Amount
}

// rowSource is everything the rows need beyond the rounds: the records, who may see amounts, and the currency.
type rowSource struct {
	records  map[string]record
	currency allwork.Currency
}

func recordsOf(scope measure.Scope, requests []PurchaseRequest) map[string]record {
	names := map[string]string{}
	for _, p := range scope.Choices.People {
		names[p.ID] = p.Label
	}
	records := map[string]record{}
	for _, n := range scope.Nodes {
		if n.Level != allwork.LevelActivity {
			continue
		}
		owner := PersonRef{Name: NothingText}
		if n.Owner != nil {
			owner = PersonRef{ID: n.Owner.ID, Name: n.Owner.Name}
		}
		records[n.ID] = record{name: n.Name, owner: owner, planID: n.PlanID, amount: n.PlannedCost}
	}
	for _, r := range requests {
		amount := r.Amount
		records[r.ID] = record{name: r.Name, planID: r.PlanID, amount: &amount,
			owner: PersonRef{ID: r.OwnerID, Name: names[r.OwnerID]}, orders: joined(r.PurchaseOrders)}
	}
	return records
}

func joined(orders []string) *string {
	if len(orders) == 0 {
		return nil
	}
	text := strings.Join(orders, ", ")
	return &text
}

// rowsOf makes one row per named approver at each level the engine did not skip (041-C2, C3).
func rowsOf(entries []entry, actions []approval.Action, src rowSource, sees func(planID string) bool) []Row {
	byRound := map[string][]approval.Action{}
	for _, a := range actions {
		byRound[a.RoundID] = append(byRound[a.RoundID], a)
	}
	rows := []Row{}
	for _, e := range entries {
		rec, found := src.records[e.round.SubjectID]
		if !found {
			continue
		}
		base := baseRow(e, rec, src.currency, sees(rec.planID))
		for _, level := range e.round.Levels {
			rows = append(rows, levelRows(base, e, level, byRound[e.round.ID])...)
		}
	}
	return rows
}

// baseRow marks every record openable: one is in view only when the person's own scope returned it.
func baseRow(e entry, rec record, currency allwork.Currency, seesMoney bool) Row {
	row := Row{ItemID: e.round.SubjectID, ItemName: rec.name, RecordType: e.kind, CanOpen: true,
		SubmittedOn: date.Today(e.round.SubmittedAt), PurchaseOrderNumber: rec.orders, Owner: rec.owner}
	if seesMoney && rec.amount != nil {
		row.Amount = &money.Money{Amount: money.FromUSD(*rec.amount, currency.Rate), Currency: currency.Code}
	}
	return row
}

func levelRows(base Row, e entry, level approval.RoundLevel, actions []approval.Action) []Row {
	base.Level, base.LevelName = level.Number, level.Name
	switch level.Resolution {
	case approval.ResolutionSkipped:
		return nil
	case approval.ResolutionAutoApproved:
		decided := date.Today(e.round.SubmittedAt)
		base.Outcome, base.DecidedOn = OutcomeAutoApproved, &decided
		return []Row{base}
	}
	decisions := decisionsAt(actions, e.round.ID, level.Number)
	slots := assign(level.Approvers, decisions)
	rows := make([]Row, 0, len(level.Approvers))
	for i, approver := range level.Approvers {
		row := base
		if slots[i] != nil {
			rows = append(rows, decidedRow(row, *slots[i]))
			continue
		}
		person := PersonRef{ID: approver.ID, Name: approver.Name}
		row.Approver, row.Outcome = &person, undecidedOutcome(e, level, len(decisions) > 0)
		rows = append(rows, row)
	}
	return rows
}

func decidedRow(row Row, a approval.Action) Row {
	person := PersonRef{ID: a.Actor.ID, Name: a.Actor.Name}
	decided := date.Today(a.At)
	row.Approver, row.DecidedOn, row.Comments = &person, &decided, a.Comment
	row.Outcome = OutcomeApproved
	if a.Action == approval.ActionDeclined {
		row.Outcome = OutcomeDeclined
	}
	return row
}

func decisionsAt(actions []approval.Action, roundID string, level int) []approval.Action {
	var found []approval.Action
	for _, a := range actions {
		if a.RoundID == roundID && a.Level == level && (a.Action == approval.ActionApproved || a.Action == approval.ActionDeclined) {
			found = append(found, a)
		}
	}
	return found
}

// undecidedOutcome is the word for a named approver with no decision of their own: Waiting while the level can
// still be decided, Not needed once it was decided or never reached, the round's own word if it was withdrawn.
func undecidedOutcome(e entry, level approval.RoundLevel, levelDecided bool) Outcome {
	open := e.outcome == OutcomeWaiting
	switch {
	case open && (!levelDecided || (level.Mode == approval.ModeAll && level.Number == e.round.CurrentLevel)):
		return OutcomeWaiting
	case !open && !levelDecided && (e.outcome == OutcomeRecalled || e.outcome == OutcomeReplaced):
		return e.outcome
	}
	return OutcomeNotNeeded
}

// assign gives each decision to the approver it was made by, then as their delegate or on their behalf, then to
// any approver left without one (a backup), so a decision is never shown twice.
func assign(approvers []actor.Actor, decisions []approval.Action) []*approval.Action {
	s := assignment{slots: make([]*approval.Action, len(approvers)), used: make([]bool, len(decisions))}
	s.take(approvers, decisions, madeBy)
	s.take(approvers, decisions, madeFor)
	s.fill(decisions)
	return s.slots
}

type assignment struct {
	slots []*approval.Action
	used  []bool
}

func (s *assignment) take(approvers []actor.Actor, decisions []approval.Action, matches func(approval.Action, actor.Actor) bool) {
	for i, approver := range approvers {
		if s.slots[i] != nil {
			continue
		}
		j := slices.IndexFunc(decisions, func(d approval.Action) bool { return matches(d, approver) })
		if j >= 0 && !s.used[j] {
			s.slots[i], s.used[j] = &decisions[j], true
		}
	}
}

func (s *assignment) fill(decisions []approval.Action) {
	for j := range decisions {
		if i := slices.Index(s.slots, nil); !s.used[j] && i >= 0 {
			s.slots[i], s.used[j] = &decisions[j], true
		}
	}
}

func madeBy(d approval.Action, a actor.Actor) bool {
	return d.Actor.ID == a.ID
}

func madeFor(d approval.Action, a actor.Actor) bool {
	if d.AsDelegateOf != nil && d.AsDelegateOf.ID == a.ID {
		return true
	}
	return slices.ContainsFunc(d.OnBehalfOf, func(b actor.Actor) bool { return b.ID == a.ID })
}
