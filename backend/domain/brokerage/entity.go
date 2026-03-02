package brokerage

import "time"

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
