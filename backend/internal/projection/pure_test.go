package projection_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPackage_StaysPure_FR015 guards what 048, 049 and 050 rely on: the engine is arithmetic, callable
// without a database, a request or a clock of its own.
func TestPackage_StaysPure_FR015(t *testing.T) {
	t.Parallel()
	forbidden := []string{"gin-gonic", "pgx", "database/sql", "net/http", "mapp/internal/platform/clock",
		"mapp/internal/platform/database"}

	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list the package: %v", err)
	}
	for _, name := range sources {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		assertNoForbiddenImport(t, name, forbidden)
	}
}

func assertNoForbiddenImport(t *testing.T, name string, forbidden []string) {
	t.Helper()
	source, err := os.ReadFile(name) //nolint:gosec // G304: name comes from a glob of this package's own directory
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), name, source, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	for _, imported := range file.Imports {
		for _, banned := range forbidden {
			if strings.Contains(imported.Path.Value, banned) {
				t.Errorf("%s imports %s, want the engine to stay pure (FR-015)", name, imported.Path.Value)
			}
		}
	}
}
