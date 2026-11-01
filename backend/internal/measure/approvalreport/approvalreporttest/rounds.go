package approvalreporttest

import (
	"time"

	"mapp/internal/approval"
	"mapp/internal/platform/actor"
)

func level(number int, mode, resolution string, approvers ...actor.Actor) approval.RoundLevel {
	return approval.RoundLevel{Number: number, Name: "Level " + string(rune('0'+number)), Approvers: approvers,
		Mode: mode, Resolution: resolution}
}

func round(id, subject string, kind approval.SubjectKind, outcome approval.Outcome, submitted time.Time,
	levels ...approval.RoundLevel) approval.Round {
	return approval.Round{ID: id, SubjectKind: kind, SubjectID: subject, Number: 1, Levels: levels, CurrentLevel: 1,
		Outcome: outcome, SubmittedAt: submitted, SubmittedBy: Erin}
}

func ended(r approval.Round, at time.Time) approval.Round {
	r.EndedAt = &at
	return r
}

func roundsAt(now time.Time) []approval.Round {
	day := func(back int) time.Time { return now.AddDate(0, 0, -back) }
	act, decide := approval.KindActivity, approval.ResolutionDecide
	automatic := ended(round("r-automatic", ActivityAutomatic, act, approval.OutcomeApproved, day(4),
		level(1, approval.ModeAny, approval.ResolutionAutoApproved)), day(4))
	automatic.AutoApproved = true
	skipped := round("r-skipped", ActivitySkippedLevel, act, approval.OutcomeApproved, day(9),
		level(1, approval.ModeAny, approval.ResolutionSkipped, Alice), level(2, approval.ModeAny, decide, Bob))
	return []approval.Round{
		round("r-waiting", ActivityWaiting, act, approval.OutcomeOpen, day(6), level(1, approval.ModeAny, decide, Alice)),
		ended(round("r-approved", ActivityApprovedAny, act, approval.OutcomeApproved, day(10),
			level(1, approval.ModeAny, decide, Bob, Carol)), day(8)),
		ended(round("r-declined", ActivityDeclinedAll, act, approval.OutcomeDeclined, day(12),
			level(1, approval.ModeAll, decide, Dan, Erin)), day(11)),
		ended(round("r-recalled", ActivityRecalled, act, approval.OutcomeRecalled, day(7),
			level(1, approval.ModeAny, decide, Alice)), day(7)),
		ended(round("r-replaced", ActivityReplaced, act, approval.OutcomeRestarted, day(5),
			level(1, approval.ModeAny, decide, Alice)), day(5)),
		ended(round("r-cancelled", ActivityCancelled, act, approval.OutcomeCancelled, day(3),
			level(1, approval.ModeAny, decide, Alice)), day(2)),
		automatic,
		ended(skipped, day(8)),
		round("r-request", RequestWaiting, approval.KindPurchaseRequest, approval.OutcomeOpen, day(2),
			level(1, approval.ModeAny, decide, Carol)),
	}
}

func decision(id, roundID, subject string, kind approval.ActionKind, lvl int, by actor.Actor, at time.Time) approval.Action {
	return approval.Action{ID: id, RoundID: roundID, RoundNumber: 1, SubjectKind: approval.KindActivity,
		SubjectID: subject, Action: kind, Level: lvl, LevelName: "Level " + string(rune('0'+lvl)), Actor: by, At: at,
		Capacity: approval.CapacityApprover}
}

func actionsAt(now time.Time) []approval.Action {
	day := func(back int) time.Time { return now.AddDate(0, 0, -back) }
	approved := decision("a-approved", "r-approved", ActivityApprovedAny, approval.ActionApproved, 1, Bob, day(8))
	approved.Comment = "Fine by me"
	delegated := decision("a-declined", "r-declined", ActivityDeclinedAll, approval.ActionDeclined, 1, Erin, day(11))
	delegated.Capacity, delegated.AsDelegateOf, delegated.Comment = approval.CapacityDelegate, &Dan, "Out of budget"
	return []approval.Action{
		approved,
		delegated,
		decision("a-skipped", "r-skipped", ActivitySkippedLevel, approval.ActionSkipped, 1, actor.System, day(9)),
		decision("a-skipped-2", "r-skipped", ActivitySkippedLevel, approval.ActionApproved, 2, Bob, day(8)),
		decision("a-automatic", "r-automatic", ActivityAutomatic, approval.ActionAutoApproved, 1, actor.System, day(4)),
	}
}
