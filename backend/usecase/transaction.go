package usecase

import (
	"context"
	"fmt"

	"github.com/lugassawan/panen/backend/domain/transaction"
)

// TransactionService handles querying and summarizing transaction history.
type TransactionService struct {
	history transaction.Repository
}

// NewTransactionService creates a new TransactionService.
func NewTransactionService(history transaction.Repository) *TransactionService {
	return &TransactionService{history: history}
}

// ListTransactions validates the filter, then returns matching records and a summary.
func (s *TransactionService) ListTransactions(
	ctx context.Context, filter transaction.Filter,
) ([]transaction.Record, *transaction.Summary, error) {
	if filter.Type != "" {
		switch transaction.Type(filter.Type) {
		case transaction.TypeBuy, transaction.TypeSell, transaction.TypeDividend:
		default:
			return nil, nil, fmt.Errorf("invalid transaction type: %s", filter.Type)
		}
	}

	records, err := s.history.List(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	summary, err := s.history.Summarize(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	return records, summary, nil
}
