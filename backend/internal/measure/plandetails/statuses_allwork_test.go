package plandetails

import (
	"fmt"
	"slices"
	"testing"

	"mapp/internal/allwork"
	"mapp/internal/platform/actor"
)

func withStatus(status, timing string) func(*allwork.Node) {
	return func(n *allwork.Node) { n.Status, n.Timing = status, timing }
}

// TestStatusWords_EqualAllWork keeps the status and timing words this report copies equal to All work's own rows.
func TestStatusWords_EqualAllWork(t *testing.T) {
	t.Parallel()
	y := newYear(t).plan("P-1").plan("P-2").on("P-2", withStatus("published", ""))
	for i, w := range progressWords() {
		y.campaign(fmt.Sprintf("C-%d", i), "P-1", withStatus(w.key, timingWords()[i].key))
	}
	for i, w := range activityStatusWords() {
		y.activity(fmt.Sprintf("A-%d", i), "C-0", "", withStatus(w.key, ""))
	}
	result, err := allWorkOf(yearSource{nodes: y.nodes}).Query(t.Context(), actor.Actor{ID: "U-1002"},
		allwork.Query{Year: testYearNumber})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	shown := labelsOf(result.Rows)
	for level, words := range map[string][]word{"plan": planStatusWords(), "campaign": progressWords(),
		"timing": timingWords(), "activity": activityStatusWords()} {
		for _, w := range words {
			if got := shown[level+" "+w.key]; got != w.label {
				t.Errorf("%s %s: All work says %q, this report %q", level, w.key, got, w.label)
			}
		}
	}
	if !slices.ContainsFunc(result.Rows, func(r allwork.Row) bool { return r.Level == allwork.LevelActivity }) {
		t.Error("All work listed no activity; the comparison checked nothing")
	}
}

// labelsOf is every status and timing All work shows, by level and value.
func labelsOf(rows []allwork.Row) map[string]string {
	shown := map[string]string{}
	for _, r := range rows {
		if r.Status != nil {
			shown[string(r.Level)+" "+r.Status.Value] = r.Status.Label
		}
		if r.Timing != nil {
			shown["timing "+r.Timing.Value] = r.Timing.Label
		}
	}
	return shown
}
