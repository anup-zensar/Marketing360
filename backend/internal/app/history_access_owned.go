package app

import (
	"cmp"
	"context"
	"encoding/json"

	"mapp/internal/history"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/memo"
)

const (
	kindCustomDashboard = "customDashboard"
	kindDataset         = "dataset"
)

type ownerStore interface {
	OwnerOf(ctx context.Context, id string) (ownerID string, found bool, err error)
}

func isOwnedKind(kind string) bool {
	return kind == kindCustomDashboard || kind == kindDataset
}

// seesOwnedLog: a private record's log entries go to its owner and to the people who see the whole log (D-375).
func (h historyAccess) seesOwnedLog(ctx context.Context, r entryRecord) bool {
	g, restricted := authz.From(ctx)
	return !restricted || g.Has(authz.ActivityLogViewAll) || h.owns(ctx, g, r)
}

// seesOwnedBin: a private record's recycle-bin entry is its owner's alone, administrators included (D-375).
func (h historyAccess) seesOwnedBin(ctx context.Context, e history.BinEntry) bool {
	g, restricted := authz.From(ctx)
	return !restricted || h.owns(ctx, g, entryRecord{kind: e.RecordType, id: e.RecordID, snapshots: e.Snapshots})
}

func (h historyAccess) mayRestoreOwned(ctx context.Context, e history.BinEntry) error {
	if !h.seesOwnedBin(ctx, e) {
		return errs.NotFound(authz.NotFoundMessage(e.RecordTypeLabel))
	}
	return nil
}

func (h historyAccess) owns(ctx context.Context, g authz.Grants, r entryRecord) bool {
	owner, found := h.ownerOf(ctx, r)
	return found && owner == g.PersonID
}

// ownerOf reads the owner from the entry's snapshots, then the stored record, then the recycle bin; a record
// deleted permanently has no owner any more.
func (h historyAccess) ownerOf(ctx context.Context, r entryRecord) (string, bool) {
	if owner, found := snapshotOwner(r.kind, r.id, r.snapshots); found {
		return owner, true
	}
	if store := h.ownerStoreOf(r.kind); store != nil {
		if owner, found, err := store.OwnerOf(ctx, r.id); err == nil && found {
			return owner, true
		}
	}
	return snapshotOwner(r.kind, r.id, h.ownedInBin(ctx, r.kind))
}

func (h historyAccess) ownerStoreOf(kind string) ownerStore {
	if kind == kindDataset {
		return h.datasets
	}
	return h.dashboards
}

func (h historyAccess) ownedInBin(ctx context.Context, kind string) []history.Snapshot {
	if h.history == nil {
		return nil
	}
	inBin, err := memo.Get(ctx, memo.Derived+":historyOwned:"+kind, func(ctx context.Context) ([]history.Snapshot, error) {
		return h.history.SnapshotsInBin(ctx, kind)
	})
	if err != nil {
		return nil
	}
	return inBin
}

// snapshotOwner reads a dashboard's snapshot, the dashboard itself, or a dataset's, which wraps it in "dataset".
func snapshotOwner(kind, id string, snapshots []history.Snapshot) (string, bool) {
	for _, s := range snapshots {
		var stored struct {
			OwnerID string `json:"ownerId"`
			Dataset struct {
				OwnerID string `json:"ownerId"`
			} `json:"dataset"`
		}
		if s.Kind == kind && s.ID == id && json.Unmarshal(s.Data, &stored) == nil {
			owner := cmp.Or(stored.OwnerID, stored.Dataset.OwnerID)
			return owner, owner != ""
		}
	}
	return "", false
}
