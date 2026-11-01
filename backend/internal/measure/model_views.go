package measure

import (
	"time"

	"mapp/internal/platform/record"
)

const KindReportView = "reportView"

// StoredView is one person's saved view of one report (039-C2): the report's canonical address query.
type StoredView struct {
	ID        string `json:"id"`
	PersonID  string `json:"personId"`
	Report    string `json:"report"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
	Query     string `json:"query"`
	record.Meta
}

type ReportView struct {
	ID        string    `json:"id"`
	Report    string    `json:"report"`
	Name      string    `json:"name"`
	IsDefault bool      `json:"isDefault"`
	Query     string    `json:"query"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ReportViewPage struct {
	Items []ReportView `json:"items"`
	Total int          `json:"total"`
}

type ViewCreate struct {
	Name      string `json:"name"`
	Query     string `json:"query"`
	IsDefault bool   `json:"isDefault"`
}

type ViewUpdate struct {
	Version   int     `json:"version"`
	Name      *string `json:"name"`
	Query     *string `json:"query"`
	IsDefault *bool   `json:"isDefault"`
}

func (v StoredView) view() ReportView {
	return ReportView{ID: v.ID, Report: v.Report, Name: v.Name, IsDefault: v.IsDefault, Query: v.Query,
		Version: v.Version, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

type viewLog struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
	Query     string `json:"query"`
}

func (v StoredView) logView() viewLog {
	return viewLog{Name: v.Name, IsDefault: v.IsDefault, Query: v.Query}
}
