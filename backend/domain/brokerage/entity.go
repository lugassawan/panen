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
	BrokerCode  string
	BuyFeePct   float64
	SellFeePct  float64
	SellTaxPct  float64
	IsManualFee bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewAccount creates a new Account with generated ID and timestamps.
func NewAccount(profileID, brokerName, brokerCode string, buyFee, sellFee, sellTax float64) *Account {
	now := time.Now().UTC()
	return &Account{
		ID:         shared.NewID(),
		ProfileID:  profileID,
		BrokerName: brokerName,
		BrokerCode: brokerCode,
		BuyFeePct:  buyFee,
		SellFeePct: sellFee,
		SellTaxPct: sellTax,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
