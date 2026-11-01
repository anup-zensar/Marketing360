package measure

import (
	"context"
	"fmt"
	"slices"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

// Report is one report of the catalogue as its feature builds it; Sheets gives its widgets for the export.
type Report interface {
	Key() string
	Available() bool
	Sheets(ctx context.Context, a actor.Actor, f Filter) ([]WidgetSheet, error)
}

// Parameterised reports have address parameters of their own (a tab, a sort) that saved views keep.
type Parameterised interface {
	Params() []string
}

type entry struct {
	key, path, name, purpose string
	requires                 authz.Permission // beyond reports.view; "" for none
}

func entries() []entry {
	return []entry{
		{KeyPlanBudget, "plan-budget", "Plan budget",
			"Budget, planned cost, committed and actual by month, and which plans, campaigns or activities lead or lag.", ""},
		{KeyApprovals, "approvals", "Approvals",
			"How many approvals wait, how long decisions take, and every approval round.", ""},
		{KeyPurchaseRequests, "purchase-requests", "Purchase requests",
			"What is committed to each vendor and owner by delivery month, and what has been invoiced.", ""},
		{KeyPlanDetails, "plan-details", "Plan details",
			"How activities and planned cost spread across each Shell attribute, and what is not filled in.", ""},
		{KeyGsapLink, "gsap-link", "GSAP link",
			"SWBS codes not received, purchase requests waiting for GSAP, and the purchase orders of each line item.", ""},
		{KeyUserAccess, "user-access", "User access",
			"Who has which role and teams, and when they last signed in.", authz.AccessManage},
		{KeySummary, "summary", "Summary",
			"Whether the year is on track: projected against goal, the leading segments and the funnel.", ""},
		{KeyRevenue, "revenue", "Revenue",
			"Revenue against goal by month, and the plans, campaigns or activities behind it.", ""},
		{KeyFunnel, "funnel", "Funnel stages",
			"One funnel stage against goal by month, and where the funnel leaks.", ""},
	}
}

func Keys() []string {
	var keys []string
	for _, e := range entries() {
		keys = append(keys, e.key)
	}
	return keys
}

func KeyOfPath(path string) (string, bool) {
	i := slices.IndexFunc(entries(), func(e entry) bool { return e.path == path })
	if i < 0 {
		return "", false
	}
	return entries()[i].key, true
}

// Registry holds the report of each key; app/measure.go fills it, one register function per report (D-369).
type Registry struct {
	reports map[string]Report
}

func NewRegistry() *Registry {
	return &Registry{reports: map[string]Report{}}
}

func (r *Registry) Register(report Report) error {
	key := report.Key()
	if !slices.Contains(Keys(), key) {
		return fmt.Errorf("register report %q: not in the catalogue", key)
	}
	if _, taken := r.reports[key]; taken {
		return fmt.Errorf("register report %q: registered twice", key)
	}
	r.reports[key] = report
	return nil
}

func (r *Registry) Get(key string) (Report, bool) {
	report, found := r.reports[key]
	return report, found
}

// Catalogue lists the reports the person may open, in order; a report not registered or a stub is not available.
func (r *Registry) Catalogue(ctx context.Context) ReportCatalogue {
	items := []ReportEntry{}
	for _, e := range entries() {
		if e.requires != "" && !authz.AvailableAnywhere(ctx, e.requires).Available {
			continue
		}
		report, found := r.reports[e.key]
		items = append(items, ReportEntry{Key: e.key, Path: e.path, Name: e.name, Purpose: e.purpose,
			Available: found && report.Available()})
	}
	return ReportCatalogue{Items: items}
}

type stub struct{ key string }

// Stub is a report not built yet: listed as coming soon, its export has only the frame's sheets.
func Stub(key string) Report { return stub{key: key} }

func (s stub) Key() string     { return s.key }
func (s stub) Available() bool { return false }

func (s stub) Sheets(context.Context, actor.Actor, Filter) ([]WidgetSheet, error) {
	return []WidgetSheet{}, nil
}

// openable refuses a report the person's roles do not open, as the catalogue leaves it out (FR-003).
func openable(ctx context.Context, report string) error {
	i := slices.IndexFunc(entries(), func(e entry) bool { return e.key == report })
	if i < 0 {
		return errs.NotFound(msgNoReport)
	}
	if p := entries()[i].requires; p != "" {
		if a := authz.AvailableAnywhere(ctx, p); !a.Available {
			return errs.Forbidden(a.Reason)
		}
	}
	return nil
}
