package presenter

// StockValuationResponse is the frontend-facing response for stock lookup.
type StockValuationResponse struct {
	Ticker         string  `json:"ticker"`
	Price          float64 `json:"price"`
	High52Week     float64 `json:"high52Week"`
	Low52Week      float64 `json:"low52Week"`
	EPS            float64 `json:"eps"`
	BVPS           float64 `json:"bvps"`
	ROE            float64 `json:"roe"`
	DER            float64 `json:"der"`
	PBV            float64 `json:"pbv"`
	PER            float64 `json:"per"`
	DividendYield  float64 `json:"dividendYield"`
	PayoutRatio    float64 `json:"payoutRatio"`
	GrahamNumber   float64 `json:"grahamNumber"`
	MarginOfSafety float64 `json:"marginOfSafety"`
	EntryPrice     float64 `json:"entryPrice"`
	ExitTarget     float64 `json:"exitTarget"`
	Verdict        string  `json:"verdict"`
	RiskProfile    string  `json:"riskProfile"`
	FetchedAt      string  `json:"fetchedAt"`
	Source         string  `json:"source"`
}

// BrokerageAccountResponse is the frontend-facing response for a brokerage account.
type BrokerageAccountResponse struct {
	ID          string  `json:"id"`
	BrokerName  string  `json:"brokerName"`
	BuyFeePct   float64 `json:"buyFeePct"`
	SellFeePct  float64 `json:"sellFeePct"`
	IsManualFee bool    `json:"isManualFee"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// PortfolioResponse is the frontend-facing response for a portfolio.
type PortfolioResponse struct {
	ID              string  `json:"id"`
	BrokerageAcctID string  `json:"brokerageAcctId"`
	Name            string  `json:"name"`
	Mode            string  `json:"mode"`
	RiskProfile     string  `json:"riskProfile"`
	Capital         float64 `json:"capital"`
	MonthlyAddition float64 `json:"monthlyAddition"`
	MaxStocks       int     `json:"maxStocks"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

// HoldingDetailResponse is the frontend-facing response for a holding with valuation.
type HoldingDetailResponse struct {
	ID             string   `json:"id"`
	Ticker         string   `json:"ticker"`
	AvgBuyPrice    float64  `json:"avgBuyPrice"`
	Lots           int      `json:"lots"`
	CurrentPrice   *float64 `json:"currentPrice,omitempty"`
	GrahamNumber   *float64 `json:"grahamNumber,omitempty"`
	EntryPrice     *float64 `json:"entryPrice,omitempty"`
	ExitTarget     *float64 `json:"exitTarget,omitempty"`
	Verdict        *string  `json:"verdict,omitempty"`
	MarginOfSafety *float64 `json:"marginOfSafety,omitempty"`
}

// PortfolioDetailResponse is the frontend-facing response for a portfolio with holdings.
type PortfolioDetailResponse struct {
	Portfolio PortfolioResponse       `json:"portfolio"`
	Holdings  []HoldingDetailResponse `json:"holdings"`
}
