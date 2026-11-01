package approvalreport

import (
	"cmp"
	"slices"
	"strings"
)

// matching applies the grid's filters and search to all rows, then orders them; paging is left to the caller.
func matching(rows []Row, q Query) []Row {
	kept := slices.DeleteFunc(slices.Clone(rows), func(r Row) bool { return !wanted(r, q) })
	key, order := cmp.Or(q.Sort, sortSubmittedOn), cmp.Or(q.Order, orderDesc)
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
