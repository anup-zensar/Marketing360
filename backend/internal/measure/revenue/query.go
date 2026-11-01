package revenue

import (
	"net/url"
	"strconv"
	"strings"

	"mapp/internal/platform/errs"
)

const (
	paramCategory = "category"
	paramStage    = "stage"
	paramSort     = "sort"
	paramOrder    = "order"
	paramGroup    = "group"
	paramSize     = "size"
	paramPage     = "page"
	paramSearch   = "q"
	paramDrill    = "drill"
)

const (
	StageINQ = "inq"
	StageMQL = "mql"
)

const (
	OrderDesc = "desc"
	OrderAsc  = "asc"
	GroupNone = "none"
)

const (
	SortActualRevenue    = "actualRevenue"
	SortProjectedRevenue = "projectedRevenue"
	SortActual           = "actual"
	SortProjected        = "projected"
)

const msgChooseListed = "Choose one of the listed values."

// categoryOf reads the category parameter: "" for all categories, "uncategorised", or a category id
// (049-C10).
func categoryOf(values url.Values) string {
	return strings.TrimSpace(values.Get(paramCategory))
}

// RankedQuery is the ranked list's own parameters, 040's shape (research R5) with this feature's sort keys.
type RankedQuery struct {
	Sort, Order, Group, Search, Drill, Stage string
	Size, Page                               int
}

func parseRanked(values url.Values, sorts []string, defaultSort string, problems *errs.Problems) RankedQuery {
	q := RankedQuery{
		Sort: orDefault(values.Get(paramSort), defaultSort), Order: orDefault(values.Get(paramOrder), OrderDesc),
		Group: orDefault(values.Get(paramGroup), GroupNone), Search: strings.TrimSpace(values.Get(paramSearch)),
		Drill: values.Get(paramDrill),
	}
	if !contains(sorts, q.Sort) {
		problems.Add(paramSort, msgChooseListed)
	}
	if q.Order != OrderDesc && q.Order != OrderAsc {
		problems.Add(paramOrder, msgChooseListed)
	}
	if len(q.Search) > 100 {
		problems.Add(paramSearch, "Search is at most 100 characters.")
	}
	q.Size = sizeOf(values.Get(paramSize))
	q.Page = pageOf(values.Get(paramPage))
	return q
}

func sizeOf(text string) int {
	n, err := strconv.Atoi(text)
	if err != nil || (n != 10 && n != 25 && n != 50) {
		return 10
	}
	return n
}

func pageOf(text string) int {
	n, err := strconv.Atoi(text)
	if err != nil || n < 1 {
		return 1
	}
	return n
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
