package purchasereport

import (
	"context"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

// Report is the Purchase requests report as the hub's registry holds it.
type Report struct {
	service *Service
}

func NewReport(s *Service) *Report {
	return &Report{service: s}
}

func (r *Report) Key() string      { return measure.KeyPurchaseRequests }
func (r *Report) Available() bool  { return true }
func (r *Report) Params() []string { return Params() }

func (r *Report) Sheets(ctx context.Context, a actor.Actor, f measure.Filter) ([]measure.WidgetSheet, error) {
	q, err := ParseQuery(f.Own)
	if err != nil {
		return nil, err
	}
	d, err := r.service.snapshot(ctx, a, f)
	if err != nil {
		return nil, err
	}
	view := d.view(q)
	return []measure.WidgetSheet{
		{Key: "cards", Sheet: cardsSheet(view.Cards, d.shown.code)},
		{Key: "status", Sheet: statusSheet(view.StatusMix, d.shown.code)},
		{Key: "byVendor", Sheet: chartSheet("Commitment by vendor", "Vendor", view.ByVendor)},
		{Key: "byOwner", Sheet: chartSheet("Commitment by owner", "Owner", view.ByOwner)},
		{Key: "grid", Sheet: d.gridSheet(filtered(d.requests, q), q)},
	}, nil
}
