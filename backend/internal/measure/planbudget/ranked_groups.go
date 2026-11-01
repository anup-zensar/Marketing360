package planbudget

import (
	"strings"

	"mapp/internal/allwork"
	"mapp/internal/platform/money"
)

const (
	keyNone          = "none"
	labelNoOwner     = "No owner"
	labelNotFilledIn = "Not filled in"
)

// share is one group an entry belongs to; part is its share of the entry under a % split, nil for all of it.
type share struct {
	key, label string
	part       *money.Percent
}

func (r *records) groupsOf(id, group string) []share {
	n := r.byID[id]
	switch {
	case group == GroupPlan:
		plan := r.byID[n.PlanID]
		return []share{{key: plan.ID, label: plan.Name}}
	case group == GroupOwner && n.Owner != nil:
		return []share{{key: n.Owner.ID, label: n.Owner.Name}}
	case group == GroupOwner:
		return []share{{key: keyNone, label: labelNoOwner}}
	case strings.HasPrefix(group, groupAttr):
		return r.attributeGroups(id, strings.TrimPrefix(group, groupAttr))
	}
	return []share{{key: GroupNone}}
}

// attributeGroups follow the value of the record or of its nearest ancestor that has one, as a campaign's value
// applies to what is below it (client 016 FR-066); a % split counts by share (D-270).
func (r *records) attributeGroups(id, attributeID string) []share {
	value, found := r.attributeValue(id, attributeID)
	if !found {
		return []share{{key: keyNone, label: labelNotFilledIn}}
	}
	if value.Yes != nil {
		if *value.Yes {
			return []share{{key: "yes", label: "Yes"}}
		}
		return []share{{key: "no", label: "No"}}
	}
	attribute := r.attributes[attributeID]
	out := make([]share, 0, len(value.OptionIDs))
	for _, optionID := range value.OptionIDs {
		g := share{key: optionID, label: optionID}
		for _, option := range attribute.Options {
			if option.ID == optionID {
				g.label = option.Label
			}
		}
		if p, split := value.Shares[optionID]; split {
			g.part = &p
		}
		out = append(out, g)
	}
	return out
}

func (r *records) attributeValue(id, attributeID string) (allwork.AttrValue, bool) {
	for n, ok := r.byID[id]; ok; n, ok = r.byID[n.ParentID] {
		if value, has := n.Attrs[attributeID]; has && (value.Yes != nil || len(value.OptionIDs) > 0) {
			return value, true
		}
	}
	return allwork.AttrValue{}, false
}

// times is the part of the sums a % split gives, each amount rounded half up to the cent.
func (s sums) times(part *money.Percent) sums {
	if part == nil {
		return s
	}
	scale := func(a money.Amount) money.Amount { return money.FromCents(part.OfWhole(a.Cents())) }
	scaled := func(a *money.Amount) *money.Amount {
		if a == nil {
			return nil
		}
		part := scale(*a)
		return &part
	}
	return sums{planned: scale(s.planned), committed: scale(s.committed), actual: scale(s.actual), financed: s.financed,
		budget: scaled(s.budget), finBudget: scaled(s.finBudget), finPlanned: scale(s.finPlanned)}
}

func (s sums) plus(t sums) sums {
	return sums{planned: s.planned.Add(t.planned), committed: s.committed.Add(t.committed),
		actual: s.actual.Add(t.actual), financed: s.financed || t.financed, budget: addAmount(s.budget, t.budget),
		finBudget: addAmount(s.finBudget, t.finBudget), finPlanned: s.finPlanned.Add(t.finPlanned)}
}
