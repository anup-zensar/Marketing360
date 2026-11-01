package performance

import (
	"context"
	"fmt"
	"strings"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

// Template is the blank file a source's mappings define, in their own order: a header row and nothing else,
// so what a loader fills in is exactly what the mappings expect (FR-008).
func (s *Service) Template(ctx context.Context, _ actor.Actor, sourceID string) (string, error) {
	if err := authz.Require(ctx, authz.PlansView); err != nil {
		return "", err
	}
	var columns []string
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		sources, readErr := s.repo.ListSources(ctx, sourceID)
		if readErr != nil {
			return readErr
		}
		if len(sources) == 0 {
			return errs.NotFound(msgSourceNotFound)
		}
		found, readErr := s.repo.ListMappings(ctx, []string{sourceID})
		for _, m := range found {
			columns = append(columns, csvField(m.ColumnName))
		}
		return readErr
	})
	if err != nil {
		return "", fmt.Errorf("build the template of source %s: %w", sourceID, err)
	}
	return strings.Join(columns, ",") + "\n", nil
}

// csvField quotes a heading holding a comma, a quote or a newline, as RFC 4180 requires.
func csvField(name string) string {
	if !strings.ContainsAny(name, ",\"\n\r") {
		return name
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
