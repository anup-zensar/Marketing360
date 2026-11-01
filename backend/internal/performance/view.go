package performance

import (
	"time"

	"mapp/internal/platform/record"
)

// SourceResponse is the contract's shape: freshness is derived at the moment of answering (FR-005).
type SourceResponse struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	OwnerID       string        `json:"ownerId"`
	Category      Category      `json:"category"`
	Status        Status        `json:"status"`
	DeliveryRoute DeliveryRoute `json:"deliveryRoute"`
	Cadence       Cadence       `json:"cadence"`
	IntegrationID string        `json:"integrationId"`
	Freshness     Freshness     `json:"freshness"`
	LastLoadedAt  *time.Time    `json:"lastLoadedAt"`
	DataThrough   *string       `json:"dataThrough"`
	FactCount     int           `json:"factCount"`
	record.Meta
}

func viewOf(v SourceView, now time.Time) SourceResponse {
	out := SourceResponse{
		ID: v.ID, Name: v.Name, Description: v.Description, OwnerID: v.OwnerID,
		Category: v.Category, Status: v.Status, DeliveryRoute: v.DeliveryRoute, Cadence: v.Cadence,
		IntegrationID: v.IntegrationID, Freshness: v.FreshnessAt(now), FactCount: v.FactCount,
		Meta: v.Meta,
	}
	if !v.LastLoadedAt.IsZero() {
		at := v.LastLoadedAt
		out.LastLoadedAt = &at
	}
	if !v.DataThrough.IsZero() {
		through := v.DataThrough.Format(time.DateOnly)
		out.DataThrough = &through
	}
	return out
}

func viewsOf(sources []SourceView, now time.Time) []SourceResponse {
	out := make([]SourceResponse, 0, len(sources))
	for _, v := range sources {
		out = append(out, viewOf(v, now))
	}
	return out
}

type SourceList struct {
	Items []SourceResponse `json:"items"`
}

type MappingList struct {
	Items []Mapping `json:"items"`
}
