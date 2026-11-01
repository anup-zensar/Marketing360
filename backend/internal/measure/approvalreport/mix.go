package approvalreport

import (
	"slices"
	"time"
)

func statusMix(entries []entry, selected RecordType) StatusMix {
	mix := StatusMix{RecordType: selected, RecordTypes: typesWithRounds(entries)}
	counts := map[Outcome]int{}
	var oldest *time.Time
	for _, e := range entries {
		if e.kind != selected {
			continue
		}
		counts[e.outcome]++
		mix.Total++
		if e.outcome == OutcomeWaiting && (oldest == nil || e.waitingSince.Before(*oldest)) {
			since := e.waitingSince
			oldest = &since
		}
	}
	mix.Waiting = Waiting{Count: counts[OutcomeWaiting], OldestSince: oldest}
	mix.Items = statusItems(counts, mix.Total)
	return mix
}

func typesWithRounds(entries []entry) []RecordType {
	types := []RecordType{}
	for _, t := range allRecordTypes() {
		if slices.ContainsFunc(entries, func(e entry) bool { return e.kind == t }) {
			types = append(types, t)
		}
	}
	return types
}

func statusItems(counts map[Outcome]int, total int) []StatusCount {
	outcomes := statusOutcomes()
	list := make([]int, len(outcomes))
	for i, o := range outcomes {
		list[i] = counts[o]
	}
	shares := wholeShares(list, total)
	items := make([]StatusCount, len(outcomes))
	for i, o := range outcomes {
		items[i] = StatusCount{Outcome: o, Count: list[i], SharePercent: shares[i]}
	}
	return items
}

// wholeShares rounds each share down and gives the rest to the largest remainders, so the shares add to 100.
func wholeShares(counts []int, total int) []int {
	shares := make([]int, len(counts))
	if total == 0 {
		return shares
	}
	order, given := make([]int, len(counts)), 0
	for i, c := range counts {
		shares[i] = c * 100 / total
		given += shares[i]
		order[i] = i
	}
	remainder := func(i int) int { return counts[i] * 100 % total }
	slices.SortStableFunc(order, func(a, b int) int { return remainder(b) - remainder(a) })
	for _, i := range order[:100-given] {
		shares[i]++
	}
	return shares
}
