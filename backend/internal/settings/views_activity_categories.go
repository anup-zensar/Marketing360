package settings

import "mapp/internal/platform/record"

type ActivityCategoryView struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	IsDefault     bool         `json:"isDefault"`
	ActivityTypes []SettingRef `json:"activityTypes"` // the types in the category, by name
	Active        bool         `json:"active"`
	Usage         Usage        `json:"usage"`
	record.Meta
}

type ActivityCategoryPage struct {
	Items []ActivityCategoryView `json:"items"`
	Total int                    `json:"total"`
}

type ActivityCategoryInput struct {
	Name            string   `json:"name"`
	IsDefault       bool     `json:"isDefault"`
	ActivityTypeIDs []string `json:"activityTypeIds"`
}

// ActivityTypeIDs, when sent, is the complete new list of types in the category; a type listed here leaves
// the category it was in.
type ActivityCategoryUpdate struct {
	Name            *string   `json:"name"`
	IsDefault       *bool     `json:"isDefault"`
	ActivityTypeIDs *[]string `json:"activityTypeIds"`
	Active          *bool     `json:"active"`
	Version         *int      `json:"version"`
}
