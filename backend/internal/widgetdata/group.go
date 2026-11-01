package widgetdata

import (
	"cmp"
	"strings"
)

const (
	labelNotSet = "Not set"
	labelOther  = "Other"
)

// groupKey is one group: a value (text, the decimal of a number, the bucket's first day), "Not set" or "Other".
type groupKey struct {
	kind  string
	text  string
	units int64
}

var (
	emptyKey = groupKey{kind: KeyEmpty} //nolint:gochecknoglobals // a constant value
	otherKey = groupKey{kind: KeyOther} //nolint:gochecknoglobals // a constant value
)

type group struct {
	key  *groupKey
	rows []int
}

func (k groupKey) label() string {
	switch k.kind {
	case KeyEmpty:
		return labelNotSet
	case KeyOther:
		return labelOther
	}
	return k.text
}

func (k groupKey) answer() *Key {
	out := &Key{Label: k.label(), Kind: k.kind}
	if k.kind == KeyValue {
		value := k.text
		out.Value = &value
	}
	return out
}

// keysOf are the groups a row counts in: one, or each value of a multi field.
func (g grouping) keysOf(r Row) []groupKey {
	v := r[g.index]
	if isEmpty(v, g.field) {
		return []groupKey{emptyKey}
	}
	switch g.field.Kind {
	case KindDate:
		return []groupKey{{kind: KeyValue, text: bucketOf(v.Date, g.bucket).String()}}
	case KindNumber:
		return []groupKey{{kind: KeyValue, text: v.Number.String(), units: v.Number.Units()}}
	case KindText, KindMoney:
	}
	var keys []groupKey
	for _, t := range textsOf(v, g.field) {
		k := groupKey{kind: KeyValue, text: t}
		if !containsKey(keys, k) {
			keys = append(keys, k)
		}
	}
	return keys
}

// groupRows groups the rows idx by g in the order first seen; without g it is one group of every row.
func groupRows(rows []Row, idx []int, g *grouping) (groups []group, multi bool) {
	if g == nil {
		return []group{{rows: idx}}, false
	}
	at := map[groupKey]int{}
	for _, i := range idx {
		keys := g.keysOf(rows[i])
		multi = multi || len(keys) > 1
		for _, k := range keys {
			pos, found := at[k]
			if !found {
				pos = len(groups)
				at[k] = pos
				key := k
				groups = append(groups, group{key: &key})
			}
			groups[pos].rows = append(groups[pos].rows, i)
		}
	}
	return groups, multi
}

// compareKeys orders values by kind, then "Not set", then "Other".
func compareKeys(a, b groupKey, kind FieldKind) int {
	if a.kind != b.kind {
		return cmp.Compare(keyRank(a.kind), keyRank(b.kind))
	}
	if kind == KindNumber {
		return cmp.Compare(a.units, b.units)
	}
	if kind == KindDate {
		return strings.Compare(a.text, b.text)
	}
	return cmp.Or(strings.Compare(strings.ToLower(a.text), strings.ToLower(b.text)), strings.Compare(a.text, b.text))
}

func keyRank(kind string) int {
	switch kind {
	case KeyEmpty:
		return 1
	case KeyOther:
		return 2
	}
	return 0
}

func containsKey(keys []groupKey, k groupKey) bool {
	for _, x := range keys {
		if x == k {
			return true
		}
	}
	return false
}
