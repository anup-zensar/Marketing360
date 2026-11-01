package access

import "mapp/internal/platform/authz"

func datasetRoutes() map[string]Rule {
	return map[string]Rule{
		"GET /api/v1/datasets":                     need(authz.DashboardsBuild),
		"POST /api/v1/datasets/preview":            need(authz.DashboardsBuild),
		"POST /api/v1/datasets":                    need(authz.DashboardsBuild),
		"GET /api/v1/datasets/{datasetId}":         need(authz.DashboardsBuild),
		"PATCH /api/v1/datasets/{datasetId}":       need(authz.DashboardsBuild),
		"DELETE /api/v1/datasets/{datasetId}":      need(authz.DashboardsBuild),
		"GET /api/v1/datasets/{datasetId}/rows":    need(authz.DashboardsBuild),
		"PUT /api/v1/datasets/{datasetId}/sharing": need(authz.DashboardsBuild),
	}
}
