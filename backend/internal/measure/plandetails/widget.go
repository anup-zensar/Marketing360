package plandetails

import (
	"cmp"
	"slices"
	"strings"

	"mapp/internal/allwork"
	"mapp/internal/platform/money"
)

const (
	labelOther  = "Other"
	labelTotal  = "Total"
	chartValues = 10 // the values a chart draws before Other (FR-017)
)

const (
	suffixCount       = "Count"
	suffixPlannedCost = "PlannedCost"
	suffixFilledIn    = "FilledIn"
)

// attrKey is a Shell attribute's widget key in 039's pattern: SA-12 and Count give attr12Count.
func attrKey(attributeID, suffix string) string {
	return "attr" + strings.TrimPrefix(attributeID, "SA-") + suffix
}

type spec struct {
	key, title string
	measure    Measure
	filter     string
	timeOrder  bool // values in their own order, never gathered into Other
}

type bucket struct {
	part
	records map[string]bool
	usd     money.Amount
}

// tally gathers the values of one widget over the records: the distinct records and the USD of each value.
type tally struct {
	buckets      map[string]*bucket
	records      map[string]bool
	usd          money.Amount
	countsInEach bool
}

func newTally() *tally {
	return &tally{buckets: map[string]*bucket{}, records: map[string]bool{}}
}

// add counts the record under its counted parts; amount, the whole the parts divide, goes to the total once.
func (t *tally) add(recordID string, parts []part, amount money.Amount) {
	t.records[recordID] = true
	t.usd = t.usd.Add(amount)
	for _, p := range parts {
		b := t.bucketOf(p)
		b.usd = b.usd.Add(p.amount)
		if p.counted {
			b.records[recordID] = true
		}
	}
}

// seed shows the value even when no record has it: every period, Yes and No, Filled in and Not filled in.
func (t *tally) seed(parts ...part) {
	for _, p := range parts {
		t.bucketOf(p)
	}
}

func (t *tally) bucketOf(p part) *bucket {
	b, found := t.buckets[p.key]
	if !found {
		b = &bucket{part: p, records: map[string]bool{}}
		t.buckets[p.key] = b
	}
	return b
}

func (t *tally) widget(s spec, c allwork.Currency) Widget {
	buckets := t.sorted(s)
	w := Widget{Key: s.key, Title: s.title, Measure: s.measure, Values: []Value{}, CountsInEach: t.countsInEach,
		Total: s.value(part{key: KeyTotal, label: labelTotal}, len(t.records), t.usd, c)}
	var other *bucket
	regular := 0
	for _, b := range buckets {
		w.Values = append(w.Values, s.value(b.part, len(b.records), b.usd, c))
		if b.order >= orderSpecial || s.timeOrder {
			continue
		}
		if regular++; regular > chartValues {
			other = gather(other, b)
		}
	}
	if other != nil {
		v := s.value(part{key: KeyOther, label: labelOther}, len(other.records), other.usd, c)
		w.Other = &v
	}
	if s.filter != "" {
		filter := s.filter
		w.FilterParam = &filter
	}
	return w
}

func gather(other, b *bucket) *bucket {
	if other == nil {
		other = &bucket{records: map[string]bool{}}
	}
	for id := range b.records {
		other.records[b.key+"/"+id] = true
	}
	other.usd = other.usd.Add(b.usd)
	return other
}

// sorted puts the largest (or, by time order, the first) values first and Not filled in and the like last.
func (t *tally) sorted(s spec) []*bucket {
	buckets := make([]*bucket, 0, len(t.buckets))
	for _, b := range t.buckets {
		buckets = append(buckets, b)
	}
	slices.SortFunc(buckets, func(x, y *bucket) int {
		xs, ys := x.order >= orderSpecial, y.order >= orderSpecial
		switch {
		case xs != ys:
			return boolOrder(xs)
		case xs || s.timeOrder:
			return cmp.Compare(x.order, y.order)
		}
		size := cmp.Compare(len(y.records), len(x.records))
		if s.measure == MeasurePlannedCost {
			size = y.usd.Cmp(x.usd)
		}
		return cmp.Or(size, cmp.Compare(x.order, y.order), strings.Compare(x.key, y.key))
	})
	return buckets
}

func boolOrder(last bool) int {
	if last {
		return 1
	}
	return -1
}

func (s spec) value(p part, count int, usd money.Amount, c allwork.Currency) Value {
	v := Value{Key: p.key, Label: p.label, Inactive: p.inactive, Count: count, Linkable: s.filter != "" && p.linkable}
	if s.measure == MeasurePlannedCost {
		m := inCurrency(usd, c)
		v.PlannedCost = &m
	}
	return v
}

// inCurrency converts a USD sum once into the report's currency (D-303).
func inCurrency(usd money.Amount, c allwork.Currency) money.Money {
	return money.Money{Amount: money.FromUSD(usd, c.Rate), Currency: c.Code}
}

func wholePercent(part, whole int) *int64 {
	if whole == 0 {
		return nil
	}
	p := (int64(part)*200 + int64(whole)) / (int64(whole) * 2)
	return &p
}
