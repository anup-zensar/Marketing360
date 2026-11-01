package measure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mapp/internal/measure"
	peoplearea "mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/database/dbtest"
	"mapp/internal/platform/record"
)

var _ measure.Store = (*measure.PostgresStore)(nil)

func TestPostgresStoreReportViewsPersonIsolation_039_FR025(t *testing.T) {
	db := dbtest.NewDB(t)
	store, directory := measure.NewPostgresStore(db), peoplearea.NewPostgresStore(db)
	meta := record.NewMeta(actor.System, time.Date(2026, time.October, 30, 9, 0, 0, 0, time.UTC))
	ctx := t.Context()
	err := db.Write(ctx, func(ctx context.Context) error {
		return errors.Join(
			directory.Put(ctx, peoplearea.Person{ID: "U-1002", Name: "Priya Raman", Email: "priya@example.com", Meta: meta}),
			directory.Put(ctx, peoplearea.Person{ID: "U-1003", Name: "Daniel Okoro", Email: "daniel@example.com", Meta: meta}),
			store.PutView(ctx, measure.StoredView{ID: "RV-1001", PersonID: "U-1002", Report: measure.KeyPlanBudget,
				Name: "Europe", Query: "year=2026", Meta: meta}),
			store.PutView(ctx, measure.StoredView{ID: "RV-1002", PersonID: "U-1003", Report: measure.KeyPlanBudget,
				Name: "Europe", Query: "", Meta: meta}),
			store.PutView(ctx, measure.StoredView{ID: "RV-1003", PersonID: "U-1002", Report: measure.KeyApprovals,
				Name: "Mine", Query: "", Meta: meta}))
	})
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	err = db.Read(ctx, func(ctx context.Context) error {
		priyas, err := store.ListViews(ctx, "U-1002", measure.KeyPlanBudget)
		if err != nil {
			return err
		}
		if len(priyas) != 1 || priyas[0].ID != "RV-1001" || priyas[0].Query != "year=2026" {
			t.Errorf("Priya's Plan budget views = %+v, want RV-1001 only", priyas)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
}
