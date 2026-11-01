package widgetdata

import (
	"mapp/internal/platform/authz"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

const (
	SourceBuiltIn = "builtIn"
	SourceDataset = "dataset"
)

type SourceFilters struct {
	Date       string
	Plan, Team bool
}

type Source struct {
	Key, Label, Description string
	Kind                    string
	Fields                  []Field
	Filters                 SourceFilters
	Owner                   *authz.PersonRef
	need                    authz.Permission
}

type Request struct {
	Type       string             `json:"type"`
	Query      Query              `json:"query"`
	Filters    *Filters           `json:"filters"`
	Comparison *ComparisonRequest `json:"comparison"`
}

type Query struct {
	Source    string    `json:"source"`
	Dimension *GroupBy  `json:"dimension"`
	Series    *GroupBy  `json:"series"`
	Measures  []Measure `json:"measures"`
	Filters   []Filter  `json:"filters"`
	Sort      Sort      `json:"sort"`
	Limit     int       `json:"limit"`
	GroupRest bool      `json:"groupRest"`
}

type GroupBy struct {
	Field     string  `json:"field"`
	DateGroup *string `json:"dateGroup"`
}

type Measure struct {
	Aggregation string  `json:"aggregation"`
	Field       *string `json:"field"`
	Label       string  `json:"label"`
}

type Filter struct {
	Field    string   `json:"field"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

type Sort struct {
	By           string `json:"by"`
	MeasureIndex int    `json:"measureIndex"`
	Direction    string `json:"direction"`
}

type Filters struct {
	DateFrom *date.Date `json:"dateFrom"`
	DateTo   *date.Date `json:"dateTo"`
	PlanIDs  []string   `json:"planIds"`
	TeamIDs  []string   `json:"teamIds"`
}

type ComparisonRequest struct {
	Kind   string  `json:"kind"`
	Target *string `json:"target"`
}

type Data struct {
	Source         string         `json:"source"`
	Columns        []Column       `json:"columns"`
	Rows           []DataRow      `json:"rows"`
	Totals         []Cell         `json:"totals"`
	GroupCount     int            `json:"groupCount"`
	Notes          []string       `json:"notes"`
	FiltersApplied FiltersApplied `json:"filtersApplied"`
	Comparison     *Comparison    `json:"comparison"`
}

type Column struct {
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	Kind      FieldKind `json:"kind"`
	DateGroup *string   `json:"dateGroup"`
}

type DataRow struct {
	Dimension *Key   `json:"dimension"`
	Series    *Key   `json:"series"`
	Values    []Cell `json:"values"`
}

const (
	KeyValue = "value"
	KeyEmpty = "empty"
	KeyOther = "other"
)

type Key struct {
	Value *string `json:"value"`
	Label string  `json:"label"`
	Kind  string  `json:"kind"`
}

type Cell struct {
	Number *Number       `json:"number"`
	Money  *money.Money  `json:"money"`
	Date   *date.Date    `json:"date"`
	Share  *SharePercent `json:"share"`
}

type FiltersApplied struct {
	Date bool `json:"date"`
	Plan bool `json:"plan"`
	Team bool `json:"team"`
}

type Comparison struct {
	Kind   string        `json:"kind"`
	Value  Cell          `json:"value"`
	Change *SharePercent `json:"change"`
	Period *Period       `json:"period"`
}

type Period struct {
	From date.Date `json:"from"`
	To   date.Date `json:"to"`
}
