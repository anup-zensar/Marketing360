package gsaplink

import (
	"context"
	"time"

	"mapp/internal/platform/database"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

type PostgresStore struct {
	db *database.DB
}

func NewPostgresStore(db *database.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) ActivitiesApprovedAt(ctx context.Context, activityIDs []string) (map[string]time.Time, error) {
	rows, err := s.db.Queries(ctx).ListActivitiesApprovedAt(ctx, activityIDs)
	if err != nil {
		return nil, err
	}
	approved := make(map[string]time.Time, len(rows))
	for _, r := range rows {
		approved[r.ActivityID] = r.ApprovedAt
	}
	return approved, nil
}

// WaitingForGsap returns one Request per purchase request, its lines in their order.
func (s *PostgresStore) WaitingForGsap(ctx context.Context, lineItemIDs []string) ([]Request, error) {
	rows, err := s.db.Queries(ctx).ListPurchaseRequestsWaitingForGsap(ctx, lineItemIDs)
	if err != nil {
		return nil, err
	}
	requests := []Request{}
	for _, r := range rows {
		if len(requests) == 0 || requests[len(requests)-1].ID != r.ID {
			requests = append(requests, Request{ID: r.ID, Name: r.Name, Vendor: r.Vendor, Owner: r.OwnerName,
				PlanID: r.PlanID, GsapPrNumber: r.GsapPrNumber, Routing: r.Routing, RoutedAt: r.RoutedAt})
		}
		last := &requests[len(requests)-1]
		last.Lines = append(last.Lines, RequestLine{LineItemID: r.LineItemID,
			Amount: money.Money{Amount: money.FromCents(r.AmountMinor), Currency: r.Currency}})
	}
	return requests, nil
}

func (s *PostgresStore) OrderTransactions(ctx context.Context, lineItemIDs []string) ([]Transaction, error) {
	rows, err := s.db.Queries(ctx).ListPurchaseOrderTransactions(ctx, lineItemIDs)
	if err != nil {
		return nil, err
	}
	list := make([]Transaction, 0, len(rows))
	for _, r := range rows {
		list = append(list, Transaction{ID: r.ID, Type: r.Type, LineItemID: r.LineItemID, PoNumber: r.PoNumber,
			LineNumber: r.LineNumber, PostingDate: date.Today(r.PostingDate), Vendor: r.VendorName,
			Amount:            money.Money{Amount: money.FromCents(r.AmountMinor), Currency: r.Currency},
			PurchaseRequestID: r.PurchaseRequestID, GsapPrNumber: r.GsapPrNumber})
	}
	return list, nil
}
