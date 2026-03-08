package presenter

import (
	"context"
	"fmt"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/usecase"
)

// BrokerageHandler handles brokerage account requests.
type BrokerageHandler struct {
	ctx        context.Context
	profileID  string
	brokerages *usecase.BrokerageService
}

// NewBrokerageHandler creates a new BrokerageHandler.
func NewBrokerageHandler(
	ctx context.Context, profileID string, brokerages *usecase.BrokerageService,
) *BrokerageHandler {
	h := &BrokerageHandler{}
	h.Bind(ctx, profileID, brokerages)
	return h
}

func (h *BrokerageHandler) Bind(ctx context.Context, profileID string, brokerages *usecase.BrokerageService) {
	h.ctx = ctx
	h.profileID = profileID
	h.brokerages = brokerages
}

// ListBrokerageAccounts returns all brokerage accounts for the current user.
func (h *BrokerageHandler) ListBrokerageAccounts() ([]*BrokerageAccountResponse, error) {
	accounts, err := h.brokerages.ListByProfileID(h.ctx, h.profileID)
	if err != nil {
		return nil, fmt.Errorf("list brokerage accounts: %w", err)
	}
	result := make([]*BrokerageAccountResponse, len(accounts))
	for i, a := range accounts {
		result[i] = newBrokerageAccountResponse(a)
	}
	return result, nil
}

// CreateBrokerageAccount creates a new brokerage account for the current user.
func (h *BrokerageHandler) CreateBrokerageAccount(
	name, brokerCode string, buyFee, sellFee, sellTax float64, isManualFee bool,
) (*BrokerageAccountResponse, error) {
	acct := brokerage.NewAccount(h.profileID, name, brokerCode, buyFee, sellFee, sellTax)
	acct.IsManualFee = isManualFee
	if err := h.brokerages.Create(h.ctx, acct); err != nil {
		return nil, fmt.Errorf("create brokerage account: %w", err)
	}
	return newBrokerageAccountResponse(acct), nil
}

// GetBrokerageAccount returns a single brokerage account by ID.
func (h *BrokerageHandler) GetBrokerageAccount(id string) (*BrokerageAccountResponse, error) {
	acct, err := h.brokerages.GetByID(h.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get brokerage account: %w", err)
	}
	return newBrokerageAccountResponse(acct), nil
}

// UpdateBrokerageAccount updates an existing brokerage account.
func (h *BrokerageHandler) UpdateBrokerageAccount(
	id, name, brokerCode string, buyFee, sellFee, sellTax float64, isManualFee bool,
) (*BrokerageAccountResponse, error) {
	acct, err := h.brokerages.GetByID(h.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("update brokerage account: %w", err)
	}
	acct.BrokerName = name
	acct.BrokerCode = brokerCode
	acct.BuyFeePct = buyFee
	acct.SellFeePct = sellFee
	acct.SellTaxPct = sellTax
	acct.IsManualFee = isManualFee
	if err := h.brokerages.Update(h.ctx, acct); err != nil {
		return nil, fmt.Errorf("update brokerage account: %w", err)
	}
	return newBrokerageAccountResponse(acct), nil
}

// DeleteBrokerageAccount removes a brokerage account by ID.
func (h *BrokerageHandler) DeleteBrokerageAccount(id string) error {
	if err := h.brokerages.Delete(h.ctx, id); err != nil {
		return toAppError(fmt.Errorf("delete brokerage account: %w", err))
	}
	return nil
}
