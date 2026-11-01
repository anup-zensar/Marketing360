package approvalreport

import (
	"strconv"
	"time"

	"mapp/internal/approval"
)

// entry is one round as the figures count it.
type entry struct {
	round        approval.Round
	kind         RecordType
	outcome      Outcome
	waitingSince time.Time
}

func entriesOf(rounds []approval.Round, actions []approval.Action) []entry {
	approvedAt := map[string]time.Time{}
	for _, a := range actions {
		if a.Action == approval.ActionApproved {
			approvedAt[levelKey(a.RoundID, a.Level)] = a.At
		}
	}
	list := make([]entry, 0, len(rounds))
	for _, r := range rounds {
		e := entry{round: r, kind: recordTypeOf(r.SubjectKind), outcome: OutcomeWord(r)}
		if e.outcome == OutcomeWaiting {
			e.waitingSince = waitingSince(r, approvedAt)
		}
		list = append(list, e)
	}
	return list
}

// waitingSince is when the current level began to wait, as feature 036 counts it, so the two reports agree.
func waitingSince(r approval.Round, approvedAt map[string]time.Time) time.Time {
	if r.CurrentLevel < 1 || r.CurrentLevel > len(r.Levels) {
		return r.SubmittedAt
	}
	level := r.Levels[r.CurrentLevel-1]
	if level.StartedAt != nil {
		return *level.StartedAt
	}
	if previous, found := approvedAt[levelKey(r.ID, level.Number-1)]; found && level.Resolution == approval.ResolutionDecide && r.Facts == nil {
		return previous
	}
	return r.SubmittedAt
}

func levelKey(roundID string, level int) string {
	return roundID + "/" + strconv.Itoa(level)
}

func decidedByPerson(e entry) bool {
	approved := e.outcome == OutcomeApproved
	return (approved || e.outcome == OutcomeDeclined) && e.round.EndedAt != nil
}
