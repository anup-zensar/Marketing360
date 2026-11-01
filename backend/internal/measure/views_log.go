package measure

import (
	"context"
	"encoding/json"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
)

const (
	labelDefaultView = "Default view"
	labelFilters     = "Filters"
)

func (s *Service) logView(ctx context.Context, a actor.Actor, v StoredView, action history.Action) error {
	_, err := s.log.Log(ctx, a, history.LogInput{Kind: KindReportView, RecordID: v.ID, RecordName: v.Name, Action: action})
	return err
}

func (s *Service) logChange(ctx context.Context, a actor.Actor, before, after StoredView) error {
	changes, err := history.Diff(before.logView(), after.logView())
	if err != nil {
		return err
	}
	_, err = s.log.Log(ctx, a, history.LogInput{Kind: KindReportView, RecordID: after.ID, RecordName: after.Name,
		Action: history.ActionChanged, Changes: changes})
	return err
}

// ViewKind tells the activity log about saved report views; they never go to the recycle bin (research R5).
func (s *Service) ViewKind() history.Kind {
	return history.Kind{
		Name: KindReportView, Singular: "saved report view", Plural: "saved report views",
		Snapshot: func(ctx context.Context, id string) (history.Snapshot, bool, error) {
			v, found, err := s.deps.Views.GetView(ctx, id)
			if err != nil || !found {
				return history.Snapshot{}, false, err
			}
			data, err := json.Marshal(v)
			if err != nil {
				return history.Snapshot{}, false, err
			}
			return history.Snapshot{ID: v.ID, Name: v.Name, Data: data}, true, nil
		},
		Remove: func(ctx context.Context, id string) error { return s.deps.Views.DeleteView(ctx, id) },
		Put: func(ctx context.Context, snapshot history.Snapshot) error {
			var v StoredView
			if err := json.Unmarshal(snapshot.Data, &v); err != nil {
				return err
			}
			return s.deps.Views.PutView(ctx, v)
		},
		Describe: describeView,
	}
}

func describeView(_ context.Context, field string, value any) (string, string, error) {
	switch field {
	case "isDefault":
		if starred, _ := value.(bool); starred {
			return labelDefaultView, "Yes", nil
		}
		return labelDefaultView, "No", nil
	case "query":
		return labelFilters, "", nil
	}
	return "", "", nil
}
