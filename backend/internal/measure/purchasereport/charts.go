package purchasereport

import (
	"cmp"
	"slices"
	"strings"

	"mapp/internal/platform/money"
)

const chartNamed = 10

// grouping says which series a request's commitment belongs to: a key and the name shown.
type grouping func(d data, r Request) (key, name string)

func byVendor(_ data, r Request) (string, string) {
	if r.Vendor == "" {
		return noVendor, notFilledIn
	}
	return r.Vendor, r.Vendor
}

func byOwner(d data, r Request) (string, string) {
	return r.OwnerID, d.owner(r.OwnerID).Name
}

// seriesTally is one vendor or owner with its commitment per month.
type seriesTally struct {
	key, name string
	cells     map[string]*tally
	total     tally
}

func (s *seriesTally) add(month string, usd money.Amount, visible bool) {
	if s.cells[month] == nil {
		s.cells[month] = &tally{}
	}
	s.cells[month].add(usd, visible)
	s.total.add(usd, visible)
}

// chart works out the commitment of the Approved and Closed lines per delivery month for each series (042-C1).
func (d data) chart(rs []Request, group grouping) Chart {
	series := map[string]*seriesTally{}
	months := map[string]*tally{}
	var all tally
	for _, r := range rs {
		if !isCommitted(r.Status) {
			continue
		}
		key, name := group(d, r)
		if series[key] == nil {
			series[key] = &seriesTally{key: key, name: name, cells: map[string]*tally{}}
		}
		visible := d.sees(r.PlanID)
		for _, l := range r.Lines {
			month := monthOf(l.DeliveryDate)
			series[key].add(month, l.Amount, visible)
			if months[month] == nil {
				months[month] = &tally{}
			}
			months[month].add(l.Amount, visible)
			all.add(l.Amount, visible)
		}
	}
	keys := make([]string, 0, len(months))
	for k := range months {
		keys = append(keys, k)
	}
	axis := monthRange(keys)
	ranked := rankedSeries(series)
	chart := Chart{Months: axis, AllSeries: d.render(ranked, axis), ChartSeries: d.render(withOther(ranked), axis),
		Totals: make([]*money.Money, 0, len(axis)), Committed: all.money(d.shown)}
	for _, m := range axis {
		t := months[m]
		chart.Totals = append(chart.Totals, d.cell(t))
	}
	return chart
}

func rankedSeries(series map[string]*seriesTally) []*seriesTally {
	list := make([]*seriesTally, 0, len(series))
	for _, s := range series {
		list = append(list, s)
	}
	slices.SortFunc(list, func(a, b *seriesTally) int {
		return cmp.Or(b.total.usd.Cmp(a.total.usd), strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name)))
	})
	return list
}

// withOther keeps the ten largest and sums the rest into Other (FR-011).
func withOther(ranked []*seriesTally) []*seriesTally {
	if len(ranked) <= chartNamed {
		return ranked
	}
	other := &seriesTally{key: "other", name: "Other", cells: map[string]*tally{}}
	for _, s := range ranked[chartNamed:] {
		for month, t := range s.cells {
			if other.cells[month] == nil {
				other.cells[month] = &tally{}
			}
			other.cells[month].merge(*t)
		}
		other.total.merge(s.total)
	}
	return append(slices.Clone(ranked[:chartNamed]), other)
}

func (d data) render(list []*seriesTally, axis []string) []Series {
	out := make([]Series, 0, len(list))
	for _, s := range list {
		values := make([]*money.Money, 0, len(axis))
		for _, m := range axis {
			values = append(values, d.cell(s.cells[m]))
		}
		out = append(out, Series{Key: s.key, Name: s.name, Values: values})
	}
	return out
}

// cell is a month's value: nothing committed is a real 0.00, and only amounts of hidden plans are a dash.
func (d data) cell(t *tally) *money.Money {
	if t == nil {
		return d.shown.of(money.Amount{})
	}
	return t.money(d.shown)
}
