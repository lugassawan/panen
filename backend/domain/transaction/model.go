package transaction

import "time"

// Type represents the kind of transaction.
type Type string

const (
	TypeBuy      Type = "BUY"
	TypeSell     Type = "SELL"
	TypeDividend Type = "DIVIDEND"
)

// Record is a read-only view of a transaction from any source
// (buy, sell, or dividend cash flow).
type Record struct {
	ID            string
	Type          Type
	Date          time.Time
	Ticker        string
	PortfolioID   string
	PortfolioName string
	Lots          int
	Price         float64
	Fee           float64
	Tax           float64
	Total         float64
	CreatedAt     time.Time
}

// Filter holds optional criteria for querying transaction records.
type Filter struct {
	PortfolioID string
	Ticker      string
	Type        string
	DateFrom    *time.Time
	DateTo      *time.Time
	SortField   string
	SortAsc     bool
}

// Summary holds aggregate totals across filtered transaction records.
type Summary struct {
	TotalBuyAmount      float64
	TotalSellAmount     float64
	TotalDividendAmount float64
	TotalFees           float64
	TransactionCount    int
}
