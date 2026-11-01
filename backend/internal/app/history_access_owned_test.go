package app

import (
	"context"
	"encoding/json"
	"testing"

	"mapp/internal/history"
	"mapp/internal/platform/authz"
)

type fakeOwners map[string]string

func (f fakeOwners) OwnerOf(_ context.Context, id string) (string, bool, error) {
	owner, found := f[id]
	return owner, found, nil
}

const (
	asker       = "U-1006"
	someoneElse = "U-1002"
)

func datasetSnapshot(t *testing.T, id, owner string) history.Snapshot {
	t.Helper()
	data, err := json.Marshal(map[string]any{"dataset": map[string]string{"id": id, "ownerId": owner}, "rows": []any{}})
	if err != nil {
		t.Fatal(err)
	}
	return history.Snapshot{Kind: kindDataset, ID: id, Data: data}
}

func TestHistoryAccess_OwnedLogEntries_FR009_FR045(t *testing.T) {
	t.Parallel()
	h := historyAccess{dashboards: fakeOwners{"DB-1001": asker, "DB-1002": someoneElse}, datasets: fakeOwners{"DS-1001": asker}}
	builder, auditor := withPermissions(authz.DashboardsBuild), withPermissions(authz.ActivityLogViewAll)

	for _, c := range []struct {
		kind, id         string
		builder, auditor bool
	}{
		{kindCustomDashboard, "DB-1001", true, true},
		{kindCustomDashboard, "DB-1002", false, true},
		{kindDataset, "DS-1001", true, true},
		{kindDataset, "DS-1009", false, true}, // deleted permanently: nobody owns it any more
	} {
		e := history.Entry{RecordType: c.kind, RecordID: c.id}
		if got := h.CanSeeEntry(builder, e); got != c.builder {
			t.Errorf("%s: the person asking sees its log = %v, want %v", c.id, got, c.builder)
		}
		if got := h.CanSeeEntry(auditor, e); got != c.auditor {
			t.Errorf("%s: a reader of the whole log sees it = %v, want %v", c.id, got, c.auditor)
		}
		if !h.CanSeeEntry(context.Background(), e) {
			t.Errorf("%s: an internal call does not see it", c.id)
		}
	}
}

func TestHistoryAccess_OwnedBinEntriesFromTheirSnapshot_FR010_FR045(t *testing.T) {
	t.Parallel()
	h := historyAccess{dashboards: fakeOwners{}, datasets: fakeOwners{}}
	mine := history.BinEntry{RecordType: kindDataset, RecordTypeLabel: "dataset", RecordID: "DS-1001",
		Snapshots: []history.Snapshot{datasetSnapshot(t, "DS-1001", asker)}}
	theirs := history.BinEntry{RecordType: kindDataset, RecordTypeLabel: "dataset", RecordID: "DS-1002",
		Snapshots: []history.Snapshot{datasetSnapshot(t, "DS-1002", someoneElse)}}
	everything := withPermissions(authz.ActivityLogViewAll, authz.RecycleBinPurge, authz.DashboardsBuild)

	if !h.CanSeeBinEntry(everything, mine) || h.CanRestore(everything, mine) != nil {
		t.Error("the owner does not see or restore their dataset's bin entry")
	}
	if h.CanSeeBinEntry(everything, theirs) {
		t.Error("someone with every log and bin permission sees another person's dataset in the bin")
	}
	if err := h.CanRestore(everything, theirs); err == nil {
		t.Error("someone else restored the owner's dataset")
	}
}
