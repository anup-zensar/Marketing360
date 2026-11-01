package access

import "mapp/internal/platform/authz"

// measureRoutes are the rules of the Measure module's wave A (D-367); 040–045 add theirs with the wave A contract.
func measureRoutes() map[string]Rule {
	return map[string]Rule{
		"GET /api/v1/reports":                            need(authz.ReportsView),
		"GET /api/v1/reports/frame":                      need(authz.ReportsView),
		"GET /api/v1/reports/{report}/export":            need(authz.DataExport),
		"GET /api/v1/reports/{report}/views":             need(authz.ReportsView),
		"POST /api/v1/reports/{report}/views":            need(authz.ReportsView),
		"PATCH /api/v1/reports/{report}/views/{viewId}":  need(authz.ReportsView),
		"DELETE /api/v1/reports/{report}/views/{viewId}": need(authz.ReportsView),

		"GET /api/v1/settings/attribution":       signedIn(),
		"PUT /api/v1/settings/attribution":       need(authz.SettingsManage),
		"GET /api/v1/reports/plan-budget":        need(authz.ReportsView),
		"GET /api/v1/reports/plan-budget/ranked": need(authz.ReportsView),
		"GET /api/v1/reports/approvals":          need(authz.ReportsView),
		"GET /api/v1/reports/purchase-requests":  need(authz.ReportsView),
		"GET /api/v1/reports/plan-details":       need(authz.ReportsView),
		"GET /api/v1/reports/gsap-link":          need(authz.ReportsView),
		"GET /api/v1/reports/user-access":        need(authz.AccessManage),
	}
}
