// Package plandetails is the Plan details report (feature 043): read-only, it works out how the records and the
// planned cost of one level spread over each Shell attribute, and what is not filled in, from All work's year.
package plandetails

import "mapp/internal/platform/money"

type Level string

const (
	LevelPlan       Level = "plan"
	LevelCampaign   Level = "campaign"
	LevelInitiative Level = "initiative"
	LevelActivity   Level = "activity"
)

type Measure string

const (
	MeasureCount       Measure = "count"
	MeasurePlannedCost Measure = "plannedCost"
	MeasureFilledIn    Measure = "filledIn"
)

const (
	KeyNotFilledIn  = "none"
	KeyDoesNotApply = "doesNotApply"
	KeyNoDates      = "noDates"
	KeyOutsideYear  = "outsideYear"
	KeyFilled       = "filled"
	KeyOther        = "other"
	KeyTotal        = "total"
	KeyYes          = "yes"
	KeyNo           = "no"
)

// The filter parameters a value opens in All work (FR-014).
const (
	FilterAttr           = "attr"
	FilterActivityType   = "activityType"
	FilterActivityStatus = "activityStatus"
)

type Report struct {
	Show            Level          `json:"show"`
	Currency        string         `json:"currency"`
	Cards           Cards          `json:"cards"`
	BuiltIn         []Widget       `json:"builtIn"`
	Attributes      []AttributeRow `json:"attributes"`
	AttributesSetUp bool           `json:"attributesSetUp"`
}

type Cards struct {
	Records            int         `json:"records"`
	PlannedCost        money.Money `json:"plannedCost"`
	AllFilledIn        int         `json:"allFilledIn"`
	AllFilledInPercent *int64      `json:"allFilledInPercent"`
}

type AttributeRow struct {
	AttributeID string   `json:"attributeId"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Widgets     []Widget `json:"widgets"`
}

// Widget is one chart with its table (contract PlanDetailsWidget); Values hold every value, Other only feeds the chart.
type Widget struct {
	Key             string  `json:"key"`
	Title           string  `json:"title"`
	Measure         Measure `json:"measure"`
	Values          []Value `json:"values"`
	Other           *Value  `json:"other"`
	Total           Value   `json:"total"`
	FilledInPercent *int64  `json:"filledInPercent"`
	CountsInEach    bool    `json:"countsInEach"`
	FilterParam     *string `json:"filterParam"`
}

type Value struct {
	Key         string       `json:"key"`
	Label       string       `json:"label"`
	Inactive    bool         `json:"inactive"`
	Count       int          `json:"count"`
	PlannedCost *money.Money `json:"plannedCost"`
	Linkable    bool         `json:"linkable"`
}
