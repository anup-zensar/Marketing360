package performance_test

import (
	"errors"
	"strings"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/errs"
)

func TestUpdateSource_AStaleSaveIsRefusedNamingWhoMovedPast_FR045(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	src := f.createSource(t, "Eloqua")

	fresh := performance.SourceInput{Name: "Eloqua EU", OwnerID: alex.ID, Category: src.Category,
		DeliveryRoute: src.DeliveryRoute, Cadence: src.Cadence, Version: src.Version}
	if _, err := f.service.UpdateSource(t.Context(), alex, src.ID, fresh); err != nil {
		t.Fatalf("UpdateSource: %v", err)
	}

	stale := fresh
	_, err := f.service.UpdateSource(t.Context(), alex, src.ID, stale)
	var refusal *errs.Error
	if !errors.As(err, &refusal) || refusal.Code != errs.CodeVersionConflict {
		t.Fatalf("a save against the old version = %v, want version_conflict", err)
	}
}

func TestCreateMapping_ASecondActiveSupplierIsRefusedNamingTheFirst_FR003(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	first := f.createSource(t, "Eloqua")
	second := f.createSource(t, "Marketo")
	stage := performance.MappingInput{ColumnName: "INQs", Meaning: performance.MeaningStage,
		Stage: "inq", ModelID: "MOD-1"}
	if _, err := f.service.CreateMapping(t.Context(), alex, first.ID, stage); err != nil {
		t.Fatalf("CreateMapping(first): %v", err)
	}

	_, err := f.service.CreateMapping(t.Context(), alex, second.ID, stage)
	if err == nil || !strings.Contains(err.Error(), "Eloqua") {
		t.Fatalf("a second active supplier = %v, want it refused naming Eloqua", err)
	}
}

func TestDeleteSource_RefusedWhileItHoldsFactsAndDeactivateOffered_FR002(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)
	f.runImport(t, src.ID, map[string]string{"r1": "5"})

	err := f.service.DeleteSource(t.Context(), alex, src.ID)
	var refusal *errs.Error
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Message, "Deactivate") {
		t.Fatalf("deleting a source with facts = %v, want the refusal offering Deactivate", err)
	}
}

func TestDeactivate_KeepsFactsAndRefusesFurtherLoads_FR004(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)
	f.runImport(t, src.ID, map[string]string{"r1": "5"})

	if err := f.service.SetSourceStatus(t.Context(), alex, src.ID, performance.StatusInactive); err != nil {
		t.Fatalf("SetSourceStatus: %v", err)
	}
	if facts := f.currentFacts(t, src.ID, "r1"); len(facts) != 1 {
		t.Fatalf("facts after deactivation = %d, want the data kept", len(facts))
	}
	_, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "late.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: importRows(map[string]string{"r2": "3"}),
	})
	if err == nil {
		t.Fatal("a deactivated source accepted a load, want it refused")
	}

	if err := f.service.SetSourceStatus(t.Context(), alex, src.ID, performance.StatusActive); err != nil {
		t.Fatalf("reactivate: %v", err)
	}
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "again.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: importRows(map[string]string{"r2": "3"}),
	}); err != nil {
		t.Fatalf("a reactivated source refused a load: %v", err)
	}
}

func TestTemplate_CarriesExactlyTheMappedColumns_FR008(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	src := f.createSource(t, "Eloqua")
	for _, in := range []performance.MappingInput{
		{ColumnName: "Row ref", Meaning: performance.MeaningReference},
		{ColumnName: "Date", Meaning: performance.MeaningFactDate},
		{ColumnName: "INQs, net", Meaning: performance.MeaningStage, Stage: "inq", ModelID: "MOD-1"},
	} {
		if _, err := f.service.CreateMapping(t.Context(), alex, src.ID, in); err != nil {
			t.Fatalf("CreateMapping(%s): %v", in.ColumnName, err)
		}
	}

	file, err := f.service.Template(t.Context(), alex, src.ID)
	if err != nil {
		t.Fatalf("Template: %v", err)
	}
	if file != "Row ref,Date,\"INQs, net\"\n" {
		t.Fatalf("template = %q, want exactly the mapped columns, RFC 4180 quoted", file)
	}
}
