// Package gsaplink is the GSAP link report (feature 044): SWBS not received, Waiting for GSAP, Purchase orders.
package gsaplink

import (
	"time"

	"mapp/internal/measure"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// The report's tabs (FR-001).
const (
	TabSwbs    = "swbs"
	TabWaiting = "waiting"
	TabOrders  = "orders"
)

// The routings GSAP answers (032).
const (
	RoutingOriginal     = "original"
	RoutingModification = "modification"
)

// The bands of the Waiting by time chart (FR-007).
const (
	BandUnder48h   = "under48h"
	BandDays2to5   = "days2to5"
	BandDays5to10  = "days5to10"
	BandOver10Days = "over10days"
)

// Report is the contract's GsapLinkReport: one tab with its cards, chart and one page of rows.
type Report struct {
	Tab      string                   `json:"tab"`
	Currency string                   `json:"currency"`
	Total    int                      `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
	Sort     string                   `json:"sort"`
	Finance  *measure.FinanceCoverage `json:"finance"`
	Swbs     *SwbsTab                 `json:"swbs"`
	Waiting  *WaitingTab              `json:"waiting"`
	Orders   *OrdersTab               `json:"orders"`
}

type Path struct {
	PlanID     string `json:"planId"`
	Plan       string `json:"plan"`
	Campaign   string `json:"campaign"`
	Initiative string `json:"initiative"`
	ActivityID string `json:"activityId"`
	Activity   string `json:"activity"`
}

type Band struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type SwbsTab struct {
	Waiting        int       `json:"waiting"`
	OverTurnaround int       `json:"overTurnaround"`
	Bands          []Band    `json:"bands"`
	Rows           []SwbsRow `json:"rows"`
}

type SwbsRow struct {
	LineItemID     string    `json:"lineItemId"`
	LineItem       string    `json:"lineItem"`
	Path           Path      `json:"path"`
	Owner          string    `json:"owner"`
	CompanyCode    string    `json:"companyCode"`
	ApprovedAt     time.Time `json:"approvedAt"`
	WaitingHours   int       `json:"waitingHours"`
	OverTurnaround bool      `json:"overTurnaround"`
}

type WaitingTab struct {
	Waiting        int          `json:"waiting"`
	Amount         *money.Money `json:"amount"`
	OverTurnaround int          `json:"overTurnaround"`
	Rows           []WaitingRow `json:"rows"`
}

type WaitingRow struct {
	PurchaseRequestID  string       `json:"purchaseRequestId"`
	Name               string       `json:"name"`
	Vendor             string       `json:"vendor"`
	Owner              string       `json:"owner"`
	Path               Path         `json:"path"`
	RoutedAt           time.Time    `json:"routedAt"`
	Routing            string       `json:"routing"`
	GsapPrNumber       string       `json:"gsapPrNumber"`
	Amount             *money.Money `json:"amount"`
	WaitingWorkingDays int          `json:"waitingWorkingDays"`
	OverTurnaround     bool         `json:"overTurnaround"`
}

type OrdersTab struct {
	Rows   []OrderRow  `json:"rows"`
	Totals OrderTotals `json:"totals"`
}

type OrderTotals struct {
	PoAmount money.Money `json:"poAmount"`
	Invoiced money.Money `json:"invoiced"`
	Paid     money.Money `json:"paid"`
}

type OrderRow struct {
	LineItemID        string      `json:"lineItemId"`
	LineItem          string      `json:"lineItem"`
	Path              Path        `json:"path"`
	PurchaseRequestID string      `json:"purchaseRequestId"`
	GsapPrNumber      string      `json:"gsapPrNumber"`
	PoNumber          string      `json:"poNumber"`
	PoLines           []string    `json:"poLines"`
	PoDate            date.Date   `json:"poDate"`
	Vendor            string      `json:"vendor"`
	PoAmount          money.Money `json:"poAmount"`
	Invoiced          money.Money `json:"invoiced"`
	Paid              money.Money `json:"paid"`
}

// Table is the tab asked for, its sort ("-" for descending, "" for the default) and its page (FR-015).
type Table struct {
	Tab      string
	Sort     string
	Page     int
	PageSize int
}
