package dataset

import (
	"context"
	"fmt"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

const maxDatasets = 50

// Preview stores nothing; it also says when the person has no room left for another dataset.
func (s *Service) Preview(ctx context.Context, a actor.Actor, up Upload) (Preview, error) {
	f, err := readFile(up)
	if err != nil {
		return Preview{}, err
	}
	p := previewOf(f)
	err = s.tx.Read(ctx, func(ctx context.Context) error {
		full, countErr := s.isFull(ctx, a)
		if full {
			p.Problems = append(p.Problems, msgTooManyDatasets)
		}
		return countErr
	})
	if err != nil {
		return Preview{}, err
	}
	return p, nil
}

// Create answers every problem at once, cells included, and stores nothing until there is none.
func (s *Service) Create(ctx context.Context, a actor.Actor, in CreateInput) (View, error) {
	f, err := readFile(Upload{FileName: in.FileName, Content: in.Content})
	if err != nil {
		return View{}, err
	}
	d := Dataset{
		Name: trimmed(in.Name), Description: trimmed(in.Description), OwnerID: a.ID, FileName: f.name,
		SheetName: f.sheet, RowCount: len(f.rows), SharedWith: []string{},
	}
	var problems errs.Problems
	checkName(d.Name, &problems)
	checkDescription(d.Description, &problems)
	d.Columns = columnsOf(in.Columns, f.width, &problems)
	rows, cellProblems := typedRows(f, d.Columns)
	return s.insert(ctx, a, checked{dataset: d, rows: rows, problems: append(problems, cellProblems...)})
}

type checked struct {
	dataset  Dataset
	rows     []Row
	problems errs.Problems
}

func (s *Service) insert(ctx context.Context, a actor.Actor, c checked) (View, error) {
	var view View
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		d := c.dataset
		if err := s.checkStorable(ctx, a, d, c.problems); err != nil {
			return err
		}
		id, err := s.ids.Next(ctx, ids.PrefixDataset)
		if err != nil {
			return err
		}
		d.ID, d.Meta = id, record.NewMeta(a, s.now())
		if err = s.repo.Insert(ctx, d, c.rows); err != nil {
			return err
		}
		if _, err = s.history.Log(ctx, a, history.LogInput{Kind: kindDataset, RecordID: d.ID, RecordName: d.Name,
			Action: history.ActionCreated, Details: fmt.Sprintf(msgUploaded, d.FileName, d.RowCount)}); err != nil {
			return err
		}
		view, err = s.view(ctx, a, d)
		return err
	})
	if err != nil {
		return View{}, err
	}
	return view, nil
}

// checkStorable adds the checks that read stored data to problems: active currencies and the person's room for
// another dataset; then a name the person already uses is name_taken.
func (s *Service) checkStorable(ctx context.Context, a actor.Actor, d Dataset, problems errs.Problems) error {
	if err := s.checkCurrencies(ctx, d.Columns, &problems); err != nil {
		return err
	}
	full, err := s.isFull(ctx, a)
	if err != nil {
		return err
	}
	if full {
		problems.Add(nameField, msgTooManyDatasets)
	}
	if err := problems.Err(); err != nil {
		return err
	}
	return s.checkNameFree(ctx, a.ID, d.Name, d.ID)
}

func (s *Service) isFull(ctx context.Context, a actor.Actor) (bool, error) {
	count, err := s.repo.CountOfOwner(ctx, a.ID)
	return count >= maxDatasets, err
}

func (s *Service) checkCurrencies(ctx context.Context, columns []Column, problems *errs.Problems) error {
	for i, c := range columns {
		if c.Kind != KindMoney || c.Currency == nil {
			continue
		}
		currency, found, err := s.currencies.Currency(ctx, *c.Currency)
		if err != nil {
			return err
		}
		if !found || !currency.Active {
			problems.Add(fmt.Sprintf(columnCurrencyField, i), msgCurrency)
		}
	}
	return nil
}

// typedRows types every cell; per column it lists the problems while fewer than 50 are listed in all, then says
// how many more the column has.
func typedRows(f file, columns []Column) ([]Row, errs.Problems) {
	rows := make([]Row, len(f.rows))
	for j := range rows {
		rows[j] = make(Row, len(columns))
	}
	var problems errs.Problems
	for i, c := range columns {
		if !c.Kind.isKnown() {
			continue
		}
		failures := 0
		for j, r := range f.rows {
			cell, ok := parseCell(c.Kind, r.Cell(i))
			rows[j][i] = cell
			if ok {
				continue
			}
			failures++
			if len(problems) < maxCellProblems {
				problems.Add(fmt.Sprintf(columnField, i), cellProblem(c.Kind, r.Number, r.Cell(i)))
			}
		}
		if more := failures - listedFor(problems, i); more > 0 {
			problems.Add(fmt.Sprintf(columnField, i), fmt.Sprintf(msgMoreProblems, more))
		}
	}
	return rows, problems
}
