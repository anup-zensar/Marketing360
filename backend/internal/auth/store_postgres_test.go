package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/people"
	"mapp/internal/platform/database"
	"mapp/internal/platform/database/dbtest"
)

// peopleDB is a migrated database holding Priya and Daniel.
func peopleDB(t *testing.T) (*database.DB, *people.PostgresStore) {
	t.Helper()
	db := dbtest.NewDB(t)
	repo := people.NewPostgresStore(db)
	if err := db.Write(t.Context(), func(ctx context.Context) error {
		if err := repo.Put(ctx, priya); err != nil {
			return err
		}
		return repo.Put(ctx, daniel)
	}); err != nil {
		t.Fatalf("put people: %v", err)
	}
	return db, repo
}

func TestRecordSignIn_OnlyTheTime_FR010(t *testing.T) {
	t.Parallel()
	db, repo := peopleDB(t)
	store := NewPostgresStore(db)
	before, _, err := repo.Get(t.Context(), priya.ID)
	if err != nil {
		t.Fatal(err)
	}

	at := time.Date(2026, 3, 12, 9, 30, 0, 0, time.UTC)
	for _, moment := range []time.Time{at.Add(-time.Hour), at} {
		record := func(ctx context.Context) error { return store.RecordSignIn(ctx, priya.ID, moment) }
		if err = db.Write(t.Context(), record); err != nil {
			t.Fatal(err)
		}
	}

	after, _, err := repo.Get(t.Context(), priya.ID)
	if diff := cmp.Diff(before, after); err != nil || diff != "" {
		t.Errorf("the person changed (-before +after):\n%s %v", diff, err)
	}
	last, err := store.LastSignIns(t.Context())
	if diff := cmp.Diff(map[string]time.Time{priya.ID: at}, last); err != nil || diff != "" {
		t.Errorf("last sign-ins (-want +got):\n%s %v", diff, err)
	}
}

func TestLastSignIns_NoneBeforeFirstSignIn_FR011(t *testing.T) {
	t.Parallel()
	db, _ := peopleDB(t)
	store := NewPostgresStore(db)

	last, err := store.LastSignIns(t.Context())
	if err != nil || len(last) != 0 {
		t.Errorf("last sign-ins = %v, %v; want none: earlier times are never invented", last, err)
	}
	since, err := store.SignInsRecordedSince(t.Context())
	if err != nil || since.IsZero() || since.After(time.Now()) {
		t.Errorf("recorded since = %v, %v; want the moment migration 00029 ran", since, err)
	}
}
