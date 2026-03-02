package presenter

import (
	"context"

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
	return &BrokerageHandler{ctx: ctx, profileID: profileID, brokerages: brokerages}
}

// ListBrokerageAccounts returns all brokerage accounts for the current user.
func (h *BrokerageHandler) ListBrokerageAccounts() ([]*BrokerageAccountResponse, error) {
	accounts, err := h.brokerages.ListByProfileID(h.ctx, h.profileID)
	if err != nil {
		return nil, err
	}
	result := make([]*BrokerageAccountResponse, len(accounts))
	for i, a := range accounts {
		result[i] = newBrokerageAccountResponse(a)
	}
	return result, nil
}

// CreateBrokerageAccount creates a new brokerage account for the current user.
func (h *BrokerageHandler) CreateBrokerageAccount(
	name string, buyFee, sellFee float64,
) (*BrokerageAccountResponse, error) {
	acct := brokerage.NewAccount(h.profileID, name, buyFee, sellFee)
	if err := h.brokerages.Create(h.ctx, acct); err != nil {
		return nil, err
	}
	return newBrokerageAccountResponse(acct), nil
}
