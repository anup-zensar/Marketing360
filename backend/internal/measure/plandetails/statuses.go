package plandetails

import (
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/platform/money"
)

// word is a status value with its words; a list of them is in lifecycle order (domain model §2.1–§2.3).
type word struct{ key, label string }

const (
	statusInProgress = "inProgress"
	labelNoDates     = "No dates"
)

func planStatusWords() []word {
	return []word{{"draft", "Draft"}, {"published", "Published"}}
}

func progressWords() []word {
	return []word{{"noActivities", "No activities"}, {statusInProgress, "Partly approved"}, {"allApproved", "All approved"}}
}

func timingWords() []word {
	return []word{{"notStarted", "Not started"}, {"running", "Running"}, {"finished", "Finished"}}
}

func activityStatusWords() []word {
	return []word{{"created", "Created"}, {"submitted", "Submitted"}, {"declined", "Declined"},
		{"approved", "Approved"}, {statusInProgress, "In progress"}, {"complete", "Complete"}}
}

func statusOf(n *allwork.Node) string { return n.Status }

func timingOf(n *allwork.Node) string { return n.Timing }

// wordPart is a value in its words and lifecycle order; a container without dates has no timing (No dates).
func wordPart(value string, words []word) part {
	if value == "" {
		return part{key: KeyNoDates, label: labelNoDates, order: orderNoDates, counted: true}
	}
	p := part{key: value, label: value, order: len(words), counted: true, linkable: true}
	if i := slices.IndexFunc(words, func(w word) bool { return w.key == value }); i >= 0 {
		p.label, p.order = words[i].label, i
	}
	return p
}

// wordsWidget counts the level's records by a status value, in lifecycle order (FR-007).
func (v view) wordsWidget(key, what string, words []word, valueOf func(*allwork.Node) string) Widget {
	t := newTally()
	for _, r := range v.records {
		t.add(r.ID, []part{wordPart(valueOf(r), words)}, money.Amount{})
	}
	return t.widget(spec{key: key, title: v.recordsWord() + " by " + what, measure: MeasureCount, timeOrder: true},
		v.currency)
}
