package usecase

import (
	"context"
	"strings"

	"github.com/lugassawan/panen/backend/domain/brokerage"
)

// BrokerageService handles brokerage account operations.
type BrokerageService struct {
	brokerages brokerage.Repository
}

// NewBrokerageService creates a new BrokerageService.
func NewBrokerageService(brokerages brokerage.Repository) *BrokerageService {
	return &BrokerageService{brokerages: brokerages}
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
	return s.brokerages.Create(ctx, a)
}
