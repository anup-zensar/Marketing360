// Package simulatedfeed produces performance rows on demand, standing in for the live connectors the
// constitution puts out of scope (ADR-0020, D-370).
package simulatedfeed

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mapp/internal/performance"
)

const (
	daysPerPull = 14
	keysInFeed  = 12
)

// Feed writes one pull deterministically from its seed and the day: a re-pull is the identical file.
type Feed struct {
	seed int64
}

func New(seed int64) *Feed {
	return &Feed{seed: seed}
}

func (f *Feed) Pull(source performance.Source, mappings []performance.Mapping, today time.Time) (string, []byte, error) {
	day := today.UTC().Truncate(24 * time.Hour)
	var out bytes.Buffer
	writer := csv.NewWriter(&out)
	headings := headingsOf(mappings)
	if err := writer.Write(headings); err != nil {
		return "", nil, fmt.Errorf("write the feed of %s: %w", source.ID, err)
	}
	for i := 0; i < daysPerPull; i++ {
		date := day.AddDate(0, 0, -(i + 1))
		if err := writer.Write(f.row(source, mappings, date, i)); err != nil {
			return "", nil, fmt.Errorf("write the feed of %s: %w", source.ID, err)
		}
	}
	writer.Flush()
	name := fmt.Sprintf("simulated-%s-%s.csv", strings.ToLower(source.ID), day.Format(time.DateOnly))
	return name, out.Bytes(), writer.Error()
}

func headingsOf(mappings []performance.Mapping) []string {
	headings := make([]string, 0, len(mappings))
	for _, m := range mappings {
		headings = append(headings, m.ColumnName)
	}
	return headings
}

func (f *Feed) row(source performance.Source, mappings []performance.Mapping, date time.Time, i int) []string {
	cells := make([]string, 0, len(mappings))
	for c, m := range mappings {
		cells = append(cells, f.cell(source, m, date, i*7+c))
	}
	return cells
}

func (f *Feed) cell(source performance.Source, m performance.Mapping, date time.Time, n int) string {
	switch m.Meaning {
	case performance.MeaningReference:
		return fmt.Sprintf("SIM-%s-%s", source.ID, date.Format(time.DateOnly))
	case performance.MeaningFactDate:
		return date.Format(time.DateOnly)
	case performance.MeaningKey:
		return f.key(source, m.KeyType, n)
	case performance.MeaningStage, performance.MeaningWebMetric:
		return strconv.FormatInt(3+(f.seed+int64(n)*31)%97, 10)
	case performance.MeaningRevenue:
		return strconv.FormatInt(250+(f.seed+int64(n)*53)%4750, 10) + ".00"
	case performance.MeaningCurrency:
		return "USD"
	case performance.MeaningJourney:
	}
	return ""
}

// key emits only the feed's own identifiers, as a real connector would; matching works once someone
// registers them on activities (046 US4).
func (f *Feed) key(source performance.Source, on performance.MatchedOn, n int) string {
	if on != performance.MatchedIntegrationID {
		return ""
	}
	return fmt.Sprintf("%s-campaign-%02d", strings.ToLower(source.IntegrationID+source.ID), (f.seed+int64(n))%keysInFeed)
}
