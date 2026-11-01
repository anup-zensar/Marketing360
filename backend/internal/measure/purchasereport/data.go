package purchasereport

import (
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/money"
)

const notFilledIn = "Not filled in"

type lineItemInfo struct {
	name, activity string
}

// data is what one Read finds: every purchase request on the line items in view and the names around them.
type data struct {
	requests   []Request
	switchedOn bool
	people     map[string]string
	items      map[string]lineItemInfo
	shown      shown
	scope      measure.Scope
}

func newData(scope measure.Scope, found Requests) data {
	people := map[string]string{}
	for _, p := range scope.Choices.People {
		people[p.ID] = p.Label
	}
	names := map[string]string{}
	for _, n := range scope.Nodes {
		names[n.ID] = n.Name
	}
	items := map[string]lineItemInfo{}
	for _, n := range scope.Nodes {
		if n.Level == allwork.LevelLineItem {
			items[n.ID] = lineItemInfo{name: n.Name, activity: names[n.ParentID]}
		}
	}
	return data{requests: found.Items, switchedOn: found.SwitchedOn, people: people, items: items, scope: scope,
		shown: shown{rate: scope.Currency.Rate, code: scope.Currency.Code}}
}

func (d data) sees(plan string) bool {
	c := d.scope.Choices
	return c.SeesFinance && (c.FinancePlans == nil || slices.Contains(c.FinancePlans, plan))
}

func (d data) owner(id string) PersonRef {
	name, known := d.people[id]
	if !known {
		name = id
	}
	return PersonRef{ID: id, Name: name}
}

func vendorLabel(vendor string) string {
	if vendor == "" {
		return notFilledIn
	}
	return vendor
}

func isCommitted(s Status) bool {
	return s == StatusApproved || s == StatusClosed
}

func requestUSD(r Request) money.Amount {
	var total money.Amount
	for _, l := range r.Lines {
		total = total.Add(l.Amount)
	}
	return total
}

func filtered(requests []Request, q Query) []Request {
	return slices.DeleteFunc(slices.Clone(requests), func(r Request) bool { return !matches(r, q) })
}

func matches(r Request, q Query) bool {
	if len(q.Statuses) > 0 && !slices.Contains(q.Statuses, r.Status) {
		return false
	}
	if len(q.RequestOwners) > 0 && !slices.Contains(q.RequestOwners, r.OwnerID) {
		return false
	}
	return len(q.Vendors) == 0 || slices.ContainsFunc(q.Vendors, func(v string) bool {
		return v == r.Vendor || (v == noVendor && r.Vendor == "")
	})
}
