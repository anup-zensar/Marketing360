package plandetails

import (
	"mapp/internal/allwork"
	"mapp/internal/platform/money"
)

// view is one level of the matched year as the widgets read it.
type view struct {
	level    Level
	year     int
	period   string
	records  []*allwork.Node            // the matching records of the level, in the year's order
	pieces   map[string][]*allwork.Node // each record's matching activities (the activity itself at its level)
	units    map[string][]string
	types    []allwork.Option
	currency allwork.Currency
}

func newView(m Matching, level Level, period string) view {
	v := view{level: level, year: m.Year, period: period, pieces: map[string][]*allwork.Node{},
		units: map[string][]string{}, types: m.ActivityTypes, currency: m.Currency}
	byID := map[string]*allwork.Node{}
	for i := range m.Nodes {
		n := &m.Nodes[i]
		byID[n.ID] = n
		if n.Level == allwork.LevelPlan {
			v.units[n.ID] = n.BusinessUnitIDs
		}
	}
	for i := range m.Nodes {
		n := &m.Nodes[i]
		if !m.Matched[n.ID] {
			continue
		}
		if n.Level == allwork.Level(level) {
			v.records = append(v.records, n)
		}
		if n.Level != allwork.LevelActivity {
			continue
		}
		if r := ancestorAt(byID, n, allwork.Level(level)); r != nil {
			v.pieces[r.ID] = append(v.pieces[r.ID], n)
		}
	}
	return v
}

// ancestorAt returns n or the ancestor of n at level; nil when it has none.
func ancestorAt(byID map[string]*allwork.Node, n *allwork.Node, level allwork.Level) *allwork.Node {
	for n != nil && n.Level != level {
		n = byID[n.ParentID]
	}
	return n
}

func (v view) cards() Cards {
	var usd money.Amount
	filled := 0
	for _, r := range v.records {
		for _, a := range v.pieces[r.ID] {
			usd = usd.Add(plannedCostOf(a))
		}
		if r.Missing == 0 {
			filled++
		}
	}
	return Cards{Records: len(v.records), PlannedCost: inCurrency(usd, v.currency), AllFilledIn: filled,
		AllFilledInPercent: wholePercent(filled, len(v.records))}
}

func plannedCostOf(activity *allwork.Node) money.Amount {
	if activity.PlannedCost == nil {
		return money.Amount{}
	}
	return *activity.PlannedCost
}

func (v view) recordsWord() string {
	return recordsWord(v.level)
}

// recordsWord names a level's records in titles and column headings.
func recordsWord(level Level) string {
	switch level {
	case LevelPlan:
		return "Plans"
	case LevelCampaign:
		return "Campaigns"
	case LevelInitiative:
		return "Initiatives"
	case LevelActivity:
		return "Activities"
	}
	return "Records"
}
