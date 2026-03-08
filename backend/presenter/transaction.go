package presenter

import (
	"context"
	"fmt"
	"time"

	"github.com/lugassawan/panen/backend/domain/transaction"
	"github.com/lugassawan/panen/backend/usecase"
)

// TransactionHandler handles transaction history requests from the frontend.
type TransactionHandler struct {
	ctx  context.Context
	txns *usecase.TransactionService
}

// NewTransactionHandler creates a new TransactionHandler.
func NewTransactionHandler(ctx context.Context, txns *usecase.TransactionService) *TransactionHandler {
	h := &TransactionHandler{}
	h.Bind(ctx, txns)
	return h
}

func (h *TransactionHandler) Bind(ctx context.Context, txns *usecase.TransactionService) {
	h.ctx = ctx
	h.txns = txns
}

// ListTransactions returns filtered transaction records with a summary.
func (h *TransactionHandler) ListTransactions(
	portfolioID, ticker, txnType, dateFrom, dateTo, sortField string, sortAsc bool,
) (*TransactionListResponse, error) {
	if h.txns == nil {
		return &TransactionListResponse{}, nil
	}

	filter := transaction.Filter{
		PortfolioID: portfolioID,
		Ticker:      ticker,
		Type:        txnType,
		SortField:   sortField,
		SortAsc:     sortAsc,
	}

	if dateFrom != "" {
		t, err := time.Parse(dateLayout, dateFrom)
		if err != nil {
			return nil, fmt.Errorf("list transactions: %w", err)
		}
		filter.DateFrom = &t
	}

	if dateTo != "" {
		t, err := time.Parse(dateLayout, dateTo)
		if err != nil {
			return nil, fmt.Errorf("list transactions: %w", err)
		}
		filter.DateTo = &t
	}

	records, summary, err := h.txns.ListTransactions(h.ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}

	items := make([]TransactionRecordResponse, len(records))
	for i, r := range records {
		items[i] = newTransactionRecordResponse(r)
	}

	return &TransactionListResponse{
		Items:   items,
		Summary: newTransactionSummaryResponse(summary),
	}, nil
}
