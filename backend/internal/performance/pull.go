package performance

import (
	"context"
	"fmt"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

// PullFeed runs the simulated feed once, by explicit action (046-C13), and hands what it produced to the
// identical pipeline an upload takes: the service cannot tell the routes apart (ADR-0020).
func (s *Service) PullFeed(ctx context.Context, a actor.Actor, sourceID string) (Import, error) {
	if err := authz.Require(ctx, PermissionImport); err != nil {
		return Import{}, err
	}
	source, mappings, err := s.pullable(ctx, sourceID)
	if err != nil {
		return Import{}, err
	}
	fileName, content, err := s.feed.Pull(source, mappings, s.now())
	if err != nil {
		return Import{}, fmt.Errorf("pull the feed of %s: %w", sourceID, err)
	}
	file, err := ReadFile(content, mappings)
	if err != nil {
		return Import{}, fmt.Errorf("read the feed of %s: %w", sourceID, err)
	}
	return s.RunImport(ctx, a, ImportInput{
		SourceID: sourceID, FileName: fileName, Origin: OriginSimulated,
		Headings: file.Headings, Rows: file.Rows,
	})
}

func (s *Service) pullable(ctx context.Context, sourceID string) (Source, []Mapping, error) {
	if s.feed == nil {
		return Source{}, nil, errs.NotFound(msgSourceNotFound)
	}
	var source Source
	var mappings []Mapping
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		found, readErr := s.loadableSource(ctx, sourceID)
		if readErr != nil {
			return readErr
		}
		if !found.CanPull() {
			return errs.NotAllowed(fmt.Sprintf(msgNoPull, found.Name))
		}
		source = found
		mappings, readErr = s.repo.ListMappings(ctx, []string{sourceID})
		return readErr
	})
	if err != nil {
		return Source{}, nil, err
	}
	return source, mappings, nil
}
