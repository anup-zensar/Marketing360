package dataset

import (
	"context"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/widgetdata"
)

var _ widgetdata.DatasetTables = (*Service)(nil)

// Visible lists the datasets the person may read as widget data, by name, with their typed columns.
func (s *Service) Visible(ctx context.Context, a actor.Actor) ([]widgetdata.DatasetInfo, error) {
	var infos []widgetdata.DatasetInfo
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		datasets, err := s.repo.ListVisible(ctx, a.ID)
		if err != nil {
			return err
		}
		persons, _, err := s.refs(ctx, datasets, nil)
		infos = make([]widgetdata.DatasetInfo, 0, len(datasets))
		for _, d := range datasets {
			infos = append(infos, infoOf(d, authz.PersonRef(persons[d.OwnerID])))
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return infos, nil
}

// Table answers not_found for a dataset the person may not see or that is deleted; money cells keep their
// column's currency.
func (s *Service) Table(ctx context.Context, a actor.Actor, id string) (widgetdata.Table, error) {
	var table widgetdata.Table
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		d, err := s.visible(ctx, a, id)
		if err != nil {
			return err
		}
		persons, _, err := s.refs(ctx, []Dataset{d}, nil)
		if err != nil {
			return err
		}
		rows, err := s.repo.Rows(ctx, id)
		table = widgetdata.Table{Info: infoOf(d, authz.PersonRef(persons[d.OwnerID])), Rows: make([]widgetdata.Row, 0, len(rows))}
		for _, r := range rows {
			table.Rows = append(table.Rows, valuesOf(d.Columns, r))
		}
		return err
	})
	if err != nil {
		return widgetdata.Table{}, err
	}
	return table, nil
}

func infoOf(d Dataset, owner authz.PersonRef) widgetdata.DatasetInfo {
	columns := make([]widgetdata.DatasetColumn, 0, len(d.Columns))
	for _, c := range d.Columns {
		columns = append(columns, widgetdata.DatasetColumn{
			Key: c.Key, Label: c.Label, Kind: widgetdata.FieldKind(c.Kind), Currency: c.currency(),
		})
	}
	return widgetdata.DatasetInfo{ID: d.ID, Name: d.Name, Owner: owner, Columns: columns}
}

func valuesOf(columns []Column, r Row) widgetdata.Row {
	values := make(widgetdata.Row, len(columns))
	for i, c := range columns {
		if i >= len(r) || r[i].Empty {
			values[i] = widgetdata.Value{Empty: true}
			continue
		}
		values[i] = valueOf(c, r[i])
	}
	return values
}

func valueOf(c Column, cell Cell) widgetdata.Value {
	switch c.Kind {
	case KindNumber:
		n, err := widgetdata.ParseNumber(cell.Text)
		return widgetdata.Value{Number: n, Empty: err != nil}
	case KindDate:
		d, err := date.Parse(cell.Text)
		return widgetdata.Value{Date: d, Empty: err != nil}
	case KindMoney:
		if cell.Minor == nil {
			return widgetdata.Value{Empty: true}
		}
		return widgetdata.Value{Money: money.Money{Amount: money.FromCents(*cell.Minor), Currency: c.currency()}}
	case KindText:
	}
	return widgetdata.Value{Text: cell.Text}
}
