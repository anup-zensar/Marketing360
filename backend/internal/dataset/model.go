// Package dataset holds the tables people upload from CSV or .xlsx files for their custom dashboards: typed
// columns, rows typed at upload, visible to the owner and to members of the teams it is shared with.
package dataset

import (
	"bytes"
	"encoding/json"
	"strconv"

	"mapp/internal/platform/record"
)

type Kind string

const (
	KindText   Kind = "text"
	KindNumber Kind = "number"
	KindDate   Kind = "date"
	KindMoney  Kind = "money"
)

func (k Kind) isKnown() bool {
	switch k {
	case KindText, KindNumber, KindDate, KindMoney:
		return true
	}
	return false
}

type Column struct {
	Key      string  `json:"key"`
	Position int     `json:"position"`
	Label    string  `json:"label"`
	Kind     Kind    `json:"kind"`
	Currency *string `json:"currency"`
}

func (c Column) currency() string {
	if c.Currency == nil {
		return ""
	}
	return *c.Currency
}

type Dataset struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	OwnerID     string   `json:"ownerId"`
	FileName    string   `json:"fileName"`
	SheetName   string   `json:"sheetName"`
	RowCount    int      `json:"rowCount"`
	Columns     []Column `json:"columns"`
	SharedWith  []string `json:"sharedWith"`
	record.Meta
}

// Cell is one stored value: empty, text (numbers as exact decimal text, dates as YYYY-MM-DD) or money in minor
// units of its column's currency. In JSON it is null, a string or an integer.
type Cell struct {
	Empty bool
	Text  string
	Minor *int64
}

type Row []Cell

func TextCell(text string) Cell {
	return Cell{Text: text}
}

func MinorCell(minor int64) Cell {
	return Cell{Minor: &minor}
}

func (c Cell) MarshalJSON() ([]byte, error) {
	switch {
	case c.Empty:
		return []byte("null"), nil
	case c.Minor != nil:
		return []byte(strconv.FormatInt(*c.Minor, 10)), nil
	}
	return json.Marshal(c.Text)
}

func (c *Cell) UnmarshalJSON(data []byte) error {
	switch {
	case bytes.Equal(data, []byte("null")):
		*c = Cell{Empty: true}
		return nil
	case len(data) > 0 && data[0] == '"':
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		*c = TextCell(text)
		return nil
	}
	minor, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*c = MinorCell(minor)
	return nil
}
