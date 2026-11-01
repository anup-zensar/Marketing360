// Package projection works out, month by month, the volumes and revenue a goal is expected to reach
// through its plan's model (feature 047). It is pure: inputs in, series out.
package projection

import (
	"time"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

const (
	StageINQ     = "inq"
	StageMQL     = "mql"
	StageDeals   = "deals"
	StageRevenue = "revenue"
)

const Months = 12

type Model struct {
	Name                    string
	InqToMqlRate            money.Percent
	MqlToDealRate           money.Percent
	AverageDealSize         money.Money
	InqToMqlVelocityMonths  int
	MqlToDealVelocityMonths int
	ParametersChangedAt     time.Time
}

type Goals struct {
	INQ                  *int64
	MQL                  *int64
	Revenue              *money.Money
	SetLaterStagesMyself bool
}

// Inputs is everything a projection needs; Today comes from the caller's clock, never from here (047-C9).
type Inputs struct {
	Goals     Goals
	Model     Model
	PlanYear  int
	StartDate *date.Date
	EndDate   *date.Date
	Ongoing   bool
	Today     date.Date
}

type MonthFigure struct {
	Month string `json:"month"`
	Value int64  `json:"value"`
}

type MoneyFigure struct {
	Month string      `json:"month"`
	Value money.Money `json:"value"`
}

type Series struct {
	Stage  string        `json:"stage"`
	Months []MonthFigure `json:"months"`
	Total  int64         `json:"total"`
}

type Stamp struct {
	ModelName           string    `json:"modelName"`
	ParametersChangedAt time.Time `json:"parametersChangedAt"`
}

// Projection is the answer; Present is false when there is no goal, so the figures show a reason, not a zero.
type Projection struct {
	Present bool
	Stages  []Series
	Revenue []MoneyFigure
	Total   money.Money
	Stamp   Stamp
}

func monthKey(year int, month time.Month) string {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
}

func (p Projection) Find(stage string) (Series, bool) {
	for _, s := range p.Stages {
		if s.Stage == stage {
			return s, true
		}
	}
	return Series{}, false
}
