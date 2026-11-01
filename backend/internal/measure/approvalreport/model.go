// Package approvalreport is the Approvals report of the Measure module (feature 041): read-only, it works its
// figures out from the approval rounds of the plans in view and stores nothing.
package approvalreport

import (
	"time"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// Outcome is the word shown for a round or a step; Not needed belongs to steps only.
type Outcome string

const (
	OutcomeWaiting      Outcome = "waiting"
	OutcomeApproved     Outcome = "approved"
	OutcomeDeclined     Outcome = "declined"
	OutcomeRecalled     Outcome = "recalled"
	OutcomeReplaced     Outcome = "replaced"
	OutcomeAutoApproved Outcome = "autoApproved"
	OutcomeNotNeeded    Outcome = "notNeeded"
)

type RecordType string

const (
	RecordActivity                RecordType = "activity"
	RecordPurchaseRequest         RecordType = "purchaseRequest"
	RecordModifiedPurchaseRequest RecordType = "modifiedPurchaseRequest"
)

// View is the contract's ApprovalsReport; a nil amount, days or date is "—".
type View struct {
	StatusMix     StatusMix     `json:"statusMix"`
	Average       Average       `json:"average"`
	Grid          Grid          `json:"grid"`
	FinanceCovers FinanceCovers `json:"financeCovers"`
}

type StatusMix struct {
	RecordType  RecordType    `json:"recordType"`
	RecordTypes []RecordType  `json:"recordTypes"`
	Total       int           `json:"total"`
	Waiting     Waiting       `json:"waiting"`
	Items       []StatusCount `json:"items"`
}

type Waiting struct {
	Count       int        `json:"count"`
	OldestSince *time.Time `json:"oldestSince"`
}

type StatusCount struct {
	Outcome      Outcome `json:"outcome"`
	Count        int     `json:"count"`
	SharePercent int     `json:"sharePercent"`
}

type Average struct {
	Days      *string           `json:"days"`
	Decisions int               `json:"decisions"`
	Approvers []ApproverAverage `json:"approvers"`
}

type ApproverAverage struct {
	Person    PersonRef `json:"person"`
	Days      string    `json:"days"`
	Decisions int       `json:"decisions"`
}

type PersonRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Grid struct {
	Items    []Row       `json:"items"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Choices  GridChoices `json:"choices"`
}

type GridChoices struct {
	RecordTypes []RecordType `json:"recordTypes"`
	Outcomes    []Outcome    `json:"outcomes"`
	Approvers   []PersonRef  `json:"approvers"`
}

type Row struct {
	ItemID              string       `json:"itemId"`
	ItemName            string       `json:"itemName"`
	RecordType          RecordType   `json:"recordType"`
	CanOpen             bool         `json:"canOpen"`
	Approver            *PersonRef   `json:"approver"`
	DecidedOn           *date.Date   `json:"decidedOn"`
	Outcome             Outcome      `json:"outcome"`
	SubmittedOn         date.Date    `json:"submittedOn"`
	Level               int          `json:"level"`
	LevelName           string       `json:"levelName"`
	PurchaseOrderNumber *string      `json:"purchaseOrderNumber"`
	Amount              *money.Money `json:"amount"`
	Owner               PersonRef    `json:"owner"`
	Comments            string       `json:"comments"`
}

type FinanceCovers struct {
	Plans int `json:"plans"`
	Of    int `json:"of"`
}
