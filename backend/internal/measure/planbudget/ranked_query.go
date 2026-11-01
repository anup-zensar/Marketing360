package planbudget

import (
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"mapp/internal/platform/errs"
)

const (
	paramSort   = "sort"
	paramOrder  = "order"
	paramGroup  = "group"
	paramSize   = "size"
	paramPage   = "page"
	paramSearch = "q"
	paramDrill  = "drill"
)

const (
	SortBudget     = "budget"
	SortPlanned    = "planned"
	SortCommitted  = "committed"
	SortActual     = "actual"
	SortLeftToPlan = "leftToPlan"
	SortUnspent    = "unspent"
	OrderDesc      = "desc"
	OrderAsc       = "asc"
	GroupNone      = "none"
	GroupPlan      = "plan"
	GroupOwner     = "owner"
	groupAttr      = "attr:"
)

const (
	msgChooseListed = "Choose one of the listed values."
	msgSearchLong   = "Use 100 characters or fewer."
	maxSearch       = 100
	defaultSize     = 10
)

var (
	groupPattern = regexp.MustCompile(`^(none|plan|owner|attr:[A-Z]+-[0-9]+)$`)
	idPattern    = regexp.MustCompile(`^[A-Z]+-[0-9]+$`)
)

// RankedQuery is the ranked list's own part of the address (research R5).
type RankedQuery struct {
	Sort, Order, Group, Search, Drill string
	Size, Page                        int
}

// ParseRanked reads the ranked list's parameters; every problem is answered at once (FR-014).
func ParseRanked(values url.Values) (RankedQuery, error) {
	var problems errs.Problems
	q := RankedQuery{Sort: orDefault(values.Get(paramSort), SortPlanned), Order: orDefault(values.Get(paramOrder), OrderDesc),
		Group: orDefault(values.Get(paramGroup), GroupNone), Search: strings.TrimSpace(values.Get(paramSearch)),
		Drill: values.Get(paramDrill)}
	listed := map[string]bool{paramSort: slices.Contains(sorts(), q.Sort),
		paramOrder: q.Order == OrderDesc || q.Order == OrderAsc, paramGroup: groupPattern.MatchString(q.Group),
		paramDrill: q.Drill == "" || idPattern.MatchString(q.Drill)}
	for _, name := range []string{paramSort, paramOrder, paramGroup, paramDrill} {
		if !listed[name] {
			problems.Add(name, msgChooseListed)
		}
	}
	if utf8.RuneCountInString(q.Search) > maxSearch {
		problems.Add(paramSearch, msgSearchLong)
	}
	q.Size, q.Page = defaultSize, 1
	if n, ok := number(values, paramSize, &problems, func(n int) bool { return slices.Contains([]int{10, 25, 50}, n) }); ok {
		q.Size = n
	}
	if n, ok := number(values, paramPage, &problems, func(n int) bool { return n >= 1 }); ok {
		q.Page = n
	}
	return q, problems.Err()
}

func sorts() []string {
	return []string{SortBudget, SortPlanned, SortCommitted, SortActual, SortLeftToPlan, SortUnspent}
}

// number is the parameter's whole number when it is given and allowed; a refused one is a problem.
func number(values url.Values, name string, problems *errs.Problems, allowed func(int) bool) (int, bool) {
	text := values.Get(name)
	if text == "" {
		return 0, false
	}
	n, err := strconv.Atoi(text)
	if err != nil || !allowed(n) {
		problems.Add(name, msgChooseListed)
		return 0, false
	}
	return n, true
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
