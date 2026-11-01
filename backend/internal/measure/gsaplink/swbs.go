package gsaplink

import (
	"slices"
	"time"

	"mapp/internal/allwork"
)

// swbsReady is All work's SWBS status of a line item Ready to send to GSAP (030 FR-001, 044-C1).
const swbsReady = "readyToSend"

func (y year) readyToSend() []allwork.Node {
	return slices.DeleteFunc(y.lineItems(), func(n allwork.Node) bool { return n.Swbs != swbsReady })
}

func activitiesOf(lineItems []allwork.Node) []string {
	ids, seen := []string{}, map[string]bool{}
	for _, li := range lineItems {
		if !seen[li.ParentID] {
			ids, seen[li.ParentID] = append(ids, li.ParentID), true
		}
	}
	return ids
}

// swbsTab lists the line items Ready to send, each waiting since its activity's latest approval (FR-005 – FR-007).
func (y year) swbsTab(ready []allwork.Node, approved map[string]time.Time) SwbsTab {
	tab := SwbsTab{Bands: bands(), Rows: make([]SwbsRow, 0, len(ready))}
	for _, li := range ready {
		activity := y.byID[li.ParentID]
		at, found := approved[activity.ID]
		if !found {
			at = activity.UpdatedAt // 044-C6: approved without an approval round
		}
		hours := waitingHours(at, y.now)
		row := SwbsRow{LineItemID: li.ID, LineItem: li.Name, Path: y.path(li), Owner: ownerName(activity),
			CompanyCode: y.companyCode(activity), ApprovedAt: at, WaitingHours: hours,
			OverTurnaround: hours > swbsTurnaroundHours}
		tab.Rows = append(tab.Rows, row)
		tab.Bands[slices.IndexFunc(tab.Bands, func(b Band) bool { return b.Key == band(hours) })].Count++
		if row.OverTurnaround {
			tab.OverTurnaround++
		}
	}
	tab.Waiting = len(tab.Rows)
	return tab
}
