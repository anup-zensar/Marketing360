package purchasereport

import (
	"cmp"
	"slices"
	"strings"
)

// sortVal is what a node is ordered by; a missing value sorts last in either order.
type sortVal struct {
	num     int64
	text    string
	missing bool
}

func compareVals(a, b sortVal, order string) int {
	if a.missing != b.missing {
		if a.missing {
			return 1
		}
		return -1
	}
	by := cmp.Or(cmp.Compare(a.num, b.num), strings.Compare(a.text, b.text))
	if order == orderDesc {
		by = -by
	}
	return by
}

func statusRank(s Status) int64 {
	return int64(slices.Index(allStatuses(), s))
}

func (d data) requestUSDShown(r Request) int64 {
	if !d.sees(r.PlanID) {
		return 0
	}
	return requestUSD(r).Cents()
}

func (d data) invoicedUSDShown(r Request) int64 {
	var cents int64
	for _, l := range r.Lines {
		if l.Invoiced != nil && d.sees(r.PlanID) {
			cents += l.Invoiced.Cents()
		}
	}
	return cents
}

func earliest(dates []int64) sortVal {
	dates = slices.DeleteFunc(slices.Clone(dates), func(n int64) bool { return n == 0 })
	if len(dates) == 0 {
		return sortVal{missing: true}
	}
	return sortVal{num: slices.Min(dates)}
}

func lineDates(lines []Line) []int64 {
	out := make([]int64, 0, len(lines))
	for _, l := range lines {
		if !l.DeliveryDate.IsZero() {
			out = append(out, l.DeliveryDate.Time().Unix())
		}
	}
	return out
}

func (d data) requestVal(rv reqView, key string) sortVal {
	r := rv.r
	switch key {
	case colStatus:
		return sortVal{num: statusRank(r.Status)}
	case colAmount:
		return sortVal{num: d.requestUSDShown(r)}
	case colInvoiced:
		return sortVal{num: d.invoicedUSDShown(r)}
	case colCreated:
		return sortVal{num: r.CreatedAt.Unix()}
	case colDelivery:
		return earliest(lineDates(r.Lines))
	}
	return sortVal{text: strings.ToLower(r.Name)}
}

func (d data) lineVal(r Request, l Line, key string) sortVal {
	switch key {
	case colAmount:
		if d.sees(r.PlanID) {
			return sortVal{num: l.Amount.Cents()}
		}
		return sortVal{}
	case colInvoiced:
		if l.Invoiced != nil && d.sees(r.PlanID) {
			return sortVal{num: l.Invoiced.Cents()}
		}
		return sortVal{missing: l.Invoiced == nil}
	case colName, colStatus, colCreated:
		return sortVal{text: strings.ToLower(d.items[l.LineItemID].name)}
	}
	return earliest(lineDates([]Line{l}))
}

func (d data) groupVal(g vendorGroup, key string) sortVal {
	var nums []int64
	var total int64
	for _, rv := range g.requests {
		v := d.requestVal(rv, key)
		total += v.num
		if !v.missing {
			nums = append(nums, v.num)
		}
	}
	switch key {
	case colAmount, colInvoiced:
		return sortVal{num: total}
	case colCreated, colDelivery:
		return earliest(nums)
	}
	return sortVal{text: strings.ToLower(vendorLabel(g.vendor))}
}

// sorted orders the vendors, the requests under each and the lines under each request by the chosen column.
func (d data) sorted(groups []vendorGroup, key, order string) []vendorGroup {
	out := slices.Clone(groups)
	for i := range out {
		out[i].requests = slices.Clone(out[i].requests)
		for j := range out[i].requests {
			rv := &out[i].requests[j]
			rv.lines = slices.Clone(rv.lines)
			slices.SortStableFunc(rv.lines, func(a, b Line) int {
				return cmp.Or(compareVals(d.lineVal(rv.r, a, key), d.lineVal(rv.r, b, key), order), strings.Compare(a.ID, b.ID))
			})
		}
		slices.SortStableFunc(out[i].requests, func(a, b reqView) int {
			return cmp.Or(compareVals(d.requestVal(a, key), d.requestVal(b, key), order), strings.Compare(a.r.ID, b.r.ID))
		})
	}
	slices.SortStableFunc(out, func(a, b vendorGroup) int {
		return cmp.Or(compareVals(d.groupVal(a, key), d.groupVal(b, key), order), strings.Compare(vendorLabel(a.vendor), vendorLabel(b.vendor)))
	})
	return out
}
