package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
)

// BrokerageService handles brokerage account operations.
type BrokerageService struct {
	brokerages brokerage.Repository
	portfolios portfolio.Repository
}

// NewBrokerageService creates a new BrokerageService.
func NewBrokerageService(brokerages brokerage.Repository, portfolios portfolio.Repository) *BrokerageService {
	return &BrokerageService{brokerages: brokerages, portfolios: portfolios}
}

// Create validates and persists a brokerage account.
func (s *BrokerageService) Create(ctx context.Context, a *brokerage.Account) error {
	if strings.TrimSpace(a.BrokerName) == "" {
		return ErrEmptyName
	}
	if a.BuyFeePct < 0 {
		return ErrInvalidFee
	}
	if a.SellFeePct < 0 {
		return ErrInvalidFee
	}
	if a.SellTaxPct < 0 {
		return ErrInvalidFee
	}
	return s.brokerages.Create(ctx, a)
}

// GetByID returns a single brokerage account by ID.
func (s *BrokerageService) GetByID(ctx context.Context, id string) (*brokerage.Account, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrEmptyID
	}
	return s.brokerages.GetByID(ctx, id)
}

// ListByProfileID returns all brokerage accounts for a profile.
func (s *BrokerageService) ListByProfileID(ctx context.Context, profileID string) ([]*brokerage.Account, error) {
	return s.brokerages.ListByProfileID(ctx, profileID)
}

// Update validates and persists changes to a brokerage account.
func (s *BrokerageService) Update(ctx context.Context, a *brokerage.Account) error {
	if strings.TrimSpace(a.BrokerName) == "" {
		return ErrEmptyName
	}
	if a.BuyFeePct < 0 {
		return ErrInvalidFee
	}
	if a.SellFeePct < 0 {
		return ErrInvalidFee
	}
	if a.SellTaxPct < 0 {
		return ErrInvalidFee
	}
	a.UpdatedAt = time.Now().UTC()
	return s.brokerages.Update(ctx, a)
}

// Delete removes a brokerage account if it has no linked portfolios.
func (s *BrokerageService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrEmptyID
	}
	linked, err := s.portfolios.ListByBrokerageAccountID(ctx, id)
	if err != nil {
		return err
	}
	if len(linked) > 0 {
		return fmt.Errorf("%w: %d portfolio(s) linked", ErrHasDependents, len(linked))
	}
	return s.brokerages.Delete(ctx, id)
}
