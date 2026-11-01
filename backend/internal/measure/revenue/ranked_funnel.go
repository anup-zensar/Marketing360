package revenue

import (
	"cmp"
	"context"
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/projection"
)

// RankedFunnel is the ranked list sorted by the selected stage's actual or projected volume (FR-006, FR-017).
// q.Stage carries the selected stage, keeping the method within the argument limit.
func (s *Service) RankedFunnel(ctx context.Context, a actor.Actor, f measure.Filter, q RankedQuery) (FunnelRanked, error) {
	var out FunnelRanked
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		sc, views, err := s.scopeAndActivities(ctx, a, f, categoryOf(f.Own))
		if err != nil {
			return err
		}
		in := funnelRankedInput{views: views, nodes: nodesForRanking(sc, views), stage: stageOf(q.Stage, availableStages(views)),
			show: showLevel(f.Show), w: window{today: date.Today(s.now()), year: sc.Year}, q: q}
		out = buildFunnelRanked(in)
		return nil
	})
	return out, err
}

// funnelRankedInput groups RankedFunnel's working parts so building the list stays within the argument limit.
type funnelRankedInput struct {
	views       []activityView
	nodes       []allwork.Node
	stage, show string
	w           window
	q           RankedQuery
}

func buildFunnelRanked(in funnelRankedInput) FunnelRanked {
	byID, below := nodeIndex(in.nodes)
	own := funnelFiguresOfActivities(in.views, in.stage, in.w)
	totals := rollUpFunnel(in.nodes, below, own)
	tops := searchByName(topLevelIDs(in.nodes, in.show), byID, in.q.Search)
	q := in.q
	show := in.show

	entries := make([]FunnelEntry, 0, len(tops))
	for _, id := range tops {
		entries = append(entries, funnelEntryOf(id, byID, below, totals))
	}
	sortFunnelEntries(entries, q.Sort, q.Order)

	total := len(entries)
	page := pageOfFunnelEntries(entries, q.Page, q.Size)
	group := FunnelGroup{Key: "", Label: "", Entries: page}
	for _, e := range page {
		group.Totals = addFunnelFigures(group.Totals, e.Figures)
	}
	return FunnelRanked{Level: show, Sort: q.Sort, Order: q.Order, Group: GroupNone, Page: q.Page, Size: q.Size,
		Total: total, Trail: []TrailStep{}, Groups: []FunnelGroup{group}}
}

func funnelFiguresOfActivities(views []activityView, stage string, w window) map[string]FunnelFigures {
	out := make(map[string]FunnelFigures, len(views))
	for _, v := range views {
		series, found := v.projection.Find(stage)
		if !found {
			out[v.node.ID] = FunnelFigures{}
			continue
		}
		split := projection.SplitCounts(v.facts.counts[stage], series.Months, w.today, w.year)
		f := FunnelFigures{}
		if v.projection.Present {
			goal := series.Total
			f.Goal = &goal
		}
		actual := split.Actual
		f.Actual = &actual
		out[v.node.ID] = f
	}
	return out
}

func rollUpFunnel(nodes []allwork.Node, below map[string][]string, ownActivity map[string]FunnelFigures) map[string]FunnelFigures {
	totals := make(map[string]FunnelFigures, len(nodes))
	var sum func(id string) FunnelFigures
	sum = func(id string) FunnelFigures {
		if f, done := totals[id]; done {
			return f
		}
		children, hasChildren := below[id]
		if !hasChildren {
			f := ownActivity[id]
			totals[id] = f
			return f
		}
		var f FunnelFigures
		for _, childID := range children {
			f = addFunnelFigures(f, sum(childID))
		}
		totals[id] = f
		return f
	}
	for _, n := range nodes {
		sum(n.ID)
	}
	return totals
}

func funnelEntryOf(id string, byID map[string]allwork.Node, below map[string][]string, figures map[string]FunnelFigures) FunnelEntry {
	n := byID[id]
	entry := FunnelEntry{ID: n.ID, Level: string(n.Level), Name: n.Name, PlanID: n.PlanID, Figures: figures[id]}
	for _, childID := range below[id] {
		child := byID[childID]
		entry.Children = append(entry.Children, FunnelEntry{
			ID: child.ID, Level: string(child.Level), Name: child.Name, PlanID: child.PlanID, Figures: figures[childID],
		})
	}
	return entry
}

func addFunnelFigures(a, b FunnelFigures) FunnelFigures {
	a.Goal = addInt64Ptr(a.Goal, b.Goal)
	a.Actual = addInt64Ptr(a.Actual, b.Actual)
	a.Projected = addInt64Ptr(a.Projected, b.Projected)
	return a
}

func addInt64Ptr(a, b *int64) *int64 {
	if a == nil && b == nil {
		return nil
	}
	var sum int64
	if a != nil {
		sum += *a
	}
	if b != nil {
		sum += *b
	}
	return &sum
}

func sortFunnelEntries(entries []FunnelEntry, sort, order string) {
	slices.SortStableFunc(entries, func(a, b FunnelEntry) int {
		c := cmp.Compare(funnelFigureValue(a.Figures, sort), funnelFigureValue(b.Figures, sort))
		if order == OrderAsc {
			return c
		}
		return -c
	})
}

func funnelFigureValue(f FunnelFigures, sort string) int64 {
	if sort == SortActual {
		return int64Of(f.Actual)
	}
	return int64Of(f.Projected)
}

func int64Of(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func pageOfFunnelEntries(entries []FunnelEntry, page, size int) []FunnelEntry {
	start := (page - 1) * size
	if start >= len(entries) {
		return []FunnelEntry{}
	}
	end := min(start+size, len(entries))
	return entries[start:end]
}
