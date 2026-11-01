package performance

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

type ImportInput struct {
	SourceID string
	FileName string
	Origin   Origin
	Headings []string
	Rows     []Row
}

// RunImport is one load: the whole-file gates, then every row classified and persisted in one transaction,
// so a load either lands completely or not at all (FR-009, FR-016).
func (s *Service) RunImport(ctx context.Context, a actor.Actor, in ImportInput) (Import, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return Import{}, err
	}
	if columns := PersonalDataIn(in.Headings); len(columns) > 0 {
		return Import{}, errs.NotAllowed(fmt.Sprintf(msgPersonalData, strings.Join(columns, ", ")))
	}
	var run Import
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		var runErr error
		run, runErr = s.runImport(ctx, a, in)
		return runErr
	})
	if err != nil {
		return Import{}, fmt.Errorf("run the import of %s: %w", in.FileName, err)
	}
	return run, nil
}

func (s *Service) runImport(ctx context.Context, a actor.Actor, in ImportInput) (Import, error) {
	_, err := s.loadableSource(ctx, in.SourceID)
	if err != nil {
		return Import{}, err
	}
	env, err := s.envOf(ctx, in)
	if err != nil {
		return Import{}, err
	}
	results := Classify(in.Rows, env)
	summary := Summarise(results)
	id, err := s.ids.Next(ctx, ids.PrefixPerformanceImport)
	if err != nil {
		return Import{}, err
	}
	run := Import{
		ID: id, SourceID: in.SourceID, FileName: in.FileName, Origin: in.Origin,
		Status: ImportCompleted, DataThrough: summary.DataThrough, Counts: summary.Counts,
		Meta: record.NewMeta(a, s.now()),
	}
	if persistErr := s.persist(ctx, a, run, results); persistErr != nil {
		return Import{}, persistErr
	}
	if putErr := s.repo.PutImport(ctx, run); putErr != nil {
		return Import{}, putErr
	}
	if putErr := s.repo.PutBreakdown(ctx, run.ID, summary.Breakdown); putErr != nil {
		return Import{}, putErr
	}
	_, err = s.history.Log(ctx, a, history.LogInput{
		Kind: importKind, RecordID: run.ID, RecordName: run.FileName, Action: history.ActionCreated,
	})
	return run, err
}

func (s *Service) loadableSource(ctx context.Context, sourceID string) (Source, error) {
	sources, err := s.repo.ListSources(ctx, sourceID)
	if err != nil {
		return Source{}, err
	}
	if len(sources) == 0 {
		return Source{}, errs.NotFound(msgSourceNotFound)
	}
	if sources[0].Status == StatusInactive {
		return Source{}, errs.NotAllowed(fmt.Sprintf(msgSourceInactive, sources[0].Name))
	}
	return sources[0].Source, nil
}

// envOf resolves everything a load needs before any row is decided: the known currencies, every key the
// file carries in one query per key type, and the approvals of every activity those keys reached.
func (s *Service) envOf(ctx context.Context, in ImportInput) (Env, error) {
	currencies, err := s.repo.CurrencyCodes(ctx)
	if err != nil {
		return Env{}, err
	}
	resolved, err := s.resolve(ctx, in.SourceID, in.Rows)
	if err != nil {
		return Env{}, err
	}
	approved, err := s.approvedOf(ctx, resolved)
	if err != nil {
		return Env{}, err
	}
	return Env{Today: s.now(), Currencies: currencies, Resolved: resolved, Approved: approved}, nil
}

func (s *Service) resolve(ctx context.Context, sourceID string, rows []Row) (Resolved, error) {
	scope := authz.PermissionScopeOf(ctx, PermissionLink)
	keys := distinctKeys(rows)
	resolved := Resolved{}
	for on, values := range keys {
		answered, err := s.matchOne(ctx, matchQuery{scope: scope, sourceID: sourceID, on: on, values: values})
		if err != nil {
			return nil, err
		}
		// The stores key their answers by the normalised key, so each raw value is mapped back to what
		// its own row carries and Decide can stay a lookup of the value as written.
		byRaw := make(map[string][]string, len(values))
		for _, raw := range values {
			byRaw[raw] = answered[normalKey(on, raw)]
		}
		resolved[on] = byRaw
	}
	return resolved, nil
}

type matchQuery struct {
	scope    authz.Scope
	sourceID string
	on       MatchedOn
	values   []string
}

func (s *Service) matchOne(ctx context.Context, q matchQuery) (map[string][]string, error) {
	switch q.on {
	case MatchedActivityID:
		return s.repo.MatchByActivityID(ctx, q.scope, normalKeys(q.on, q.values))
	case MatchedMediaCode:
		return s.repo.MatchByMediaCode(ctx, q.scope, mediaCodesOf(q.values))
	case MatchedIntegrationID:
		return s.repo.MatchByIntegrationID(ctx, q.scope, q.sourceID, q.values)
	case MatchedExternalName:
		return s.repo.MatchByExternalName(ctx, q.scope, q.values)
	case MatchedNone:
	}
	return nil, nil
}

func (s *Service) approvedOf(ctx context.Context, resolved Resolved) (func(string) bool, error) {
	if s.approvals == nil {
		return func(string) bool { return false }, nil
	}
	var reached []string
	seen := map[string]bool{}
	for _, byKey := range resolved {
		for _, activities := range byKey {
			for _, id := range activities {
				if !seen[id] {
					seen[id] = true
					reached = append(reached, id)
				}
			}
		}
	}
	approved, err := s.approvals.ApprovedIDs(ctx, reached)
	if err != nil {
		return nil, err
	}
	return func(id string) bool { return approved[id] }, nil
}

func distinctKeys(rows []Row) map[MatchedOn][]string {
	seen := map[MatchedOn]map[string]bool{}
	keys := map[MatchedOn][]string{}
	for _, row := range rows {
		for _, on := range MatchOrder() {
			value := row.Keys.valueOf(on)
			if value == "" {
				continue
			}
			if seen[on] == nil {
				seen[on] = map[string]bool{}
			}
			if !seen[on][value] {
				seen[on][value] = true
				keys[on] = append(keys[on], value)
			}
		}
	}
	return keys
}

func normalKey(on MatchedOn, raw string) string {
	if on == MatchedMediaCode {
		if code, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); err == nil {
			return strconv.FormatInt(code, 10)
		}
		return strings.TrimSpace(raw)
	}
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalKeys(on MatchedOn, values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, normalKey(on, v))
	}
	return out
}

func mediaCodesOf(values []string) []int64 {
	codes := make([]int64, 0, len(values))
	for _, v := range values {
		if code, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			codes = append(codes, code)
		}
	}
	return codes
}
