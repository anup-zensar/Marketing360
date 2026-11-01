package access

import "mapp/internal/platform/authz"

// customDashboardRoutes leave sight and ownership to the service: a dashboard the person may not see answers 404.
func customDashboardRoutes() map[string]Rule {
	return map[string]Rule{
		"GET /api/v1/custom-dashboards":                       need(authz.DashboardsView),
		"POST /api/v1/custom-dashboards":                      need(authz.DashboardsBuild),
		"GET /api/v1/custom-dashboards/{dashboardId}":         need(authz.DashboardsView),
		"PATCH /api/v1/custom-dashboards/{dashboardId}":       need(authz.DashboardsBuild),
		"DELETE /api/v1/custom-dashboards/{dashboardId}":      need(authz.DashboardsBuild),
		"POST /api/v1/custom-dashboards/{dashboardId}/copy":   need(authz.DashboardsBuild),
		"PUT /api/v1/custom-dashboards/{dashboardId}/sharing": need(authz.DashboardsBuild),
	}
}
