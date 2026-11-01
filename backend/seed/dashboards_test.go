package seed_test

import (
	"context"
	"fmt"
	"testing"

	"mapp/internal/customdashboard"
	"mapp/internal/customdashboard/customdashboardtest"
)

// DB-1001 is Priya's dashboard shared with Lubricants Europe, valid as the API would save it, with a bar,
// a line, a key figure, a table, a text note and a bar over her dataset DS-1001.
func TestLoad_DemoDashboard_FR008(t *testing.T) {
	t.Parallel()
	db := loadedStore(t)
	d := loadDemo(t)
	var stored customdashboard.Dashboard
	var found bool
	read(t, db, func(ctx context.Context) error {
		var err error
		stored, found, err = customdashboardtest.NewStore(db).Get(ctx, "DB-1001")
		return err
	})
	if !found {
		t.Fatal("DB-1001 is not stored")
	}
	names := map[string]string{}
	for _, p := range d.people {
		names[p.ID] = p.Name
	}
	for _, team := range d.teams {
		names[team.ID] = team.Name
	}
	types := ""
	for _, w := range stored.Widgets {
		types += w.Type + " "
	}
	got := fmt.Sprintf("%s by %s, shared with %v: %s| %v", stored.Name, names[stored.OwnerID], names[stored.SharedWith[0]],
		types, customdashboard.DatasetSources(stored.Widgets))
	if want := "Lubricants Europe overview by Priya Raman, shared with Lubricants Europe: bar line kpi table text bar | [DS-1001]"; got != want {
		t.Errorf("demo dashboard = %s\nwant %s", got, want)
	}
	if err := customdashboard.Validate(stored); err != nil {
		t.Errorf("Validate(DB-1001) = %v", err)
	}
}
