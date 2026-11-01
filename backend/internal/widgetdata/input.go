package widgetdata

import (
	"context"
	"fmt"
	"strings"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/money"
)

// input is one widget's request with what it reads: a built-in source's reader, or a dataset's rows in USD.
type input struct {
	service *Service
	actor   actor.Actor
	request Request
	plan    *compiled
	reader  SourceReader
	rows    []Row
	notes   []string
}

func (in *input) answer(ctx context.Context) (Data, error) {
	rows, notes, err := in.read(ctx, in.prefilter())
	if err != nil {
		return Data{}, err
	}
	out, err := in.plan.run(rows)
	if err != nil {
		return Data{}, err
	}
	src := in.plan.source
	data := Data{
		Source: src.Key, Columns: in.plan.columns(), Rows: out.rows, Totals: out.totals, GroupCount: out.groupCount,
		Notes:          append(in.filterNotes(), in.plan.notes(notes, out)...),
		FiltersApplied: FiltersApplied{Date: src.Filters.Date != "", Plan: src.Filters.Plan, Team: src.Filters.Team},
	}
	if data.Rows == nil {
		data.Rows = []DataRow{}
	}
	data.Comparison, err = in.comparison(ctx, out)
	return data, err
}

func (in *input) read(ctx context.Context, f Prefilter) ([]Row, []string, error) {
	if in.reader == nil {
		return in.rows, in.notes, nil
	}
	found, err := in.reader.Rows(ctx, in.actor, f)
	return found.Items, found.Notes, err
}

// prefilter keeps the dashboard filters the source applies; datasets take none.
func (in *input) prefilter() Prefilter {
	f, src := in.request.Filters, in.plan.source
	var p Prefilter
	if f == nil || src.Kind == SourceDataset {
		return p
	}
	if src.Filters.Date != "" {
		p.From, p.To = f.DateFrom, f.DateTo
	}
	if src.Filters.Plan {
		p.PlanIDs = f.PlanIDs
	}
	if src.Filters.Team {
		p.TeamIDs = f.TeamIDs
	}
	return p
}

func (in *input) filterNotes() []string {
	f, applies := in.request.Filters, in.plan.source.Filters
	notes := []string{}
	if f == nil {
		return notes
	}
	if (f.DateFrom != nil || f.DateTo != nil) && applies.Date == "" {
		notes = append(notes, noteNoDate)
	}
	if len(f.PlanIDs) > 0 && !applies.Plan {
		notes = append(notes, noteNoPlan)
	}
	if len(f.TeamIDs) > 0 && !applies.Team {
		notes = append(notes, noteNoTeam)
	}
	return notes
}

func (in *input) comparison(ctx context.Context, out outcome) (*Comparison, error) {
	c := in.request.Comparison
	if c == nil {
		return nil, nil
	}
	m, current := in.plan.measures[0], out.values[0]
	if c.Kind == compareTarget {
		return againstTarget(c, m, current), nil
	}
	f := in.request.Filters
	period := previousPeriod(*f.DateFrom, *f.DateTo)
	p := in.prefilter()
	p.From, p.To = &period.From, &period.To
	rows, _, err := in.read(ctx, p)
	if err != nil {
		return nil, err
	}
	before, err := in.plan.run(rows)
	if err != nil {
		return nil, err
	}
	return againstPrevious(m, current, before.values[0], period), nil
}

func (p *compiled) notes(fromReader []string, out outcome) []string {
	var notes []string
	for _, note := range fromReader {
		if !strings.HasPrefix(note, financeNotePrefix) || p.usesFinance() {
			notes = append(notes, note)
		}
	}
	if out.multi {
		notes = append(notes, noteSeveralTeams)
	}
	if out.limited && !p.groupRest {
		notes = append(notes, fmt.Sprintf(noteShowing, min(p.limit, out.groupCount), out.groupCount))
	}
	return notes
}

func (p *compiled) usesFinance() bool {
	used := []Field{}
	for _, g := range []*grouping{p.dimension, p.series} {
		if g != nil {
			used = append(used, g.field)
		}
	}
	for _, m := range p.measures {
		used = append(used, m.field)
	}
	for _, c := range p.conditions {
		used = append(used, c.field)
	}
	for _, f := range used {
		if f.Finance {
			return true
		}
	}
	return false
}

func (p *compiled) columns() []Column {
	var out []Column
	for _, g := range []struct {
		key string
		by  *grouping
	}{{"dimension", p.dimension}, {"series", p.series}} {
		if g.by != nil {
			out = append(out, Column{Key: g.key, Label: g.by.field.Label, Kind: g.by.field.Kind, DateGroup: bucketOrNil(g.by.bucket)})
		}
	}
	for i, m := range p.measures {
		out = append(out, Column{Key: fmt.Sprintf("m%d", i), Label: m.label, Kind: m.kind()})
	}
	return out
}

func bucketOrNil(bucket string) *string {
	if bucket == "" {
		return nil
	}
	return &bucket
}

// inUSD converts a dataset's money columns to USD at the current rate; an amount in a currency without a
// rate is left out, and said so.
func (s *Service) inUSD(ctx context.Context, t Table) ([]Row, []string, error) {
	if s.deps.Rates == nil {
		return t.Rows, nil, nil
	}
	rates, err := s.deps.Rates.Rates(ctx)
	if err != nil {
		return nil, nil, err
	}
	var notes []string
	rows := make([]Row, len(t.Rows))
	for r, row := range t.Rows {
		rows[r] = make(Row, len(row))
		for c, v := range row {
			converted, known := inUSD(v, columnCurrency(t.Info, c), rates)
			if !known {
				notes = []string{noteUnknownCurrency}
			}
			rows[r][c] = converted
		}
	}
	return rows, notes, nil
}

// columnCurrency is the currency of a money column, "" for any other.
func columnCurrency(info DatasetInfo, c int) string {
	if c >= len(info.Columns) || info.Columns[c].Kind != KindMoney {
		return ""
	}
	return info.Columns[c].Currency
}

func inUSD(v Value, currency string, rates money.Rates) (Value, bool) {
	if currency == "" || v.Empty {
		return v, true
	}
	converted, err := rates.Convert(money.Money{Amount: v.Money.Amount, Currency: currency}, money.BaseCurrency)
	if err != nil {
		return noValue(), false
	}
	return Value{Money: converted}, true
}
