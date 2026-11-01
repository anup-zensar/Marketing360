package planbudget

import (
	"cmp"
	"slices"
	"strings"

	"mapp/internal/allwork"
	"mapp/internal/platform/money"
)

// placed is one top entry in one group, with its share when a % split puts it in several (R6).
type placed struct {
	id    string
	share *money.Percent
	s     sums
	entry Entry
}

type groupAcc struct {
	share
	items []placed
	total sums
}

// ranked is one page of the level's entries, ranked, grouped and drilled (FR-008 – FR-015).
func (r *records) ranked(level string, q RankedQuery) Ranked {
	out := Ranked{Level: level, Sort: q.Sort, Order: q.Order, Group: q.Group, Page: q.Page, Size: q.Size,
		Trail: []TrailStep{}, Groups: []Group{}}
	tops := r.levelIDs(level)
	if q.Drill != "" {
		if drilled, ok := r.byID[q.Drill]; ok && r.inView[q.Drill] && drilled.Level != allwork.LevelLineItem {
			tops, out.Trail = []string{q.Drill}, r.trailOf(q.Drill)
		} else {
			out.DrillLost = true
		}
	}
	groups := r.grouped(r.searched(tops, q.Search), q)
	for _, g := range groups {
		out.Total += len(g.items)
	}
	out.Groups = r.page(groups, q)
	return out
}

func (r *records) levelIDs(level string) []string {
	m := r.matched
	ids, found := map[allwork.Level][]string{allwork.LevelPlan: m.Plans, allwork.LevelInitiative: m.Initiatives,
		allwork.LevelActivity: m.Activities}[allwork.Level(level)]
	if !found {
		return m.Campaigns
	}
	return ids
}

func (r *records) searched(ids []string, search string) []string {
	if search == "" {
		return ids
	}
	wanted := strings.ToLower(search)
	return slices.DeleteFunc(slices.Clone(ids), func(id string) bool {
		return !strings.Contains(strings.ToLower(r.byID[id].Name), wanted)
	})
}

// grouped puts the entries in their groups, each group ranked, the groups ordered by their total (FR-013).
func (r *records) grouped(ids []string, q RankedQuery) []*groupAcc {
	var groups []*groupAcc
	byKey := map[string]*groupAcc{}
	for _, id := range ids {
		whole := r.sumsOf(id)
		for _, g := range r.groupsOf(id, q.Group) {
			acc, found := byKey[g.key]
			if !found {
				acc = &groupAcc{share: g}
				byKey[g.key], groups = acc, append(groups, acc)
			}
			s := whole.times(g.part)
			acc.items = append(acc.items, placed{id: id, share: g.part, s: s, entry: r.entryOf(id, s)})
			acc.total = acc.total.plus(s)
		}
	}
	for _, g := range groups {
		slices.SortStableFunc(g.items, func(x, y placed) int { return r.compare(x.entry, y.entry, q) })
	}
	slices.SortStableFunc(groups, func(x, y *groupAcc) int {
		return r.compare(Entry{Name: x.label, Figures: r.figures(x.total)}, Entry{Name: y.label, Figures: r.figures(y.total)}, q)
	})
	return groups
}

// page keeps one page of top entries across the groups; a group's totals stay those of the whole group (FR-014).
func (r *records) page(groups []*groupAcc, q RankedQuery) []Group {
	first, last := (q.Page-1)*q.Size, q.Page*q.Size
	out := []Group{}
	seen := 0
	for _, g := range groups {
		group := Group{Key: g.key, Label: g.label, Totals: r.figures(g.total), Entries: []Entry{}}
		for _, p := range g.items {
			if seen >= first && seen < last {
				p.entry.Children = r.childrenOf(p.id, p.share, q)
				group.Entries = append(group.Entries, p.entry)
			}
			seen++
		}
		if len(group.Entries) > 0 {
			out = append(out, group)
		}
	}
	return out
}

func (r *records) childrenOf(id string, part *money.Percent, q RankedQuery) []Entry {
	children := []Entry{}
	for _, child := range r.childrenInView(id) {
		children = append(children, r.entryOf(child, r.sumsOf(child).times(part)))
	}
	slices.SortStableFunc(children, func(x, y Entry) int { return r.compare(x, y, q) })
	return children
}

func (r *records) entryOf(id string, s sums) Entry {
	n := r.byID[id]
	return Entry{ID: id, Level: string(n.Level), Name: n.Name, PlanID: n.PlanID, Figures: r.figures(s)}
}

// compare ranks by the sort figure in the chosen order, entries without it last, then by name.
func (r *records) compare(x, y Entry, q RankedQuery) int {
	a, b := figureOf(x.Figures, q.Sort), figureOf(y.Figures, q.Sort)
	switch {
	case a == nil && b == nil:
		return strings.Compare(x.Name, y.Name)
	case a == nil:
		return 1
	case b == nil:
		return -1
	}
	order := a.Amount.Cmp(b.Amount)
	if q.Order == OrderDesc {
		order = -order
	}
	return cmp.Or(order, strings.Compare(x.Name, y.Name))
}

func figureOf(f Figures, sort string) *money.Money {
	return map[string]*money.Money{SortBudget: f.Budget, SortPlanned: f.Planned, SortCommitted: f.Committed,
		SortActual: f.Actual, SortLeftToPlan: f.LeftToPlan, SortUnspent: f.Unspent}[sort]
}

// trailOf is the drilled record and its ancestors, from the plan down (FR-010).
func (r *records) trailOf(id string) []TrailStep {
	var trail []TrailStep
	for n, ok := r.byID[id]; ok; n, ok = r.byID[n.ParentID] {
		trail = append(trail, TrailStep{ID: n.ID, Level: string(n.Level), Name: n.Name})
	}
	slices.Reverse(trail)
	return trail
}
