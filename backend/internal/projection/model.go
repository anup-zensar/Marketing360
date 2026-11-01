// Package projection turns an activity's goal into the stage volumes and revenue it is expected to reach,
// month by month, through its plan's model (feature 047, client 020 FR-048–FR-053). It is pure: no HTTP,
// no SQL, no clock of its own — inputs in, series out.
package projection

import (
	"time"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// The stages a projection carries, in funnel order. Deals is the volume behind revenue, which the
// waterfall already counts (domain model §4.5).
const (
	StageINQ     = "inq"
	StageMQL     = "mql"
	StageDeals   = "deals"
	StageRevenue = "revenue"
)

// Months is the number of months a projection covers: the plan year, nothing beyond it (FR-008).
const Months = 12

// Model is the plan's rulebook as the projection reads it (047-C1, 047-C2).
type Model struct {
	Name                    string
	InqToMqlRate            money.Percent
	MqlToDealRate           money.Percent
	AverageDealSize         money.Money
	InqToMqlVelocityMonths  int
	MqlToDealVelocityMonths int
	ParametersChangedAt     time.Time
}

// Goals are the activity's goals as its own screens show them: INQ always, MQL and Revenue typed only when
// the activity sets its later stages itself (047-C6).
type Goals struct {
	INQ                  *int64
	MQL                  *int64
	Revenue              *money.Money
	SetLaterStagesMyself bool
}

// Inputs is everything a projection needs. Today comes from the caller's clock, never from the engine
// (047-C9).
type Inputs struct {
	Goals     Goals
	Model     Model
	PlanYear  int
	StartDate *date.Date
	EndDate   *date.Date
	Ongoing   bool
	Today     date.Date
}

// MonthFigure is one month of a series: "2026-01" and the whole number projected in it.
type MonthFigure struct {
	Month string `json:"month"`
	Value int64  `json:"value"`
}

// MoneyFigure is one month of the revenue series.
type MoneyFigure struct {
	Month string      `json:"month"`
	Value money.Money `json:"value"`
}

// Series is one stage's projection across the plan year.
type Series struct {
	Stage  string        `json:"stage"`
	Months []MonthFigure `json:"months"`
	Total  int64         `json:"total"`
}

// Stamp lets a reviewer reproduce a figure: which model produced it and when its parameters last changed
// (FR-012, client 020 FR-052).
type Stamp struct {
	ModelName           string    `json:"modelName"`
	ParametersChangedAt time.Time `json:"parametersChangedAt"`
}

// Projection is the whole answer. Present is false when the activity has no goal to project from: the
// figures then show nothing with a reason, never a zero (FR-013).
type Projection struct {
	Present bool
	Stages  []Series
	Revenue []MoneyFigure
	Total   money.Money
	Stamp   Stamp
}

// monthKey is the month label a fact's month carries too ("2026-01"), so the two series line up.
func monthKey(year int, month time.Month) string {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
}

// Find returns the series of the stage, and false when the projection does not carry it.
func (p Projection) Find(stage string) (Series, bool) {
	for _, s := range p.Stages {
		if s.Stage == stage {
			return s, true
		}
	}
	return Series{}, false
}
