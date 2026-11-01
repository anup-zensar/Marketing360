package gsaplink

import (
	"slices"
	"strings"
	"time"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/money"
)

// companyCodeLabel names the activity-level Shell attribute the SWBS request sends to GSAP (030).
const companyCodeLabel = "company code"

// year indexes the frame's scope for the tabs: the nodes by ID and the matched line items in the year's order.
type year struct {
	scope  measure.Scope
	byID   map[string]allwork.Node
	inView map[string]bool // the matched line items
	now    time.Time
}

func newYear(sc measure.Scope, now time.Time) year {
	y := year{scope: sc, byID: make(map[string]allwork.Node, len(sc.Nodes)),
		inView: make(map[string]bool, len(sc.Matched.LineItems)), now: now}
	for _, n := range sc.Nodes {
		y.byID[n.ID] = n
	}
	for _, id := range sc.Matched.LineItems {
		y.inView[id] = true
	}
	return y
}

func (y year) lineItems() []allwork.Node {
	list := make([]allwork.Node, 0, len(y.scope.Matched.LineItems))
	for _, id := range y.scope.Matched.LineItems {
		if n, found := y.byID[id]; found {
			list = append(list, n)
		}
	}
	return list
}

// financeLineItems are the matched line items of plans whose financial data the person may see (D-237).
func (y year) financeLineItems() []string {
	ids := []string{}
	for _, n := range y.lineItems() {
		if y.seesFinance(n.PlanID) {
			ids = append(ids, n.ID)
		}
	}
	return ids
}

func (y year) seesFinance(planID string) bool {
	c := y.scope.Choices
	return c.SeesFinance && (c.FinancePlans == nil || slices.Contains(c.FinancePlans, planID))
}

func (y year) coverage() *measure.FinanceCoverage {
	coverage := measure.FinanceCoverage{Of: len(y.scope.Matched.Plans)}
	for _, id := range y.scope.Matched.Plans {
		if y.seesFinance(id) {
			coverage.Plans++
		}
	}
	return &coverage
}

func (y year) path(n allwork.Node) Path {
	var p Path
	for found := true; found; n, found = y.byID[n.ParentID] {
		switch n.Level {
		case allwork.LevelPlan:
			p.PlanID, p.Plan = n.ID, n.Name
		case allwork.LevelCampaign:
			p.Campaign = n.Name
		case allwork.LevelInitiative:
			p.Initiative = n.Name
		case allwork.LevelActivity:
			p.ActivityID, p.Activity = n.ID, n.Name
		}
	}
	return p
}

func (y year) inCurrency(usd money.Amount) money.Money {
	return money.Money{Amount: money.FromUSD(usd, y.scope.Currency.Rate), Currency: y.scope.Currency.Code}
}

// companyCode reads the option chosen for the activity's Company code, as 030 sends it; "" when none.
func (y year) companyCode(activity allwork.Node) string {
	i := slices.IndexFunc(y.scope.Attributes, func(a allwork.Attribute) bool {
		return a.Level == allwork.LevelActivity && strings.EqualFold(strings.TrimSpace(a.Label), companyCodeLabel)
	})
	if i < 0 || len(activity.Attrs[y.scope.Attributes[i].ID].OptionIDs) == 0 {
		return ""
	}
	chosen := activity.Attrs[y.scope.Attributes[i].ID].OptionIDs[0]
	for _, o := range y.scope.Attributes[i].Options {
		if o.ID == chosen {
			return strings.TrimSpace(o.Label)
		}
	}
	return ""
}

func ownerName(n allwork.Node) string {
	if n.Owner == nil {
		return ""
	}
	return n.Owner.Name
}

// usdOf brings an amount to USD at its currency's rate; one without a rate adds nothing, as in 015.
func usdOf(m money.Money, rates money.Rates) money.Amount {
	usd, err := rates.ToUSD(m)
	if err != nil {
		return money.Amount{}
	}
	return usd
}
