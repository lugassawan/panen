package portfolio

import (
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
)

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

// NewPortfolio creates a new Portfolio with generated ID, empty universe, and timestamps.
func NewPortfolio(
	brokerageAcctID, name string,
	mode Mode,
	riskProfile RiskProfile,
	capital, monthlyAddition float64,
	maxStocks int,
) *Portfolio {
	now := time.Now().UTC()
	return &Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: brokerageAcctID,
		Name:               name,
		Mode:               mode,
		RiskProfile:        riskProfile,
		Capital:            capital,
		MonthlyAddition:    monthlyAddition,
		MaxStocks:          maxStocks,
		Universe:           []string{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
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

// NewHolding creates a new Holding with generated ID and timestamps.
func NewHolding(portfolioID, ticker string, avgBuyPrice float64, lots int) *Holding {
	now := time.Now().UTC()
	return &Holding{
		ID:          shared.NewID(),
		PortfolioID: portfolioID,
		Ticker:      ticker,
		AvgBuyPrice: avgBuyPrice,
		Lots:        lots,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
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

// NewBuyTransaction creates a new BuyTransaction with generated ID and timestamp.
func NewBuyTransaction(holdingID string, date time.Time, price float64, lots int, fee float64) *BuyTransaction {
	return &BuyTransaction{
		ID:        shared.NewID(),
		HoldingID: holdingID,
		Date:      date,
		Price:     price,
		Lots:      lots,
		Fee:       fee,
		CreatedAt: time.Now().UTC(),
	}
}
