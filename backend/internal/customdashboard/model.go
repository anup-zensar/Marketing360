// Package customdashboard holds the dashboards people build from widgets, owned by one person and shared
// view-only with teams (feature 051).
package customdashboard

import (
	"time"

	"mapp/internal/people"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/date"
	"mapp/internal/platform/record"
)

// Dashboard is the stored record: SharedWith holds team ids, Widgets are in their saved order.
type Dashboard struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	OwnerID     string   `json:"ownerId"`
	SharedWith  []string `json:"sharedWith"`
	Filters     Filters  `json:"filters"`
	Widgets     []Widget `json:"widgets"`
	record.Meta
}

type Filters struct {
	DateFrom *date.Date `json:"dateFrom"`
	DateTo   *date.Date `json:"dateTo"`
	PlanIDs  []string   `json:"planIds"`
	TeamIDs  []string   `json:"teamIds"`
}

type Widget struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	Title    string  `json:"title"`
	Subtitle string  `json:"subtitle"`
	Layout   Layout  `json:"layout"`
	Query    *Query  `json:"query"`
	Display  Display `json:"display"`
	Note     string  `json:"note"`
}

type Layout struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
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

// Display holds only the options a widget sets: an option left out takes its default in the web app.
type Display struct {
	Palette       *string        `json:"palette,omitempty"`
	Legend        *Legend        `json:"legend,omitempty"`
	DataLabels    *bool          `json:"dataLabels,omitempty"`
	AxisTitles    *AxisTitles    `json:"axisTitles,omitempty"`
	Gridlines     *bool          `json:"gridlines,omitempty"`
	Smooth        *bool          `json:"smooth,omitempty"`
	Stack         *string        `json:"stack,omitempty"`
	Orientation   *string        `json:"orientation,omitempty"`
	NumberFormat  *string        `json:"numberFormat,omitempty"`
	Decimals      *int           `json:"decimals,omitempty"`
	Zoom          *bool          `json:"zoom,omitempty"`
	ReferenceLine *ReferenceLine `json:"referenceLine,omitempty"`
	MeasureStyles []MeasureStyle `json:"measureStyles,omitempty"`
	Gauge         *GaugeOptions  `json:"gauge,omitempty"`
	Kpi           *KpiOptions    `json:"kpi,omitempty"`
	Table         *TableOptions  `json:"table,omitempty"`
}

type Legend struct {
	Show     bool   `json:"show"`
	Position string `json:"position"`
}

type AxisTitles struct {
	X  string `json:"x"`
	Y  string `json:"y"`
	Y2 string `json:"y2"`
}

type ReferenceLine struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type MeasureStyle struct {
	As   string `json:"as"`
	Axis string `json:"axis"`
}

type GaugeOptions struct {
	Min string  `json:"min"`
	Max *string `json:"max"`
}

type KpiOptions struct {
	Comparison     string  `json:"comparison"`
	Target         *string `json:"target"`
	HigherIsBetter bool    `json:"higherIsBetter"`
}

type TableOptions struct {
	Columns []string `json:"columns"`
	Totals  bool     `json:"totals"`
}

// View is a dashboard as the API answers it (contract CustomDashboard).
type View struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Owner          people.PersonRef `json:"owner"`
	IsOwner        bool             `json:"isOwner"`
	SharedWith     []people.TeamRef `json:"sharedWith"`
	ShareableTeams []people.TeamRef `json:"shareableTeams"`
	Filters        Filters          `json:"filters"`
	Widgets        []Widget         `json:"widgets"`
	Access         Access           `json:"access"`
	record.Meta
}

type Access struct {
	Edit   authz.Availability `json:"edit"`
	Share  authz.Availability `json:"share"`
	Delete authz.Availability `json:"delete"`
	Copy   authz.Availability `json:"copy"`
}

type Summary struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Owner       people.PersonRef `json:"owner"`
	SharedWith  []people.TeamRef `json:"sharedWith"`
	WidgetCount int              `json:"widgetCount"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	Access      Access           `json:"access"`
}

type List struct {
	Mine         []Summary `json:"mine"`
	SharedWithMe []Summary `json:"sharedWithMe"`
}

type Ref struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Filters     Filters  `json:"filters"`
	Widgets     []Widget `json:"widgets"`
}

// UpdateInput changes only the fields sent; Widgets, when sent, replace the whole list.
type UpdateInput struct {
	Version     *int      `json:"version"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Filters     *Filters  `json:"filters"`
	Widgets     *[]Widget `json:"widgets"`
}

type CopyInput struct {
	Name *string `json:"name"`
}

type SharingInput struct {
	Version *int     `json:"version"`
	TeamIDs []string `json:"teamIds"`
}
