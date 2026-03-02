package presenter

import (
	"context"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/shared"
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

// CreateBrokerageAccount creates a new brokerage account for the current user.
func (h *BrokerageHandler) CreateBrokerageAccount(
	name string, buyFee, sellFee float64,
) (*BrokerageAccountResponse, error) {
	now := time.Now().UTC()
	acct := &brokerage.Account{
		ID:         shared.NewID(),
		ProfileID:  h.profileID,
		BrokerName: name,
		BuyFeePct:  buyFee,
		SellFeePct: sellFee,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := h.brokerages.Create(h.ctx, acct); err != nil {
		return nil, err
	}
	return buildBrokerageResponse(acct), nil
}
