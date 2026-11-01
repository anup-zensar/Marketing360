package gsaplink

import (
	"cmp"
	"maps"
	"slices"
	"strings"

	"mapp/internal/platform/money"
)

// Paging of every tab (FR-015).
const (
	DefaultPageSize = 20
	MinPageSize     = 5
	MaxPageSize     = 500
	descending      = "-"
)

const (
	sortLineItem        = "lineItem"
	sortPlan            = "plan"
	sortOwner           = "owner"
	sortCompanyCode     = "companyCode"
	sortApprovedAt      = "approvedAt"
	sortWaiting         = "waiting"
	sortPurchaseRequest = "purchaseRequest"
	sortVendor          = "vendor"
	sortRoutedAt        = "routedAt"
	sortAmount          = "amount"
	sortPoNumber        = "poNumber"
	sortPoDate          = "poDate"
	sortPoAmount        = "poAmount"
	sortInvoiced        = "invoiced"
	sortPaid            = "paid"
	sortGsapPrNumber    = "gsapPrNumber"
)

type compare[R any] func(a, b R) int

// DefaultSort is longest waiting first on the first two tabs, line item then PO number on the third (FR-015).
func DefaultSort(tab string) string {
	if tab == TabOrders {
		return sortLineItem
	}
	return descending + sortWaiting
}

// ValidSort says whether sort names a column of the tab, "-" first for descending.
func ValidSort(tab, sort string) bool {
	key := strings.TrimPrefix(sort, descending)
	switch tab {
	case TabSwbs:
		return swbsColumns()[key] != nil
	case TabWaiting:
		return waitingColumns()[key] != nil
	case TabOrders:
		return orderColumns()[key] != nil
	}
	return false
}

func SortKeys(tab string) []string {
	switch tab {
	case TabWaiting:
		return slices.Sorted(maps.Keys(waitingColumns()))
	case TabOrders:
		return slices.Sorted(maps.Keys(orderColumns()))
	}
	return slices.Sorted(maps.Keys(swbsColumns()))
}

// swbsColumns are the SWBS tab's sorts; on both waiting tabs a wait is longer the earlier it started.
func swbsColumns() map[string]compare[SwbsRow] {
	return map[string]compare[SwbsRow]{
		sortLineItem:    func(a, b SwbsRow) int { return compareText(a.LineItem, b.LineItem) },
		sortPlan:        func(a, b SwbsRow) int { return compareText(pathKey(a.Path), pathKey(b.Path)) },
		sortOwner:       func(a, b SwbsRow) int { return compareText(a.Owner, b.Owner) },
		sortCompanyCode: func(a, b SwbsRow) int { return compareText(a.CompanyCode, b.CompanyCode) },
		sortApprovedAt:  func(a, b SwbsRow) int { return a.ApprovedAt.Compare(b.ApprovedAt) },
		sortWaiting:     func(a, b SwbsRow) int { return b.ApprovedAt.Compare(a.ApprovedAt) },
	}
}

func waitingColumns() map[string]compare[WaitingRow] {
	return map[string]compare[WaitingRow]{
		sortPurchaseRequest: func(a, b WaitingRow) int { return compareText(a.Name, b.Name) },
		sortVendor:          func(a, b WaitingRow) int { return compareText(a.Vendor, b.Vendor) },
		sortOwner:           func(a, b WaitingRow) int { return compareText(a.Owner, b.Owner) },
		sortPlan:            func(a, b WaitingRow) int { return compareText(pathKey(a.Path), pathKey(b.Path)) },
		sortGsapPrNumber:    func(a, b WaitingRow) int { return compareIDs(a.GsapPrNumber, b.GsapPrNumber) },
		sortRoutedAt:        func(a, b WaitingRow) int { return a.RoutedAt.Compare(b.RoutedAt) },
		sortAmount:          func(a, b WaitingRow) int { return compareMoney(a.Amount, b.Amount) },
		sortWaiting:         func(a, b WaitingRow) int { return b.RoutedAt.Compare(a.RoutedAt) },
	}
}

func orderColumns() map[string]compare[OrderRow] {
	return map[string]compare[OrderRow]{
		sortLineItem:        func(a, b OrderRow) int { return compareText(a.LineItem, b.LineItem) },
		sortPlan:            func(a, b OrderRow) int { return compareText(pathKey(a.Path), pathKey(b.Path)) },
		sortPurchaseRequest: func(a, b OrderRow) int { return compareIDs(a.PurchaseRequestID, b.PurchaseRequestID) },
		sortGsapPrNumber:    func(a, b OrderRow) int { return compareIDs(a.GsapPrNumber, b.GsapPrNumber) },
		sortPoNumber:        func(a, b OrderRow) int { return compareIDs(a.PoNumber, b.PoNumber) },
		sortPoDate:          func(a, b OrderRow) int { return a.PoDate.Compare(b.PoDate) },
		sortVendor:          func(a, b OrderRow) int { return compareText(a.Vendor, b.Vendor) },
		sortPoAmount:        func(a, b OrderRow) int { return a.PoAmount.Amount.Cmp(b.PoAmount.Amount) },
		sortInvoiced:        func(a, b OrderRow) int { return a.Invoiced.Amount.Cmp(b.Invoiced.Amount) },
		sortPaid:            func(a, b OrderRow) int { return a.Paid.Amount.Cmp(b.Paid.Amount) },
	}
}

func swbsID(a, b SwbsRow) int       { return compareIDs(a.LineItemID, b.LineItemID) }
func waitingID(a, b WaitingRow) int { return compareIDs(a.PurchaseRequestID, b.PurchaseRequestID) }
func orderID(a, b OrderRow) int {
	return cmp.Or(compareIDs(a.LineItemID, b.LineItemID), compareIDs(a.PoNumber, b.PoNumber))
}

// sortRows orders rows by one column, "-" first for descending; ties break by ID, ascending (FR-015).
func sortRows[R any](rows []R, columns map[string]compare[R], sort string, id compare[R]) {
	by := columns[strings.TrimPrefix(sort, descending)]
	sign := 1
	if strings.HasPrefix(sort, descending) {
		sign = -1
	}
	slices.SortFunc(rows, func(a, b R) int { return cmp.Or(sign*by(a, b), id(a, b)) })
}

func pageOf[R any](rows []R, page, size int) []R {
	from := min((page-1)*size, len(rows))
	return rows[from:min(from+size, len(rows))]
}

func compareText(a, b string) int {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

// compareIDs puts L-999 before L-1000.
func compareIDs(a, b string) int {
	return cmp.Or(cmp.Compare(len(a), len(b)), strings.Compare(a, b))
}

// compareMoney puts an amount the person may not see ("—") first.
func compareMoney(a, b *money.Money) int {
	if a == nil || b == nil {
		return cmp.Compare(boolInt(a != nil), boolInt(b != nil))
	}
	return a.Amount.Cmp(b.Amount)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// pathKey sorts by the path as the screen shows it: plan, then campaign, initiative and activity.
func pathKey(p Path) string {
	return strings.Join([]string{p.Plan, p.Campaign, p.Initiative, p.Activity}, "\x00")
}
