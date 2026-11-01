package app

import (
	"context"
	"slices"

	"mapp/internal/measure"
	"mapp/internal/measure/approvalreport"
	"mapp/internal/platform/authz"
	"mapp/internal/purchaserequest"
	"mapp/internal/stores"
)

func registerApprovals(w *wiring, r *measure.Registry) error {
	service := approvalreport.NewService(w.set.Tx, approvalreport.Deps{
		Scope:     w.reports,
		Rounds:    w.set.Approvals,
		Purchases: approvalPurchases{service: w.prs, repo: w.set.PurchaseRequests},
	})
	mount(w.signedIn, approvalreport.NewHandler(service))
	return r.Register(approvalreport.NewReport(service))
}

// approvalPurchases gives the purchase requests on the line items in view: the amount as feature 032 counts it
// from its service, and the name, owner and purchase orders from one scoped read of the store.
type approvalPurchases struct {
	service *purchaserequest.Service
	repo    stores.PurchaseRequestStore
}

func (p approvalPurchases) OnLineItems(ctx context.Context, lineItemIDs []string) ([]approvalreport.PurchaseRequest, error) {
	summaries, err := p.service.Summaries(ctx, lineItemIDs)
	if err != nil {
		return nil, err
	}
	scope := purchaserequest.ViewScope{Plans: authz.ScopeOf(ctx), List: authz.PermissionScopeOf(ctx, authz.PurchaseRequestsList)}
	found, err := p.repo.OnLineItems(ctx, scope, lineItemIDs)
	if err != nil {
		return nil, err
	}
	list := make([]approvalreport.PurchaseRequest, 0, len(found))
	for _, pr := range found {
		request := approvalreport.PurchaseRequest{ID: pr.ID, Name: pr.Name, PlanID: pr.PlanID, OwnerID: pr.OwnerID,
			PurchaseOrders: orderNumbers(pr.Lines)}
		if i := slices.IndexFunc(summaries.Items, func(s purchaserequest.Summary) bool { return s.ID == pr.ID }); i >= 0 {
			request.Amount = summaries.Items[i].AmountUSD
		}
		list = append(list, request)
	}
	return list, nil
}

func orderNumbers(lines []purchaserequest.Line) []string {
	numbers := []string{}
	for _, l := range lines {
		if l.PO.Number != "" && !slices.Contains(numbers, l.PO.Number) {
			numbers = append(numbers, l.PO.Number)
		}
	}
	return numbers
}
