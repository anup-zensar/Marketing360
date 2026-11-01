package purchasereport

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"mapp/internal/platform/errs"
)

const (
	paramVendor   = "vendor"
	paramOwner    = "requestOwner"
	paramStatus   = "status"
	paramColumns  = "columns"
	paramQ        = "q"
	paramSort     = "sort"
	paramOrder    = "order"
	paramPage     = "page"
	paramPageSize = "pageSize"
	orderAsc      = "asc"
	orderDesc     = "desc"
	noVendor      = "none"

	colName         = "name"
	colID           = "id"
	colStatus       = "status"
	colAmount       = "amount"
	colOwner        = "owner"
	colCreated      = "created"
	colDelivery     = "delivery"
	colSent         = "sent"
	colPONumber     = "poNumber"
	colPOLine       = "poLine"
	colPOLineStatus = "poLineStatus"
	colInvoiced     = "invoiced"
	colWorkflow     = "workflow"

	msgChooseListed = "Choose one of the listed values."
	msgPage         = "Choose a page from 1."
	msgColumns      = "Choose columns from the listed ones, each once."
)

// Query is the report's own part of the address; unknown filter values are kept and match nothing (042-C6).
type Query struct {
	Vendors       []string
	RequestOwners []string
	Statuses      []Status
	Search        string
	Sort          string
	Order         string
	Page          int
	PageSize      int
	Columns       []string
}

func Params() []string {
	return []string{paramVendor, paramOwner, paramStatus, paramQ, paramSort, paramOrder, paramPage, paramPageSize, paramColumns}
}

func sortKeys() []string {
	return []string{colName, colStatus, colAmount, colInvoiced, colCreated, colDelivery}
}

func pageSizes() []int {
	return []int{5, 10, 25, 50, 100}
}

func allColumns() []string {
	return []string{colName, colID, colStatus, colAmount, colOwner, colCreated, colDelivery, colSent, colPONumber, colPOLine,
		colPOLineStatus, colInvoiced, colWorkflow}
}

// ParseQuery reads the report's own parameters; a malformed sort, order, page, page size or column is a problem.
func ParseQuery(own url.Values) (Query, error) {
	var problems errs.Problems
	q := Query{Vendors: own[paramVendor], RequestOwners: own[paramOwner], Search: strings.TrimSpace(own.Get(paramQ)),
		Sort: own.Get(paramSort), Order: own.Get(paramOrder)}
	for _, s := range own[paramStatus] {
		q.Statuses = append(q.Statuses, Status(s))
	}
	if q.Sort != "" && !slices.Contains(sortKeys(), q.Sort) {
		problems.Add(paramSort, msgChooseListed)
	}
	if q.Order != "" && q.Order != orderAsc && q.Order != orderDesc {
		problems.Add(paramOrder, msgChooseListed)
	}
	q.Page = number(own, numberRule{paramPage, msgPage, func(n int) bool { return n >= 1 }}, &problems)
	q.PageSize = number(own, numberRule{paramPageSize, msgChooseListed, func(n int) bool { return slices.Contains(pageSizes(), n) }}, &problems)
	q.Columns = columnsOf(own.Get(paramColumns), &problems)
	return q, problems.Err()
}

type numberRule struct {
	name, message string
	valid         func(int) bool
}

func number(own url.Values, rule numberRule, problems *errs.Problems) int {
	raw := own.Get(rule.name)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || !rule.valid(n) {
		problems.Add(rule.name, rule.message)
		return 0
	}
	return n
}

func columnsOf(raw string, problems *errs.Problems) []string {
	if raw == "" {
		return allColumns()
	}
	var chosen []string
	for _, c := range strings.Split(raw, ",") {
		if !slices.Contains(allColumns(), c) || slices.Contains(chosen, c) {
			problems.Add(paramColumns, msgColumns)
			return allColumns()
		}
		chosen = append(chosen, c)
	}
	return chosen
}
