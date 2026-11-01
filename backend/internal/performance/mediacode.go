package performance

import (
	"context"
	"fmt"
	"strings"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

// MediaCodeArea is the activity's Media codes section: present only where the activity's type requires
// codes (FR-040), and the API says so rather than leaving the web app to work it out.
type MediaCodeArea struct {
	RequiresCodes bool        `json:"requiresCodes"`
	Items         []MediaCode `json:"items"`
}

type MediaCodeInput struct {
	Attributes  map[string]string `json:"attributes"`
	LandingPage string            `json:"landingPage"`
	Version     int               `json:"version"`
}

func (s *Service) MediaCodes(ctx context.Context, _ actor.Actor, activityID string) (MediaCodeArea, error) {
	var area MediaCodeArea
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		if _, readErr := s.visibleActivity(ctx, activityID); readErr != nil {
			return readErr
		}
		var readErr error
		area, readErr = s.mediaCodeArea(ctx, activityID)
		return readErr
	})
	if err != nil {
		return MediaCodeArea{}, fmt.Errorf("list the media codes of %s: %w", activityID, err)
	}
	return area, nil
}

func (s *Service) mediaCodeArea(ctx context.Context, activityID string) (MediaCodeArea, error) {
	required, _, err := s.requiresCodes(ctx, activityID)
	if err != nil {
		return MediaCodeArea{}, err
	}
	codes, err := s.repo.ListMediaCodes(ctx, authz.ScopeOf(ctx), activityID)
	if err != nil {
		return MediaCodeArea{}, err
	}
	if codes == nil {
		codes = []MediaCode{}
	}
	return MediaCodeArea{RequiresCodes: required, Items: codes}, nil
}

func (s *Service) requiresCodes(ctx context.Context, activityID string) (bool, Configuration, error) {
	config, err := s.configurationOrStart(ctx)
	if err != nil {
		return false, Configuration{}, err
	}
	if s.figures == nil {
		return false, config, nil
	}
	goals, found, err := s.figures.GoalFacts(ctx, activityID)
	if err != nil || !found {
		return false, config, err
	}
	return config.RequiresCodes(goals.ActivityTypeID), config, nil
}

// GenerateMediaCode works at any activity status (FR-041): a code is prepared before launch, long before
// approval, and the missing required attributes are reported together.
func (s *Service) GenerateMediaCode(ctx context.Context, a actor.Actor, activityID string, in MediaCodeInput) (MediaCode, error) {
	if err := authz.Require(ctx, PermissionLink); err != nil {
		return MediaCode{}, err
	}
	var created MediaCode
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		var writeErr error
		created, writeErr = s.generate(ctx, a, activityID, in)
		return writeErr
	})
	if err != nil {
		return MediaCode{}, fmt.Errorf("generate a media code on %s: %w", activityID, err)
	}
	return created, nil
}

func (s *Service) generate(ctx context.Context, a actor.Actor, activityID string, in MediaCodeInput) (MediaCode, error) {
	if _, err := s.visibleActivity(ctx, activityID); err != nil {
		return MediaCode{}, err
	}
	required, config, err := s.requiresCodes(ctx, activityID)
	if err != nil {
		return MediaCode{}, err
	}
	if !required {
		return MediaCode{}, errs.NotAllowed(msgNoCodesHere)
	}
	if problems := attributeProblems(config, in.Attributes); len(problems) > 0 {
		return MediaCode{}, errs.Validation(problems...)
	}
	id, err := s.ids.Next(ctx, ids.PrefixMediaCode)
	if err != nil {
		return MediaCode{}, err
	}
	number, takeErr := s.repo.TakeNextCode(ctx, a, s.now())
	if takeErr != nil {
		return MediaCode{}, takeErr
	}
	created := MediaCode{ID: id, Code: number, ActivityID: activityID, Attributes: attributesOf(in),
		LandingPage: strings.TrimSpace(in.LandingPage), Status: StatusActive, Meta: record.NewMeta(a, s.now())}
	if putErr := s.repo.PutMediaCode(ctx, created); putErr != nil {
		return MediaCode{}, putErr
	}
	_, err = s.history.Log(ctx, a, history.LogInput{
		Kind: mediaCodeKind, RecordID: created.ID, RecordName: codeText(created.Code),
		ParentID: activityID, Action: history.ActionCreated,
	})
	return created, err
}

// attributeProblems reports every missing required attribute at once (FR-041).
func attributeProblems(config Configuration, given map[string]string) []errs.FieldError {
	var problems []errs.FieldError
	for _, name := range config.MissingAttributes(given) {
		problems = append(problems, errs.FieldError{Field: "attributes." + name,
			Message: fmt.Sprintf(msgAttributeNeeded, name)})
	}
	return problems
}

func attributesOf(in MediaCodeInput) map[string]string {
	if in.Attributes == nil {
		return map[string]string{}
	}
	return in.Attributes
}
