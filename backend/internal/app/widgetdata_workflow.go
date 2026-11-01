package app

import (
	"context"
	"maps"
	"math/big"
	"slices"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/approval"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/purchaserequest"
	"mapp/internal/settings"
	"mapp/internal/stores"
	"mapp/internal/widgetdata"
)

const tenThousandthsPerDay = 86400

var (
	recordTypeWords = map[approval.SubjectKind]string{ //nolint:gochecknoglobals // the source's record types
		approval.KindActivity: "Activity", approval.KindPurchaseRequest: "Purchase request",
		approval.KindModifiedPurchaseRequest: "Modified purchase request",
	}
	prStatusWords = map[purchaserequest.Status]string{ //nolint:gochecknoglobals // the purchase request statuses
		purchaserequest.StatusCreated: wordCreated, purchaserequest.StatusSubmitted: wordSubmitted,
		purchaserequest.StatusInReview: "In review", purchaserequest.StatusDeclined: wordDeclined,
		purchaserequest.StatusApproved: wordApproved, purchaserequest.StatusClosed: "Closed",
	}
)

// workflowReader answers Approvals or Purchase requests: the rounds and requests of the records in scope only.
type workflowReader struct {
	planning  widgetPlanning
	approvals stores.ApprovalStore
	purchases stores.PurchaseRequestStore
	settings  *settings.Service
	rounds    bool
}

func (r workflowReader) Rows(ctx context.Context, a actor.Actor, f widgetdata.Prefilter) (widgetdata.Rows, error) {
	d, err := r.planning.read(ctx, a, f, false)
	if err != nil {
		return widgetdata.Rows{}, err
	}
	scope := purchaserequest.ViewScope{Plans: authz.ScopeOf(ctx), List: authz.PermissionScopeOf(ctx, authz.PurchaseRequestsList)}
	prs, err := r.purchases.OnLineItems(ctx, scope, d.idsOf(allwork.LevelLineItem))
	if err != nil {
		return widgetdata.Rows{}, err
	}
	if r.rounds {
		return r.roundRows(ctx, d, prs, f)
	}
	return r.purchaseRows(ctx, d, prs, f)
}

func (r workflowReader) roundRows(ctx context.Context, d *planningData, prs []purchaserequest.PurchaseRequest,
	f widgetdata.Prefilter) (widgetdata.Rows, error) {
	names, plans := map[string]string{}, map[string]string{}
	for _, n := range d.nodes {
		if n.Level == allwork.LevelActivity {
			names[n.ID], plans[n.ID] = n.Name, n.PlanID
		}
	}
	for _, pr := range prs {
		names[pr.ID], plans[pr.ID] = pr.Name, pr.PlanID
	}
	rounds, err := r.approvals.RoundsOfSubjects(ctx, slices.Sorted(maps.Keys(names)))
	if err != nil {
		return widgetdata.Rows{}, err
	}
	fields := widgetdata.BuiltInFields("approvals")
	out := widgetdata.Rows{Items: []widgetdata.Row{}, Notes: d.notes}
	for _, round := range rounds {
		submitted := date.Today(round.SubmittedAt)
		if within(&submitted, f) {
			v := roundValues(round)
			v[fieldRecord], v[fieldPlan] = text(names[round.SubjectID]), d.nameOf(plans[round.SubjectID])
			v[fieldTeam] = d.teamsOfPlan(plans[round.SubjectID])
			out.Items = append(out.Items, widgetdata.RowOf(fields, v))
		}
	}
	return out, nil
}

func roundValues(r approval.Round) values {
	kind := r.SubjectKind
	if kind == "" {
		kind = approval.KindActivity
	}
	submitted := date.Today(r.SubmittedAt)
	v := values{fieldRecordType: text(recordTypeWords[kind]), fieldStatus: text(roundStatus(r.Outcome)),
		fieldAutomatic: widgetdata.YesNo(r.AutoApproved), fieldWorkflow: text(r.WorkflowName), fieldSubmittedOn: widgetdata.DateValue(&submitted)}
	decided := r.Outcome == approval.OutcomeApproved || r.Outcome == approval.OutcomeDeclined
	if decided && r.EndedAt != nil {
		on := date.Today(*r.EndedAt)
		v[fieldDecidedOn] = widgetdata.DateValue(&on)
		if !r.AutoApproved {
			v[fieldDaysToDecide] = daysBetween(r.SubmittedAt, *r.EndedAt)
		}
	}
	return v
}

