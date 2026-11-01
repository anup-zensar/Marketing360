package allwork

import "cmp"

// Matched are the records a filter set matches in a year as All work shows and counts them (039 R1): the shown and
// matching records of each level in the year's order, the line items of the matching activities and what was left out.
type Matched struct {
	Plans       []string
	Campaigns   []string
	Initiatives []string
	Activities  []string
	LineItems   []string
	LeftOut     []LeftOut
}

// Match runs All work's own filters on the year's nodes, so a report and All work cover the same records (039 FR-012).
func Match(nodes []Node, q Query, attributes []Attribute, c Choices) Matched {
	q, leftOut := knownValues(withDefaults(q), YearData{Nodes: nodes}, c)
	q.attrs = resolveAttrs(q.Attributes, attributes, nodes)
	t := buildTree(nodes)
	if c.SeesFinance {
		t.withFinance(c.FinancePlans)
	}
	applyFilters(t, q)
	m := Matched{Plans: []string{}, Campaigns: []string{}, Initiatives: []string{}, Activities: []string{},
		LineItems: []string{}, LeftOut: leftOut}
	for i := range nodes {
		if n, found := t.byID[nodes[i].ID]; found && n.shown && n.match {
			m.add(n)
		}
	}
	return m
}

func withDefaults(q Query) Query {
	if q.Strategic == "" {
		q.Strategic = StrategicInclude
	}
	if q.BCPIR == "" {
		q.BCPIR = StrategicInclude
	}
	if q.ShowLevels == "" {
		q.ShowLevels = ShowEverything
	}
	return q
}

func (m *Matched) add(n *tnode) {
	switch n.n.Level {
	case LevelPlan:
		m.Plans = append(m.Plans, n.n.ID)
	case LevelCampaign:
		m.Campaigns = append(m.Campaigns, n.n.ID)
	case LevelInitiative:
		m.Initiatives = append(m.Initiatives, n.n.ID)
	case LevelActivity:
		m.Activities = append(m.Activities, n.n.ID)
		for _, child := range n.children {
			if child.n.Level == LevelLineItem {
				m.LineItems = append(m.LineItems, child.n.ID)
			}
		}
	case LevelLineItem, LevelUnassigned:
	}
}

// PlanTypeWord and ActivityStatusWord name a plan type and an activity status as All work shows them (039 FR-022).
func PlanTypeWord(code string) string { return cmp.Or(planTypeWords[code], code) }

func ActivityStatusWord(code string) string { return cmp.Or(activityStatuses[code].label, code) }
