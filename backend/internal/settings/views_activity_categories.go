package settings

import "mapp/internal/platform/record"

// ActivityCategoryView is an activity category as the API returns it (contract ActivityCategory, 047-C7).
type ActivityCategoryView struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	IsDefault     bool         `json:"isDefault"`
	ActivityTypes []SettingRef `json:"activityTypes"` // the types in the category, by name
	Active        bool         `json:"active"`
	Usage         Usage        `json:"usage"`
	record.Meta
}

// ActivityCategoryPage is the answer of GET /api/v1/settings/activity-categories.
type ActivityCategoryPage struct {
	Items []ActivityCategoryView `json:"items"`
	Total int                    `json:"total"`
}

// ActivityCategoryInput is the body of POST /api/v1/settings/activity-categories.
type ActivityCategoryInput struct {
	Name            string   `json:"name"`
	IsDefault       bool     `json:"isDefault"`
	ActivityTypeIDs []string `json:"activityTypeIds"`
}

// ActivityCategoryUpdate is the body of PATCH /api/v1/settings/activity-categories/{categoryId}.
// ActivityTypeIDs, when sent, is the complete new list of types in the category; a type listed here leaves
// the category it was in.
type ActivityCategoryUpdate struct {
	Name            *string   `json:"name"`
	IsDefault       *bool     `json:"isDefault"`
	ActivityTypeIDs *[]string `json:"activityTypeIds"`
	Active          *bool     `json:"active"`
	Version         *int      `json:"version"`
}
