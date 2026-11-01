package performance_test

import (
	"context"
	"errors"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/errs"
)

type codeFigures struct {
	typeID string
}

func (f codeFigures) GoalFacts(context.Context, string) (performance.GoalFacts, bool, error) {
	return performance.GoalFacts{ActivityTypeID: f.typeID, PlanYear: 2026}, true, nil
}

func codesFixture(t *testing.T) *fixture {
	t.Helper()
	f := newFixture(t)
	f.store.Activities = map[string]performance.KeyFacts{"A-1": {}}
	f.service.UseFigures(codeFigures{typeID: "AT-1"})
	if _, err := f.service.SaveConfiguration(t.Context(), alex, performance.ConfigurationInput{
		ActivityTypeIDs: []string{"AT-1"},
		CodeAttributes: []performance.CodeAttribute{
			{Name: "Channel", Required: true}, {Name: "Market", Required: true}, {Name: "Note", Required: false},
		},
		MediaCodeNext: 9000001,
	}); err != nil {
		t.Fatalf("SaveConfiguration: %v", err)
	}
	return f
}

func TestGenerateMediaCode_NumbersAreUniqueAndNeverReused_FR042(t *testing.T) {
	t.Parallel()
	f := codesFixture(t)
	given := map[string]string{"Channel": "paid-social", "Market": "DE"}

	first, err := f.service.GenerateMediaCode(t.Context(), alex, "A-1",
		performance.MediaCodeInput{Attributes: given})
	if err != nil || first.Code != 9000001 {
		t.Fatalf("first code = %+v (%v), want 9000001: generation starts above Hive9's numbers", first, err)
	}
	second, secondErr := f.service.GenerateMediaCode(t.Context(), alex, "A-1",
		performance.MediaCodeInput{Attributes: given})
	if secondErr != nil || second.Code != 9000002 {
		t.Fatalf("second code = %+v (%v), want 9000002", second, secondErr)
	}

	if deleteErr := f.service.DeleteMediaCode(t.Context(), alex, second.ID); deleteErr != nil {
		t.Fatalf("DeleteMediaCode: %v", deleteErr)
	}
	third, thirdErr := f.service.GenerateMediaCode(t.Context(), alex, "A-1",
		performance.MediaCodeInput{Attributes: given})
	if thirdErr != nil || third.Code != 9000003 {
		t.Fatalf("after a delete the next code = %+v (%v), want 9000003: numbers are never reused", third, thirdErr)
	}
}

func TestGenerateMediaCode_EveryMissingAttributeReportedAtOnce_FR041(t *testing.T) {
	t.Parallel()
	f := codesFixture(t)

	_, err := f.service.GenerateMediaCode(t.Context(), alex, "A-1", performance.MediaCodeInput{})

	var refusal *errs.Error
	if !errors.As(err, &refusal) || len(refusal.Fields) != 2 {
		t.Fatalf("GenerateMediaCode = %v, want Channel and Market reported together", err)
	}
}

func TestGenerateMediaCode_OnlyWhereTheTypeRequiresCodes_FR040(t *testing.T) {
	t.Parallel()
	f := codesFixture(t)
	f.service.UseFigures(codeFigures{typeID: "AT-other"})

	_, err := f.service.GenerateMediaCode(t.Context(), alex, "A-1",
		performance.MediaCodeInput{Attributes: map[string]string{"Channel": "x", "Market": "y"}})
	if err == nil {
		t.Fatal("a code was generated on a type that does not use codes, want a refusal")
	}
	area, areaErr := f.service.MediaCodes(t.Context(), alex, "A-1")
	if areaErr != nil || area.RequiresCodes {
		t.Errorf("area = %+v (%v), want requiresCodes false so the section stays hidden", area, areaErr)
	}
}

func TestDeleteMediaCode_RefusedOnceUsedAndDeactivateWorks_FR041(t *testing.T) {
	t.Parallel()
	f := codesFixture(t)
	given := map[string]string{"Channel": "paid-social", "Market": "DE"}
	code, err := f.service.GenerateMediaCode(t.Context(), alex, "A-1",
		performance.MediaCodeInput{Attributes: given})
	if err != nil {
		t.Fatalf("GenerateMediaCode: %v", err)
	}
	src := f.createSource(t, "Web analytics")
	f.store.Keys[performance.MatchedMediaCode] = map[string][]string{"9000001": {"A-1"}}
	f.service.UseApprovals(approvedStub{"A-1": true})
	rows := importRows(map[string]string{"r1": "5"})
	rows[0].Keys = performance.RowKeys{MediaCode: "9000001"}
	if _, runErr := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "web.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	}); runErr != nil {
		t.Fatalf("RunImport: %v", runErr)
	}

	err = f.service.DeleteMediaCode(t.Context(), alex, code.ID)
	var refusal *errs.Error
	if !errors.As(err, &refusal) || refusal.Code != errs.CodeNotAllowed {
		t.Fatalf("deleting a used code = %v, want the 409 offering Deactivate", err)
	}
	deactivated, err := f.service.DeactivateMediaCode(t.Context(), alex, code.ID)
	if err != nil || deactivated.Status != performance.StatusInactive {
		t.Fatalf("DeactivateMediaCode = %+v (%v), want it inactive", deactivated, err)
	}
}

func TestSaveConfiguration_TheStartNumberNeverMovesBack_FR042(t *testing.T) {
	t.Parallel()
	f := codesFixture(t)

	config, err := f.service.PerformanceConfiguration(t.Context(), alex)
	if err != nil {
		t.Fatalf("PerformanceConfiguration: %v", err)
	}
	_, err = f.service.SaveConfiguration(t.Context(), alex, performance.ConfigurationInput{
		ActivityTypeIDs: config.ActivityTypeIDs, CodeAttributes: config.CodeAttributes,
		MediaCodeNext: 100, Version: config.Version,
	})
	if err == nil {
		t.Fatal("the start number moved backwards, want it refused: codes are never reused")
	}
}

func TestConfiguration_LivesInThisFeaturesSettings_046C12(t *testing.T) {
	t.Parallel()
	f := codesFixture(t)

	config, err := f.service.PerformanceConfiguration(t.Context(), alex)
	if err != nil || len(config.ActivityTypeIDs) != 1 || len(config.CodeAttributes) != 3 {
		t.Fatalf("configuration = %+v (%v), want the types and catalogue stored here, not on the type", config, err)
	}
}
