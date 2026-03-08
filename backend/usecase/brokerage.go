package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/brokerconfig"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
)

// BrokerageService handles brokerage account operations.
type BrokerageService struct {
	brokerages brokerage.Repository
	portfolios portfolio.Repository
	emitter    EventEmitter
}

// NewBrokerageService creates a new BrokerageService.
func NewBrokerageService(
	brokerages brokerage.Repository,
	portfolios portfolio.Repository,
	emitter EventEmitter,
) *BrokerageService {
	return &BrokerageService{brokerages: brokerages, portfolios: portfolios, emitter: emitter}
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

// SyncFeesFromConfig updates fees on non-manual accounts whose BrokerCode
// matches a config entry. Returns the number of accounts updated.
func (s *BrokerageService) SyncFeesFromConfig(
	ctx context.Context,
	profileID string,
	configs []*brokerconfig.BrokerConfig,
) (int, error) {
	accounts, err := s.brokerages.ListNonManualByProfileID(ctx, profileID)
	if err != nil {
		return 0, err
	}
	configMap := make(map[string]*brokerconfig.BrokerConfig, len(configs))
	for _, c := range configs {
		configMap[c.Code] = c
	}

	var count int
	for _, a := range accounts {
		c, ok := configMap[a.BrokerCode]
		if !ok {
			continue
		}
		if a.BuyFeePct == c.BuyFeePct && a.SellFeePct == c.SellFeePct && a.SellTaxPct == c.SellTaxPct {
			continue
		}
		a.BuyFeePct = c.BuyFeePct
		a.SellFeePct = c.SellFeePct
		a.SellTaxPct = c.SellTaxPct
		a.UpdatedAt = time.Now().UTC()
		if err := s.brokerages.Update(ctx, a); err != nil {
			return count, err
		}
		count++
	}
	if count > 0 && s.emitter != nil {
		s.emitter.Emit(shared.EventBrokerFeesSynced, map[string]any{"count": count})
	}
	return count, nil
}

// Delete removes a brokerage account if it has no linked portfolios.
func (s *BrokerageService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrEmptyID
	}
	linked, err := s.portfolios.ListByBrokerageAccountID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete brokerage: %w", err)
	}
	if len(linked) > 0 {
		return fmt.Errorf("%w: %d portfolio(s) linked", ErrHasDependents, len(linked))
	}
	if err := s.brokerages.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete brokerage: %w", err)
	}
	return nil
}
