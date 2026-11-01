package simulatedfeed_test

import (
	"bytes"
	"testing"
	"time"

	"mapp/internal/performance"
	"mapp/internal/performance/simulatedfeed"
)

var day = time.Date(2026, 3, 12, 15, 30, 0, 0, time.UTC)

func feedSource() performance.Source {
	return performance.Source{ID: "PS-1001", Name: "Eloqua", Status: performance.StatusActive,
		DeliveryRoute: performance.RouteSimulated, IntegrationID: "eloqua"}
}

func feedMappings() []performance.Mapping {
	return []performance.Mapping{
		{ColumnName: "Ref", Meaning: performance.MeaningReference},
		{ColumnName: "Date", Meaning: performance.MeaningFactDate},
		{ColumnName: "Campaign", Meaning: performance.MeaningKey, KeyType: performance.MatchedIntegrationID},
		{ColumnName: "INQs", Meaning: performance.MeaningStage, Stage: "inq"},
		{ColumnName: "Revenue", Meaning: performance.MeaningRevenue},
		{ColumnName: "Currency", Meaning: performance.MeaningCurrency},
	}
}

func TestPull_IsDeterministicOnItsSeedAndDay_FR010(t *testing.T) {
	t.Parallel()
	feed := simulatedfeed.New(46)

	name1, content1, err1 := feed.Pull(feedSource(), feedMappings(), day)
	name2, content2, err2 := feed.Pull(feedSource(), feedMappings(), day.Add(3*time.Hour))

	if err1 != nil || err2 != nil {
		t.Fatalf("Pull: %v %v", err1, err2)
	}
	if name1 != name2 || !bytes.Equal(content1, content2) {
		t.Errorf("two pulls on one day differ (%s vs %s): a re-pull must be the identical file", name1, name2)
	}
}

func TestPull_RowsTakeTheIdenticalPathAsAnUpload_046C13(t *testing.T) {
	t.Parallel()
	feed := simulatedfeed.New(46)

	_, content, err := feed.Pull(feedSource(), feedMappings(), day)
	if err != nil {
		t.Fatalf("Pull: %v", err)
	}
	file, err := performance.ReadFile(content, feedMappings())
	if err != nil {
		t.Fatalf("ReadFile refused the feed's own file: %v", err)
	}
	results := performance.Classify(file.Rows, performance.Env{
		Today:      day,
		Currencies: map[string]bool{"USD": true},
		Resolved:   performance.Resolved{},
		Approved:   func(string) bool { return true },
	})
	summary := performance.Summarise(results)
	if !summary.Counts.Reconciles() || summary.Counts.Read != 14 {
		t.Fatalf("counts = %+v, want 14 rows read, reconciling", summary.Counts)
	}
	if summary.Counts.Rejected != 0 {
		for _, r := range results {
			if r.Outcome == performance.OutcomeRejected {
				t.Errorf("the feed produced a row its own pipeline rejects: %+v", r.Faults)
			}
		}
	}
}

func TestPull_EmitsOnlyItsOwnIdentifiers_046C23(t *testing.T) {
	t.Parallel()
	feed := simulatedfeed.New(46)

	_, content, err := feed.Pull(feedSource(), feedMappings(), day)
	if err != nil {
		t.Fatalf("Pull: %v", err)
	}
	file, err := performance.ReadFile(content, feedMappings())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	for _, row := range file.Rows {
		if row.Keys.IntegrationID == "" || row.Keys.ActivityID != "" {
			t.Fatalf("keys = %+v, want only the feed's own integration id: a feed cannot know MAPP's ids", row.Keys)
		}
	}
}