func roundStatus(o approval.Outcome) string {
	switch o {
	case approval.OutcomeOpen:
		return "Waiting"
	case approval.OutcomeApproved:
		return wordApproved
	case approval.OutcomeDeclined:
		return wordDeclined
	case approval.OutcomeRecalled, approval.OutcomeRestarted, approval.OutcomeCancelled, approval.OutcomeSuperseded:
	}
	return "Ended"
}

// daysBetween is the time from from to to in days, exact to four decimals, half up.
func daysBetween(from, to time.Time) widgetdata.Value {
	seconds := big.NewInt(int64(to.Sub(from) / time.Second))
	num := new(big.Int).Mul(seconds, big.NewInt(10000))
	q, rem := new(big.Int).QuoRem(num, big.NewInt(tenThousandthsPerDay), new(big.Int))
	if rem.Sign() > 0 && 2*rem.Int64() >= tenThousandthsPerDay {
		q.Add(q, big.NewInt(1))
	}
	n := widgetdata.NumberOfUnits(q.Int64())
	return widgetdata.NumberValue(&n)
}

func (r workflowReader) purchaseRows(ctx context.Context, d *planningData, prs []purchaserequest.PurchaseRequest,
	f widgetdata.Prefilter) (widgetdata.Rows, error) {
	rates, err := r.settings.Rates(ctx)
	if err != nil {
		return widgetdata.Rows{}, err
	}
	owners, err := r.planning.personNames(ctx)
	if err != nil {
		return widgetdata.Rows{}, err
	}
	fields := widgetdata.BuiltInFields("purchaseRequests")
	out := widgetdata.Rows{Items: []widgetdata.Row{}, Notes: d.notes}
	for _, pr := range prs {
		created := date.Today(pr.CreatedAt)
		if !within(&created, f) {
			continue
		}
		v := d.purchaseValues(pr, rates)
		v[fieldOwner], v[fieldCreatedOn] = text(owners[pr.OwnerID]), widgetdata.DateValue(&created)
		out.Items = append(out.Items, widgetdata.RowOf(fields, v))
	}
	return out, nil
}

func (d *planningData) purchaseValues(pr purchaserequest.PurchaseRequest, rates money.Rates) values {
	var total money.Amount
	sent := false
	for _, l := range pr.Lines {
		usd, err := rates.Convert(money.Money{Amount: l.Amount, Currency: l.Currency}, money.BaseCurrency)
		if err != nil {
			usd = money.USD(l.Amount)
		}
		total, sent = total.Add(usd.Amount), sent || l.SentOn != nil
	}
	v := values{fieldPurchaseRequest: text(pr.Name), fieldStatus: text(prStatusWords[pr.Status]), fieldVendor: text(pr.Vendor),
		fieldAmount: widgetdata.MoneyValue(&total), fieldSentToGsap: widgetdata.YesNo(sent), fieldPlan: d.nameOf(pr.PlanID),
		fieldTeam: d.teamsOfPlan(pr.PlanID)}
	if len(pr.Lines) > 0 {
		activity := d.byID[parentOf(d.byID[pr.Lines[0].LineItemID])]
		v[fieldActivity], v[fieldCampaign] = d.nameOf(parentOf(d.byID[pr.Lines[0].LineItemID])), d.parentName(d.byID[parentOf(activity)])
	}
	return v
}

func (d *planningData) teamsOfPlan(planID string) widgetdata.Value {
	if p := d.byID[planID]; p != nil {
		return widgetdata.TextsValue(d.teamNamesOf(p.TeamIDs))
	}
	return widgetdata.TextsValue(nil)
}

func (d *planningData) idsOf(level allwork.Level) []string {
	ids := []string{}
	for _, n := range d.nodes {
		if n.Level == level {
			ids = append(ids, n.ID)
		}
	}
	return ids
}

func (p widgetPlanning) personNames(ctx context.Context) (map[string]string, error) {
	everyone, err := p.people.List(ctx)
	names := make(map[string]string, len(everyone))
	for _, person := range everyone {
		names[person.ID] = person.Name
	}
	return names, err
}
