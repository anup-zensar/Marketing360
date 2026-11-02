package revenue

import (
	"cmp"
	"slices"
	"strings"

	"mapp/internal/allwork"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
)

// revenueRankedInput groups revenueRanked's working parts so building it stays within the argument limit.
type revenueRankedInput struct {
	views []activityView
	nodes []allwork.Node
	show  string
	q     RankedQuery
	w     window
}

// revenueRanked builds one page of the Show level's entries, each with its children one level below
// (FR-063), sorted and searched. Grouping and drill follow in a later pass (plan.md R5; tasks.md T014).
func revenueRanked(in revenueRankedInput) RevenueRanked {
	byID, below := nodeIndex(in.nodes)
	own := revenueFiguresOfActivities(in.views, in.w)
	totals := rollUpRevenue(in.nodes, below, own)
	tops := searchByName(topLevelIDs(in.nodes, in.show), byID, in.q.Search)
	q := in.q

	entries := make([]RevenueEntry, 0, len(tops))
	for _, id := range tops {
		entries = append(entries, revenueEntryOf(id, byID, below, totals))
	}
	sortRevenueEntries(entries, q.Sort, q.Order)

	total := len(entries)
	page := pageOfEntries(entries, q.Page, q.Size)
	group := RevenueGroup{Key: "", Label: "", Entries: page}
	for _, e := range page {
		group.Totals = addRevenueFigures(group.Totals, e.Figures)
	}
	group.Totals = withPerformance(group.Totals)
	return RevenueRanked{Level: in.show, Sort: q.Sort, Order: q.Order, Group: GroupNone, Page: q.Page, Size: q.Size,
		Total: total, Trail: []TrailStep{}, Groups: []RevenueGroup{group}}
}

func nodeIndex(nodes []allwork.Node) (map[string]allwork.Node, map[string][]string) {
	byID := make(map[string]allwork.Node, len(nodes))
	below := map[string][]string{}
	for _, n := range nodes {
		byID[n.ID] = n
	}
	for _, n := range nodes {
		if n.ParentID != "" {
			below[n.ParentID] = append(below[n.ParentID], n.ID)
		}
	}
	return byID, below
}

func topLevelIDs(nodes []allwork.Node, show string) []string {
	ids := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if string(n.Level) == show {
			ids = append(ids, n.ID)
		}
	}
	return ids
}

func searchByName(ids []string, byID map[string]allwork.Node, search string) []string {
	if search == "" {
		return ids
	}
	wanted := strings.ToLower(search)
	found := make([]string, 0, len(ids))
	for _, id := range ids {
		if strings.Contains(strings.ToLower(byID[id].Name), wanted) {
			found = append(found, id)
		}
	}
	return found
}

func revenueEntryOf(id string, byID map[string]allwork.Node, below map[string][]string, figures map[string]RevenueFigures) RevenueEntry {
	n := byID[id]
	entry := RevenueEntry{ID: n.ID, Level: string(n.Level), Name: n.Name, PlanID: n.PlanID, Figures: withPerformance(figures[id])}
	for _, childID := range below[id] {
		child := byID[childID]
		entry.Children = append(entry.Children, RevenueEntry{
			ID: child.ID, Level: string(child.Level), Name: child.Name, PlanID: child.PlanID,
			Figures: withPerformance(figures[childID]),
		})
	}
	return entry
}

// revenueFiguresOfActivities is each activity's own figures from its projection and counted facts: Goal is
// the full plan-year projection, Actual every counted fact, and Projected the same "actual or projected"
// total the cards show, split at today (FR-004, FR-012).
func revenueFiguresOfActivities(views []activityView, w window) map[string]RevenueFigures {
	out := make(map[string]RevenueFigures, len(views))
	for _, v := range views {
		f := RevenueFigures{ActualCostReason: ReasonAttributionNotConfigured, ROIReason: ReasonAttributionNotConfigured}
		if v.projection.Present {
			goal := sumMoney(v.projection.Revenue)
			f.Goal = &goal
		}
		if len(v.facts.revenue) > 0 {
			actual := sumMoney(v.facts.revenue)
			f.Actual = &actual
		}
		split := projectionSplit(v, w)
		if f.Actual != nil || v.projection.Present {
			total := money.USD(split.Actual.Amount.Add(split.Projected.Amount))
			f.Projected = &total
		}
		out[v.node.ID] = f
	}
	return out
}

func projectionSplit(v activityView, w window) projection.SplitMoney {
	return projection.SplitAmounts(v.facts.revenue, v.projection.Revenue, w.today, w.year)
}

// withPerformance derives Performance from Projected ÷ Goal (049-C15, 049-C16), the same formula the
// cards use, so it stays correct whether f is one activity's own figures or a rolled-up total.
func withPerformance(f RevenueFigures) RevenueFigures {
	if f.Goal == nil {
		return f
	}
	var total money.Amount
	if f.Projected != nil {
		total = f.Projected.Amount
	}
	f.Performance = performanceOf(total, f.Goal.Amount, true).Value
	return f
}

// rollUpRevenue sums every node's figures from its activity descendants, so a plan or campaign row shows
// its activities' totals (FR-063, SC-002); a leaf activity keeps its own.
func rollUpRevenue(nodes []allwork.Node, below map[string][]string, ownActivity map[string]RevenueFigures) map[string]RevenueFigures {
	totals := make(map[string]RevenueFigures, len(nodes))
	var sum func(id string) RevenueFigures
	sum = func(id string) RevenueFigures {
		if f, done := totals[id]; done {
			return f
		}
		children, hasChildren := below[id]
		if !hasChildren {
			f := ownActivity[id]
			totals[id] = f
			return f
		}
		var f RevenueFigures
		for _, childID := range children {
			f = addRevenueFigures(f, sum(childID))
		}
		totals[id] = f
		return f
	}
	for _, n := range nodes {
		sum(n.ID)
	}
	return totals
}

func sumMoney(months []projection.MoneyFigure) money.Money {
	var total money.Amount
	for _, m := range months {
		total = total.Add(m.Value.Amount)
	}
	return money.USD(total)
}

func addRevenueFigures(a, b RevenueFigures) RevenueFigures {
	a.Goal = addMoneyPtr(a.Goal, b.Goal)
	a.Actual = addMoneyPtr(a.Actual, b.Actual)
	a.Projected = addMoneyPtr(a.Projected, b.Projected)
	return a
}

func addMoneyPtr(a, b *money.Money) *money.Money {
	if a == nil && b == nil {
		return nil
	}
	var total money.Amount
	if a != nil {
		total = total.Add(a.Amount)
	}
	if b != nil {
		total = total.Add(b.Amount)
	}
	sum := money.USD(total)
	return &sum
}

func centsOf(m *money.Money) int64 {
	if m == nil {
		return 0
	}
	return m.Amount.Cents()
}

func sortRevenueEntries(entries []RevenueEntry, sort, order string) {
	slices.SortStableFunc(entries, func(a, b RevenueEntry) int {
		c := cmp.Compare(figureValue(a.Figures, sort), figureValue(b.Figures, sort))
		if order == OrderAsc {
			return c
		}
		return -c
	})
}

func figureValue(f RevenueFigures, sort string) int64 {
	if sort == SortActualRevenue {
		return centsOf(f.Actual)
	}
	return centsOf(f.Projected)
}

func pageOfEntries(entries []RevenueEntry, page, size int) []RevenueEntry {
	start := (page - 1) * size
	if start >= len(entries) {
		return []RevenueEntry{}
	}
	end := min(start+size, len(entries))
	return entries[start:end]
}
