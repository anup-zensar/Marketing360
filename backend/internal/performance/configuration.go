package performance

import (
	"context"
	"fmt"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/record"
)

// ConfigurationInput is Settings › Performance data's one record (046-C12): which activity types require
// media codes, the attribute catalogue those codes carry, and the number generation continues from.
type ConfigurationInput struct {
	ActivityTypeIDs []string        `json:"activityTypeIds"`
	CodeAttributes  []CodeAttribute `json:"codeAttributes"`
	MediaCodeNext   int64           `json:"mediaCodeNext"`
	Version         int             `json:"version"`
}

func (s *Service) PerformanceConfiguration(ctx context.Context, _ actor.Actor) (Configuration, error) {
	var config Configuration
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		var readErr error
		config, readErr = s.configurationOrStart(ctx)
		return readErr
	})
	if err != nil {
		return Configuration{}, fmt.Errorf("read the performance configuration: %w", err)
	}
	return config, nil
}

// configurationOrStart answers the stored row, or the defaults a first save starts from: generation begins
// above Hive9's numbers so a new code cannot collide (FR-042).
func (s *Service) configurationOrStart(ctx context.Context) (Configuration, error) {
	config, found, err := s.repo.GetConfiguration(ctx)
	if err != nil {
		return Configuration{}, err
	}
	if !found {
		config = Configuration{MediaCodeNext: firstMediaCode, CodeAttributes: []CodeAttribute{},
			ActivityTypeIDs: []string{}}
	}
	if config.ActivityTypeIDs == nil {
		config.ActivityTypeIDs = []string{}
	}
	if config.CodeAttributes == nil {
		config.CodeAttributes = []CodeAttribute{}
	}
	return config, nil
}

func (s *Service) SaveConfiguration(ctx context.Context, a actor.Actor, in ConfigurationInput) (Configuration, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return Configuration{}, err
	}
	var saved Configuration
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		stored, writeErr := s.configurationOrStart(ctx)
		if writeErr != nil {
			return writeErr
		}
		if stored.Version != in.Version {
			return errs.VersionConflict(configurationKind, stored.UpdatedBy.Name, stored.UpdatedAt)
		}
		if in.MediaCodeNext < stored.MediaCodeNext {
			return errs.Validation(errs.FieldError{Field: "mediaCodeNext",
				Message: fmt.Sprintf(msgCodeNumberBack, stored.MediaCodeNext)})
		}
		saved = stored
		saved.ActivityTypeIDs, saved.CodeAttributes = in.ActivityTypeIDs, in.CodeAttributes
		saved.MediaCodeNext = in.MediaCodeNext
		if stored.Version == 0 {
			saved.Meta = record.NewMeta(a, s.now())
		} else {
			saved.Version++
			saved.UpdatedAt, saved.UpdatedBy = s.now(), a
		}
		if putErr := s.repo.PutConfiguration(ctx, saved); putErr != nil {
			return putErr
		}
		_, logErr := s.history.Log(ctx, a, history.LogInput{
			Kind: configurationKind, RecordID: "performance", RecordName: "Performance data",
			Action: history.ActionChanged,
		})
		return logErr
	})
	if err != nil {
		return Configuration{}, fmt.Errorf("save the performance configuration: %w", err)
	}
	return saved, nil
}
