// Package attribution credits revenue to a journey's activities and works out ROI (feature 050) for 048 and 049.
package attribution

import (
	"time"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/platform/record"
)

type Model string

const (
	FirstTouch  Model = "firstTouch"
	LastTouch   Model = "lastTouch"
	Equal       Model = "equallyDistributed"
	LinearDecay Model = "linearTimeDecay"
	UShaped     Model = "uShaped"
)

type ModelChoice struct {
	Key         Model  `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func Models() []ModelChoice {
	return []ModelChoice{
		{FirstTouch, "First touch", "All the revenue goes to the first activity the customer met."},
		{LastTouch, "Last touch", "All the revenue goes to the last activity before the conversion."},
		{Equal, "Equally distributed", "Every activity of the journey gets the same share."},
		{LinearDecay, "Linear time decay", "Later activities get more: the share grows with each touch."},
		{UShaped, "U-shaped", "40% to the first and the last activity each, 20% shared by the ones between."},
	}
}

func (m Model) Name() string {
	for _, c := range Models() {
		if c.Key == m {
			return c.Name
		}
	}
	return string(m)
}

type Settings struct {
	Model Model `json:"model"`
	record.Meta
}

func DefaultSettings() Settings {
	return Settings{Model: Equal, Meta: record.Meta{Version: 1, CreatedBy: actor.System, UpdatedBy: actor.System}}
}

type SettingsView struct {
	Model     Model              `json:"model"`
	Models    []ModelChoice      `json:"models"`
	Version   int                `json:"version"`
	UpdatedAt *time.Time         `json:"updatedAt"`
	UpdatedBy *actor.Actor       `json:"updatedBy"`
	CanChange authz.Availability `json:"canChange"`
}

type SettingsInput struct {
	Version int   `json:"version"`
	Model   Model `json:"model"`
}

// Touch is one activity of a journey; ActivityID is "" for a touch that matched no activity (client 020 FR-057).
type Touch struct {
	ActivityID string
	Date       date.Date
}

// Conversion is converted revenue with its journey; a conversion without a journey reference has Journey false.
type Conversion struct {
	ActivityID string
	Date       date.Date
	Revenue    money.Amount
	Journey    bool
	Touches    []Touch
}

// Credit is an activity's share of a conversion; ActivityID "" is Unattributed.
type Credit struct {
	ActivityID string
	Amount     money.Amount
}
