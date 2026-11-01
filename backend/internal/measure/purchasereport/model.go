// Package purchasereport is the Purchase requests report of the Measure module (feature 042): read-only, it works
// its figures out from the purchase requests on the line items in view and stores nothing.
package purchasereport

import (
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

type Status string

const (
	StatusCreated   Status = "created"
	StatusSubmitted Status = "submitted"
	StatusInReview  Status = "inReview"
	StatusDeclined  Status = "declined"
	StatusApproved  Status = "approved"
	StatusClosed    Status = "closed"
)

// View is the contract's PurchaseRequestsReport; a nil amount or date is "—".
type View struct {
	PurchaseSupport bool          `json:"purchaseSupport"`
	Cards           Cards         `json:"cards"`
	StatusMix       StatusMix     `json:"statusMix"`
	ByVendor        Chart         `json:"byVendor"`
	ByOwner         Chart         `json:"byOwner"`
	Grid            Grid          `json:"grid"`
	FinanceCovers   FinanceCovers `json:"financeCovers"`
}

type Cards struct {
	PurchaseRequests int          `json:"purchaseRequests"`
	Lines            int          `json:"lines"`
	Vendors          int          `json:"vendors"`
	Committed        *money.Money `json:"committed"`
	Invoiced         *money.Money `json:"invoiced"`
}

type StatusMix struct {
	Total          int           `json:"total"`
	TotalAmount    *money.Money  `json:"totalAmount"`
	WaitingForGSAP int           `json:"waitingForGsap"`
	Items          []StatusCount `json:"items"`
}

type StatusCount struct {
	Status       Status       `json:"status"`
	Count        int          `json:"count"`
	Amount       *money.Money `json:"amount"`
	SharePercent int          `json:"sharePercent"`
}

type Chart struct {
	Months      []string       `json:"months"`
	ChartSeries []Series       `json:"chartSeries"`
	AllSeries   []Series       `json:"allSeries"`
	Totals      []*money.Money `json:"totals"`
	Committed   *money.Money   `json:"committed"`
}

type Series struct {
	Key    string         `json:"key"`
	Name   string         `json:"name"`
	Values []*money.Money `json:"values"`
}

type PersonRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Grid struct {
	Vendors  []VendorNode `json:"vendors"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
	Columns  []string     `json:"columns"`
	Choices  GridChoices  `json:"choices"`
}

type GridChoices struct {
	Vendors  []string    `json:"vendors"`
	Owners   []PersonRef `json:"owners"`
	Statuses []Status    `json:"statuses"`
}

type VendorNode struct {
	Vendor           string        `json:"vendor"`
	Label            string        `json:"label"`
	PurchaseRequests int           `json:"purchaseRequests"`
	Lines            int           `json:"lines"`
	Committed        *money.Money  `json:"committed"`
	Invoiced         *money.Money  `json:"invoiced"`
	Requests         []RequestNode `json:"requests"`
}

type RequestNode struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Status         Status       `json:"status"`
	Amount         *money.Money `json:"amount"`
	Invoiced       *money.Money `json:"invoiced"`
	Owner          PersonRef    `json:"owner"`
	CreatedOn      date.Date    `json:"createdOn"`
	SentOn         *date.Date   `json:"sentOn"`
	DeliveryDate   *date.Date   `json:"deliveryDate"`
	PONumber       *string      `json:"poNumber"`
	Workflow       *string      `json:"workflow"`
	GSAPNumber     *string      `json:"gsapNumber"`
	WaitingForGSAP bool         `json:"waitingForGsap"`
	Lines          []LineNode   `json:"lines"`
}

type LineNode struct {
	ID           string       `json:"id"`
	LineItem     LineItemRef  `json:"lineItem"`
	Activity     *string      `json:"activity"`
	Amount       *money.Money `json:"amount"`
	DeliveryDate *date.Date   `json:"deliveryDate"`
	SentOn       *date.Date   `json:"sentOn"`
	PONumber     *string      `json:"poNumber"`
	POLine       *string      `json:"poLine"`
	POLineStatus *string      `json:"poLineStatus"`
	Invoiced     *money.Money `json:"invoiced"`
}

type LineItemRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FinanceCovers struct {
	Plans int `json:"plans"`
	Of    int `json:"of"`
}
