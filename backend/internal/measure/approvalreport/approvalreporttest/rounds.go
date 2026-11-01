package approvalreporttest

import (
	"strconv"
	"time"

	"mapp/internal/approval"
	"mapp/internal/platform/actor"
)

func level(number int, mode, resolution string, approvers ...actor.Actor) approval.RoundLevel {
	return approval.RoundLevel{Number: number, Name: "Level " + strconv.Itoa(number), Approvers: approvers,
		Mode: mode, Resolution: resolution}
}

type roundSpec struct {
	id, subject string
	kind        approval.SubjectKind
	outcome     approval.Outcome
	submitted   time.Time
}

func round(s roundSpec, levels ...approval.RoundLevel) approval.Round {
	return approval.Round{ID: s.id, SubjectKind: s.kind, SubjectID: s.subject, Number: 1, Levels: levels, CurrentLevel: 1,
		Outcome: s.outcome, SubmittedAt: s.submitted, SubmittedBy: Erin()}
}

func ended(r approval.Round, at time.Time) approval.Round {
	r.EndedAt = &at
	return r
}

func roundsAt(now time.Time) []approval.Round {
	day := func(back int) time.Time { return now.AddDate(0, 0, -back) }
	activity := func(id, subject string, outcome approval.Outcome, submitted time.Time) roundSpec {
		return roundSpec{id, subject, approval.KindActivity, outcome, submitted}
	}
	decide, anyOf := approval.ResolutionDecide, approval.ModeAny
	automatic := ended(round(activity("r-automatic", ActivityAutomatic, approval.OutcomeApproved, day(4)),
		level(1, anyOf, approval.ResolutionAutoApproved)), day(4))
	automatic.AutoApproved = true
	skipped := round(activity("r-skipped", ActivitySkippedLevel, approval.OutcomeApproved, day(9)),
		level(1, anyOf, approval.ResolutionSkipped, Alice()), level(2, anyOf, decide, Bob()))
	request := roundSpec{"r-request", RequestWaiting, approval.KindPurchaseRequest, approval.OutcomeOpen, day(2)}
	return []approval.Round{
		round(activity("r-waiting", ActivityWaiting, approval.OutcomeOpen, day(6)), level(1, anyOf, decide, Alice())),
		ended(round(activity("r-approved", ActivityApprovedAny, approval.OutcomeApproved, day(10)),
			level(1, anyOf, decide, Bob(), Carol())), day(8)),
		ended(round(activity("r-declined", ActivityDeclinedAll, approval.OutcomeDeclined, day(12)),
			level(1, approval.ModeAll, decide, Dan(), Erin())), day(11)),
		ended(round(activity("r-recalled", ActivityRecalled, approval.OutcomeRecalled, day(7)),
			level(1, anyOf, decide, Alice())), day(7)),
		ended(round(activity("r-replaced", ActivityReplaced, approval.OutcomeRestarted, day(5)),
			level(1, anyOf, decide, Alice())), day(5)),
		ended(round(activity("r-cancelled", ActivityCancelled, approval.OutcomeCancelled, day(3)),
			level(1, anyOf, decide, Alice())), day(2)),
		automatic,
		ended(skipped, day(8)),
		round(request, level(1, anyOf, decide, Carol())),
	}
}

type decisionSpec struct {
	id, roundID, subject string
	kind                 approval.ActionKind
	level                int
	by                   actor.Actor
}

func decision(s decisionSpec, at time.Time) approval.Action {
	return approval.Action{ID: s.id, RoundID: s.roundID, RoundNumber: 1, SubjectKind: approval.KindActivity,
		SubjectID: s.subject, Action: s.kind, Level: s.level, LevelName: "Level " + strconv.Itoa(s.level), Actor: s.by, At: at,
		Capacity: approval.CapacityApprover}
}

func actionsAt(now time.Time) []approval.Action {
	day := func(back int) time.Time { return now.AddDate(0, 0, -back) }
	approved := decision(decisionSpec{"a-approved", "r-approved", ActivityApprovedAny, approval.ActionApproved, 1, Bob()}, day(8))
	approved.Comment = "Fine by me"
	delegated := decision(decisionSpec{"a-declined", "r-declined", ActivityDeclinedAll, approval.ActionDeclined, 1, Erin()}, day(11))
	dan := Dan()
	delegated.Capacity, delegated.AsDelegateOf, delegated.Comment = approval.CapacityDelegate, &dan, "Out of budget"
	return []approval.Action{
		approved,
		delegated,
		decision(decisionSpec{"a-skipped", "r-skipped", ActivitySkippedLevel, approval.ActionSkipped, 1, actor.System}, day(9)),
		decision(decisionSpec{"a-skipped-2", "r-skipped", ActivitySkippedLevel, approval.ActionApproved, 2, Bob()}, day(8)),
		decision(decisionSpec{"a-automatic", "r-automatic", ActivityAutomatic, approval.ActionAutoApproved, 1, actor.System}, day(4)),
	}
}
