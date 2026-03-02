package brokerage

import (
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// Account represents a securities brokerage account belonging to a user profile.
type Account struct {
	ID          string
	ProfileID   string
	BrokerName  string
	BuyFeePct   float64
	SellFeePct  float64
	IsManualFee bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewAccount creates a new Account with generated ID and timestamps.
func NewAccount(profileID, brokerName string, buyFee, sellFee float64) *Account {
	now := time.Now().UTC()
	return &Account{
		ID:         shared.NewID(),
		ProfileID:  profileID,
		BrokerName: brokerName,
		BuyFeePct:  buyFee,
		SellFeePct: sellFee,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
