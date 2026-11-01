package dataset

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

// BenchmarkDatasetUpload checks and saves a 10,000-row, 50-column file on the fakes; the spec allows 5 s and 10 s.
func BenchmarkDatasetUpload_SC005(b *testing.B) {
	var csv strings.Builder
	columns := make([]ColumnInput, 0, maxColumns)
	for i := range maxColumns {
		fmt.Fprintf(&csv, "Column %d,", i+1)
		columns = append(columns, ColumnInput{Position: i, Label: fmt.Sprintf("Column %d", i+1), Kind: KindNumber})
	}
	csv.WriteString("\n")
	row := strings.TrimSuffix(strings.Repeat("1234.5,", maxColumns), ",") + "\n"
	csv.WriteString(strings.Repeat(row, maxRows))
	content := base64.StdEncoding.EncodeToString([]byte(csv.String()))
	for b.Loop() {
		f := newFixture(b)
		if _, err := f.service.Preview(b.Context(), priya, Upload{FileName: "big.csv", Content: content}); err != nil {
			b.Fatalf("Preview: %v", err)
		}
		in := CreateInput{FileName: "big.csv", Content: content, Name: "Big", Columns: columns}
		if _, err := f.service.Create(b.Context(), priya, in); err != nil {
			b.Fatalf("Create: %v", err)
		}
	}
}
