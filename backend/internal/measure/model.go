// Package measure is what every report of the Reports area shares (feature 039): the catalogue and the registry
// reports plug into, the report filter, the frame of choices and counts, saved report views and the export.
package measure

import (
	"mapp/internal/allwork"
	"mapp/internal/platform/sheet"
)

// The report keys of the catalogue (data-model §3), in its order.
const (
	KeyPlanBudget       = "planBudget"
	KeyApprovals        = "approvals"
	KeyPurchaseRequests = "purchaseRequests"
	KeyPlanDetails      = "planDetails"
	KeyGsapLink         = "gsapLink"
	KeyUserAccess       = "userAccess"
	KeySummary          = "summary"
	KeyRevenue          = "revenue"
	KeyFunnel           = "funnel"
)

const (
	PeriodMonth   = "month"
	PeriodQuarter = "quarter"

	ShowPlan       = "plan"
	ShowCampaign   = "campaign"
	ShowInitiative = "initiative"
	ShowActivity   = "activity"
)

// Filter is the state a report is shown with: All work's filters plus the report's own controls (research R1).
type Filter struct {
	Query    allwork.Query
	Period   string
	Show     string
	Currency string
	Tables   []string
}

// WidgetSheet is one widget of a report as a sheet of the export; Key is the widget's key in the address.
type WidgetSheet struct {
	Key   string
	Sheet sheet.Sheet
}

type ReportEntry struct {
	Key       string `json:"key"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Purpose   string `json:"purpose"`
	Available bool   `json:"available"`
}

type ReportCatalogue struct {
	Items []ReportEntry `json:"items"`
}

type InView struct {
	Plans       int `json:"plans"`
	Campaigns   int `json:"campaigns"`
	Initiatives int `json:"initiatives"`
	Activities  int `json:"activities"`
	LineItems   int `json:"lineItems"`
}

type FinanceCoverage struct {
	Plans int `json:"plans"`
	Of    int `json:"of"`
}
