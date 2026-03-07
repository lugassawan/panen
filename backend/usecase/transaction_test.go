package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/transaction"
)

type mockTransactionRepo struct {
	records []transaction.Record
	summary *transaction.Summary
}

func (m *mockTransactionRepo) List(_ context.Context, _ transaction.Filter) ([]transaction.Record, error) {
	return m.records, nil
}

func (m *mockTransactionRepo) Summarize(_ context.Context, _ transaction.Filter) (*transaction.Summary, error) {
	return m.summary, nil
}

func TestTransactionServiceListTransactions(t *testing.T) {
	now := time.Now().UTC()
	repo := &mockTransactionRepo{
		records: []transaction.Record{
			{ID: "1", Type: transaction.TypeBuy, Date: now, Ticker: "BBCA", Lots: 5, Price: 8500, Total: 4250000},
			{ID: "2", Type: transaction.TypeSell, Date: now, Ticker: "BBRI", Lots: 3, Price: 4500, Total: 1350000},
		},
		summary: &transaction.Summary{
			TotalBuyAmount:   4250000,
			TotalSellAmount:  1350000,
			TotalFees:        12000,
			TransactionCount: 2,
		},
	}
	svc := NewTransactionService(repo)
	ctx := context.Background()

	tests := []struct {
		name      string
		filter    transaction.Filter
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty filter returns all",
			filter:    transaction.Filter{},
			wantCount: 2,
		},
		{
			name:      "valid type BUY",
			filter:    transaction.Filter{Type: "BUY"},
			wantCount: 2,
		},
		{
			name:      "valid type SELL",
			filter:    transaction.Filter{Type: "SELL"},
			wantCount: 2,
		},
		{
			name:      "valid type DIVIDEND",
			filter:    transaction.Filter{Type: "DIVIDEND"},
			wantCount: 2,
		},
		{
			name:    "invalid type returns error",
			filter:  transaction.Filter{Type: "INVALID"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, summary, err := svc.ListTransactions(ctx, tt.filter)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(records) != tt.wantCount {
				t.Errorf("got %d records, want %d", len(records), tt.wantCount)
			}
			if summary == nil {
				t.Fatal("summary is nil")
			}
			if summary.TransactionCount != 2 {
				t.Errorf("TransactionCount = %d, want 2", summary.TransactionCount)
			}
		})
	}
}
