// Package attributiontest is the in-memory fake of the attribution store: the designated model.
package attributiontest

import (
	"context"

	"mapp/internal/measure/attribution"
	"mapp/internal/platform/fake"
)

const settingsID = "attribution"

type row struct {
	ID string `json:"id"`
	attribution.Settings
}

type Store struct {
	settings *fake.Table[row]
}

func NewStore(db *fake.DB) Store {
	return Store{settings: fake.TableOf(db, "attribution_settings", func(r row) string { return r.ID })}
}

func (s Store) Settings(context.Context) (attribution.Settings, error) {
	r, found, err := s.settings.Get(settingsID)
	if err != nil || !found {
		return attribution.DefaultSettings(), err
	}
	return r.Settings, nil
}

func (s Store) PutSettings(ctx context.Context, st attribution.Settings) error {
	return s.settings.Put(ctx, row{ID: settingsID, Settings: st})
}
