package measure

import (
	"slices"

	"mapp/internal/allwork"
)

const (
	filterAttribute = "attribute"
	filterCurrency  = "currency"
)

// narrowed matches f in sc with only the values the report offers, so the frame, the chips and the export never
// name what was dropped: an owner or team outside the scope, an unknown, inactive or unfilterable attribute, an
// unknown option and an unknown currency are left out and counted (FR-015, SC-002).
func narrowed(sc Scope, f Filter) Scope {
	q, left := knownAttributes(f.Query, sc.Attributes)
	sc.Matched = allwork.Match(sc.Nodes, q, sc.Attributes, inReportScope(sc.Choices, sc.Nodes))
	sc.Matched.LeftOut = slices.Concat(sc.Matched.LeftOut, left)
	if f.Currency != "" && f.Currency != sc.Currency.Code {
		sc.Matched.LeftOut = append(sc.Matched.LeftOut, allwork.LeftOut{Filter: filterCurrency, ID: f.Currency})
	}
	sc.Query = withoutLeftOut(q, sc.Matched.LeftOut)
	return sc
}

func inReportScope(c allwork.Choices, nodes []allwork.Node) allwork.Choices {
	owners, teams := map[string]bool{}, map[string]bool{}
	for _, n := range nodes {
		if n.Owner != nil {
			owners[n.Owner.ID] = true
		}
		for _, id := range n.TeamIDs {
			teams[id] = true
		}
	}
	c.People = slices.DeleteFunc(slices.Clone(c.People), func(o allwork.Option) bool { return !owners[o.ID] })
	c.Teams = slices.DeleteFunc(slices.Clone(c.Teams), func(o allwork.Option) bool { return !teams[o.ID] })
	return c
}

func knownAttributes(q allwork.Query, attributes []allwork.Attribute) (allwork.Query, []allwork.LeftOut) {
	var left []allwork.LeftOut
	kept := []allwork.AttrFilter{}
	for _, filter := range q.Attributes {
		i := slices.IndexFunc(attributes, func(a allwork.Attribute) bool {
			return a.ID == filter.AttributeID && a.Active && filterable(a.Type)
		})
		if i < 0 {
			left = append(left, allwork.LeftOut{Filter: filterAttribute, ID: filter.AttributeID})
			continue
		}
		var values []string
		for _, value := range filter.Values {
			if knownValue(attributes[i], value) {
				values = append(values, value)
			} else {
				left = append(left, allwork.LeftOut{Filter: filterAttribute, ID: value})
			}
		}
		if len(values) > 0 {
			kept = append(kept, allwork.AttrFilter{AttributeID: filter.AttributeID, Values: values})
		}
	}
	q.Attributes = kept
	return q, left
}

func knownValue(a allwork.Attribute, value string) bool {
	switch value {
	case allwork.AttrNotFilledIn:
		return true
	case allwork.AttrYes, allwork.AttrNo:
		return a.Type == allwork.TypeYesNo
	}
	return slices.ContainsFunc(a.Options, func(o allwork.Option) bool { return o.ID == value })
}

func withoutLeftOut(q allwork.Query, left []allwork.LeftOut) allwork.Query {
	drop := func(filter string, ids []string) []string {
		return slices.DeleteFunc(slices.Clone(ids), func(id string) bool {
			return slices.Contains(left, allwork.LeftOut{Filter: filter, ID: id})
		})
	}
	q.Plans, q.Owners, q.Teams = drop("plan", q.Plans), drop("owner", q.Owners), drop("team", q.Teams)
	q.BusinessUnits, q.ActivityTypes = drop("businessUnit", q.BusinessUnits), drop("activityType", q.ActivityTypes)
	return q
}
