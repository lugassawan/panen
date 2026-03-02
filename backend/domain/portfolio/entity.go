package portfolio

import "time"

// Portfolio represents an investment portfolio within a brokerage account.
type Portfolio struct {
	ID                 string
	BrokerageAccountID string
	Name               string
	Mode               Mode
	RiskProfile        RiskProfile
	Capital            float64
	MonthlyAddition    float64
	MaxStocks          int
	Universe           []string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Holding represents a stock position within a portfolio.
type Holding struct {
	ID          string
	PortfolioID string
	Ticker      string
	AvgBuyPrice float64
	Lots        int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BuyTransaction represents an immutable record of a stock purchase.
type BuyTransaction struct {
	ID        string
	HoldingID string
	Date      time.Time
	Price     float64
	Lots      int
	Fee       float64
	CreatedAt time.Time
}
