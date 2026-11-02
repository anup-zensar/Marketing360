package approvalreport

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"mapp/internal/approval"
)

// averages gives the overall figure of Approvals › Monitoring's rule (041-C7) and, per person, the time from the
// submission to that person's own decision.
func averages(entries []entry, actions []approval.Action) Average {
	var decisions, seconds int64
	submitted := map[string]time.Time{}
	for _, e := range entries {
		submitted[e.round.ID] = e.round.SubmittedAt
		if decidedByPerson(e) {
			decisions++
			seconds += int64(e.round.EndedAt.Sub(e.round.SubmittedAt) / time.Second)
		}
	}
	return Average{Days: approval.AverageDecisionDays(decisions, seconds), Decisions: int(decisions),
		Approvers: approverAverages(actions, submitted)}
}

type tally struct {
	person    PersonRef
	decisions int64
	seconds   int64
}

func approverAverages(actions []approval.Action, submitted map[string]time.Time) []ApproverAverage {
	byPerson := map[string]*tally{}
	for _, a := range actions {
		since, inView := submitted[a.RoundID]
		if !inView || (a.Action != approval.ActionApproved && a.Action != approval.ActionDeclined) {
			continue
		}
		t := byPerson[a.Actor.ID]
		if t == nil {
			t = &tally{person: PersonRef{ID: a.Actor.ID, Name: a.Actor.Name}}
			byPerson[a.Actor.ID] = t
		}
		t.decisions++
		t.seconds += int64(a.At.Sub(since) / time.Second)
	}
	list := make([]*tally, 0, len(byPerson))
	for _, t := range byPerson {
		list = append(list, t)
	}
	slices.SortFunc(list, slowestFirst)
	out := make([]ApproverAverage, len(list))
	for i, t := range list {
		days := approval.AverageDecisionDays(t.decisions, t.seconds)
		out[i] = ApproverAverage{Person: t.person, Days: *days, Decisions: int(t.decisions)}
	}
	return out
}

func slowestFirst(a, b *tally) int {
	return cmp.Or(cmp.Compare(b.seconds*a.decisions, a.seconds*b.decisions), strings.Compare(a.person.Name, b.person.Name))
}
