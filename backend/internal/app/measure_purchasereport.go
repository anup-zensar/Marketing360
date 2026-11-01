package app

import (
	"context"
	"slices"

	"mapp/internal/measure"
	"mapp/internal/measure/purchasereport"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/money"
	"mapp/internal/purchaserequest"
	"mapp/internal/stores"
)

func registerPurchaseRequests(w *wiring, r *measure.Registry) error {
	service := purchasereport.NewService(w.set.Tx, purchasereport.Deps{
		Scope: w.reports,
		Purchases: reportPurchases{repo: w.set.PurchaseRequests, workflows: w.set.Approvals, rates: w.settings.Rates,
			invoices: prInvoices{repo: w.set.Financials, rates: w.settings.Rates}},
	})
	mount(w.signedIn, purchasereport.NewHandler(service))
	return r.Register(purchasereport.NewReport(service))
}

// reportPurchases reads the purchase requests on the line items in view as the Purchase requests page does: the
// same scoped store read, the invoices of 015 on their PO lines and today's rates, every amount in USD.
type reportPurchases struct {
	repo      stores.PurchaseRequestStore
	workflows stores.ApprovalStore
	invoices  prInvoices
	rates     func(ctx context.Context) (money.Rates, error)
}

func (p reportPurchases) OnLineItems(ctx context.Context, lineItemIDs []string) (purchasereport.Requests, error) {
	config, err := p.repo.Configuration(ctx)
	if err != nil || !config.PurchaseSupport {
		return purchasereport.Requests{Items: []purchasereport.Request{}}, err
	}
	scope := purchaserequest.ViewScope{Plans: authz.ScopeOf(ctx), List: authz.PermissionScopeOf(ctx, authz.PurchaseRequestsList)}
	found, err := p.repo.OnLineItems(ctx, scope, lineItemIDs)
	if err != nil {
		return purchasereport.Requests{}, err
	}
	rates, err := p.rates(ctx)
	if err != nil {
		return purchasereport.Requests{}, err
	}
	invoiced, err := p.invoices.InvoicedOn(ctx, poLinesOf(found))
	if err != nil {
		return purchasereport.Requests{}, err
	}
	names, err := p.workflowNames(ctx, found)
	if err != nil {
		return purchasereport.Requests{}, err
	}
	items := make([]purchasereport.Request, 0, len(found))
	for _, pr := range found {
		items = append(items, requestOf(pr, names[pr.WorkflowID], rates, invoiced))
	}
	return purchasereport.Requests{Items: items, SwitchedOn: true}, nil
}

func (p reportPurchases) workflowNames(ctx context.Context, found []purchaserequest.PurchaseRequest) (map[string]string, error) {
	names := map[string]string{}
	for _, pr := range found {
		if _, known := names[pr.WorkflowID]; known || pr.WorkflowID == "" {
			continue
		}
		w, ok, err := p.workflows.GetWorkflow(ctx, pr.WorkflowID)
		if err != nil {
			return nil, err
		}
		if ok {
			names[pr.WorkflowID] = w.Name
		}
	}
	return names, nil
}

func poLinesOf(found []purchaserequest.PurchaseRequest) []purchaserequest.POLineRef {
	var refs []purchaserequest.POLineRef
	for _, pr := range found {
		for _, l := range pr.Lines {
			if l.PO.Line != "" {
				refs = append(refs, purchaserequest.POLineRef{PONumber: l.PO.Number, Line: l.PO.Line})
			}
		}
	}
	return refs
}

func requestOf(pr purchaserequest.PurchaseRequest, workflow string, rates money.Rates,
	invoiced map[purchaserequest.POLineRef]money.Amount) purchasereport.Request {
	r := purchasereport.Request{ID: pr.ID, Name: pr.Name, Vendor: pr.Vendor, Status: purchasereport.Status(pr.Status),
		PlanID: pr.PlanID, OwnerID: pr.OwnerID, Workflow: workflow, GSAPNumber: pr.GSAP.PRNumber, CreatedAt: pr.CreatedAt,
		WaitingForGSAP: waitingForGSAP(pr), Lines: make([]purchasereport.Line, 0, len(pr.Lines))}
	for _, l := range pr.Lines {
		line := purchasereport.Line{ID: l.ID, LineItemID: l.LineItemID, Amount: lineUSD(rates, l), DeliveryDate: l.DeliveryDate,
			SentOn: l.SentOn, PONumber: l.PO.Number, POLine: l.PO.Line, POLineStatus: l.PO.Status}
		if l.PO.Line != "" {
			amount := invoiced[purchaserequest.POLineRef{PONumber: l.PO.Number, Line: l.PO.Line}]
			line.Invoiced = &amount
		}
		r.Lines = append(r.Lines, line)
	}
	return r
}

// lineUSD converts a line to USD at today's rates; a currency without a rate counts at 1, as 032 counts it.
func lineUSD(rates money.Rates, l purchaserequest.Line) money.Amount {
	converted, err := rates.Convert(money.Money{Amount: l.Amount, Currency: l.Currency}, money.BaseCurrency)
	if err != nil {
		return l.Amount
	}
	return converted.Amount
}

// waitingForGSAP is 036's rule: an approved request sent to GSAP with a line GSAP has not answered yet (032).
func waitingForGSAP(pr purchaserequest.PurchaseRequest) bool {
	if pr.Status != purchaserequest.StatusApproved || !slices.ContainsFunc(pr.Lines, func(l purchaserequest.Line) bool { return l.SentOn != nil }) {
		return false
	}
	return slices.ContainsFunc(pr.Lines, func(l purchaserequest.Line) bool {
		return l.PO.Line == "" || l.PO.Value == nil || l.PO.Value.Currency != l.Currency || l.PO.Value.Amount.Cmp(l.Amount) != 0
	})
}
