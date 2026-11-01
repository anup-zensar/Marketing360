package widgetdata

import (
	"mapp/internal/platform/authz"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

type FieldKind string

const (
	KindText   FieldKind = "text"
	KindNumber FieldKind = "number"
	KindMoney  FieldKind = "money"
	KindDate   FieldKind = "date"
)

type Field struct {
	Key, Label     string
	Kind           FieldKind
	Multi, Finance bool // several values (a plan's teams); empty outside the finance scope
}

// Value is one cell; Empty wins over the other fields, and only the field of the column's kind is read.
type Value struct {
	Empty  bool
	Text   string
	Texts  []string
	Number Number
	Money  money.Money // USD for built-in sources; the column's currency for datasets
	Date   date.Date
}

type Row []Value

type DatasetColumn struct {
	Key, Label string
	Kind       FieldKind
	Currency   string
}

type DatasetInfo struct {
	ID, Name string
	Owner    authz.PersonRef
	Columns  []DatasetColumn
}

type Table struct {
	Info DatasetInfo
	Rows []Row
}
