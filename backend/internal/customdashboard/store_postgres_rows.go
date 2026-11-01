package customdashboard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"mapp/internal/platform/database/sqlcdb"
)

func widgetRow(dashboardID string, position int, w Widget) (sqlcdb.InsertCustomDashboardWidgetsParams, error) {
	display, err := json.Marshal(w.Display)
	if err != nil {
		return sqlcdb.InsertCustomDashboardWidgetsParams{}, fmt.Errorf("encode the display of widget %s: %w", w.ID, err)
	}
	row := sqlcdb.InsertCustomDashboardWidgetsParams{
		DashboardID: dashboardID, ID: w.ID, Position: position, Type: w.Type, Title: w.Title, Subtitle: w.Subtitle,
		GridX: w.Layout.X, GridY: w.Layout.Y, GridW: w.Layout.W, GridH: w.Layout.H, Display: display, Note: w.Note,
	}
	if w.Query != nil {
		source := w.Query.Source
		row.Source = &source
		if row.Query, err = json.Marshal(w.Query); err != nil {
			return sqlcdb.InsertCustomDashboardWidgetsParams{}, fmt.Errorf("encode the query of widget %s: %w", w.ID, err)
		}
	}
	return row, nil
}

func widgetsByDashboard(rows []sqlcdb.CustomDashboardWidget) (map[string][]Widget, error) {
	widgets := map[string][]Widget{}
	for _, row := range rows {
		w := Widget{ID: row.ID, Type: row.Type, Title: row.Title, Subtitle: row.Subtitle, Note: row.Note,
			Layout: Layout{X: row.GridX, Y: row.GridY, W: row.GridW, H: row.GridH}}
		if err := decodeStrict(row.Display, &w.Display); err != nil {
			return nil, fmt.Errorf("decode the display of widget %s of %s: %w", row.ID, row.DashboardID, err)
		}
		if row.Query != nil {
			w.Query = &Query{}
			if err := decodeStrict(row.Query, w.Query); err != nil {
				return nil, fmt.Errorf("decode the query of widget %s of %s: %w", row.ID, row.DashboardID, err)
			}
		}
		widgets[row.DashboardID] = append(widgets[row.DashboardID], w)
	}
	return widgets, nil
}

// decodeStrict refuses unknown fields, so a stored query or display always has the contract's shape.
func decodeStrict(data []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("data after the JSON value")
	}
	return nil
}
