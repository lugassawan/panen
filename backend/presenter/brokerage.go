package presenter

import (
	"time"

	"github.com/lugassawan/panen/backend/internal/domain/brokerage"
	"github.com/lugassawan/panen/backend/internal/domain/shared"
)

// CreateBrokerageAccount creates a new brokerage account for the current user.
func (a *App) CreateBrokerageAccount(name string, buyFee, sellFee float64) (*BrokerageAccountResponse, error) {
	now := time.Now().UTC()
	acct := &brokerage.Account{
		ID:         shared.NewID(),
		ProfileID:  a.profileID,
		BrokerName: name,
		BuyFeePct:  buyFee,
		SellFeePct: sellFee,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := a.brokerages.Create(a.ctx, acct); err != nil {
		return nil, err
	}
	return buildBrokerageResponse(acct), nil
}
