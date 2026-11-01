package performance

import (
	"context"
	"fmt"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

// UndoImport takes a whole load back; it restates figures, so settings.manage joins performance.import.
func (s *Service) UndoImport(ctx context.Context, a actor.Actor, importID string) (Import, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return Import{}, err
	}
	if err := authz.Require(ctx, authz.SettingsManage); err != nil {
		return Import{}, err
	}
	var undone Import
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		var undoErr error
		undone, undoErr = s.undoImport(ctx, a, importID)
		return undoErr
	})
	if err != nil {
		return Import{}, fmt.Errorf("undo import %s: %w", importID, err)
	}
	return undone, nil
}

func (s *Service) undoImport(ctx context.Context, a actor.Actor, importID string) (Import, error) {
	run, err := s.importByID(ctx, importID)
	if err != nil {
		return Import{}, err
	}
	if run.Status != ImportCompleted {
		return Import{}, errs.NotAllowed(fmt.Sprintf(msgNotUndoable, run.FileName))
	}
	facts, err := s.repo.ListFactsOfImport(ctx, importID)
	if err != nil {
		return Import{}, err
	}
	// Deleting before restoring keeps the identity index satisfied at every moment.
	replaced, err := s.repo.FactsReplacedBy(ctx, factIDs(facts))
	if err != nil {
		return Import{}, err
	}
	if deleteErr := s.repo.DeleteFacts(ctx, factIDs(facts)); deleteErr != nil {
		return Import{}, deleteErr
	}
	if restoreErr := s.restoreReplaced(ctx, a, facts, replaced); restoreErr != nil {
		return Import{}, restoreErr
	}
	if closeErr := s.closeRowsOfUndoneImport(ctx, a, importID); closeErr != nil {
		return Import{}, closeErr
	}
	run.Status = ImportUndone
	run.Version++
	run.UpdatedAt, run.UpdatedBy = s.now(), a
	if putErr := s.repo.PutImport(ctx, run); putErr != nil {
		return Import{}, putErr
	}
	_, err = s.history.Log(ctx, a, history.LogInput{
		Kind: importKind, RecordID: run.ID, RecordName: run.FileName,
		Action: history.ActionStatusChanged, Verb: undoneVerb,
	})
	return run, err
}

// restoreReplaced brings back what this import superseded; a version whose replacement was itself replaced
// later is repointed at the later fact, never resurrected: two current facts would double a figure.
func (s *Service) restoreReplaced(ctx context.Context, a actor.Actor, facts, replaced []Fact) error {
	replacerOf := make(map[string]string, len(facts))
	for _, f := range facts {
		replacerOf[f.ID] = f.ReplacedByID
	}
	for _, old := range replaced {
		old.ReplacedByID = replacerOf[old.ReplacedByID]
		old.Version++
		old.UpdatedAt, old.UpdatedBy = s.now(), a
		if err := s.repo.PutFact(ctx, old); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) importByID(ctx context.Context, importID string) (Import, error) {
	found, err := s.repo.ListImports(ctx, importID, "", 1)
	if err != nil {
		return Import{}, err
	}
	if len(found) == 0 {
		return Import{}, errs.NotFound(msgImportNotFound)
	}
	return found[0], nil
}

func factIDs(facts []Fact) []string {
	ids := make([]string, 0, len(facts))
	for _, f := range facts {
		ids = append(ids, f.ID)
	}
	return ids
}

// closeRowsOfUndoneImport resolves the import's waiting rows with it: they describe a load that no longer
// stands, and a re-check of one would hang new facts on an undone upload.
func (s *Service) closeRowsOfUndoneImport(ctx context.Context, a actor.Actor, importID string) error {
	rows, err := s.repo.ListReviewRows(ctx, ReviewFilter{StatusNot: string(ReviewResolved), ImportID: importID, Limit: maxRows})
	if err != nil {
		return err
	}
	for _, row := range rows {
		if resolveErr := s.resolveRow(ctx, a, row.ReviewRow, undoneVerb); resolveErr != nil {
			return resolveErr
		}
	}
	return nil
}
