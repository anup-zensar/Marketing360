package dataset

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"mapp/internal/platform/errs"
	"mapp/internal/platform/record"
)

const (
	maxNameRunes        = 80
	maxDescriptionRunes = 500
	maxCellProblems     = 50
	maxQuotedRunes      = 40
	nameField           = "name"
	descriptionField    = "description"
	columnsField        = "columns"
	columnField         = "columns[%d]"
	columnLabelField    = "columns[%d].label"
	columnKindField     = "columns[%d].kind"
	columnCurrencyField = "columns[%d].currency"
)

func trimmed(text string) string {
	return strings.TrimSpace(text)
}

func checkName(name string, problems *errs.Problems) {
	switch {
	case name == "":
		problems.Add(nameField, msgNameRequired)
	case utf8.RuneCountInString(name) > maxNameRunes:
		problems.Add(nameField, msgNameTooLong)
	case record.CheckSingleLine(name) != "":
		problems.Add(nameField, record.CheckSingleLine(name))
	}
}

func checkDescription(description string, problems *errs.Problems) {
	if utf8.RuneCountInString(description) > maxDescriptionRunes {
		problems.Add(descriptionField, msgDescription)
	}
}

// columnsOf turns one input per file column, positions 0… in order, into the columns to store.
func columnsOf(inputs []ColumnInput, width int, problems *errs.Problems) []Column {
	if len(inputs) != width {
		problems.Add(columnsField, fmt.Sprintf(msgColumnsMismatch, width))
		return nil
	}
	columns := make([]Column, 0, width)
	labels := map[string]bool{}
	for i, in := range inputs {
		if in.Position != i {
			problems.Add(columnsField, fmt.Sprintf(msgColumnsMismatch, width))
			return nil
		}
		c := Column{Key: keyOf(i), Position: i, Label: trimmed(in.Label), Kind: in.Kind}
		checkLabel(c.Label, i, labels, problems)
		switch {
		case !c.Kind.isKnown():
			problems.Add(fmt.Sprintf(columnKindField, i), msgKind)
		case c.Kind == KindMoney && in.Currency == nil:
			problems.Add(fmt.Sprintf(columnCurrencyField, i), msgCurrency)
		case c.Kind == KindMoney:
			code := strings.ToUpper(trimmed(*in.Currency))
			c.Currency = &code
		}
		columns = append(columns, c)
	}
	return columns
}

func keyOf(position int) string {
	return fmt.Sprintf("c%d", position+1)
}

// checkLabel also remembers the label in taken, so the next column with the same name ignoring case is refused.
func checkLabel(label string, i int, taken map[string]bool, problems *errs.Problems) {
	field := fmt.Sprintf(columnLabelField, i)
	switch {
	case label == "":
		problems.Add(field, msgLabelRequired)
	case utf8.RuneCountInString(label) > maxLabelRunes:
		problems.Add(field, msgLabelTooLong)
	case taken[strings.ToLower(label)]:
		problems.Add(field, fmt.Sprintf(msgLabelTaken, label))
	}
	taken[strings.ToLower(label)] = true
}

func cellProblem(c Column, row int, text string) string {
	quoted := text
	if utf8.RuneCountInString(text) > maxQuotedRunes {
		quoted = string([]rune(text)[:maxQuotedRunes]) + "…"
	}
	switch c.Kind {
	case KindNumber:
		return fmt.Sprintf(msgNotANumber, row, c.Label, quoted)
	case KindDate:
		return fmt.Sprintf(msgNotADate, row, c.Label, quoted)
	case KindMoney:
		return fmt.Sprintf(msgNotAnAmount, row, c.Label, quoted)
	case KindText:
	}
	return fmt.Sprintf(msgTextTooLong, row, c.Label)
}

func listedFor(problems errs.Problems, column int) int {
	field, count := fmt.Sprintf(columnField, column), 0
	for _, p := range problems {
		if p.Field == field {
			count++
		}
	}
	return count
}

// checkNameFree answers name_taken when the owner has another dataset of that name, ignoring case.
func (s *Service) checkNameFree(ctx context.Context, ownerID, name, exceptID string) error {
	datasets, err := s.repo.ListVisible(ctx, ownerID)
	if err != nil {
		return err
	}
	for _, d := range datasets {
		if d.OwnerID == ownerID && d.ID != exceptID && strings.EqualFold(d.Name, name) {
			return errs.NameTaken(fmt.Sprintf(msgNameTaken, d.Name))
		}
	}
	return nil
}
