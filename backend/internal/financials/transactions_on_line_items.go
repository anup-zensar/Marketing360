package financials

import (
	"context"
	"fmt"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/money"
)

type LinkedTransaction struct {
	Transaction
	USD *money.Amount // nil when its currency has no rate
}

// TransactionsOnLineItems needs finance.view; the caller chose the line items from what a may see (D-356).
func (s *Service) TransactionsOnLineItems(ctx context.Context, a actor.Actor, lineItemIDs []string) ([]LinkedTransaction, error) {
	var out []LinkedTransaction
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		if err := s.access.SeeFinance(ctx, a); err != nil {
			return err
		}
		stored, err := s.repo.TransactionsOnLineItems(ctx, lineItemIDs)
		if err != nil {
			return err
		}
		rates, err := s.rates.Rates(ctx)
		if err != nil {
			return err
		}
		out = make([]LinkedTransaction, 0, len(stored))
		for i := range stored {
			out = append(out, LinkedTransaction{Transaction: stored[i], USD: usdOf(&stored[i], rates)})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("transactions of line items for %s: %w", a.ID, err)
	}
	return out, nil
}
