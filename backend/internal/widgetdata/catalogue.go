package widgetdata

import (
	"mapp/internal/platform/authz"
)

const (
	keyPlans            = "plans"
	keyCampaigns        = "campaigns"
	keyInitiatives      = "initiatives"
	keyActivities       = "activities"
	keyLineItems        = "lineItems"
	keyPlannedByMonth   = "plannedByMonth"
	keyTransactions     = "financialTransactions"
	keyApprovals        = "approvals"
	keyPurchaseRequests = "purchaseRequests"
	keyGlobalBudgets    = "globalBudgets"
)

func fText(key, label string) Field   { return Field{Key: key, Label: label, Kind: KindText} }
func fNumber(key, label string) Field { return Field{Key: key, Label: label, Kind: KindNumber} }
func fMoney(key, label string) Field  { return Field{Key: key, Label: label, Kind: KindMoney} }
func fDate(key, label string) Field   { return Field{Key: key, Label: label, Kind: KindDate} }

func teams() Field { return Field{Key: "team", Label: "Team", Kind: KindText, Multi: true} }

func fFinance(key, label string) Field {
	return Field{Key: key, Label: label, Kind: KindMoney, Finance: true}
}

func planned() []Field {
	return []Field{fMoney("budget", "Budget"), fMoney("plannedCost", "Planned cost"), fMoney("leftToPlan", "Left to plan")}
}

func spent() []Field {
	return []Field{fFinance("committed", "Committed"), fFinance("actual", "Actual")}
}

