package approvalreport

import (
	"cmp"
	"slices"
	"strings"
)

const (
	sortItemID        = "itemId"
	sortItemName      = "itemName"
	sortRecordType    = "recordType"
	sortApprover      = "approver"
	sortDecidedOn     = "decidedOn"
	sortOutcome       = "outcome"
	sortSubmittedOn   = "submittedOn"
	sortLevel         = "level"
	sortPurchaseOrder = "purchaseOrder"
	sortAmount        = "amount"
	sortOwner         = "owner"
	sortComments      = "comments"

	orderAsc  = "asc"
	orderDesc = "desc"
)

// column is one sortable column of the grid; missing says a row has no value in it, and such rows sort last.
type column struct {
	key     string
	missing func(Row) bool
	compare func(a, b Row) int
}

func never(Row) bool { return false }

func byText(text func(Row) string) func(a, b Row) int {
	return func(a, b Row) int { return strings.Compare(strings.ToLower(text(a)), strings.ToLower(text(b))) }
}

func columns() []column {
	return []column{
		{sortItemID, never, byText(func(r Row) string { return r.ItemID })},
		{sortItemName, never, byText(func(r Row) string { return r.ItemName })},
		{sortRecordType, never, byText(func(r Row) string { return recordTypeLabel(r.RecordType) })},
		{sortApprover, func(r Row) bool { return r.Approver == nil }, byText(func(r Row) string { return r.Approver.Name })},
		{sortDecidedOn, func(r Row) bool { return r.DecidedOn == nil }, func(a, b Row) int { return a.DecidedOn.Compare(*b.DecidedOn) }},
		{sortOutcome, never, byText(func(r Row) string { return outcomeLabel(r.Outcome) })},
		{sortSubmittedOn, never, func(a, b Row) int { return a.SubmittedOn.Compare(b.SubmittedOn) }},
		{sortLevel, never, func(a, b Row) int { return cmp.Compare(a.Level, b.Level) }},
		{sortPurchaseOrder, func(r Row) bool { return r.PurchaseOrderNumber == nil }, byText(func(r Row) string { return *r.PurchaseOrderNumber })},
		{sortAmount, func(r Row) bool { return r.Amount == nil }, func(a, b Row) int { return a.Amount.Amount.Cmp(b.Amount.Amount) }},
		{sortOwner, never, byText(func(r Row) string { return r.Owner.Name })},
		{sortComments, never, byText(func(r Row) string { return r.Comments })},
	}
}

func sortKeys() []string {
	keys := []string{}
	for _, c := range columns() {
		keys = append(keys, c.key)
	}
	return keys
}

// compareRows orders by the chosen column, a missing value last in either order, ties by item then level.
func compareRows(a, b Row, key, order string) int {
	i := slices.IndexFunc(columns(), func(c column) bool { return c.key == key })
	if i < 0 {
		return 0
	}
	c := columns()[i]
	if missingA, missingB := c.missing(a), c.missing(b); missingA != missingB {
		if missingA {
			return 1
		}
		return -1
	}
	by := c.compare(a, b)
	if order == orderDesc {
		by = -by
	}
	return cmp.Or(by, strings.Compare(a.ItemID, b.ItemID), cmp.Compare(a.Level, b.Level))
}
