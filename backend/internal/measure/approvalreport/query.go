package approvalreport

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"mapp/internal/platform/errs"
)

const (
	msgChooseListed = "Choose one of the listed values."
	msgPage         = "Choose a page from 1."
)

// Query is the report's own part of the address. Unknown filter values are kept and match nothing (041-C10).
type Query struct {
	StatusType  RecordType
	RecordTypes []RecordType
	Approvers   []string
	Outcomes    []Outcome
	Search      string
	Sort        string
	Order       string
	Page        int
	PageSize    int
}

func Params() []string {
	return []string{"statusType", "recordType", "approver", "outcome", "q", "sort", "order", "page", "pageSize"}
}

func sortKeys() []string {
	return []string{"itemId", "itemName", "recordType", "approver", "decidedOn", "outcome", "submittedOn", "level",
		"purchaseOrder", "amount", "owner", "comments"}
}

func pageSizes() []int {
	return []int{5, 10, 25, 50, 100}
}

// ParseQuery reads the report's own parameters; a malformed sort, order, page or page size is a problem.
func ParseQuery(own url.Values) (Query, error) {
	var problems errs.Problems
	q := Query{StatusType: RecordType(own.Get("statusType")), Approvers: own["approver"],
		Search: strings.TrimSpace(own.Get("q")), Sort: own.Get("sort"), Order: own.Get("order")}
	for _, t := range own["recordType"] {
		q.RecordTypes = append(q.RecordTypes, RecordType(t))
	}
	for _, o := range own["outcome"] {
		q.Outcomes = append(q.Outcomes, Outcome(o))
	}
	if q.Sort != "" && !slices.Contains(sortKeys(), q.Sort) {
		problems.Add("sort", msgChooseListed)
	}
	if q.Order != "" && q.Order != "asc" && q.Order != "desc" {
		problems.Add("order", msgChooseListed)
	}
	q.Page = number(own, "page", msgPage, func(n int) bool { return n >= 1 }, &problems)
	q.PageSize = number(own, "pageSize", msgChooseListed, func(n int) bool { return slices.Contains(pageSizes(), n) }, &problems)
	return q, problems.Err()
}

func number(own url.Values, name, message string, valid func(int) bool, problems *errs.Problems) int {
	raw := own.Get(name)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || !valid(n) {
		problems.Add(name, message)
		return 0
	}
	return n
}