func joined(parts ...[]Field) []Field {
	var out []Field
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// about names a source: its key, label and description.
type about [3]string

func planning(a about, dateField string, fields []Field) Source {
	return Source{Key: a[0], Label: a[1], Description: a[2], Kind: SourceBuiltIn, Fields: fields,
		Filters: SourceFilters{Date: dateField, Plan: true, Team: true}, need: authz.PlansView}
}

// builtIns are data model §5 in its order; readers answer rows in the order of each source's fields.
func builtIns() []Source {
	return []Source{
		planning(about{keyPlans, "Plans", "Every plan you can see, with its budget and costs."}, "", joined([]Field{
			fText("plan", "Plan"), fNumber("year", "Year"), fText("status", "Status"), fText("planType", "Plan type"),
			fText("owner", "Owner"), teams(), fText("currency", "Currency"),
		}, planned(), spent())),
		planning(about{keyCampaigns, "Campaigns", "Campaigns of the plans you can see."}, "startDate", joined([]Field{
			fText("campaign", "Campaign"), fText("plan", "Plan"), fNumber("year", "Year"), teams(), fText("owner", "Owner"),
			fText("status", "Status"), fText("timing", "Timing"), fDate("startDate", "Start date"), fDate("endDate", "End date"),
		}, planned(), spent())),
		planning(about{keyInitiatives, "Initiatives", "Initiatives of the plans you can see."}, "startDate", joined([]Field{
			fText("initiative", "Initiative"), fText("campaign", "Campaign"), fText("plan", "Plan"), fNumber("year", "Year"),
			teams(), fText("owner", "Owner"), fText("status", "Status"), fDate("startDate", "Start date"), fDate("endDate", "End date"),
		}, planned())),
		planning(about{keyActivities, "Activities", "Activities of the plans you can see."}, "startDate", joined([]Field{
			fText("activity", "Activity"), fText("initiative", "Initiative"), fText("campaign", "Campaign"), fText("plan", "Plan"),
			fNumber("year", "Year"), teams(), fText("owner", "Owner"), fText("activityType", "Activity type"),
			fText("status", "Status"), fText("complete", "Complete"), fDate("startDate", "Start date"), fDate("endDate", "End date"),
			fMoney("plannedCost", "Planned cost"),
		}, spent())),
		planning(about{keyLineItems, "Line items", "Line items of the plans you can see."}, "", joined([]Field{
			fText("lineItem", "Line item"), fText("activity", "Activity"), fText("campaign", "Campaign"), fText("plan", "Plan"),
			fNumber("year", "Year"), teams(), fText("currency", "Currency"), fText("financeBudgetItem", "Finance budget item"),
			fText("swbsStatus", "SWBS status"), fText("complete", "Complete"), fMoney("plannedCost", "Planned cost"),
		}, spent())),
		planning(about{keyPlannedByMonth, "Planned cost by month", "Planned cost spread over the months, as Budget & cost shows it."},
			"month", []Field{
				fDate("month", "Month"), fText("plan", "Plan"), fText("campaign", "Campaign"), fText("initiative", "Initiative"),
				fText("activity", "Activity"), fText("activityType", "Activity type"), teams(), fMoney("plannedCost", "Planned cost"),
			}),
		transactionsSource(),
		planning(about{keyApprovals, "Approvals", "Approval rounds of activities and purchase requests."}, "submittedOn", []Field{
			fText("record", "Record"), fText("recordType", "Record type"), fText("status", "Status"), fText("automatic", "Automatic"),
			fText("workflow", "Workflow"), fDate("submittedOn", "Submitted on"), fDate("decidedOn", "Decided on"),
			fNumber("daysToDecide", "Days to decide"), fText("plan", "Plan"), teams(),
		}),
		planning(about{keyPurchaseRequests, "Purchase requests", "The purchase requests you can see."}, "createdOn", []Field{
			fText("purchaseRequest", "Purchase request"), fText("status", "Status"), fText("owner", "Owner"), fText("vendor", "Vendor"),
			fMoney("amount", "Amount"), fDate("createdOn", "Created on"), fText("sentToGsap", "Sent to GSAP"),
			fText("activity", "Activity"), fText("campaign", "Campaign"), fText("plan", "Plan"), teams(),
		}),
		globalBudgetsSource(),
	}
}

func transactionsSource() Source {
	s := planning(about{keyTransactions, "Financial transactions", "GSAP transactions of line items whose financial data you see."},
		"postingDate", []Field{
			fText("type", "Type"), fDate("postingDate", "Posting date"), fMoney("amount", "Amount"), fText("vendor", "Vendor"),
			fText("documentNumber", "Document number"), fText("lineItem", "Line item"), fText("activity", "Activity"),
			fText("campaign", "Campaign"), fText("plan", "Plan"), teams(),
		})
	s.need = authz.FinanceView
	return s
}

func globalBudgetsSource() Source {
	return Source{Key: keyGlobalBudgets, Label: "Global budgets", Description: "Global budgets, as the Global budgets page shows them.",
		Kind: SourceBuiltIn, need: authz.GlobalBudgetsView, Fields: []Field{
			fText("globalBudget", "Global budget"), fNumber("year", "Year"), fText("owner", "Owner"), fMoney("amount", "Amount"),
			fMoney("allocated", "Allocated"), fMoney("unallocated", "Unallocated"), fMoney("plannedCost", "Planned cost"),
		}}
}

// BuiltInFields are the fields of a built-in source in row order; nil for an unknown key.
func BuiltInFields(key string) []Field {
	if s, found := builtIn(key); found {
		return s.Fields
	}
	return nil
}

func builtIn(key string) (Source, bool) {
	for _, s := range builtIns() {
		if s.Key == key {
			return s, true
		}
	}
	return Source{}, false
}

func datasetSource(info DatasetInfo) Source {
	fields := make([]Field, 0, len(info.Columns))
	for _, c := range info.Columns {
		fields = append(fields, Field{Key: c.Key, Label: c.Label, Kind: c.Kind})
	}
	owner := info.Owner
	return Source{Key: info.ID, Label: info.Name, Description: "Uploaded data", Kind: SourceDataset, Fields: fields, Owner: &owner}
}

func IsGroupable(f Field) bool {
	return f.Kind != KindMoney
}

func AggregationsOf(f Field) []string {
	switch f.Kind {
	case KindNumber, KindMoney:
		return []string{aggCount, aggCountDistinct, aggSum, aggAvg, aggMin, aggMax}
	case KindDate:
		return []string{aggCount, aggCountDistinct, aggMin, aggMax}
	case KindText:
	}
	return []string{aggCount, aggCountDistinct}
}
