package access

import "mapp/internal/platform/authz"

func activityCategoryRoutes() map[string]Rule {
	return map[string]Rule{
		// 047 Activity categories: readable by everyone, changed with settings.manage (FR-023).
		"GET /api/v1/settings/activity-categories":                 signedIn(),
		"POST /api/v1/settings/activity-categories":                need(authz.SettingsManage),
		"GET /api/v1/settings/activity-categories/{categoryId}":    signedIn(),
		"PATCH /api/v1/settings/activity-categories/{categoryId}":  need(authz.SettingsManage),
		"DELETE /api/v1/settings/activity-categories/{categoryId}": need(authz.SettingsManage),
	}
}
