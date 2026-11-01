package purchasereport

import (
	"slices"

	"mapp/internal/platform/money"
)

func allStatuses() []Status {
	return []Status{StatusCreated, StatusSubmitted, StatusInReview, StatusDeclined, StatusApproved, StatusClosed}
}

func (d data) statusMix(rs []Request) StatusMix {
	counts := map[Status]int{}
	amounts := map[Status]*tally{}
	var all tally
	mix := StatusMix{Total: len(rs)}
	for _, r := range rs {
		counts[r.Status]++
		if amounts[r.Status] == nil {
			amounts[r.Status] = &tally{}
		}
		amounts[r.Status].add(requestUSD(r), d.sees(r.PlanID))
		all.add(requestUSD(r), d.sees(r.PlanID))
		if r.WaitingForGSAP {
			mix.WaitingForGSAP++
		}
	}
	mix.TotalAmount = all.money(d.shown)
	list := make([]int, 0, len(allStatuses()))
	for _, s := range allStatuses() {
		list = append(list, counts[s])
	}
	shares := wholeShares(list, len(rs))
	for i, s := range allStatuses() {
		item := StatusCount{Status: s, Count: counts[s], SharePercent: shares[i], Amount: d.shown.of(money.Amount{})}
		if t := amounts[s]; t != nil {
			item.Amount = t.money(d.shown)
		}
		mix.Items = append(mix.Items, item)
	}
	return mix
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
