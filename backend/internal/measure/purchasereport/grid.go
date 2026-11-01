package purchasereport

import (
	"cmp"
	"slices"
	"strings"

	"mapp/internal/platform/date"
)

const defaultPageSize = 25

func words(search string) []string {
	return strings.Fields(strings.ToLower(search))
}

// grid is one page of vendors, each with every request, line and line item the search kept (042-C3).
func (d data) grid(rs, all []Request, q Query) Grid {
	key, order := cmp.Or(q.Sort, colName), cmp.Or(q.Order, orderAsc)
	groups := d.sorted(d.searched(groupsOf(rs), words(q.Search)), key, order)
	size, page := cmp.Or(q.PageSize, defaultPageSize), cmp.Or(q.Page, 1)
	from := min((page-1)*size, len(groups))
	to := min(from+size, len(groups))
	grid := Grid{Vendors: make([]VendorNode, 0, to-from), Total: len(groups), Page: page, PageSize: size,
		Columns: columnsOr(q.Columns), Choices: d.choices(all)}
	for _, g := range groups[from:to] {
		grid.Vendors = append(grid.Vendors, d.vendorNode(g))
	}
	return grid
}

func (d data) choices(all []Request) GridChoices {
	choices := GridChoices{Vendors: []string{}, Owners: []PersonRef{}, Statuses: allStatuses()}
	for _, r := range all {
		vendor := cmp.Or(r.Vendor, noVendor)
		if !slices.Contains(choices.Vendors, vendor) {
			choices.Vendors = append(choices.Vendors, vendor)
		}
		if owner := d.owner(r.OwnerID); !slices.Contains(choices.Owners, owner) {
			choices.Owners = append(choices.Owners, owner)
		}
	}
	slices.SortFunc(choices.Vendors, func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) })
	slices.SortFunc(choices.Owners, func(a, b PersonRef) int {
		return cmp.Or(strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), strings.Compare(a.ID, b.ID))
	})
	return choices
}

func (d data) vendorNode(g vendorGroup) VendorNode {
	var committed, invoiced tally
	node := VendorNode{Vendor: g.vendor, Label: vendorLabel(g.vendor), PurchaseRequests: len(g.requests),
		Requests: make([]RequestNode, 0, len(g.requests))}
	for _, rv := range g.requests {
		node.Lines += len(rv.r.Lines)
		for _, l := range rv.r.Lines {
			if isCommitted(rv.r.Status) {
				committed.add(l.Amount, d.sees(rv.r.PlanID))
			}
			if l.Invoiced != nil {
				invoiced.add(*l.Invoiced, d.sees(rv.r.PlanID))
			}
		}
		node.Requests = append(node.Requests, d.requestNode(rv))
	}
	node.Committed, node.Invoiced = committed.money(d.shown), invoiced.money(d.shown)
	return node
}

func (d data) requestNode(rv reqView) RequestNode {
	r := rv.r
	var amount, invoiced tally
	var orders []string
	var delivery, sent date.Date
	for _, l := range r.Lines {
		amount.add(l.Amount, d.sees(r.PlanID))
		if l.Invoiced != nil {
			invoiced.add(*l.Invoiced, d.sees(r.PlanID))
		}
		if l.PONumber != "" && !slices.Contains(orders, l.PONumber) {
			orders = append(orders, l.PONumber)
		}
		delivery, sent = earlierDate(delivery, l.DeliveryDate), earlierDate(sent, sentOf(l))
	}
	node := RequestNode{ID: r.ID, Name: r.Name, Status: r.Status, Amount: amount.money(d.shown), Invoiced: invoiced.money(d.shown),
		Owner: d.owner(r.OwnerID), CreatedOn: date.Today(r.CreatedAt), SentOn: datePtr(sent), DeliveryDate: datePtr(delivery),
		PONumber: textPtr(strings.Join(orders, ", ")), Workflow: textPtr(r.Workflow), GSAPNumber: textPtr(r.GSAPNumber),
		WaitingForGSAP: r.WaitingForGSAP, Lines: make([]LineNode, 0, len(rv.lines))}
	for _, l := range rv.lines {
		node.Lines = append(node.Lines, d.lineNode(r, l))
	}
	return node
}

func (d data) lineNode(r Request, l Line) LineNode {
	info, known := d.items[l.LineItemID]
	name := cmp.Or(info.name, l.LineItemID)
	node := LineNode{ID: l.ID, LineItem: LineItemRef{ID: l.LineItemID, Name: name}, DeliveryDate: datePtr(l.DeliveryDate),
		SentOn: l.SentOn, PONumber: textPtr(l.PONumber), POLine: textPtr(l.POLine), POLineStatus: textPtr(l.POLineStatus)}
	if known {
		node.Activity = textPtr(info.activity)
	}
	if d.sees(r.PlanID) {
		node.Amount = d.shown.of(l.Amount)
		if l.Invoiced != nil {
			node.Invoiced = d.shown.of(*l.Invoiced)
		}
	}
	return node
}

func sentOf(l Line) date.Date {
	if l.SentOn == nil {
		return date.Date{}
	}
	return *l.SentOn
}

func earlierDate(a, b date.Date) date.Date {
	if b.IsZero() || (!a.IsZero() && !b.Before(a)) {
		return a
	}
	return b
}

func datePtr(d date.Date) *date.Date {
	if d.IsZero() {
		return nil
	}
	return &d
}

func textPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func columnsOr(chosen []string) []string {
	if len(chosen) == 0 {
		return allColumns()
	}
	return chosen
}
