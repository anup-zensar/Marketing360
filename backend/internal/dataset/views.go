package dataset

import (
	"time"

	"mapp/internal/people"
	"mapp/internal/platform/record"
)

type Availability struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason"`
}

type Access struct {
	Edit   Availability `json:"edit"`
	Share  Availability `json:"share"`
	Delete Availability `json:"delete"`
}

type DashboardRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type View struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Owner          people.PersonRef `json:"owner"`
	IsOwner        bool             `json:"isOwner"`
	SharedWith     []people.TeamRef `json:"sharedWith"`
	ShareableTeams []people.TeamRef `json:"shareableTeams"`
	FileName       string           `json:"fileName"`
	SheetName      string           `json:"sheetName"`
	RowCount       int              `json:"rowCount"`
	Columns        []Column         `json:"columns"`
	UsedBy         []DashboardRef   `json:"usedBy"`
	Access         Access           `json:"access"`
	record.Meta
}

type Summary struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Owner       people.PersonRef `json:"owner"`
	SharedWith  []people.TeamRef `json:"sharedWith"`
	RowCount    int              `json:"rowCount"`
	ColumnCount int              `json:"columnCount"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	Access      Access           `json:"access"`
}

type List struct {
	Mine         []Summary `json:"mine"`
	SharedWithMe []Summary `json:"sharedWithMe"`
}

type PreviewColumn struct {
	Position         int      `json:"position"`
	Label            string   `json:"label"`
	ProposedKind     Kind     `json:"proposedKind"`
	ProposedCurrency *string  `json:"proposedCurrency"`
	Samples          []string `json:"samples"`
	EmptyCount       int      `json:"emptyCount"`
}

type Preview struct {
	FileName  string          `json:"fileName"`
	SheetName string          `json:"sheetName"`
	RowCount  int             `json:"rowCount"`
	Columns   []PreviewColumn `json:"columns"`
	Rows      [][]string      `json:"rows"`
	Problems  []string        `json:"problems"`
}

type RowPage struct {
	Columns []Column    `json:"columns"`
	Rows    [][]*string `json:"rows"`
	Offset  int         `json:"offset"`
	Limit   int         `json:"limit"`
	Total   int         `json:"total"`
}

type Upload struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

type ColumnInput struct {
	Position int     `json:"position"`
	Label    string  `json:"label"`
	Kind     Kind    `json:"kind"`
	Currency *string `json:"currency"`
}

type CreateInput struct {
	FileName    string        `json:"fileName"`
	Content     string        `json:"content"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Columns     []ColumnInput `json:"columns"`
}

type LabelInput struct {
	Position int    `json:"position"`
	Label    string `json:"label"`
}

type UpdateInput struct {
	Version     *int         `json:"version"`
	Name        *string      `json:"name"`
	Description *string      `json:"description"`
	Columns     []LabelInput `json:"columns"`
}

type SharingInput struct {
	Version *int     `json:"version"`
	TeamIDs []string `json:"teamIds"`
}

type RowsQuery struct {
	Offset *int
	Limit  *int
}
