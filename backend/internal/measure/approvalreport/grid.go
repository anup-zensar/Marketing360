package approvalreport

import (
	"cmp"
	"slices"
	"strings"

	"mapp/internal/platform/date"
)

// matching applies the grid's filters and search to all rows, then orders them; paging is left to the caller.
func matching(rows []Row, q Query) []Row {
	kept := slices.DeleteFunc(slices.Clone(rows), func(r Row) bool { return !wanted(r, q) })
	key, order := cmp.Or(q.Sort, "submittedOn"), cmp.Or(q.Order, "desc")
	slices.SortStableFunc(kept, func(a, b Row) int { return compareRows(a, b, key, order) })
	return kept
}

func wanted(r Row, q Query) bool {
	if len(q.RecordTypes) > 0 && !slices.Contains(q.RecordTypes, r.RecordType) {
		return false
	}
	if len(q.Outcomes) > 0 && !slices.Contains(q.Outcomes, r.Outcome) {
		return false
	}
	if len(q.Approvers) > 0 && (r.Approver == nil || !slices.Contains(q.Approvers, r.Approver.ID)) {
		return false
	}
	text := strings.ToLower(searchText(r))
	return !slices.ContainsFunc(strings.Fields(strings.ToLower(q.Search)), func(word string) bool {
		return !strings.Contains(text, word)
	})
}

func searchText(r Row) string {
	parts := []string{r.ItemID, r.ItemName, recordTypeLabel(r.RecordType), outcomeLabel(r.Outcome), r.LevelName,
		r.Owner.Name, r.Comments}
	if r.Approver != nil {
		parts = append(parts, r.Approver.Name)
	}
	if r.PurchaseOrderNumber != nil {
		parts = append(parts, *r.PurchaseOrderNumber)
	}
	return strings.Join(parts, "\n")
}

// compareRows orders by the chosen column, a missing value last in either order, ties by item then level.
func compareRows(a, b Row, key, order string) int {
	missingA, missingB := isMissing(a, key), isMissing(b, key)
	if missingA != missingB {
		if missingA {
			return 1
		}
		return -1
	}
	by := compareBy(a, b, key)
	if order == "desc" {
		by = -by
	}
	return cmp.Or(by, strings.Compare(a.ItemID, b.ItemID), cmp.Compare(a.Level, b.Level))
}

func isMissing(r Row, key string) bool {
	switch key {
	case "approver":
		return r.Approver == nil
	case "decidedOn":
		return r.DecidedOn == nil
	case "purchaseOrder":
		return r.PurchaseOrderNumber == nil
	case "amount":
		return r.Amount == nil
	}
	return false
}

func compareBy(a, b Row, key string) int {
	switch key {
	case "itemId":
		return strings.Compare(a.ItemID, b.ItemID)
	case "itemName":
		return strings.Compare(strings.ToLower(a.ItemName), strings.ToLower(b.ItemName))
	case "recordType":
		return strings.Compare(recordTypeLabel(a.RecordType), recordTypeLabel(b.RecordType))
	case "approver":
		return strings.Compare(strings.ToLower(a.Approver.Name), strings.ToLower(b.Approver.Name))
	case "decidedOn":
		return a.DecidedOn.Compare(*b.DecidedOn)
	case "outcome":
		return strings.Compare(outcomeLabel(a.Outcome), outcomeLabel(b.Outcome))
	case "submittedOn":
		return date.Date.Compare(a.SubmittedOn, b.SubmittedOn)
	case "level":
		return cmp.Compare(a.Level, b.Level)
	case "purchaseOrder":
		return strings.Compare(*a.PurchaseOrderNumber, *b.PurchaseOrderNumber)
	case "amount":
		return a.Amount.Amount.Cmp(b.Amount.Amount)
	case "owner":
		return strings.Compare(strings.ToLower(a.Owner.Name), strings.ToLower(b.Owner.Name))
	case "comments":
		return strings.Compare(strings.ToLower(a.Comments), strings.ToLower(b.Comments))
	}
	return 0
}

// pageOf cuts the page the person asked for out of the ordered rows.
func pageOf(rows []Row, q Query) Grid {
	size, page := cmp.Or(q.PageSize, defaultPageSize), cmp.Or(q.Page, 1)
	from := min((page-1)*size, len(rows))
	to := min(from+size, len(rows))
	return Grid{Items: slices.Clone(rows[from:to]), Total: len(rows), Page: page, PageSize: size}
}

// choicesOf offers what the filters can pick from, whatever is filtered now.
func choicesOf(rows []Row) GridChoices {
	choices := GridChoices{RecordTypes: []RecordType{}, Outcomes: append(statusOutcomes(), OutcomeNotNeeded), Approvers: []PersonRef{}}
	for _, t := range allRecordTypes() {
		if slices.ContainsFunc(rows, func(r Row) bool { return r.RecordType == t }) {
			choices.RecordTypes = append(choices.RecordTypes, t)
		}
	}
	for _, r := range rows {
		if r.Approver != nil && !slices.Contains(choices.Approvers, *r.Approver) {
			choices.Approvers = append(choices.Approvers, *r.Approver)
		}
	}
	slices.SortFunc(choices.Approvers, func(a, b PersonRef) int {
		return cmp.Or(strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), strings.Compare(a.ID, b.ID))
	})
	return choices
}
