package access

import "mapp/internal/platform/authz"

// dataSourceRoutes are custom: the service accepts dashboards.view or dashboards.build, then each source's permission.
func dataSourceRoutes() map[string]Rule {
	return map[string]Rule{
		"GET /api/v1/data-sources":        customWith(authz.DashboardsView),
		"POST /api/v1/data-sources/query": customWith(authz.DashboardsView),
	}
}
