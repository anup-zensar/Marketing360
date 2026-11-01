package seed

import (
	"context"
	"fmt"
	"strings"

	"mapp/internal/dataset"
	"mapp/internal/platform/ids"
)

const (
	seedDatasetName     = "Event leads 2026"
	seedDatasetOwner    = "Priya Raman"
	seedDatasetCurrency = "EUR"
	seedDatasetColumns  = "Event,Date,Leads,Agency cost"
)

type EventLead struct {
	Event, Date, Leads string
	CostCents          int64
}

// EventLeads are DS-1001's rows, and those of the end-to-end upload fixtures: agency costs in EUR cents.
func EventLeads() []EventLead {
	return []EventLead{
		{"Lubricants Expo Hamburg", "2026-01-22", "140", 1850000},
		{"Fleet Day Rotterdam", "2026-02-12", "85", 920000},
		{"Workshop Tour Lyon", "2026-03-05", "62", 740050},
		{"Motorsport Hospitality Monza", "2026-03-26", "210", 3200000},
		{"Trade Fair Milan", "2026-04-16", "175", 2175000},
		{"Dealer Summit Madrid", "2026-05-07", "98", 1230000},
		{"Truck Show Birmingham", "2026-05-28", "120", 1580000},
		{"Agri Fair Poznan", "2026-06-18", "54", 610000},
		{"Mining Expo Vienna", "2026-07-09", "47", 890000},
		{"Autumn Roadshow Brussels", "2026-09-10", "133", 1425000},
		{"Workshop Tour Porto", "2026-10-01", "71", 795000},
		{"Partner Event Amsterdam", "2026-11-19", "160", 1960000},
	}
}

// addDatasets adds DS-1001, Priya Raman's uploaded event leads, shared with nobody (D-376).
func (d *demo) addDatasets(ctx context.Context) error {
	owner, err := find(d.people, "person", seedDatasetOwner)
	if err != nil {
		return err
	}
	id, err := d.next(ctx, ids.PrefixDataset)
	if err != nil {
		return err
	}
	currency := seedDatasetCurrency
	columns := []dataset.Column{}
	for i, kind := range []dataset.Kind{dataset.KindText, dataset.KindDate, dataset.KindNumber, dataset.KindMoney} {
		columns = append(columns, dataset.Column{Key: fmt.Sprintf("c%d", i+1), Position: i,
			Label: strings.Split(seedDatasetColumns, ",")[i], Kind: kind})
	}
	columns[3].Currency = &currency
	leads := EventLeads()
	rows := make([]dataset.Row, 0, len(leads))
	for _, l := range leads {
		rows = append(rows, dataset.Row{dataset.TextCell(l.Event), dataset.TextCell(l.Date), dataset.TextCell(l.Leads),
			dataset.MinorCell(l.CostCents)})
	}
	return d.stores.Datasets.Insert(ctx, dataset.Dataset{
		ID: id, Name: seedDatasetName, Description: "Leads and agency costs of the 2026 events, from the agency's sheet.",
		OwnerID: owner, FileName: "event-leads.xlsx", SheetName: "Leads", RowCount: len(rows), Columns: columns,
		SharedWith: []string{}, Meta: d.meta(),
	}, rows)
}
