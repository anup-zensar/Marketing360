package attribution

import (
	"context"
	"encoding/json"
	"errors"

	"mapp/internal/history"
)

var errSettingsKept = errors.New("attribution: the attribution setting is never removed")

// Kinds returns the history kind of Settings › Attribution, which is only ever logged.
func (s *Service) Kinds() []history.Kind {
	return []history.Kind{{Name: settingsKind, Singular: settingsNoun, Plural: "attribution settings",
		Snapshot: s.snapshot, Remove: func(context.Context, string) error { return errSettingsKept }, Put: s.put,
		Describe: describe}}
}

func (s *Service) snapshot(ctx context.Context, _ string) (history.Snapshot, bool, error) {
	st, err := s.repo.Settings(ctx)
	if err != nil {
		return history.Snapshot{}, false, err
	}
	data, err := json.Marshal(st)
	return history.Snapshot{ID: settingsID, Name: settingsName, Data: data}, true, err
}

func (s *Service) put(ctx context.Context, snap history.Snapshot) error {
	var st Settings
	if err := json.Unmarshal(snap.Data, &st); err != nil {
		return err
	}
	return s.repo.PutSettings(ctx, st)
}

func describe(_ context.Context, field string, value any) (string, string, error) {
	if field != fieldModel {
		return "", "", nil
	}
	key, _ := value.(string)
	return "Designated model", Model(key).Name(), nil
}
