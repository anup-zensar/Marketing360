// Package measure is what every report of the Reports area shares (feature 039): the catalogue and the registry
// reports plug into, the report filter, the frame of choices and counts, saved report views and the export.
package measure

import (
	"net/url"

	"time"

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
	Own      url.Values
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

// Frame is the contract's ReportFrame: what every report page shares (FR-006 – FR-019).
type Frame struct {
	Report          string          `json:"report"`
	Filter          AppliedFilter   `json:"filter"`
	Choices         Choices         `json:"choices"`
	InView          InView          `json:"inView"`
	Finance         FinanceCoverage `json:"finance"`
	LeftOut         int             `json:"leftOut"`
	DisplayCurrency string          `json:"displayCurrency"`
	GeneratedAt     time.Time       `json:"generatedAt"`
	FinanceAsOf     *time.Time      `json:"financeAsOf"`
}

type AppliedFilter struct {
	Year             int                  `json:"year"`
	Plans            []string             `json:"plans"`
	PlanTypes        []string             `json:"planTypes"`
	Teams            []string             `json:"teams"`
	Owners           []string             `json:"owners"`
	ActivityTypes    []string             `json:"activityTypes"`
	ActivityStatuses []string             `json:"activityStatuses"`
	Attributes       []allwork.AttrFilter `json:"attributes"`
	Period           string               `json:"period"`
	Show             *string              `json:"show"`
	Currency         string               `json:"currency"`
	Tables           []string             `json:"tables"`
}

type Choices struct {
	Years            []int            `json:"years"`
	Plans            []PlanChoice     `json:"plans"`
	PlanTypes        []string         `json:"planTypes"`
	Teams            []TeamChoice     `json:"teams"`
	Owners           []NamedChoice    `json:"owners"`
	ActivityTypes    []NamedChoice    `json:"activityTypes"`
	ActivityStatuses []string         `json:"activityStatuses"`
	AttributeGroups  []AttributeGroup `json:"attributeGroups"`
	Currencies       []Currency       `json:"currencies"`
}

type NamedChoice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PlanChoice struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	PlanType string   `json:"planType"`
	TeamIDs  []string `json:"teamIds"`
}

type TeamChoice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Mine bool   `json:"mine"`
}

// AttributeGroup is one Shell attribute of one level in the filter panel; never merged with another by label.
type AttributeGroup struct {
	Level       string           `json:"level"`
	AttributeID string           `json:"attributeId"`
	Label       string           `json:"label"`
	Values      []AttributeValue `json:"values"`
}

type AttributeValue struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}
