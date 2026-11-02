// Package planbudget is the Plan budget report (feature 040): spend pacing by period split at today, the summary
// cards and the ranked list with drill, over 039's records in view.
package planbudget

import "mapp/internal/platform/money"

const (
	StatePast       = "actual"
	StateInProgress = "inProgress"
	StateFuture     = "planned"
)

type Figures struct {
	Budget        *money.Money `json:"budget"`
	Planned       *money.Money `json:"planned"`
	Committed     *money.Money `json:"committed"`
	Actual        *money.Money `json:"actual"`
	LeftToPlan    *money.Money `json:"leftToPlan"`
	Unspent       *money.Money `json:"unspent"`
	OverAllocated bool         `json:"overAllocated"`
}

type Cards struct {
	Figures
	BudgetUsed      *int         `json:"budgetUsed"`
	BudgetSpent     *int         `json:"budgetSpent"`
	CommittedShare  *int         `json:"committedShare"`
	NotYetCommitted *money.Money `json:"notYetCommitted"`
}

type Period struct {
	Key           string       `json:"key"`
	State         string       `json:"state"`
	Budget        money.Money  `json:"budget"`
	Planned       money.Money  `json:"planned"`
	Actual        *money.Money `json:"actual"`
	BudgetToDate  money.Money  `json:"budgetToDate"`
	PlannedToDate money.Money  `json:"plannedToDate"`
	ActualToDate  *money.Money `json:"actualToDate"`
}

type Notes struct {
	NotPhased        *money.Money `json:"notPhased"`
	NoDates          *money.Money `json:"noDates"`
	OtherYears       *money.Money `json:"otherYears"`
	ActualOtherYears *money.Money `json:"actualOtherYears"`
}

type Report struct {
	Cards   Cards    `json:"cards"`
	Period  string   `json:"period"`
	Periods []Period `json:"periods"`
	Notes   Notes    `json:"notes"`
}

type Entry struct {
	ID       string  `json:"id"`
	Level    string  `json:"level"`
	Name     string  `json:"name"`
	PlanID   string  `json:"planId"`
	Figures  Figures `json:"figures"`
	Children []Entry `json:"children,omitempty"`
}

type Group struct {
	Key     string  `json:"key"`
	Label   string  `json:"label"`
	Totals  Figures `json:"totals"`
	Entries []Entry `json:"entries"`
}

type TrailStep struct {
	ID    string `json:"id"`
	Level string `json:"level"`
	Name  string `json:"name"`
}

type Ranked struct {
	Level     string      `json:"level"`
	Sort      string      `json:"sort"`
	Order     string      `json:"order"`
	Group     string      `json:"group"`
	Page      int         `json:"page"`
	Size      int         `json:"size"`
	Total     int         `json:"total"`
	Trail     []TrailStep `json:"trail"`
	Groups    []Group     `json:"groups"`
	DrillLost bool        `json:"drillLost"`
}
