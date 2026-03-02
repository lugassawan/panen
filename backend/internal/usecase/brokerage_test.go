package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/internal/domain/brokerage"
	"github.com/lugassawan/panen/backend/internal/domain/shared"
)

func TestBrokerageServiceCreateHappy(t *testing.T) {
	repo := newMockBrokerageRepo()
	svc := NewBrokerageService(repo)

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Ajaib",
		BuyFeePct: 0.15, SellFeePct: 0.25,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := svc.Create(context.Background(), acct); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), acct.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.BrokerName != "Ajaib" {
		t.Errorf("BrokerName = %q, want Ajaib", got.BrokerName)
	}
}

func TestBrokerageServiceCreateEmptyName(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo())

	acct := &brokerage.Account{ID: shared.NewID(), BrokerName: "  "}
	err := svc.Create(context.Background(), acct)
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("Create() error = %v, want ErrEmptyName", err)
	}
}

func TestBrokerageServiceCreateNegativeFee(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo())

	tests := []struct {
		name    string
		buyFee  float64
		sellFee float64
	}{
		{name: "negative buy fee", buyFee: -0.1, sellFee: 0},
		{name: "negative sell fee", buyFee: 0, sellFee: -0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acct := &brokerage.Account{
				ID: shared.NewID(), BrokerName: "Broker",
				BuyFeePct: tt.buyFee, SellFeePct: tt.sellFee,
			}
			err := svc.Create(context.Background(), acct)
			if !errors.Is(err, ErrInvalidFee) {
				t.Errorf("Create() error = %v, want ErrInvalidFee", err)
			}
		})
	}
}
